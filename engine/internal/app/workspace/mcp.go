package workspace

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MCPToolBinding struct {
	ToolID     string
	ServerID   string
	ServerTool string
}

type mcpToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Store) ListMCPServers() []MCPServerConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]MCPServerConfig, 0, len(s.servers))
	for _, server := range s.servers {
		result = append(result, server.safe())
	}
	return sortedValuesFromSlice(result, func(item MCPServerConfig) string { return item.DisplayName })
}

func (s *Store) SaveMCPServer(input SaveMCPServerInput) (MCPServerConfig, *AppError) {
	if strings.TrimSpace(input.ServerID) == "" {
		return MCPServerConfig{}, BadRequest("BAD_REQUEST", "MCP serverId is required", "complete the MCP server id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	record, exists := s.servers[input.ServerID]
	if !exists {
		record.CreatedAt = now
		record.Secrets = map[string]string{}
	}
	record.ServerID = strings.TrimSpace(input.ServerID)
	record.DisplayName = strings.TrimSpace(input.DisplayName)
	record.Command = strings.TrimSpace(input.Command)
	record.Args = append([]string{}, input.Args...)
	record.URL = input.URL
	record.TrustLevel = fallback(input.TrustLevel, "local_unverified")
	record.Enabled = input.Enabled
	record.UpdatedAt = now
	if input.Secrets != nil {
		record.Secrets = cloneStringMap(input.Secrets)
	}
	record.EnvKeys, record.MaskedSecrets = secretSummaries(record.Secrets)
	record.HasSecrets = len(record.Secrets) > 0
	s.servers[record.ServerID] = record
	return record.safe(), nil
}

func (s *Store) DeleteMCPServer(serverID string) (DeleteResult, *AppError) {
	if serverID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "missing serverId", "select an MCP server")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.servers, serverID)
	for toolID, binding := range s.mcpTools {
		if binding.ServerID == serverID {
			delete(s.mcpTools, toolID)
			delete(s.tools, toolID)
		}
	}
	return DeleteResult{OK: true, DeletedID: serverID}, nil
}

func (s *Store) TestMCPServer(serverID string) (TestResult, *AppError) {
	if serverID == "" {
		return TestResult{}, BadRequest("BAD_REQUEST", "missing serverId", "select an MCP server")
	}
	s.mu.Lock()
	server, ok := s.servers[serverID]
	s.mu.Unlock()
	if !ok {
		return TestResult{}, NotFound("MCP_NOT_FOUND", "MCP server was not found", "refresh MCP servers")
	}
	startedAt := time.Now()
	if !server.Enabled {
		return TestResult{
			OK:        false,
			TargetID:  serverID,
			Message:   "MCP server is disabled",
			LatencyMS: latencyMS(startedAt),
			TraceID:   s.traceID(),
		}, nil
	}
	tools, err := mcpListTools(context.Background(), server)
	if err != nil {
		return TestResult{
			OK:        false,
			TargetID:  serverID,
			Message:   redactSecrets(err.Error()),
			LatencyMS: latencyMS(startedAt),
			TraceID:   s.traceID(),
		}, nil
	}
	return TestResult{
		OK:        true,
		TargetID:  serverID,
		Message:   fmt.Sprintf("MCP stdio connection ready; discovered %d tools", len(tools)),
		LatencyMS: latencyMS(startedAt),
		TraceID:   s.traceID(),
	}, nil
}

func (s *Store) RefreshMCPTools(serverID string) ([]ToolConfig, *AppError) {
	if serverID == "" {
		return nil, BadRequest("BAD_REQUEST", "missing serverId", "select an MCP server")
	}
	s.mu.Lock()
	server, ok := s.servers[serverID]
	s.mu.Unlock()
	if !ok {
		return nil, NotFound("MCP_NOT_FOUND", "MCP server was not found", "refresh MCP servers")
	}
	tools, err := mcpListTools(context.Background(), server)
	if err != nil {
		return nil, BadRequest("MCP_DISCOVERY_FAILED", redactSecrets(err.Error()), "check the MCP command and refresh again")
	}
	discovered := make([]ToolConfig, 0, len(tools))
	s.mu.Lock()
	defer s.mu.Unlock()
	for toolID, binding := range s.mcpTools {
		if binding.ServerID == serverID {
			delete(s.mcpTools, toolID)
			delete(s.tools, toolID)
		}
	}
	for _, item := range tools {
		toolID := "mcp_" + sanitizeID(serverID) + "_" + sanitizeID(item.Name)
		risk := mcpRiskLevel(server.TrustLevel)
		tool := ToolConfig{
			ToolID:      toolID,
			DisplayName: fallback(item.Name, toolID),
			Description: redactSecrets(fallback(item.Description, "MCP discovered tool")),
			Category:    "project",
			RiskLevel:   risk,
			Enabled:     true,
			BuiltIn:     false,
		}
		s.tools[toolID] = tool
		s.mcpTools[toolID] = MCPToolBinding{ToolID: toolID, ServerID: serverID, ServerTool: item.Name}
		discovered = append(discovered, tool)
	}
	sort.Slice(discovered, func(i, j int) bool { return discovered[i].DisplayName < discovered[j].DisplayName })
	return discovered, nil
}

func (s MCPServerRecord) safe() MCPServerConfig {
	safe := s.MCPServerConfig
	safe.EnvKeys, safe.MaskedSecrets = secretSummaries(s.Secrets)
	safe.HasSecrets = len(s.Secrets) > 0
	return safe
}

func (s *Store) callMCPTool(ctx context.Context, binding MCPToolBinding, arguments string) (string, error) {
	s.mu.Lock()
	server, ok := s.servers[binding.ServerID]
	s.mu.Unlock()
	if !ok || !server.Enabled {
		return "", errors.New("MCP server is unavailable")
	}
	args := decodeToolArguments(arguments)
	result, err := mcpCallTool(ctx, server, binding.ServerTool, args)
	if err != nil {
		return "", err
	}
	return result, nil
}

func mcpListTools(ctx context.Context, server MCPServerRecord) ([]mcpToolInfo, error) {
	client, err := startMCPClient(ctx, server)
	if err != nil {
		return nil, err
	}
	defer client.close()
	if err := client.initialize(); err != nil {
		return nil, err
	}
	var response struct {
		Result struct {
			Tools []mcpToolInfo `json:"tools"`
		} `json:"result"`
		Error *mcpError `json:"error"`
	}
	if err := client.request("tools/list", map[string]any{}, &response); err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, errors.New(response.Error.Message)
	}
	return response.Result.Tools, nil
}

func mcpCallTool(ctx context.Context, server MCPServerRecord, toolName string, arguments map[string]any) (string, error) {
	client, err := startMCPClient(ctx, server)
	if err != nil {
		return "", err
	}
	defer client.close()
	if err := client.initialize(); err != nil {
		return "", err
	}
	var response struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
		Error *mcpError `json:"error"`
	}
	params := map[string]any{"name": toolName, "arguments": arguments}
	if err := client.request("tools/call", params, &response); err != nil {
		return "", err
	}
	if response.Error != nil {
		return "", errors.New(response.Error.Message)
	}
	parts := make([]string, 0, len(response.Result.Content))
	for _, item := range response.Result.Content {
		if item.Text != "" {
			parts = append(parts, redactSecrets(item.Text))
		}
	}
	if len(parts) == 0 {
		return "MCP tool completed with an empty result.", nil
	}
	return strings.Join(parts, "\n"), nil
}

type mcpClient struct {
	cancel context.CancelFunc
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	nextID int
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func startMCPClient(ctx context.Context, server MCPServerRecord) (*mcpClient, error) {
	if server.URL != nil && strings.TrimSpace(*server.URL) != "" {
		return nil, errors.New("HTTP/SSE MCP is not supported in this phase")
	}
	if strings.TrimSpace(server.Command) == "" {
		return nil, errors.New("MCP command is required")
	}
	runCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	cmd := exec.CommandContext(runCtx, server.Command, server.Args...)
	cmd.Env = append(os.Environ(), mcpEnv(server.Secrets)...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start MCP command: %w", err)
	}
	return &mcpClient{cancel: cancel, cmd: cmd, stdin: stdin, reader: bufio.NewReader(stdout)}, nil
}

func (client *mcpClient) close() {
	_ = client.stdin.Close()
	client.cancel()
	_ = client.cmd.Wait()
}

func (client *mcpClient) initialize() error {
	var response struct {
		Error *mcpError `json:"error"`
	}
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "DreamWorker Engine",
			"version": "0.1.0",
		},
	}
	if err := client.request("initialize", params, &response); err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(response.Error.Message)
	}
	return client.notify("notifications/initialized", map[string]any{})
}

func (client *mcpClient) request(method string, params any, target any) error {
	client.nextID++
	requestID := client.nextID
	if err := writeMCPMessage(client.stdin, map[string]any{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  method,
		"params":  params,
	}); err != nil {
		return err
	}
	for {
		payload, err := readMCPMessage(client.reader)
		if err != nil {
			return err
		}
		var envelope struct {
			ID any `json:"id"`
		}
		if err := json.Unmarshal(payload, &envelope); err != nil {
			continue
		}
		if intID(envelope.ID) != requestID {
			continue
		}
		return json.Unmarshal(payload, target)
	}
}

func (client *mcpClient) notify(method string, params any) error {
	return writeMCPMessage(client.stdin, map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
}

func writeMCPMessage(writer io.Writer, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func readMCPMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			contentLength, _ = strconv.Atoi(strings.TrimSpace(value))
		}
	}
	if contentLength <= 0 {
		return nil, errors.New("MCP response missing Content-Length")
	}
	payload := make([]byte, contentLength)
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func intID(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return 0
	}
}

func mcpEnv(secrets map[string]string) []string {
	keys := make([]string, 0, len(secrets))
	for key, value := range secrets {
		if key == "" {
			continue
		}
		keys = append(keys, key+"="+value)
	}
	sort.Strings(keys)
	return keys
}

func mcpRiskLevel(trustLevel string) string {
	switch trustLevel {
	case "trusted_builtin", "verified_partner":
		return "low"
	case "community", "local_unverified":
		return "medium"
	default:
		return "high"
	}
}

func sanitizeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z':
			builder.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			builder.WriteRune(ch)
		default:
			builder.WriteRune('_')
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "tool"
	}
	return result
}
