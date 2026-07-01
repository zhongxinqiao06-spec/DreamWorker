package modelgateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

type Gateway struct{}

type modelMessage = ports.ChatModelMessage
type modelStreamChunk = ports.ChatModelStreamChunk
type ToolExecutionRequest = ports.ToolExecutionRequest
type ModelProviderRecord = ports.ChatModelProvider
type ModelProfile = ports.ChatModelProfile
type ChatModelUsage = ports.ChatModelUsage
type ChatStreamError = ports.ChatStreamError

const (
	ProviderOpenAICompatible = "openai_compatible"
	ProviderDeepSeek         = "deepseek"
	ProviderOpenAI           = "openai"
	ProviderAnthropic        = "anthropic"
	ProviderGLM              = "glm"
	ProviderVolcano          = "volcano"
	ProviderSiliconFlow      = "siliconflow"
	ProviderGemini           = "gemini"
	ProviderOllama           = "ollama"
	ProviderCustom           = "custom"
)

func NewGateway() *Gateway {
	return &Gateway{}
}

func (Gateway) DiscoverModels(ctx context.Context, provider ports.ChatModelProvider) ports.ProviderModelDiscoveryResult {
	start := time.Now()
	models, err := discoverProviderModels(ctx, provider)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return ports.ProviderModelDiscoveryResult{
			Models:    append([]string{}, provider.AvailableModels...),
			LatencyMS: latency,
			ErrorCode: modelErrorCode(err),
			LastError: sanitizeProviderError(err.Error()),
		}
	}
	return ports.ProviderModelDiscoveryResult{
		Models:     models,
		LatencyMS:  latency,
		Discovered: true,
	}
}

func (Gateway) HealthCheck(ctx context.Context, provider ports.ChatModelProvider) ports.ProviderHealth {
	start := time.Now()
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return ports.ProviderHealth{
			OK:                provider.Enabled,
			Status:            "connected",
			Message:           "local deterministic streaming provider is ready",
			LatencyMS:         int(time.Since(start).Milliseconds()),
			StreamingVerified: true,
		}
	}
	if !provider.Enabled {
		return ports.ProviderHealth{
			Status:    "unknown",
			Message:   "provider is disabled",
			LatencyMS: int(time.Since(start).Milliseconds()),
			ErrorCode: "MODEL_PROVIDER_DISABLED",
		}
	}
	if provider.ProviderType != ProviderOllama && provider.APIKey == "" {
		return ports.ProviderHealth{
			Status:    "error",
			Message:   "provider api key is missing",
			LatencyMS: int(time.Since(start).Milliseconds()),
			ErrorCode: "MODEL_API_KEY_MISSING",
		}
	}
	if provider.APIKey == "sk-local-demo" {
		return ports.ProviderHealth{
			Status:    "error",
			Message:   "demo key cannot call real provider",
			LatencyMS: int(time.Since(start).Milliseconds()),
			ErrorCode: "MODEL_API_KEY_DEMO",
		}
	}
	err := checkProviderHealth(ctx, provider)
	if err != nil {
		return ports.ProviderHealth{
			Status:    "error",
			Message:   sanitizeProviderError(err.Error()),
			LatencyMS: int(time.Since(start).Milliseconds()),
			ErrorCode: modelErrorCode(err),
		}
	}
	return ports.ProviderHealth{
		OK:                true,
		Status:            "connected",
		Message:           "provider connection is ready",
		LatencyMS:         int(time.Since(start).Milliseconds()),
		StreamingVerified: providerSupportsStreaming(provider.ProviderType),
	}
}

func checkProviderHealth(ctx context.Context, provider ports.ChatModelProvider) error {
	switch provider.ProviderType {
	case ProviderOpenAICompatible, ProviderDeepSeek, ProviderGLM, ProviderVolcano, ProviderSiliconFlow, ProviderGemini, ProviderCustom:
		return probeOpenAICompatibleChat(ctx, provider)
	default:
		_, err := fetchProviderModels(ctx, provider)
		return err
	}
}

func (Gateway) StreamChat(
	ctx context.Context,
	provider ports.ChatModelProvider,
	profile ports.ChatModelProfile,
	messages []ports.ChatModelMessage,
) <-chan ports.ChatModelStreamChunk {
	return streamProviderModel(ctx, provider, profile, messages)
}

func streamProviderModel(
	ctx context.Context,
	provider ModelProviderRecord,
	profile ModelProfile,
	messages []modelMessage,
) <-chan modelStreamChunk {
	out := make(chan modelStreamChunk, 16)
	go func() {
		defer close(out)
		if profile.Model == "model_generate_stub" || provider.ProviderID == "provider_local_stub" {
			streamLocalStub(ctx, profile, messages, out)
			return
		}
		if !provider.Enabled {
			out <- modelStreamChunk{Error: streamError("MODEL_PROVIDER_DISABLED", "provider is disabled", true)}
			return
		}
		if provider.ProviderType != ProviderOllama && provider.APIKey == "" {
			out <- modelStreamChunk{Error: streamError("MODEL_API_KEY_MISSING", "provider api key is missing", true)}
			return
		}
		if provider.APIKey == "sk-local-demo" {
			out <- modelStreamChunk{Error: streamError("MODEL_API_KEY_DEMO", "demo key cannot call real provider", true)}
			return
		}
		var err error
		switch provider.ProviderType {
		case ProviderOpenAI:
			err = streamOpenAIResponses(ctx, provider, profile, messages, out)
		case ProviderAnthropic:
			err = streamAnthropicMessages(ctx, provider, profile, messages, out)
		case ProviderOllama:
			err = streamOllamaChat(ctx, provider, profile, messages, out)
		case ProviderOpenAICompatible, ProviderDeepSeek, ProviderGLM, ProviderVolcano, ProviderSiliconFlow, ProviderGemini, ProviderCustom:
			err = streamOpenAICompatibleChat(ctx, provider, profile, messages, out)
		default:
			err = errors.New("model provider is not supported")
		}
		if err != nil {
			out <- modelStreamChunk{Error: streamError(modelErrorCode(err), err.Error(), true)}
		}
	}()
	return out
}

func streamLocalStub(ctx context.Context, profile ModelProfile, messages []modelMessage, out chan<- modelStreamChunk) {
	last := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			last = messages[i].Content
			break
		}
	}
	reply := fmt.Sprintf("Local streaming model received: %s\n\nRuntime path: PLAN -> GRAPH -> EXECUTE -> OBSERVE -> REPLAN. Configure a real provider in Resource Center to call external streaming models.", last)
	parts := splitForStreaming(reply)
	for _, part := range parts {
		select {
		case <-ctx.Done():
			out <- modelStreamChunk{FinishReason: "cancelled"}
			return
		case out <- modelStreamChunk{Delta: part}:
		}
	}
	usage := estimateChatUsage(messages, reply)
	out <- modelStreamChunk{Usage: &usage, FinishReason: "stop"}
	_ = profile
}

func streamOpenAIResponses(
	ctx context.Context,
	provider ModelProviderRecord,
	profile ModelProfile,
	messages []modelMessage,
	out chan<- modelStreamChunk,
) error {
	input := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		input = append(input, map[string]any{
			"role": message.Role,
			"content": []map[string]string{{
				"type": "input_text",
				"text": message.Content,
			}},
		})
	}
	body := map[string]any{
		"model":             profile.Model,
		"input":             input,
		"temperature":       profile.Temperature,
		"max_output_tokens": profile.MaxTokens,
		"stream":            true,
	}
	return streamSSERequest(ctx, provider, openAIResponsesEndpoint(provider.BaseURL), body, openAIHeaders(provider), func(_ string, data string) error {
		if strings.TrimSpace(data) == "[DONE]" {
			return nil
		}
		var event struct {
			Type     string `json:"type"`
			Delta    string `json:"delta"`
			Response struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
					TotalTokens  int `json:"total_tokens"`
				} `json:"usage"`
				Error *struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			} `json:"response"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil
		}
		switch event.Type {
		case "response.output_text.delta":
			if event.Delta != "" {
				out <- modelStreamChunk{Delta: event.Delta}
			}
		case "response.completed":
			usage := ChatModelUsage{
				InputTokens:  event.Response.Usage.InputTokens,
				OutputTokens: event.Response.Usage.OutputTokens,
				TotalTokens:  event.Response.Usage.TotalTokens,
			}
			out <- modelStreamChunk{Usage: &usage, FinishReason: "stop"}
		case "response.failed", "error":
			code := "MODEL_STREAM_FAILED"
			message := "provider stream failed"
			if event.Error != nil {
				code = event.Error.Code
				message = event.Error.Message
			}
			if event.Response.Error != nil {
				code = event.Response.Error.Code
				message = event.Response.Error.Message
			}
			out <- modelStreamChunk{Error: streamError(code, message, true)}
		}
		return nil
	})
}

func streamOpenAICompatibleChat(
	ctx context.Context,
	provider ModelProviderRecord,
	profile ModelProfile,
	messages []modelMessage,
	out chan<- modelStreamChunk,
) error {
	body := map[string]any{
		"model":       profile.Model,
		"messages":    messages,
		"temperature": profile.Temperature,
		"max_tokens":  profile.MaxTokens,
		"stream":      true,
	}
	return streamSSERequest(ctx, provider, chatCompletionsEndpoint(provider), body, bearerHeaders(provider.APIKey), func(_ string, data string) error {
		if strings.TrimSpace(data) == "[DONE]" {
			return nil
		}
		var event struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
					Reasoning        string `json:"reasoning"`
					ToolCalls        []struct {
						ID       string `json:"id"`
						Index    int    `json:"index"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil
		}
		if event.Error != nil {
			out <- modelStreamChunk{Error: streamError(event.Error.Code, event.Error.Message, true)}
			return nil
		}
		for _, choice := range event.Choices {
			reasoningDelta := strings.TrimSpace(choice.Delta.ReasoningContent)
			if reasoningDelta == "" {
				reasoningDelta = strings.TrimSpace(choice.Delta.Reasoning)
			}
			if reasoningDelta != "" {
				out <- modelStreamChunk{ReasoningDelta: reasoningDelta}
			}
			if choice.Delta.Content != "" {
				out <- modelStreamChunk{Delta: choice.Delta.Content}
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				if toolCall.Function.Name == "" && toolCall.ID == "" {
					continue
				}
				out <- modelStreamChunk{ToolCall: &ToolExecutionRequest{
					CallID:      fallback(toolCall.ID, fmt.Sprintf("tool_call_%d", toolCall.Index)),
					ToolID:      toolCall.Function.Name,
					DisplayName: toolCall.Function.Name,
					Arguments:   toolCall.Function.Arguments,
				}}
			}
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				out <- modelStreamChunk{FinishReason: *choice.FinishReason}
			}
		}
		if event.Usage != nil {
			usage := ChatModelUsage{
				InputTokens:  event.Usage.PromptTokens,
				OutputTokens: event.Usage.CompletionTokens,
				TotalTokens:  event.Usage.TotalTokens,
			}
			out <- modelStreamChunk{Usage: &usage}
		}
		return nil
	})
}

func streamAnthropicMessages(
	ctx context.Context,
	provider ModelProviderRecord,
	profile ModelProfile,
	messages []modelMessage,
	out chan<- modelStreamChunk,
) error {
	system := ""
	anthropicMessages := make([]modelMessage, 0, len(messages))
	for _, message := range messages {
		if message.Role == "system" {
			if system != "" {
				system += "\n\n"
			}
			system += message.Content
			continue
		}
		role := message.Role
		if role != "assistant" {
			role = "user"
		}
		anthropicMessages = append(anthropicMessages, modelMessage{Role: role, Content: message.Content})
	}
	body := map[string]any{
		"model":       profile.Model,
		"messages":    anthropicMessages,
		"max_tokens":  profile.MaxTokens,
		"temperature": profile.Temperature,
		"stream":      true,
	}
	if system != "" {
		body["system"] = system
	}
	return streamSSERequest(ctx, provider, anthropicMessagesEndpoint(provider.BaseURL), body, anthropicHeaders(provider), func(_ string, data string) error {
		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type       string `json:"type"`
				Text       string `json:"text"`
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
			Message struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"message"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
			Error *struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil
		}
		switch event.Type {
		case "content_block_delta":
			if event.Delta.Text != "" {
				out <- modelStreamChunk{Delta: event.Delta.Text}
			}
		case "message_delta":
			usage := ChatModelUsage{
				InputTokens:  event.Usage.InputTokens,
				OutputTokens: event.Usage.OutputTokens,
				TotalTokens:  event.Usage.InputTokens + event.Usage.OutputTokens,
			}
			out <- modelStreamChunk{Usage: &usage, FinishReason: event.Delta.StopReason}
		case "message_stop":
			out <- modelStreamChunk{FinishReason: "stop"}
		case "error":
			code := "ANTHROPIC_STREAM_ERROR"
			message := "anthropic stream failed"
			if event.Error != nil {
				code = event.Error.Type
				message = event.Error.Message
			}
			out <- modelStreamChunk{Error: streamError(code, message, true)}
		}
		if event.Type == "message_start" {
			usage := ChatModelUsage{
				InputTokens: event.Message.Usage.InputTokens,
				TotalTokens: event.Message.Usage.InputTokens,
			}
			out <- modelStreamChunk{Usage: &usage}
		}
		return nil
	})
}

func streamOllamaChat(
	ctx context.Context,
	provider ModelProviderRecord,
	profile ModelProfile,
	messages []modelMessage,
	out chan<- modelStreamChunk,
) error {
	body := map[string]any{
		"model":    profile.Model,
		"messages": messages,
		"stream":   true,
		"options": map[string]any{
			"temperature": profile.Temperature,
			"num_predict": profile.MaxTokens,
		},
	}
	resp, err := jsonPost(ctx, ollamaChatEndpoint(provider.BaseURL), body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done            bool   `json:"done"`
			DoneReason      string `json:"done_reason"`
			PromptEvalCount int    `json:"prompt_eval_count"`
			EvalCount       int    `json:"eval_count"`
			Error           string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Error != "" {
			out <- modelStreamChunk{Error: streamError("OLLAMA_STREAM_ERROR", event.Error, true)}
			continue
		}
		if event.Message.Content != "" {
			out <- modelStreamChunk{Delta: event.Message.Content}
		}
		if event.Done {
			usage := ChatModelUsage{
				InputTokens:  event.PromptEvalCount,
				OutputTokens: event.EvalCount,
				TotalTokens:  event.PromptEvalCount + event.EvalCount,
			}
			out <- modelStreamChunk{Usage: &usage, FinishReason: fallback(event.DoneReason, "stop")}
		}
	}
	return scanner.Err()
}

func streamSSERequest(
	ctx context.Context,
	provider ModelProviderRecord,
	endpoint string,
	body any,
	headers map[string]string,
	handle func(event string, data string) error,
) error {
	resp, err := jsonPost(ctx, endpoint, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}
	return scanSSE(resp.Body, handle)
}

func scanSSE(body io.Reader, handle func(event string, data string) error) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	eventName := ""
	var dataLines []string
	dispatch := func() error {
		if len(dataLines) == 0 {
			return nil
		}
		data := strings.Join(dataLines, "\n")
		dataLines = nil
		currentEvent := eventName
		eventName = ""
		return handle(currentEvent, data)
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if value, ok := strings.CutPrefix(line, "event:"); ok {
			eventName = strings.TrimSpace(value)
			continue
		}
		if value, ok := strings.CutPrefix(line, "data:"); ok {
			dataLines = append(dataLines, strings.TrimSpace(value))
		}
	}
	if err := dispatch(); err != nil {
		return err
	}
	return scanner.Err()
}

func jsonPost(ctx context.Context, endpoint string, body any, headers map[string]string) (*http.Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "text/event-stream, application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{Timeout: 0}
	return client.Do(req)
}

func discoverProviderModels(ctx context.Context, provider ModelProviderRecord) ([]string, error) {
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return []string{"model_generate_stub"}, nil
	}
	if !provider.Enabled {
		return nil, errors.New("provider is disabled")
	}
	if provider.APIKey == "sk-local-demo" {
		return nil, errors.New("demo key cannot call real provider")
	}
	if provider.ProviderType != ProviderOllama && provider.APIKey == "" {
		return nil, errors.New("provider api key is missing")
	}
	return fetchProviderModels(ctx, provider)
}

func fetchProviderModels(ctx context.Context, provider ModelProviderRecord) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var endpoint string
	headers := map[string]string{}
	switch provider.ProviderType {
	case ProviderAnthropic:
		endpoint = anthropicModelsEndpoint(provider.BaseURL)
		headers = anthropicHeaders(provider)
	case ProviderOllama:
		endpoint = ollamaTagsEndpoint(provider.BaseURL)
	case ProviderSiliconFlow:
		endpoint = siliconFlowModelsEndpoint(provider.BaseURL)
		headers = bearerHeaders(provider.APIKey)
	default:
		endpoint = modelsEndpoint(provider.BaseURL)
		headers = bearerHeaders(provider.APIKey)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError(resp)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Models []struct {
			Name  string `json:"name"`
			Model string `json:"model"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	var models []string
	for _, item := range payload.Data {
		if item.ID != "" {
			models = append(models, item.ID)
		}
	}
	for _, item := range payload.Models {
		if item.Name != "" {
			models = append(models, item.Name)
		} else if item.Model != "" {
			models = append(models, item.Model)
		}
	}
	if len(models) == 0 {
		return nil, errors.New("provider returned no models")
	}
	return models, nil
}

func probeOpenAICompatibleChat(ctx context.Context, provider ModelProviderRecord) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	model := provider.DefaultModel
	if model == "" && len(provider.AvailableModels) > 0 {
		model = provider.AvailableModels[0]
	}
	if model == "" {
		return errors.New("provider default model is missing")
	}
	body := map[string]any{
		"model":       model,
		"messages":    []modelMessage{{Role: "user", Content: "ping"}},
		"temperature": 0,
		"max_tokens":  4,
		"stream":      false,
	}
	resp, err := jsonPost(ctx, chatCompletionsEndpoint(provider), body, bearerHeaders(provider.APIKey))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Error != nil {
		return fmt.Errorf("%s: %s", payload.Error.Code, payload.Error.Message)
	}
	if len(payload.Choices) == 0 {
		return errors.New("provider returned no chat choices")
	}
	return nil
}

func responseError(resp *http.Response) error {
	limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	message := sanitizeProviderError(string(limited))
	if message == "" {
		message = resp.Status
	}
	return fmt.Errorf("provider returned %s: %s", resp.Status, message)
}

func bearerHeaders(apiKey string) map[string]string {
	if apiKey == "" {
		return nil
	}
	return map[string]string{"Authorization": "Bearer " + apiKey}
}

func openAIHeaders(provider ModelProviderRecord) map[string]string {
	headers := bearerHeaders(provider.APIKey)
	if provider.Organization != nil && *provider.Organization != "" {
		headers["OpenAI-Organization"] = *provider.Organization
	}
	if provider.Project != nil && *provider.Project != "" {
		headers["OpenAI-Project"] = *provider.Project
	}
	return headers
}

func anthropicHeaders(provider ModelProviderRecord) map[string]string {
	headers := map[string]string{
		"x-api-key":         provider.APIKey,
		"anthropic-version": "2023-06-01",
	}
	return headers
}

func modelsEndpoint(baseURL string) string {
	return joinURLPath(baseURL, "models")
}

func siliconFlowModelsEndpoint(baseURL string) string {
	endpoint := modelsEndpoint(baseURL)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := parsed.Query()
	if query.Get("type") == "" {
		query.Set("type", "text")
	}
	if query.Get("sub_type") == "" {
		query.Set("sub_type", "chat")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func openAIResponsesEndpoint(baseURL string) string {
	return joinURLPath(ensureVersionedBaseURL(baseURL, "v1"), "responses")
}

func chatCompletionsEndpoint(provider ModelProviderRecord) string {
	baseURL := provider.BaseURL
	if provider.ProviderType == ProviderOpenAI || provider.ProviderType == ProviderOpenAICompatible || provider.ProviderType == ProviderCustom {
		baseURL = ensureVersionedBaseURL(baseURL, "v1")
	}
	return joinURLPath(baseURL, "chat/completions")
}

func anthropicMessagesEndpoint(baseURL string) string {
	return joinURLPath(ensureVersionedBaseURL(baseURL, "v1"), "messages")
}

func anthropicModelsEndpoint(baseURL string) string {
	return joinURLPath(ensureVersionedBaseURL(baseURL, "v1"), "models")
}

func ollamaChatEndpoint(baseURL string) string {
	return joinURLPath(defaultBaseURL(baseURL, "http://127.0.0.1:11434"), "api/chat")
}

func ollamaTagsEndpoint(baseURL string) string {
	return joinURLPath(defaultBaseURL(baseURL, "http://127.0.0.1:11434"), "api/tags")
}

func ensureVersionedBaseURL(raw string, version string) string {
	value := defaultBaseURL(raw, "https://api.openai.com")
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		parsed.Path = "/" + version
		return parsed.String()
	}
	segments := strings.Split(path, "/")
	last := segments[len(segments)-1]
	if strings.HasPrefix(last, "v") {
		return parsed.String()
	}
	return parsed.String()
}

func joinURLPath(baseURL string, suffix string) string {
	parsed, err := url.Parse(defaultBaseURL(baseURL, "http://127.0.0.1"))
	if err != nil {
		return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(suffix, "/")
	}
	basePath := strings.TrimRight(parsed.Path, "/")
	suffix = strings.TrimLeft(suffix, "/")
	if basePath == "" {
		parsed.Path = "/" + suffix
	} else {
		parsed.Path = basePath + "/" + suffix
	}
	return parsed.String()
}

func defaultBaseURL(value string, fallbackValue string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallbackValue
	}
	return value
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func splitForStreaming(value string) []string {
	words := strings.Split(value, " ")
	if len(words) <= 1 {
		return []string{value}
	}
	parts := make([]string, 0, len(words))
	for i, word := range words {
		if i == 0 {
			parts = append(parts, word)
		} else {
			parts = append(parts, " "+word)
		}
	}
	return parts
}

func estimateChatUsage(messages []modelMessage, output string) ChatModelUsage {
	input := 0
	for _, message := range messages {
		input += estimateTokens(message.Content)
	}
	outputTokens := estimateTokens(output)
	return ChatModelUsage{InputTokens: input, OutputTokens: outputTokens, TotalTokens: input + outputTokens}
}

func estimateTokens(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	return len([]rune(trimmed))/4 + 1
}

func streamError(code string, message string, recoverable bool) *ChatStreamError {
	if code == "" {
		code = "MODEL_STREAM_FAILED"
	}
	if message == "" {
		message = "model stream failed"
	}
	return &ChatStreamError{Code: code, Message: sanitizeProviderError(message), Recoverable: recoverable}
}

func modelErrorCode(err error) string {
	if errors.Is(err, context.Canceled) {
		return "MODEL_STREAM_CANCELLED"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "MODEL_STREAM_TIMEOUT"
	}
	return "MODEL_STREAM_FAILED"
}

func providerSupportsStreaming(providerType string) bool {
	switch providerType {
	case ProviderOpenAI, ProviderOpenAICompatible, ProviderDeepSeek, ProviderGLM, ProviderVolcano, ProviderSiliconFlow, ProviderAnthropic, ProviderOllama:
		return true
	default:
		return false
	}
}

func sanitizeProviderError(message string) string {
	value := strings.TrimSpace(message)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	for _, marker := range []string{"sk-", "sk_live_", "sk_test_", "Bearer "} {
		for {
			index := strings.Index(value, marker)
			if index < 0 {
				break
			}
			end := index + len(marker)
			for end < len(value) {
				ch := value[end]
				if !(ch == '-' || ch == '_' || ch == '.' || ch == ':' || ch == '/' || ch == '=' || ch == '+' || (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')) {
					break
				}
				end++
			}
			value = value[:index] + "[redacted]" + value[end:]
		}
	}
	if len([]rune(value)) > 280 {
		runes := []rune(value)
		value = string(runes[:280]) + "..."
	}
	return value
}
