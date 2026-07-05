package modelgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

func TestOpenAICompatibleStreamChatParsesSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-real-test" {
			t.Fatalf("missing bearer token")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2,\"total_tokens\":5}}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), testMessages()))

	if content := concat(chunks); content != "Hello world" {
		t.Fatalf("expected content, got %q", content)
	}
	if usage := lastUsage(chunks); usage == nil || usage.TotalTokens != 5 {
		t.Fatalf("expected usage, got %#v", usage)
	}
}

func TestOpenAICompatibleStreamChatSendsImageContentParts(t *testing.T) {
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests <- payload
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	messages := []ports.ChatModelMessage{
		{Role: "system", Content: "System"},
		{
			Role:    "user",
			Content: "Describe this image",
			Parts: []ports.ChatModelContentPart{
				{Type: "text", Text: "Describe this image"},
				{Type: "image_url", ImageURL: &ports.ChatModelImageURL{URL: "data:image/png;base64,iVBORw0KGgo=", Detail: "low"}},
			},
		},
	}
	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), messages))
	if content := concat(chunks); content != "ok" {
		t.Fatalf("expected content, got %q", content)
	}

	var payload map[string]any
	select {
	case payload = <-requests:
	default:
		t.Fatal("expected request payload")
	}
	requestMessages, ok := payload["messages"].([]any)
	if !ok || len(requestMessages) != 2 {
		t.Fatalf("expected messages array, got %#v", payload["messages"])
	}
	userMessage, ok := requestMessages[1].(map[string]any)
	if !ok {
		t.Fatalf("expected user message object, got %#v", requestMessages[1])
	}
	parts, ok := userMessage["content"].([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("expected multimodal content parts, got %#v", userMessage["content"])
	}
	imagePart, ok := parts[1].(map[string]any)
	if !ok || imagePart["type"] != "image_url" {
		t.Fatalf("expected image_url part, got %#v", parts[1])
	}
	imageURL, ok := imagePart["image_url"].(map[string]any)
	if !ok || imageURL["url"] != "data:image/png;base64,iVBORw0KGgo=" || imageURL["detail"] != "low" {
		t.Fatalf("expected image URL payload, got %#v", imagePart["image_url"])
	}
}

func TestOpenAIResponsesStreamChatParsesSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"A\"}\n\n")
		_, _ = fmt.Fprint(w, "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"I\"}\n\n")
		_, _ = fmt.Fprint(w, "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":4,\"output_tokens\":2,\"total_tokens\":6}}}\n\n")
	}))
	defer server.Close()

	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL, ProviderOpenAI), testProfile(), testMessages()))

	if content := concat(chunks); content != "AI" {
		t.Fatalf("expected content, got %q", content)
	}
	if usage := lastUsage(chunks); usage == nil || usage.TotalTokens != 6 {
		t.Fatalf("expected usage, got %#v", usage)
	}
}

func TestOpenAICompatibleStreamChatParsesToolCallDelta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"id\":\"call_1\",\"index\":0,\"type\":\"function\",\"function\":{\"name\":\"tool_model_generate_stub\",\"arguments\":\"{\\\"prompt\\\":\\\"hi\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), testMessages()))

	if len(chunks) == 0 || chunks[0].ToolCall == nil {
		t.Fatalf("expected tool call chunk, got %#v", chunks)
	}
	if chunks[0].ToolCall.ToolID != "tool_model_generate_stub" || chunks[0].ToolCall.CallID != "call_1" {
		t.Fatalf("unexpected tool call: %#v", chunks[0].ToolCall)
	}
}

func TestOpenAICompatibleStreamChatParsesReasoningDelta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"先判断上下文\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"结论\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), testMessages()))

	if reasoning := concatReasoning(chunks); reasoning != "先判断上下文" {
		t.Fatalf("expected reasoning delta, got %q", reasoning)
	}
	if content := concat(chunks); content != "结论" {
		t.Fatalf("expected content, got %q", content)
	}
}

func TestAnthropicStreamChatParsesSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "sk-real-test" {
			t.Fatalf("missing anthropic key")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":5}}}\n\n")
		_, _ = fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Claude\"}}\n\n")
		_, _ = fmt.Fprint(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"input_tokens\":5,\"output_tokens\":1}}\n\n")
		_, _ = fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer server.Close()

	chunks := collectChunks(NewGateway().StreamChat(context.Background(), testProvider(server.URL, ProviderAnthropic), testProfile(), testMessages()))

	if content := concat(chunks); content != "Claude" {
		t.Fatalf("expected content, got %q", content)
	}
	if usage := lastUsage(chunks); usage == nil || usage.TotalTokens != 6 {
		t.Fatalf("expected usage, got %#v", usage)
	}
}

func TestOllamaStreamChatParsesJSONL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = fmt.Fprintln(w, "{\"message\":{\"content\":\"Local\"}}")
		_, _ = fmt.Fprintln(w, "{\"message\":{\"content\":\" model\"},\"done\":true,\"done_reason\":\"stop\",\"prompt_eval_count\":3,\"eval_count\":2}")
	}))
	defer server.Close()

	provider := testProvider(server.URL, ProviderOllama)
	provider.APIKey = ""
	chunks := collectChunks(NewGateway().StreamChat(context.Background(), provider, testProfile(), testMessages()))

	if content := concat(chunks); content != "Local model" {
		t.Fatalf("expected content, got %q", content)
	}
	if usage := lastUsage(chunks); usage == nil || usage.TotalTokens != 5 {
		t.Fatalf("expected usage, got %#v", usage)
	}
}

func TestProviderErrorIsRedactedAndTruncated(t *testing.T) {
	message := sanitizeProviderError("bad key sk-test-secret-token-with-long-body " + strings.Repeat("x", 500))
	if strings.Contains(message, "sk-test-secret") {
		t.Fatalf("error leaked key: %s", message)
	}
	if len([]rune(message)) > 283 {
		t.Fatalf("expected truncated error, got %d runes", len([]rune(message)))
	}
}

func TestOpenAICompatibleGenerateImageParsesBase64Image(t *testing.T) {
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("expected JSON accept header, got %q", r.Header.Get("Accept"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests <- payload
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":[{"b64_json":"aW1hZ2U=","revised_prompt":"radar"}]}`)
	}))
	defer server.Close()

	result, err := NewGateway().GenerateImage(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), ports.ImageGenerationInput{
		Prompt:         "radar",
		Size:           "512x512",
		ResponseFormat: "b64_json",
	})
	if err != nil {
		t.Fatalf("generate image: %v", err)
	}
	if len(result.Images) != 1 || result.Images[0].DataURL != "data:image/png;base64,aW1hZ2U=" {
		t.Fatalf("expected generated data URL, got %#v", result.Images)
	}
	var payload map[string]any
	select {
	case payload = <-requests:
	default:
		t.Fatal("expected request payload")
	}
	if payload["prompt"] != "radar" || payload["size"] != "512x512" || payload["response_format"] != "b64_json" {
		t.Fatalf("unexpected image generation request: %#v", payload)
	}
}

func TestOpenAICompatibleGenerateImageUsesNineRouterImageModel(t *testing.T) {
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests <- payload
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":[{"b64_json":"aW1hZ2U="}]}`)
	}))
	defer server.Close()

	provider := testProvider(server.URL+"/v1", ProviderOpenAICompatible)
	provider.ProviderID = "provider_9router_local"
	provider.AvailableModels = []string{"cx/gpt-5.5", "cx/gpt-5.5-image"}
	profile := testProfile()
	profile.Model = "cx/gpt-5.5"
	_, err := NewGateway().GenerateImage(context.Background(), provider, profile, ports.ImageGenerationInput{Prompt: "radar"})
	if err != nil {
		t.Fatalf("generate image: %v", err)
	}
	payload := <-requests
	if payload["model"] != "cx/gpt-5.5-image" {
		t.Fatalf("expected 9Router image model, got %#v", payload["model"])
	}
}

func TestOpenAICompatibleGenerateImageReportsInvalidProviderResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprint(w, "entitlement required")
	}))
	defer server.Close()

	_, err := NewGateway().GenerateImage(context.Background(), testProvider(server.URL+"/v1", ProviderOpenAICompatible), testProfile(), ports.ImageGenerationInput{
		Prompt: "radar",
	})
	if err == nil || !strings.Contains(err.Error(), "entitlement required") {
		t.Fatalf("expected provider response in error, got %v", err)
	}
}

func testProvider(baseURL string, providerType string) ports.ChatModelProvider {
	return ports.ChatModelProvider{
		ProviderID:      "provider_test",
		ProviderType:    providerType,
		DisplayName:     "Test provider",
		BaseURL:         baseURL,
		DefaultModel:    "test-model",
		AvailableModels: []string{"test-model"},
		Enabled:         true,
		APIKey:          "sk-real-test",
	}
}

func testProfile() ports.ChatModelProfile {
	return ports.ChatModelProfile{
		ProfileID:   "profile_test",
		DisplayName: "Test profile",
		ProviderID:  "provider_test",
		Model:       "test-model",
		Temperature: 0,
		MaxTokens:   128,
		TimeoutMS:   30000,
	}
}

func testMessages() []ports.ChatModelMessage {
	return []ports.ChatModelMessage{
		{Role: "system", Content: "System"},
		{Role: "user", Content: "Hello"},
	}
}

func collectChunks(stream <-chan ports.ChatModelStreamChunk) []ports.ChatModelStreamChunk {
	var chunks []ports.ChatModelStreamChunk
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func concat(chunks []ports.ChatModelStreamChunk) string {
	var builder strings.Builder
	for _, chunk := range chunks {
		builder.WriteString(chunk.Delta)
	}
	return builder.String()
}

func concatReasoning(chunks []ports.ChatModelStreamChunk) string {
	var builder strings.Builder
	for _, chunk := range chunks {
		builder.WriteString(chunk.ReasoningDelta)
	}
	return builder.String()
}

func lastUsage(chunks []ports.ChatModelStreamChunk) *ports.ChatModelUsage {
	var usage *ports.ChatModelUsage
	for _, chunk := range chunks {
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
	}
	return usage
}
