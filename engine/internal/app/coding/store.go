package coding

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

const maxReadFileBytes int64 = 512 * 1024

type Store struct {
	state *resources.Store
	mu    sync.Mutex

	sessions map[string]Session
	streams  map[string]context.CancelFunc
}

type adapterRequest struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

type adapterResponse struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Event  *adapterEvent   `json:"event,omitempty"`
	Error  *adapterError   `json:"error,omitempty"`
}

type adapterEvent struct {
	Type           string          `json:"type"`
	Delta          string          `json:"delta,omitempty"`
	Message        string          `json:"message,omitempty"`
	CallID         string          `json:"callId,omitempty"`
	ToolName       string          `json:"toolName,omitempty"`
	Arguments      json.RawMessage `json:"arguments,omitempty"`
	Command        string          `json:"command,omitempty"`
	Output         string          `json:"output,omitempty"`
	Path           string          `json:"path,omitempty"`
	Status         string          `json:"status,omitempty"`
	EngineThreadID string          `json:"engineThreadId,omitempty"`
	Error          *StreamError    `json:"error,omitempty"`
}

type adapterError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type adapterProvider struct {
	ProviderID   string `json:"providerId"`
	ProviderType string `json:"providerType"`
	DisplayName  string `json:"displayName"`
	BaseURL      string `json:"baseURL"`
	APIKey       string `json:"apiKey,omitempty"`
}

type adapterTurnParams struct {
	StreamID       string          `json:"streamId"`
	SessionID      string          `json:"sessionId"`
	EngineID       EngineID        `json:"engineId"`
	Provider       adapterProvider `json:"provider"`
	Model          string          `json:"model"`
	Prompt         string          `json:"prompt"`
	CWD            string          `json:"cwd"`
	EngineThreadID string          `json:"engineThreadId,omitempty"`
}

func NewStore(state *resources.Store) *Store {
	return &Store{
		state:    state,
		sessions: make(map[string]Session),
		streams:  make(map[string]context.CancelFunc),
	}
}

func (s *Store) ListEngines() RuntimeStatus {
	status := resolveRuntimeStatus()
	status.Engines = engineDescriptors()
	return status
}

func (s *Store) CreateSession(input CreateSessionInput) (Session, *AppError) {
	if strings.TrimSpace(input.ProjectID) == "" {
		return Session{}, resources.BadRequest("BAD_REQUEST", "missing projectId", "select a project")
	}
	if input.EngineID == "" {
		input.EngineID = EngineClaudeAgent
	}
	if !isSupportedEngine(input.EngineID) {
		return Session{}, resources.BadRequest("UNSUPPORTED_CODING_ENGINE", "unsupported coding engine", "select Claude Agent, Codex, or OpenCode")
	}
	project, root, appErr := s.resolveProjectRoot(input.ProjectID)
	if appErr != nil {
		return Session{}, appErr
	}
	provider, appErr := s.resolveProvider(input.ProviderID, input.Model)
	if appErr != nil {
		return Session{}, appErr
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = provider.DefaultModel
	}
	now := s.state.Now()
	session := Session{
		SessionID:     s.nextID("coding"),
		ProjectID:     project.ProjectID,
		EngineID:      input.EngineID,
		ProviderID:    provider.ProviderID,
		Model:         model,
		Title:         fallback(input.Title, "Coding Agent"),
		LocalRootPath: root,
		Status:        "ready",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.mu.Lock()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()
	return session, nil
}

func (s *Store) GetSession(sessionID string) (Session, *AppError) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return Session{}, resources.NotFound("CODING_SESSION_NOT_FOUND", "coding session not found", "create a new coding session")
	}
	return session, nil
}

func (s *Store) StreamTurn(ctx context.Context, input TurnInput) (<-chan StreamEvent, *AppError) {
	if strings.TrimSpace(input.Prompt) == "" {
		return nil, resources.BadRequest("BAD_REQUEST", "prompt is required", "enter an instruction for the coding agent")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	session, appErr := s.resolveSessionForTurn(input)
	if appErr != nil {
		return nil, appErr
	}
	provider, appErr := s.resolveProvider(session.ProviderID, session.Model)
	if appErr != nil {
		return nil, appErr
	}
	runtime := resolveRuntimeStatus()
	if !runtime.Available {
		return nil, resources.BadRequest("CODING_RUNTIME_UNAVAILABLE", runtime.Message, "run npm run coding:build in dev or package the coding runtime before release")
	}
	streamID := strings.TrimSpace(input.StreamID)
	if streamID == "" {
		streamID = s.nextID("coding_stream")
	}
	traceID := s.state.TraceID()
	streamCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.streams[streamID] = cancel
	session.Status = "running"
	session.UpdatedAt = s.state.Now()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()

	out := make(chan StreamEvent, 32)
	go s.runAdapterTurn(streamCtx, out, runtime, session, provider, input.Prompt, streamID, traceID)
	return out, nil
}

func (s *Store) CancelTurn(input CancelTurnInput) (DeleteResult, *AppError) {
	if strings.TrimSpace(input.StreamID) == "" {
		return DeleteResult{}, resources.BadRequest("BAD_REQUEST", "missing streamId", "select an active coding turn")
	}
	s.mu.Lock()
	cancel, ok := s.streams[input.StreamID]
	if ok {
		delete(s.streams, input.StreamID)
	}
	s.mu.Unlock()
	if !ok {
		return DeleteResult{}, resources.NotFound("CODING_STREAM_NOT_FOUND", "coding stream not found", "the turn may already be finished")
	}
	cancel()
	return DeleteResult{OK: true, DeletedID: input.StreamID}, nil
}

func (s *Store) ListFiles(input ListFilesInput) ([]FileEntry, *AppError) {
	_, root, appErr := s.resolveProjectRoot(input.ProjectID)
	if appErr != nil {
		return nil, appErr
	}
	limit := input.Limit
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	query := strings.ToLower(strings.TrimSpace(input.Query))
	status := s.gitStatusMap(root)
	entries := make([]FileEntry, 0, 128)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if path == root {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() && shouldSkipDir(name) {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if query != "" && !strings.Contains(strings.ToLower(rel), query) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		entries = append(entries, FileEntry{
			Path:       rel,
			Name:       name,
			IsDir:      entry.IsDir(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
			GitStatus:  status[rel],
		})
		if len(entries) >= limit {
			return errStopWalk
		}
		return nil
	})
	if err != nil && !errors.Is(err, errStopWalk) {
		return nil, resources.BadRequest("FILE_TREE_FAILED", "failed to list project files", "check the local project directory")
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func (s *Store) ReadFile(input ReadFileInput) (ReadFileResult, *AppError) {
	_, root, appErr := s.resolveProjectRoot(input.ProjectID)
	if appErr != nil {
		return ReadFileResult{}, appErr
	}
	path, rel, appErr := safeProjectPath(root, input.Path)
	if appErr != nil {
		return ReadFileResult{}, appErr
	}
	info, err := os.Stat(path)
	if err != nil {
		return ReadFileResult{}, resources.NotFound("FILE_NOT_FOUND", "file not found", "select another project file")
	}
	if info.IsDir() {
		return ReadFileResult{}, resources.BadRequest("FILE_IS_DIRECTORY", "path is a directory", "select a file")
	}
	limit := maxReadFileBytes
	size := info.Size()
	truncated := size > limit
	if truncated {
		size = limit
	}
	file, err := os.Open(path)
	if err != nil {
		return ReadFileResult{}, resources.BadRequest("FILE_READ_FAILED", "failed to open file", "check file permissions")
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return ReadFileResult{}, resources.BadRequest("FILE_READ_FAILED", "failed to read file", "check file permissions")
	}
	if int64(len(payload)) > limit {
		payload = payload[:limit]
		truncated = true
	}
	content := string(payload)
	mimeType := "text/plain"
	if !utf8.Valid(payload) {
		content = ""
		mimeType = "application/octet-stream"
		truncated = true
	}
	return ReadFileResult{
		ProjectID: input.ProjectID,
		Path:      rel,
		Content:   content,
		Size:      info.Size(),
		Truncated: truncated,
		MimeType:  mimeType,
	}, nil
}

func (s *Store) FileStatus(input FileStatusInput) (FileStatus, *AppError) {
	_, root, appErr := s.resolveProjectRoot(input.ProjectID)
	if appErr != nil {
		return FileStatus{}, appErr
	}
	changes := s.gitChanges(root)
	branch := s.gitBranch(root)
	message := "git status ready"
	if branch == "" && len(changes) == 0 {
		message = "not a git repository or no git executable available"
	}
	return FileStatus{
		ProjectID: input.ProjectID,
		Branch:    branch,
		Changes:   changes,
		Clean:     len(changes) == 0,
		Message:   message,
	}, nil
}

func (s *Store) runAdapterTurn(
	ctx context.Context,
	out chan<- StreamEvent,
	runtime RuntimeStatus,
	session Session,
	provider resources.ModelProviderRecord,
	prompt string,
	streamID string,
	traceID string,
) {
	defer close(out)
	defer func() {
		s.mu.Lock()
		delete(s.streams, streamID)
		if current, ok := s.sessions[session.SessionID]; ok && current.Status == "running" {
			current.Status = "ready"
			current.UpdatedAt = s.state.Now()
			s.sessions[session.SessionID] = current
		}
		s.mu.Unlock()
	}()
	seq := 0
	emit := func(event StreamEvent) {
		seq++
		event.StreamID = streamID
		event.SessionID = session.SessionID
		event.EngineID = session.EngineID
		event.ProviderID = provider.ProviderID
		event.Model = session.Model
		event.TraceID = traceID
		event.Sequence = seq
		event.Timestamp = s.state.Now()
		out <- event
	}

	cmd := exec.CommandContext(ctx, runtime.NodeBin, runtime.AdapterPath)
	cmd.Dir = session.LocalRootPath
	cmd.Env = os.Environ()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		emit(failedEvent("ADAPTER_STDIN_FAILED", err.Error()))
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		emit(failedEvent("ADAPTER_STDOUT_FAILED", err.Error()))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		emit(failedEvent("ADAPTER_STDERR_FAILED", err.Error()))
		return
	}
	if err := cmd.Start(); err != nil {
		emit(failedEvent("ADAPTER_START_FAILED", err.Error()))
		return
	}

	stderrDone := make(chan string, 1)
	go func() {
		payload, _ := io.ReadAll(stderr)
		stderrDone <- strings.TrimSpace(string(payload))
	}()

	request := adapterRequest{
		ID:     streamID,
		Method: "turn.run",
		Params: adapterTurnParams{
			StreamID:       streamID,
			SessionID:      session.SessionID,
			EngineID:       session.EngineID,
			Provider:       toAdapterProvider(provider),
			Model:          session.Model,
			Prompt:         prompt,
			CWD:            session.LocalRootPath,
			EngineThreadID: session.EngineThreadID,
		},
	}
	if err := json.NewEncoder(stdin).Encode(request); err != nil {
		_ = cmd.Process.Kill()
		emit(failedEvent("ADAPTER_WRITE_FAILED", err.Error()))
		return
	}
	_ = stdin.Close()

	terminalSeen := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024), 2*1024*1024)
	for scanner.Scan() {
		var response adapterResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			continue
		}
		if response.Error != nil {
			terminalSeen = true
			emit(failedEvent(response.Error.Code, response.Error.Message))
			continue
		}
		if response.Event == nil {
			continue
		}
		event := mapAdapterEvent(*response.Event)
		if event.EngineThreadID != "" {
			s.updateSessionThread(session.SessionID, event.EngineThreadID)
		}
		if event.Type == "completed" || event.Type == "cancelled" || event.Type == "error" {
			terminalSeen = true
		}
		emit(event)
	}
	waitErr := cmd.Wait()
	stderrText := <-stderrDone
	if ctx.Err() != nil {
		emit(StreamEvent{Type: "cancelled", Message: "coding turn cancelled"})
		return
	}
	if err := scanner.Err(); err != nil {
		emit(failedEvent("ADAPTER_READ_FAILED", err.Error()))
		return
	}
	if waitErr != nil && !terminalSeen {
		message := waitErr.Error()
		if stderrText != "" {
			message = message + ": " + resources.RedactSecrets(stderrText)
		}
		emit(failedEvent("ADAPTER_EXIT_FAILED", message))
		return
	}
	if !terminalSeen {
		emit(StreamEvent{Type: "completed", Message: "coding turn completed"})
	}
}

func (s *Store) resolveSessionForTurn(input TurnInput) (Session, *AppError) {
	if input.SessionID != "" {
		session, appErr := s.GetSession(input.SessionID)
		if appErr != nil {
			return Session{}, appErr
		}
		if input.EngineID != "" {
			session.EngineID = input.EngineID
		}
		if input.ProviderID != "" {
			session.ProviderID = input.ProviderID
		}
		if input.Model != "" {
			session.Model = input.Model
		}
		_, root, appErr := s.resolveProjectRoot(session.ProjectID)
		if appErr != nil {
			return Session{}, appErr
		}
		session.LocalRootPath = root
		s.mu.Lock()
		s.sessions[session.SessionID] = session
		s.mu.Unlock()
		return session, nil
	}
	return s.CreateSession(CreateSessionInput{
		ProjectID:  input.ProjectID,
		EngineID:   input.EngineID,
		ProviderID: input.ProviderID,
		Model:      input.Model,
		Title:      "Coding Agent",
	})
}

func (s *Store) resolveProjectRoot(projectID string) (resources.Project, string, *AppError) {
	s.state.Mu.Lock()
	project, ok := s.state.Projects[projectID]
	s.state.Mu.Unlock()
	if !ok {
		return resources.Project{}, "", resources.NotFound("PROJECT_NOT_FOUND", "project not found", "select another project")
	}
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		return resources.Project{}, "", resources.BadRequest("LOCAL_DIRECTORY_NOT_SET", "project has no localRootPath", "bind a local project directory first")
	}
	root := filepath.Clean(*project.LocalRootPath)
	realRoot, err := filepath.EvalSymlinks(root)
	if err == nil {
		root = realRoot
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return resources.Project{}, "", resources.BadRequest("LOCAL_DIRECTORY_INVALID", "project localRootPath is not an available directory", "check project settings")
	}
	if !directoryWritable(root) {
		return resources.Project{}, "", resources.BadRequest("LOCAL_DIRECTORY_NOT_WRITABLE", "project localRootPath is not writable", "check directory permissions")
	}
	return project, root, nil
}

func (s *Store) resolveProvider(providerID string, model string) (resources.ModelProviderRecord, *AppError) {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()
	if providerID == "" {
		for _, provider := range s.state.Providers {
			if provider.Enabled {
				if model == "" || contains(provider.AvailableModels, model) || provider.DefaultModel == model {
					return provider, nil
				}
			}
		}
		return resources.ModelProviderRecord{}, resources.NotFound("PROVIDER_NOT_FOUND", "no enabled model provider found", "configure a provider")
	}
	provider, ok := s.state.Providers[providerID]
	if !ok {
		return resources.ModelProviderRecord{}, resources.NotFound("PROVIDER_NOT_FOUND", "provider not found", "select another provider")
	}
	if !provider.Enabled {
		return resources.ModelProviderRecord{}, resources.BadRequest("PROVIDER_DISABLED", "provider is disabled", "enable the provider")
	}
	return provider, nil
}

func (s *Store) nextID(prefix string) string {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()
	return s.state.NextIDLocked(prefix)
}

func (s *Store) updateSessionThread(sessionID string, threadID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return
	}
	session.EngineThreadID = threadID
	session.UpdatedAt = s.state.Now()
	s.sessions[sessionID] = session
}

func resolveRuntimeStatus() RuntimeStatus {
	runtimeDir := strings.TrimSpace(os.Getenv("DREAMWORKER_CODING_AGENT_RUNTIME_DIR"))
	nodeBin := strings.TrimSpace(os.Getenv("DREAMWORKER_CODING_AGENT_NODE_BIN"))
	adapterPath := ""
	if runtimeDir != "" {
		for _, candidate := range []string{
			filepath.Join(runtimeDir, "adapter", "dist", "index.js"),
			filepath.Join(runtimeDir, "dist", "index.js"),
		} {
			if fileExists(candidate) {
				adapterPath = candidate
				break
			}
		}
		if nodeBin == "" {
			packagedNode := filepath.Join(runtimeDir, "node", nodeExecutableName())
			if fileExists(packagedNode) {
				nodeBin = packagedNode
			}
		}
	}
	if nodeBin == "" {
		nodeBin = "node"
	}
	status := RuntimeStatus{
		RuntimeDir:  runtimeDir,
		NodeBin:     nodeBin,
		AdapterPath: adapterPath,
		Available:   adapterPath != "",
		Message:     "coding agent runtime is ready",
	}
	if adapterPath == "" {
		status.Message = "coding agent adapter was not found"
	}
	return status
}

func engineDescriptors() []EngineDescriptor {
	return []EngineDescriptor{
		{
			EngineID:               EngineClaudeAgent,
			DisplayName:            "Claude Agent",
			Description:            "Anthropic Claude Agent SDK, cwd scoped to project localRootPath.",
			SupportedProviderTypes: []string{"anthropic"},
			PreferredProviderIDs:   []string{"provider_anthropic"},
			DirectWrite:            true,
			Streaming:              true,
		},
		{
			EngineID:               EngineCodex,
			DisplayName:            "Codex",
			Description:            "OpenAI Codex SDK thread run with workspace-write sandbox.",
			SupportedProviderTypes: []string{"openai", "openai_compatible", "deepseek", "siliconflow", "glm", "custom"},
			PreferredProviderIDs:   []string{"provider_9router_local", "provider_openai"},
			DirectWrite:            true,
			Streaming:              false,
		},
		{
			EngineID:               EngineOpenCode,
			DisplayName:            "OpenCode",
			Description:            "OpenCode SDK server/client prompt runtime.",
			SupportedProviderTypes: []string{"openai", "openai_compatible", "deepseek", "siliconflow", "glm", "ollama", "custom"},
			PreferredProviderIDs:   []string{"provider_9router_local"},
			DirectWrite:            true,
			Streaming:              true,
		},
	}
}

func isSupportedEngine(engineID EngineID) bool {
	return engineID == EngineClaudeAgent || engineID == EngineCodex || engineID == EngineOpenCode
}

func mapAdapterEvent(event adapterEvent) StreamEvent {
	result := StreamEvent{
		Type:           event.Type,
		Delta:          event.Delta,
		Message:        event.Message,
		Command:        event.Command,
		Output:         event.Output,
		Path:           event.Path,
		Status:         event.Status,
		EngineThreadID: event.EngineThreadID,
		Error:          event.Error,
	}
	if result.Type == "error" {
		result.Type = "error"
	}
	if event.ToolName != "" || len(event.Arguments) > 0 {
		var arguments any
		if len(event.Arguments) > 0 {
			_ = json.Unmarshal(event.Arguments, &arguments)
		}
		result.ToolCall = &ToolCall{
			CallID:    event.CallID,
			ToolName:  event.ToolName,
			Arguments: arguments,
		}
	}
	if event.Path != "" || event.Status != "" {
		result.File = &FileChange{Path: event.Path, Status: event.Status}
	}
	return result
}

func failedEvent(code string, message string) StreamEvent {
	return StreamEvent{
		Type:    "error",
		Message: message,
		Error:   &StreamError{Code: fallback(code, "CODING_AGENT_FAILED"), Message: resources.RedactSecrets(message), Recoverable: true},
	}
}

func toAdapterProvider(provider resources.ModelProviderRecord) adapterProvider {
	return adapterProvider{
		ProviderID:   provider.ProviderID,
		ProviderType: string(provider.ProviderType),
		DisplayName:  provider.DisplayName,
		BaseURL:      provider.BaseURL,
		APIKey:       provider.APIKey,
	}
}

func safeProjectPath(root string, raw string) (string, string, *AppError) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if trimmed == "" || strings.HasPrefix(trimmed, "/") || strings.Contains(trimmed, "\x00") {
		return "", "", resources.BadRequest("PATH_OUTSIDE_PROJECT", "file path must be relative to the project", "select a project file")
	}
	cleanRel := filepath.Clean(filepath.FromSlash(trimmed))
	if cleanRel == "." || strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return "", "", resources.BadRequest("PATH_OUTSIDE_PROJECT", "file path escapes the project root", "select a file inside the project")
	}
	joined := filepath.Join(root, cleanRel)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", resources.BadRequest("LOCAL_DIRECTORY_INVALID", "project root is invalid", "check project settings")
	}
	absPath, err := filepath.Abs(joined)
	if err != nil {
		return "", "", resources.BadRequest("PATH_OUTSIDE_PROJECT", "file path is invalid", "select another file")
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", resources.BadRequest("PATH_OUTSIDE_PROJECT", "file path escapes the project root", "select a file inside the project")
	}
	if realPath, err := filepath.EvalSymlinks(absPath); err == nil {
		realRoot, rootErr := filepath.EvalSymlinks(absRoot)
		if rootErr == nil {
			realRel, relErr := filepath.Rel(realRoot, realPath)
			if relErr != nil || realRel == ".." || strings.HasPrefix(realRel, ".."+string(filepath.Separator)) {
				return "", "", resources.BadRequest("PATH_OUTSIDE_PROJECT", "resolved file path escapes the project root", "select a file inside the project")
			}
		}
	}
	return absPath, filepath.ToSlash(rel), nil
}

func (s *Store) gitStatusMap(root string) map[string]string {
	changes := s.gitChanges(root)
	result := make(map[string]string, len(changes))
	for _, change := range changes {
		result[change.Path] = change.Status
	}
	return result
}

func (s *Store) gitChanges(root string) []FileChange {
	cmd := exec.Command("git", "status", "--short", "--porcelain=v1")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	changes := []FileChange{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[len(parts)-1]
		}
		changes = append(changes, FileChange{Path: filepath.ToSlash(path), Status: status})
	}
	return changes
}

func (s *Store) gitBranch(root string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

var errStopWalk = errors.New("stop walking project files")

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "dist", "out", "release", "coverage", ".cache", ".vite", "tmp":
		return true
	default:
		return false
	}
}

func directoryWritable(path string) bool {
	file, err := os.CreateTemp(path, ".dreamworker-coding-write-test-*")
	if err != nil {
		return false
	}
	name := file.Name()
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(name)
		return false
	}
	return os.Remove(name) == nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func nodeExecutableName() string {
	if os.PathSeparator == '\\' {
		return "node.exe"
	}
	return "node"
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func fallback(value string, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
