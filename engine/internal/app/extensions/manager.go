package extensions

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	NineRouterExtensionID = "extension_9router"
	NineRouterProviderID  = "provider_9router_local"
	NineRouterSecretKey   = "NINEROUTER_API_KEY"
)

type CommandRunner interface {
	Run(ctx context.Context, workDir string, env []string, command string, args ...string) CommandResult
}

type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

type realCommandRunner struct{}

func (realCommandRunner) Run(ctx context.Context, workDir string, env []string, command string, args ...string) CommandResult {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		var exitErr *exec.ExitError
		if ok := errorAs(err, &exitErr); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return CommandResult{
		Stdout:   redact(stdout.String()),
		Stderr:   redact(stderr.String()),
		ExitCode: exitCode,
		Err:      err,
	}
}

type processHandle struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

type NodeExtensionManager struct {
	mu        sync.Mutex
	now       func() string
	baseDir   string
	runner    CommandRunner
	settings  AppSettings
	specs     map[string]ExtensionSpec
	statuses  map[string]ExtensionStatus
	secrets   map[string]string
	persist   bool
	processes map[string]processHandle
	logs      map[string][]ExtensionLogLine
	restarts  map[string]int
}

func NewNodeExtensionManager(options ...Option) *NodeExtensionManager {
	manager := &NodeExtensionManager{
		now: func() string {
			return time.Now().UTC().Format(time.RFC3339)
		},
		baseDir:   defaultExtensionBaseDir(),
		runner:    realCommandRunner{},
		settings:  DefaultSettings(),
		specs:     map[string]ExtensionSpec{},
		statuses:  map[string]ExtensionStatus{},
		secrets:   map[string]string{},
		persist:   false,
		processes: map[string]processHandle{},
		logs:      map[string][]ExtensionLogLine{},
		restarts:  map[string]int{},
	}
	for _, option := range options {
		option(manager)
	}
	if manager.persist {
		manager.loadState()
	}
	if key := strings.TrimSpace(os.Getenv(NineRouterSecretKey)); validSecretToken(key) {
		manager.secrets[NineRouterExtensionID] = key
	}
	manager.registerSpec(manager.nineRouterSpec())
	return manager
}

type Option func(*NodeExtensionManager)

func WithClock(now func() string) Option {
	return func(manager *NodeExtensionManager) {
		if now != nil {
			manager.now = now
		}
	}
}

func WithBaseDir(baseDir string) Option {
	return func(manager *NodeExtensionManager) {
		if strings.TrimSpace(baseDir) != "" {
			manager.baseDir = filepath.Clean(baseDir)
		}
	}
}

func WithCommandRunner(runner CommandRunner) Option {
	return func(manager *NodeExtensionManager) {
		if runner != nil {
			manager.runner = runner
		}
	}
}

func WithPersistence(enabled bool) Option {
	return func(manager *NodeExtensionManager) {
		manager.persist = enabled
	}
}

func DefaultSettings() AppSettings {
	return AppSettings{
		EnableNineRouterIntegration:     true,
		NineRouterRunMode:               RunModeExternal,
		NineRouterBaseURL:               "http://localhost:20128/v1",
		NineRouterDashboardURL:          "http://localhost:20128",
		NineRouterDefaultModel:          "kr/claude-sonnet-4.5",
		NineRouterAutoDetectOnStart:     true,
		NineRouterManagedAutoStart:      false,
		NineRouterManagedAutoRestart:    false,
		NineRouterManagedInstallVersion: "latest",
		NineRouterManagedPackageName:    "9router",
		NineRouterManagedCommand:        "9router",
		NineRouterManagedTimeoutMS:      30000,
		AllowNineRouterAsFreeRoute:      true,
		AllowAgentsUseNineRouter:        true,
	}
}

func (m *NodeExtensionManager) registerSpec(spec ExtensionSpec) {
	m.specs[spec.ExtensionID] = spec
	m.statuses[spec.ExtensionID] = m.defaultStatus(spec)
}

func (m *NodeExtensionManager) nineRouterSpec() ExtensionSpec {
	runtimeDir, logDir, configDir := m.extensionDirs(NineRouterExtensionID)
	settings := m.settings
	return ExtensionSpec{
		ExtensionID: NineRouterExtensionID,
		Name:        "9Router 本地模型路由器",
		Kind:        "node_managed_provider",
		RuntimeKind: "node",
		Description: "将 9Router 作为 OpenAI 兼容的本地上游模型路由器使用。",
		Install: ExtensionInstallSpec{
			PackageName:    fallback(settings.NineRouterManagedPackageName, "9router"),
			PackageVersion: fallback(settings.NineRouterManagedInstallVersion, "latest"),
			RuntimeDir:     fallback(settings.NineRouterManagedWorkDir, runtimeDir),
			LogDir:         fallback(settings.NineRouterManagedLogDir, logDir),
			ConfigDir:      configDir,
		},
		Process: ExtensionProcessSpec{
			DefaultCommand: fallback(settings.NineRouterManagedCommand, "9router"),
			DefaultArgs:    []string{"--host", "127.0.0.1", "--no-browser", "--skip-update", "--tray"},
			Port:           20128,
			Env:            []string{"PORT=20128", "NEXT_PUBLIC_BASE_URL=http://localhost:20128"},
		},
		Health: ExtensionHealthSpec{
			DashboardURL: fallback(settings.NineRouterDashboardURL, "http://localhost:20128"),
			BaseURL:      fallback(settings.NineRouterBaseURL, "http://localhost:20128/v1"),
			ModelsPath:   "/models",
			ChatPath:     "/chat/completions",
		},
		ProviderBridge: &ExtensionProviderBridge{
			ProviderID:    NineRouterProviderID,
			ProviderType:  "openai_compatible",
			DisplayName:   "9Router 免费模型路由",
			BaseURL:       fallback(settings.NineRouterBaseURL, "http://localhost:20128/v1"),
			DefaultModel:  fallback(settings.NineRouterDefaultModel, "kr/claude-sonnet-4.5"),
			SortOrder:     999,
			SystemPreset:  true,
			AllowDeletion: false,
		},
		Capabilities: []string{"model_gateway", "openai_compatible_chat", "model_discovery", "streaming", "vision", "image_generation"},
		Security: ExtensionSecuritySpec{
			RiskLevel:       "medium",
			AllowedHosts:    []string{"localhost", "127.0.0.1", "::1"},
			SecretKeys:      []string{NineRouterSecretKey},
			EnvAllowList:    []string{"PATH", "SystemRoot", "COMSPEC", "TEMP", "TMP", "PORT", "NEXT_PUBLIC_BASE_URL"},
			ManagedRequires: true,
		},
		SystemPreset: true,
		Enabled:      settings.EnableNineRouterIntegration,
	}
}

func (m *NodeExtensionManager) ListExtensions() []ExtensionSpec {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshSpecsLocked()
	result := make([]ExtensionSpec, 0, len(m.specs))
	for _, spec := range m.specs {
		result = append(result, spec)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ExtensionID < result[j].ExtensionID
	})
	return result
}

func (m *NodeExtensionManager) GetExtensionStatus(extensionID string) (ExtensionStatus, *Error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	status, ok := m.statuses[extensionID]
	if !ok {
		return ExtensionStatus{}, notFound(extensionID)
	}
	return status, nil
}

func (m *NodeExtensionManager) GetSettings() AppSettings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settings
}

func (m *NodeExtensionManager) UpdateSettings(input UpdateSettingsInput) (AppSettings, *Error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings
	if input.EnableNineRouterIntegration != nil {
		settings.EnableNineRouterIntegration = *input.EnableNineRouterIntegration
	}
	if input.NineRouterRunMode != nil {
		settings.NineRouterRunMode = normalizeRunMode(*input.NineRouterRunMode)
	}
	if input.NineRouterBaseURL != nil {
		settings.NineRouterBaseURL = strings.TrimSpace(*input.NineRouterBaseURL)
	}
	if input.NineRouterDashboardURL != nil {
		settings.NineRouterDashboardURL = strings.TrimSpace(*input.NineRouterDashboardURL)
	}
	if input.NineRouterDefaultModel != nil {
		settings.NineRouterDefaultModel = strings.TrimSpace(*input.NineRouterDefaultModel)
	}
	if input.NineRouterAutoDetectOnStart != nil {
		settings.NineRouterAutoDetectOnStart = *input.NineRouterAutoDetectOnStart
	}
	if input.NineRouterManagedAutoStart != nil {
		settings.NineRouterManagedAutoStart = *input.NineRouterManagedAutoStart
	}
	if input.NineRouterManagedAutoRestart != nil {
		settings.NineRouterManagedAutoRestart = *input.NineRouterManagedAutoRestart
	}
	if input.NineRouterManagedInstallVersion != nil {
		settings.NineRouterManagedInstallVersion = strings.TrimSpace(*input.NineRouterManagedInstallVersion)
	}
	if input.NineRouterManagedPackageName != nil {
		settings.NineRouterManagedPackageName = strings.TrimSpace(*input.NineRouterManagedPackageName)
	}
	if input.NineRouterManagedCommand != nil {
		settings.NineRouterManagedCommand = strings.TrimSpace(*input.NineRouterManagedCommand)
	}
	if input.NineRouterManagedWorkDir != nil {
		settings.NineRouterManagedWorkDir = strings.TrimSpace(*input.NineRouterManagedWorkDir)
	}
	if input.NineRouterManagedLogDir != nil {
		settings.NineRouterManagedLogDir = strings.TrimSpace(*input.NineRouterManagedLogDir)
	}
	if input.NineRouterManagedTimeoutMS != nil {
		settings.NineRouterManagedTimeoutMS = normalizeTimeout(*input.NineRouterManagedTimeoutMS)
	}
	if input.AllowNineRouterAsFreeRoute != nil {
		settings.AllowNineRouterAsFreeRoute = *input.AllowNineRouterAsFreeRoute
	}
	if input.AllowAgentsUseNineRouter != nil {
		settings.AllowAgentsUseNineRouter = *input.AllowAgentsUseNineRouter
	}
	m.settings = normalizeSettings(settings)
	m.refreshSpecsLocked()
	if err := m.persistStateLocked(); err != nil {
		return m.settings, err
	}
	return m.settings, nil
}

func (m *NodeExtensionManager) ResetExtensionSettings(extensionID string) (AppSettings, *Error) {
	if extensionID != "" && extensionID != NineRouterExtensionID {
		return AppSettings{}, notFound(extensionID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = DefaultSettings()
	m.refreshSpecsLocked()
	if err := m.persistStateLocked(); err != nil {
		return m.settings, err
	}
	return m.settings, nil
}

func (m *NodeExtensionManager) SetSecret(extensionID string, value string) *Error {
	if extensionID != NineRouterExtensionID {
		return notFound(extensionID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if !validSecretToken(value) {
		return &Error{
			Code:       "EXTENSION_SECRET_INVALID",
			Message:    "9Router API Key contains invalid characters.",
			UserAction: "Copy the Endpoint Key from 9Router Dashboard again, then save it.",
		}
	}
	m.secrets[extensionID] = value
	status := m.statuses[extensionID]
	status.HasAPIKey = true
	status.MaskedKey = maskSecret(value)
	m.statuses[extensionID] = status
	if err := m.persistStateLocked(); err != nil {
		return err
	}
	return nil
}

func (m *NodeExtensionManager) Secret(extensionID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.secrets[extensionID]
}

func (m *NodeExtensionManager) DetectExtension(ctx context.Context, extensionID string) (ExtensionActionResult, *Error) {
	m.mu.Lock()
	spec, ok := m.specs[extensionID]
	if !ok {
		m.mu.Unlock()
		return ExtensionActionResult{}, notFound(extensionID)
	}
	m.mu.Unlock()
	runtimeInfo := m.detectRuntime(ctx, spec)
	m.appendLog(extensionID, "detect", detectMessage(runtimeInfo))
	m.mu.Lock()
	defer m.mu.Unlock()
	status := m.statuses[extensionID]
	status.Runtime = runtimeInfo
	status.NodeAvailable = runtimeInfo.NodeAvailable
	status.NPMAvailable = runtimeInfo.NPMAvailable
	status.NodeVersion = runtimeInfo.NodeVersion
	status.NPMVersion = runtimeInfo.NPMVersion
	status.Command = runtimeInfo.Command
	status.InstallSource = runtimeInfo.InstallSource
	status.Installed = runtimeInfo.CommandAvailable || status.Installed
	status.LastCheckedAt = m.now()
	status.LastErrorCode = runtimeInfo.LastErrorCode
	status.LastErrorMessage = runtimeInfo.LastErrorMessage
	m.statuses[extensionID] = status
	return ExtensionActionResult{OK: runtimeInfo.LastErrorCode == "", ExtensionID: extensionID, Message: detectMessage(runtimeInfo), Status: status}, nil
}

func (m *NodeExtensionManager) InstallExtension(ctx context.Context, input InstallExtensionInput) (ExtensionActionResult, *Error) {
	extensionID := strings.TrimSpace(input.ExtensionID)
	m.mu.Lock()
	spec, ok := m.specs[extensionID]
	settings := m.settings
	m.mu.Unlock()
	if !ok {
		return ExtensionActionResult{}, notFound(extensionID)
	}
	if settings.NineRouterRunMode != RunModeManaged {
		return ExtensionActionResult{}, &Error{Code: "EXTENSION_MANAGED_MODE_REQUIRED", Message: "请先切换到 DreamWorker 受管模式。", UserAction: "在拓展设置中选择受管模式后再安装。"}
	}
	version := fallback(input.Version, spec.Install.PackageVersion)
	if version == "" {
		version = "latest"
	}
	if err := os.MkdirAll(spec.Install.RuntimeDir, 0o755); err != nil {
		return ExtensionActionResult{}, &Error{Code: "EXTENSION_INSTALL_DIR_FAILED", Message: "无法创建拓展安装目录。", UserAction: "检查本机目录权限后重试。"}
	}
	if err := os.MkdirAll(spec.Install.LogDir, 0o755); err != nil {
		return ExtensionActionResult{}, &Error{Code: "EXTENSION_LOG_DIR_FAILED", Message: "无法创建拓展日志目录。", UserAction: "检查本机目录权限后重试。"}
	}
	timeout := time.Duration(normalizeTimeout(settings.NineRouterManagedTimeoutMS)) * time.Millisecond
	installCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	m.appendLog(extensionID, "install", "开始本地安装 "+spec.Install.PackageName+"@"+version)
	initResult := m.runner.Run(installCtx, spec.Install.RuntimeDir, safeEnv(spec), "npm", "init", "-y")
	m.appendCommandResult(extensionID, "install", initResult)
	if initResult.Err != nil {
		return m.failAction(extensionID, "EXTENSION_NPM_INIT_FAILED", "npm init 执行失败。", "检查 Node/npm 环境和目录权限后重试。")
	}
	packageName := fallback(spec.Install.PackageName, "9router") + "@" + version
	installResult := m.runner.Run(installCtx, spec.Install.RuntimeDir, safeEnv(spec), "npm", "install", packageName)
	m.appendCommandResult(extensionID, "install", installResult)
	if installResult.Err != nil {
		return m.failAction(extensionID, "EXTENSION_NPM_INSTALL_FAILED", "npm install 执行失败。", "查看安装日志，修复网络或依赖构建问题后重试。")
	}
	runtimeInfo := m.detectRuntime(ctx, spec)
	m.mu.Lock()
	status := m.statuses[extensionID]
	status.Installed = true
	status.InstallSource = "managed_local"
	status.ProcessState = "stopped"
	status.Runtime = runtimeInfo
	status.Command = runtimeInfo.Command
	status.LastErrorCode = ""
	status.LastErrorMessage = ""
	status.LastCheckedAt = m.now()
	m.statuses[extensionID] = status
	m.mu.Unlock()
	return ExtensionActionResult{OK: true, ExtensionID: extensionID, Message: "9Router 已安装到本地受管目录。", Status: status}, nil
}

func (m *NodeExtensionManager) StartExtension(ctx context.Context, extensionID string) (ExtensionActionResult, *Error) {
	m.mu.Lock()
	spec, ok := m.specs[extensionID]
	settings := m.settings
	if !ok {
		m.mu.Unlock()
		return ExtensionActionResult{}, notFound(extensionID)
	}
	if _, running := m.processes[extensionID]; running {
		status := m.statuses[extensionID]
		m.mu.Unlock()
		return ExtensionActionResult{OK: true, ExtensionID: extensionID, Message: "拓展进程已在运行。", Status: status}, nil
	}
	m.mu.Unlock()

	if settings.NineRouterRunMode == RunModeExternal {
		m.appendLog(extensionID, "process", "当前为外部服务模式，DreamWorker 不启动本地进程。")
		result, appErr := m.TestExtension(ctx, extensionID)
		if appErr != nil {
			return result, appErr
		}
		if !result.OK {
			message := "当前为外部服务模式，DreamWorker 不会启动受管 9Router；未检测到外部 9Router 服务。请切换到 DreamWorker 受管模式后再启动，或先手动启动 9Router。"
			status := m.updateError(extensionID, "EXTENSION_EXTERNAL_SERVICE_UNREACHABLE", message)
			result.Message = message
			result.Status = status
		}
		return result, nil
	}
	if !settings.EnableNineRouterIntegration {
		return ExtensionActionResult{}, &Error{Code: "EXTENSION_DISABLED", Message: "9Router 集成已关闭。", UserAction: "先在设置中启用 9Router 集成。"}
	}
	if m.portReachable(spec.Health.DashboardURL) {
		return m.failAction(extensionID, "EXTENSION_PORT_OCCUPIED", "检测到 9Router 端口已有服务，DreamWorker 不会接管未知进程。", "切换到外部服务模式，或释放端口后再启动受管模式。")
	}
	runtimeInfo := m.detectRuntime(ctx, spec)
	if !runtimeInfo.CommandAvailable {
		return m.failAction(extensionID, "EXTENSION_COMMAND_NOT_FOUND", "未找到 9router 启动命令。", "先安装 9Router，或配置自定义命令。")
	}
	m.appendLog(extensionID, "process", "准备启动受管进程："+runtimeInfo.Command)
	if err := os.MkdirAll(spec.Install.LogDir, 0o755); err != nil {
		return m.failAction(extensionID, "EXTENSION_LOG_DIR_FAILED", "无法创建拓展日志目录。", "检查本机目录权限后重试。")
	}
	startCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(startCtx, runtimeInfo.Command, spec.Process.DefaultArgs...)
	cmd.Dir = spec.Install.RuntimeDir
	cmd.Env = safeEnv(spec)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		cancel()
		return m.failAction(extensionID, "EXTENSION_START_FAILED", "9Router 启动失败。", "检查命令路径、Node 环境和日志后重试。")
	}
	now := m.now()
	handle := processHandle{cmd: cmd, cancel: cancel}
	m.mu.Lock()
	m.processes[extensionID] = handle
	status := m.statuses[extensionID]
	status.ProcessState = "starting"
	status.PID = cmd.Process.Pid
	status.StartedByDreamWorker = true
	status.LastStartedAt = now
	status.Command = runtimeInfo.Command
	status.Runtime = runtimeInfo
	status.WorkDir = spec.Install.RuntimeDir
	status.LogDir = spec.Install.LogDir
	m.statuses[extensionID] = status
	m.mu.Unlock()
	go m.collectOutput(extensionID, "stdout", stdout)
	go m.collectOutput(extensionID, "stderr", stderr)
	go m.waitProcess(extensionID, cmd)

	deadline := time.Now().Add(time.Duration(normalizeTimeout(settings.NineRouterManagedTimeoutMS)) * time.Millisecond)
	for time.Now().Before(deadline) {
		if m.portReachable(spec.Health.DashboardURL) {
			break
		}
		select {
		case <-ctx.Done():
			return m.failAction(extensionID, "EXTENSION_START_CANCELLED", "启动已取消。", "重新启动 9Router。")
		case <-time.After(500 * time.Millisecond):
		}
	}
	result, appErr := m.TestExtension(ctx, extensionID)
	m.mu.Lock()
	status = m.statuses[extensionID]
	if appErr == nil && (status.HealthStatus == "connected" || status.HealthStatus == "disconnected") {
		status.ProcessState = "running"
		status.PID = cmd.Process.Pid
		status.StartedByDreamWorker = true
		m.statuses[extensionID] = status
	}
	m.mu.Unlock()
	if appErr != nil {
		return result, appErr
	}
	result.Message = "9Router 受管进程已启动。"
	return result, nil
}

func (m *NodeExtensionManager) StopExtension(extensionID string) (ExtensionActionResult, *Error) {
	m.mu.Lock()
	handle, ok := m.processes[extensionID]
	status, exists := m.statuses[extensionID]
	if !exists {
		m.mu.Unlock()
		return ExtensionActionResult{}, notFound(extensionID)
	}
	if !ok {
		status.ProcessState = "stopped"
		status.StartedByDreamWorker = false
		status.PID = 0
		status.LastStoppedAt = m.now()
		m.statuses[extensionID] = status
		m.mu.Unlock()
		return ExtensionActionResult{OK: true, ExtensionID: extensionID, Message: "没有由 DreamWorker 启动的进程。", Status: status}, nil
	}
	delete(m.processes, extensionID)
	status.ProcessState = "stopping"
	m.statuses[extensionID] = status
	m.mu.Unlock()
	handle.cancel()
	time.Sleep(500 * time.Millisecond)
	if handle.cmd.Process != nil {
		_ = handle.cmd.Process.Kill()
	}
	m.mu.Lock()
	status = m.statuses[extensionID]
	status.ProcessState = "stopped"
	status.PID = 0
	status.StartedByDreamWorker = false
	status.LastStoppedAt = m.now()
	status.HealthStatus = "disconnected"
	m.statuses[extensionID] = status
	m.mu.Unlock()
	m.appendLog(extensionID, "process", "受管进程已停止。")
	return ExtensionActionResult{OK: true, ExtensionID: extensionID, Message: "9Router 受管进程已停止。", Status: status}, nil
}

func (m *NodeExtensionManager) RestartExtension(ctx context.Context, extensionID string) (ExtensionActionResult, *Error) {
	if _, err := m.StopExtension(extensionID); err != nil {
		return ExtensionActionResult{}, err
	}
	return m.StartExtension(ctx, extensionID)
}

func (m *NodeExtensionManager) TestExtension(ctx context.Context, extensionID string) (ExtensionActionResult, *Error) {
	m.appendLog(extensionID, "health", "开始测试 9Router 连接。")
	status, err := m.healthCheck(ctx, extensionID, false)
	if err != nil {
		return ExtensionActionResult{OK: false, ExtensionID: extensionID, Message: err.Message, Status: status}, nil
	}
	return ExtensionActionResult{OK: status.HealthStatus == "connected", ExtensionID: extensionID, Message: healthMessage(status), Status: status}, nil
}

func (m *NodeExtensionManager) RefreshModels(ctx context.Context, extensionID string) (ExtensionModelRefreshResult, *Error) {
	status, err := m.healthCheck(ctx, extensionID, true)
	if err != nil {
		return ExtensionModelRefreshResult{OK: false, ExtensionID: extensionID, Models: status.Models, Status: status}, nil
	}
	return ExtensionModelRefreshResult{OK: true, ExtensionID: extensionID, Models: status.Models, Status: status}, nil
}

func (m *NodeExtensionManager) VerifyStreaming(ctx context.Context, extensionID string) (ExtensionStreamingResult, *Error) {
	start := time.Now()
	m.mu.Lock()
	spec, ok := m.specs[extensionID]
	settings := m.settings
	key := m.secrets[extensionID]
	m.mu.Unlock()
	if !ok {
		return ExtensionStreamingResult{}, notFound(extensionID)
	}
	body := map[string]any{
		"model":  fallback(settings.NineRouterDefaultModel, "kr/claude-sonnet-4.5"),
		"stream": true,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	}
	payload, _ := json.Marshal(body)
	request, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(spec.Health.BaseURL, spec.Health.ChatPath), bytes.NewReader(payload))
	if reqErr != nil {
		status, _ := m.GetExtensionStatus(extensionID)
		return ExtensionStreamingResult{OK: false, ExtensionID: extensionID, Message: "无法创建流式验证请求。", Status: status}, nil
	}
	request.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(key) != "" {
		request.Header.Set("Authorization", "Bearer "+key)
	}
	response, httpErr := http.DefaultClient.Do(request)
	if httpErr != nil {
		status := m.updateError(extensionID, "EXTENSION_STREAM_FAILED", "流式验证请求失败。")
		return ExtensionStreamingResult{OK: false, ExtensionID: extensionID, Message: "流式验证请求失败。", LatencyMS: latencyMS(start), Status: status}, nil
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "流式验证返回异常状态。"
		if detail := responseErrorMessage(response); detail != "" {
			message = fmt.Sprintf("流式验证返回 HTTP %d：%s", response.StatusCode, detail)
		}
		status := m.updateError(extensionID, fmt.Sprintf("HTTP_%d", response.StatusCode), message)
		return ExtensionStreamingResult{OK: false, ExtensionID: extensionID, Message: message, LatencyMS: latencyMS(start), Status: status}, nil
	}
	reader := bufio.NewReader(response.Body)
	for {
		line, readErr := reader.ReadString('\n')
		if strings.TrimSpace(line) != "" {
			m.mu.Lock()
			status := m.statuses[extensionID]
			status.StreamingVerified = true
			status.LastCheckedAt = m.now()
			status.LastErrorCode = ""
			status.LastErrorMessage = ""
			m.statuses[extensionID] = status
			m.mu.Unlock()
			return ExtensionStreamingResult{OK: true, ExtensionID: extensionID, Message: "流式输出验证成功。", LatencyMS: latencyMS(start), Status: status}, nil
		}
		if readErr != nil {
			break
		}
	}
	status := m.updateError(extensionID, "EXTENSION_STREAM_EMPTY", "流式验证未收到有效数据。")
	return ExtensionStreamingResult{OK: false, ExtensionID: extensionID, Message: "流式验证未收到有效数据。", LatencyMS: latencyMS(start), Status: status}, nil
}

func (m *NodeExtensionManager) TailLogs(input TailLogsInput) ([]ExtensionLogLine, *Error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.specs[input.ExtensionID]; !ok {
		return nil, notFound(input.ExtensionID)
	}
	lines := append([]ExtensionLogLine{}, m.logs[input.ExtensionID]...)
	limit := input.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

func (m *NodeExtensionManager) ClearLogs(extensionID string) (ExtensionActionResult, *Error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	status, ok := m.statuses[extensionID]
	if !ok {
		return ExtensionActionResult{}, notFound(extensionID)
	}
	m.logs[extensionID] = nil
	return ExtensionActionResult{OK: true, ExtensionID: extensionID, Message: "拓展日志已清空。", Status: status}, nil
}

func (m *NodeExtensionManager) healthCheck(ctx context.Context, extensionID string, refreshModels bool) (ExtensionStatus, *Error) {
	m.mu.Lock()
	spec, ok := m.specs[extensionID]
	key := m.secrets[extensionID]
	m.mu.Unlock()
	if !ok {
		return ExtensionStatus{}, notFound(extensionID)
	}
	status := m.defaultStatus(spec)
	if !m.serverReachable(ctx, spec.Health.DashboardURL) {
		status.HealthStatus = "disconnected"
		status.LastErrorCode = "EXTENSION_SERVICE_UNREACHABLE"
		status.LastErrorMessage = "9Router 服务未启动或端口不可访问。"
		m.storeStatus(status)
		return status, &Error{Code: status.LastErrorCode, Message: status.LastErrorMessage, UserAction: "启动 9Router 或切换运行模式后重试。"}
	}
	models, code, message := m.fetchModels(ctx, spec, key)
	if code != "" {
		status.HealthStatus = "error"
		status.LastErrorCode = code
		status.LastErrorMessage = message
		m.storeStatus(status)
		return status, nil
	}
	status.HealthStatus = "connected"
	status.ProcessState = preserveRunningState(m.currentStatus(extensionID).ProcessState)
	status.Models = models
	status.ModelCount = len(models)
	status.LastErrorCode = ""
	status.LastErrorMessage = ""
	if !refreshModels && len(models) == 0 {
		status.Models = m.currentStatus(extensionID).Models
		status.ModelCount = len(status.Models)
	}
	m.storeStatus(status)
	return status, nil
}

func (m *NodeExtensionManager) fetchModels(ctx context.Context, spec ExtensionSpec, apiKey string) ([]string, string, string) {
	models, code, message := m.fetchModelsAtPath(ctx, spec, apiKey, spec.Health.ModelsPath)
	if code != "" {
		return nil, code, message
	}
	if spec.ExtensionID == NineRouterExtensionID {
		imageModels, imageCode, imageMessage := m.fetchModelsAtPath(ctx, spec, apiKey, "/models/image")
		if imageCode != "" && len(models) == 0 {
			return nil, imageCode, imageMessage
		}
		models = appendUniqueStrings(models, imageModels...)
	}
	sort.Strings(models)
	return models, "", ""
}

func (m *NodeExtensionManager) fetchModelsAtPath(ctx context.Context, spec ExtensionSpec, apiKey string, modelPath string) ([]string, string, string) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(spec.Health.BaseURL, modelPath), nil)
	if err != nil {
		return nil, "EXTENSION_MODELS_REQUEST_INVALID", "模型刷新请求无效。"
	}
	if strings.TrimSpace(apiKey) != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "EXTENSION_MODELS_FAILED", "无法连接 9Router 模型接口。"
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
	default:
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			return nil, "EXTENSION_API_KEY_INVALID", "9Router API Key 无效或权限不足，请在 9Router Dashboard 重新复制 Endpoint Key。"
		}
		if response.StatusCode == http.StatusNotFound {
			return nil, "EXTENSION_BASE_URL_INVALID", "当前 Base URL 不正确，OpenAI 兼容接口应使用 http://localhost:20128/v1。"
		}
		return nil, fmt.Sprintf("HTTP_%d", response.StatusCode), "9Router 模型接口返回异常状态。"
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&payload); err != nil {
		return nil, "EXTENSION_MODELS_PARSE_FAILED", "无法解析 9Router 模型列表。"
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) != "" {
			models = append(models, item.ID)
		}
	}
	return models, "", ""
}

func appendUniqueStrings(values []string, extras ...string) []string {
	seen := make(map[string]bool, len(values)+len(extras))
	result := make([]string, 0, len(values)+len(extras))
	for _, value := range append(values, extras...) {
		item := strings.TrimSpace(value)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}

func responseErrorMessage(response *http.Response) string {
	if response == nil || response.Body == nil {
		return ""
	}
	limited, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	message := strings.TrimSpace(string(limited))
	if message == "" {
		return ""
	}
	var payload struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(limited, &payload); err == nil {
		if payload.Error != nil && payload.Error.Message != "" {
			if payload.Error.Code != "" {
				return redact(payload.Error.Code + ": " + payload.Error.Message)
			}
			return redact(payload.Error.Message)
		}
		if payload.Message != "" {
			return redact(payload.Message)
		}
	}
	return redact(message)
}

func (m *NodeExtensionManager) detectRuntime(ctx context.Context, spec ExtensionSpec) NodeRuntimeInfo {
	info := NodeRuntimeInfo{}
	if nodePath, err := exec.LookPath("node"); err == nil {
		info.NodeAvailable = true
		info.NodeVersion = strings.TrimSpace(m.runner.Run(ctx, "", safeEnv(spec), nodePath, "--version").Stdout)
	}
	if npmPath, err := exec.LookPath("npm"); err == nil {
		info.NPMAvailable = true
		info.NPMVersion = strings.TrimSpace(m.runner.Run(ctx, "", safeEnv(spec), npmPath, "--version").Stdout)
	}
	info.ManagedLocalBin = managedBinPath(spec)
	packageCommand := fallback(spec.Install.PackageName, "9router")
	configuredCommand := strings.TrimSpace(spec.Process.DefaultCommand)
	type commandCandidate struct {
		source  string
		command string
	}
	candidates := []commandCandidate{}
	if configuredCommand != "" && configuredCommand != packageCommand {
		candidates = append(candidates, commandCandidate{"custom_command", configuredCommand})
	}
	candidates = append(candidates,
		commandCandidate{"managed_local", info.ManagedLocalBin},
		commandCandidate{"system_path", packageCommand},
	)
	for _, candidate := range candidates {
		if candidate.command == "" {
			continue
		}
		if filepath.IsAbs(candidate.command) {
			if fileExists(candidate.command) {
				info.CommandAvailable = true
				info.Command = candidate.command
				info.InstallSource = candidate.source
				return info
			}
			continue
		}
		if found, err := exec.LookPath(candidate.command); err == nil {
			info.CommandAvailable = true
			info.Command = found
			info.InstallSource = candidate.source
			return info
		}
	}
	if !info.NodeAvailable {
		info.LastErrorCode = "NODE_NOT_FOUND"
		info.LastErrorMessage = "未检测到 Node.js。"
		return info
	}
	if !info.NPMAvailable {
		info.LastErrorCode = "NPM_NOT_FOUND"
		info.LastErrorMessage = "未检测到 npm。"
		return info
	}
	info.LastErrorCode = "EXTENSION_COMMAND_NOT_FOUND"
	info.LastErrorMessage = "未检测到 9router 命令。"
	return info
}

func (m *NodeExtensionManager) defaultStatus(spec ExtensionSpec) ExtensionStatus {
	settings := m.settings
	secret := m.secrets[spec.ExtensionID]
	status := m.statuses[spec.ExtensionID]
	models := append([]string{}, status.Models...)
	if len(models) == 0 {
		models = []string{fallback(settings.NineRouterDefaultModel, "kr/claude-sonnet-4.5")}
	}
	processState := status.ProcessState
	if processState == "" {
		processState = "stopped"
	}
	installSource := status.InstallSource
	if installSource == "" {
		installSource = "none"
	}
	return ExtensionStatus{
		ExtensionID:          spec.ExtensionID,
		Installed:            status.Installed,
		InstallSource:        installSource,
		NodeAvailable:        status.NodeAvailable,
		NPMAvailable:         status.NPMAvailable,
		NodeVersion:          status.NodeVersion,
		NPMVersion:           status.NPMVersion,
		Command:              status.Command,
		RunMode:              settings.NineRouterRunMode,
		ProcessState:         processState,
		PID:                  status.PID,
		StartedByDreamWorker: status.StartedByDreamWorker,
		BaseURL:              spec.Health.BaseURL,
		DashboardURL:         spec.Health.DashboardURL,
		HealthStatus:         fallback(status.HealthStatus, "unknown"),
		ModelCount:           len(models),
		Models:               models,
		StreamingVerified:    status.StreamingVerified,
		HasAPIKey:            secret != "",
		MaskedKey:            maskSecret(secret),
		LogDir:               spec.Install.LogDir,
		WorkDir:              spec.Install.RuntimeDir,
		LastStartedAt:        status.LastStartedAt,
		LastStoppedAt:        status.LastStoppedAt,
		LastCheckedAt:        fallback(status.LastCheckedAt, m.now()),
		LastErrorCode:        status.LastErrorCode,
		LastErrorMessage:     status.LastErrorMessage,
		Runtime:              status.Runtime,
	}
}

func (m *NodeExtensionManager) refreshSpecsLocked() {
	m.settings = normalizeSettings(m.settings)
	m.specs[NineRouterExtensionID] = m.nineRouterSpec()
	status := m.statuses[NineRouterExtensionID]
	status.RunMode = m.settings.NineRouterRunMode
	status.BaseURL = m.settings.NineRouterBaseURL
	status.DashboardURL = m.settings.NineRouterDashboardURL
	status.HasAPIKey = m.secrets[NineRouterExtensionID] != ""
	status.MaskedKey = maskSecret(m.secrets[NineRouterExtensionID])
	status.LogDir = fallback(m.settings.NineRouterManagedLogDir, m.specs[NineRouterExtensionID].Install.LogDir)
	status.WorkDir = fallback(m.settings.NineRouterManagedWorkDir, m.specs[NineRouterExtensionID].Install.RuntimeDir)
	m.statuses[NineRouterExtensionID] = status
}

func (m *NodeExtensionManager) storeStatus(status ExtensionStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.statuses[status.ExtensionID]
	status.Installed = current.Installed || status.Installed
	status.InstallSource = fallback(status.InstallSource, current.InstallSource)
	status.NodeAvailable = current.NodeAvailable
	status.NPMAvailable = current.NPMAvailable
	status.NodeVersion = current.NodeVersion
	status.NPMVersion = current.NPMVersion
	status.Command = fallback(status.Command, current.Command)
	status.PID = current.PID
	status.StartedByDreamWorker = current.StartedByDreamWorker
	status.LastStartedAt = current.LastStartedAt
	status.LastStoppedAt = current.LastStoppedAt
	status.Runtime = current.Runtime
	status.LastCheckedAt = m.now()
	status.HasAPIKey = m.secrets[status.ExtensionID] != ""
	status.MaskedKey = maskSecret(m.secrets[status.ExtensionID])
	m.statuses[status.ExtensionID] = status
}

func (m *NodeExtensionManager) currentStatus(extensionID string) ExtensionStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statuses[extensionID]
}

func (m *NodeExtensionManager) updateError(extensionID string, code string, message string) ExtensionStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := m.statuses[extensionID]
	status.HealthStatus = "error"
	status.LastErrorCode = code
	status.LastErrorMessage = message
	status.LastCheckedAt = m.now()
	m.statuses[extensionID] = status
	return status
}

func (m *NodeExtensionManager) failAction(extensionID string, code string, message string, userAction string) (ExtensionActionResult, *Error) {
	status := m.updateError(extensionID, code, message)
	status.ProcessState = "failed"
	m.mu.Lock()
	m.statuses[extensionID] = status
	m.mu.Unlock()
	return ExtensionActionResult{OK: false, ExtensionID: extensionID, Message: message, Status: status}, &Error{Code: code, Message: message, UserAction: userAction}
}

func (m *NodeExtensionManager) collectOutput(extensionID string, stream string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		m.appendLog(extensionID, stream, scanner.Text())
	}
}

func (m *NodeExtensionManager) waitProcess(extensionID string, cmd *exec.Cmd) {
	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if ok := errorAs(err, &exitErr); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	m.mu.Lock()
	if _, running := m.processes[extensionID]; running {
		delete(m.processes, extensionID)
		status := m.statuses[extensionID]
		status.ProcessState = "failed"
		status.PID = 0
		status.StartedByDreamWorker = false
		status.LastStoppedAt = m.now()
		status.LastErrorCode = "EXTENSION_PROCESS_EXITED"
		status.LastErrorMessage = fmt.Sprintf("9Router 进程已退出，退出码 %d。", exitCode)
		m.statuses[extensionID] = status
	}
	m.mu.Unlock()
}

func (m *NodeExtensionManager) appendCommandResult(extensionID string, stream string, result CommandResult) {
	if result.Stdout != "" {
		m.appendLog(extensionID, stream, result.Stdout)
	}
	if result.Stderr != "" {
		m.appendLog(extensionID, stream, result.Stderr)
	}
}

func (m *NodeExtensionManager) appendLog(extensionID string, stream string, line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, part := range strings.Split(redact(line), "\n") {
		part = strings.TrimRight(part, "\r")
		if strings.TrimSpace(part) == "" {
			continue
		}
		m.logs[extensionID] = append(m.logs[extensionID], ExtensionLogLine{
			ExtensionID: extensionID,
			Timestamp:   m.now(),
			Stream:      stream,
			Line:        part,
		})
	}
	if len(m.logs[extensionID]) > 1000 {
		m.logs[extensionID] = m.logs[extensionID][len(m.logs[extensionID])-1000:]
	}
}

func (m *NodeExtensionManager) serverReachable(ctx context.Context, target string) bool {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false
	}
	client := http.Client{Timeout: 1500 * time.Millisecond}
	response, err := client.Do(request)
	if err == nil {
		_ = response.Body.Close()
		return response.StatusCode < 500
	}
	return m.portReachable(target)
}

func (m *NodeExtensionManager) portReachable(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil || parsed.Host == "" {
		return false
	}
	host := parsed.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	conn, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (m *NodeExtensionManager) extensionDirs(extensionID string) (string, string, string) {
	root := filepath.Join(m.baseDir, extensionID)
	return filepath.Join(root, "runtime"), filepath.Join(root, "logs"), filepath.Join(root, "config")
}

type persistedState struct {
	Settings AppSettings       `json:"settings"`
	Secrets  map[string]string `json:"secrets,omitempty"`
}

func (m *NodeExtensionManager) statePath() string {
	return filepath.Join(m.baseDir, "extensions.config.json")
}

func (m *NodeExtensionManager) loadState() {
	data, err := os.ReadFile(m.statePath())
	if err != nil {
		return
	}
	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	m.settings = normalizeSettings(state.Settings)
	for extensionID, secret := range state.Secrets {
		secret = strings.TrimSpace(secret)
		if extensionID == "" || !validSecretToken(secret) {
			continue
		}
		m.secrets[extensionID] = secret
	}
}

func (m *NodeExtensionManager) persistStateLocked() *Error {
	if !m.persist {
		return nil
	}
	state := persistedState{
		Settings: normalizeSettings(m.settings),
		Secrets:  map[string]string{},
	}
	for extensionID, secret := range m.secrets {
		secret = strings.TrimSpace(secret)
		if extensionID == "" || !validSecretToken(secret) {
			continue
		}
		state.Secrets[extensionID] = secret
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return &Error{Code: "EXTENSION_CONFIG_ENCODE_FAILED", Message: "无法序列化拓展配置。", UserAction: "检查 9Router 设置后重试。"}
	}
	if err := os.MkdirAll(m.baseDir, 0o700); err != nil {
		return &Error{Code: "EXTENSION_CONFIG_DIR_FAILED", Message: "无法创建拓展配置目录。", UserAction: "检查 DreamWorker 配置目录权限后重试。"}
	}
	if err := os.WriteFile(m.statePath(), data, 0o600); err != nil {
		return &Error{Code: "EXTENSION_CONFIG_WRITE_FAILED", Message: "无法写入拓展配置文件。", UserAction: "检查 DreamWorker 配置目录权限后重试。"}
	}
	return nil
}

func defaultExtensionBaseDir() string {
	if configured := strings.TrimSpace(os.Getenv("DREAMWORKER_EXTENSION_DATA_DIR")); configured != "" {
		return filepath.Clean(configured)
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "DreamWorker", "extensions")
	}
	return filepath.Join(os.TempDir(), "DreamWorker", "extensions")
}

func managedBinPath(spec ExtensionSpec) string {
	binName := fallback(spec.Install.PackageName, spec.Process.DefaultCommand)
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(binName), ".cmd") {
		binName += ".cmd"
	}
	return filepath.Join(spec.Install.RuntimeDir, "node_modules", ".bin", binName)
}

func safeEnv(spec ExtensionSpec) []string {
	allow := map[string]bool{}
	for _, item := range spec.Security.EnvAllowList {
		allow[strings.ToUpper(item)] = true
	}
	env := []string{}
	for _, item := range os.Environ() {
		key := strings.SplitN(item, "=", 2)[0]
		if allow[strings.ToUpper(key)] {
			env = append(env, item)
		}
	}
	env = append(env, spec.Process.Env...)
	return env
}

func joinURL(base string, path string) string {
	base = strings.TrimRight(base, "/")
	path = "/" + strings.TrimLeft(path, "/")
	return base + path
}

func normalizeSettings(settings AppSettings) AppSettings {
	defaults := DefaultSettings()
	if settings.NineRouterRunMode == "" {
		settings.NineRouterRunMode = defaults.NineRouterRunMode
	}
	settings.NineRouterRunMode = normalizeRunMode(settings.NineRouterRunMode)
	if settings.NineRouterBaseURL == "" {
		settings.NineRouterBaseURL = defaults.NineRouterBaseURL
	}
	settings.NineRouterBaseURL = normalizeLocalHTTPURL(settings.NineRouterBaseURL)
	if settings.NineRouterDashboardURL == "" {
		settings.NineRouterDashboardURL = defaults.NineRouterDashboardURL
	}
	settings.NineRouterDashboardURL = normalizeLocalHTTPURL(settings.NineRouterDashboardURL)
	if settings.NineRouterDefaultModel == "" {
		settings.NineRouterDefaultModel = defaults.NineRouterDefaultModel
	}
	settings.NineRouterDefaultModel = normalizeNineRouterModelID(settings.NineRouterDefaultModel)
	if settings.NineRouterManagedInstallVersion == "" {
		settings.NineRouterManagedInstallVersion = defaults.NineRouterManagedInstallVersion
	}
	if settings.NineRouterManagedPackageName == "" {
		settings.NineRouterManagedPackageName = defaults.NineRouterManagedPackageName
	}
	if settings.NineRouterManagedCommand == "" {
		settings.NineRouterManagedCommand = defaults.NineRouterManagedCommand
	}
	settings.NineRouterManagedTimeoutMS = normalizeTimeout(settings.NineRouterManagedTimeoutMS)
	return settings
}

func normalizeLocalHTTPURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" || !isLoopbackHostname(parsed.Hostname()) {
		return value
	}
	parsed.Scheme = "http"
	return parsed.String()
}

func normalizeNineRouterModelID(model string) string {
	model = strings.TrimSpace(model)
	if strings.HasPrefix(strings.ToLower(model), "kiro/") {
		return "kr/" + strings.TrimSpace(model[len("kiro/"):])
	}
	return model
}

func validSecretToken(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, char := range value {
		if char < 33 || char > 126 {
			return false
		}
	}
	return true
}

func isLoopbackHostname(hostname string) bool {
	host := strings.Trim(strings.ToLower(hostname), "[]")
	if host == "localhost" || host == "0.0.0.0" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func normalizeRunMode(mode RunMode) RunMode {
	if mode == RunModeManaged {
		return RunModeManaged
	}
	return RunModeExternal
}

func normalizeTimeout(value int) int {
	if value <= 0 {
		return 30000
	}
	if value < 5000 {
		return 5000
	}
	if value > 300000 {
		return 300000
	}
	return value
}

func preserveRunningState(value string) string {
	if value == "running" || value == "starting" {
		return value
	}
	return "stopped"
}

func detectMessage(info NodeRuntimeInfo) string {
	if info.LastErrorCode != "" {
		return info.LastErrorMessage
	}
	return "Node 运行时与拓展命令检测完成。"
}

func healthMessage(status ExtensionStatus) string {
	if status.HealthStatus == "connected" {
		return fmt.Sprintf("9Router 已连接，发现 %d 个模型。", status.ModelCount)
	}
	if status.LastErrorMessage != "" {
		return status.LastErrorMessage
	}
	return "9Router 服务可访问，但模型接口尚未验证。"
}

func notFound(extensionID string) *Error {
	return &Error{Code: "EXTENSION_NOT_FOUND", Message: "未找到拓展 " + extensionID + "。", UserAction: "刷新拓展能力列表后重试。"}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return strings.TrimSpace(value)
}

func maskSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func redact(value string) string {
	if value == "" {
		return ""
	}
	replacers := []string{"Authorization", "api_key", "token", "password", "secret", "cookie"}
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lower := strings.ToLower(line)
		for _, key := range replacers {
			if strings.Contains(lower, strings.ToLower(key)) {
				lines[index] = key + ": ***"
				break
			}
		}
		lines[index] = redactSkTokens(lines[index])
	}
	return strings.Join(lines, "\n")
}

func redactSkTokens(value string) string {
	fields := strings.Fields(value)
	for index, field := range fields {
		if strings.HasPrefix(field, "sk-") && len(field) > 10 {
			fields[index] = field[:4] + "***"
		}
	}
	if len(fields) == 0 {
		return value
	}
	return strings.Join(fields, " ")
}

func latencyMS(start time.Time) int {
	return int(time.Since(start).Milliseconds())
}

func errorAs(err error, target any) bool {
	switch typed := target.(type) {
	case **exec.ExitError:
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			*typed = exitErr
		}
		return ok
	default:
		return false
	}
}
