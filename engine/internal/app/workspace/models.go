package workspace

import (
	"context"
	"strings"
)

func (s *Store) ListProviders() []SafeModelProvider {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]SafeModelProvider, 0, len(s.providers))
	for _, provider := range s.providers {
		result = append(result, provider.safe())
	}
	return sortedValuesFromSlice(result, func(item SafeModelProvider) string { return item.DisplayName })
}

func (s *Store) SaveProvider(input SaveModelProviderInput) (SafeModelProvider, *AppError) {
	if strings.TrimSpace(input.ProviderID) == "" || strings.TrimSpace(input.DisplayName) == "" {
		return SafeModelProvider{}, BadRequest("INVALID_PROVIDER", "provider requires providerId and displayName", "complete provider basics")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	record, exists := s.providers[input.ProviderID]
	if !exists {
		record.CreatedAt = now
	}
	record.ProviderID = input.ProviderID
	record.ProviderType = input.ProviderType
	record.DisplayName = input.DisplayName
	record.BaseURL = strings.TrimSpace(input.BaseURL)
	record.Organization = input.Organization
	record.Project = input.Project
	record.DefaultModel = strings.TrimSpace(input.DefaultModel)
	record.AvailableModels = append([]string{}, input.AvailableModels...)
	record.Enabled = input.Enabled
	record.Capabilities = normalizeCapabilities(input.Capabilities)
	if len(record.Capabilities) == 0 {
		record.Capabilities = defaultProviderCapabilities(input.ProviderType)
	}
	if record.Status == "" {
		record.Status = "unknown"
	}
	if record.HealthStatus == "" {
		record.HealthStatus = record.Status
	}
	record.SupportsStream = providerSupportsStreaming(input.ProviderType)
	record.ModelCount = len(record.AvailableModels)
	record.UpdatedAt = now
	if input.APIKey != "" {
		record.APIKey = input.APIKey
	}
	record.HasAPIKey = record.APIKey != ""
	record.MaskedKey = nil
	if record.HasAPIKey {
		masked := maskSecret(record.APIKey)
		record.MaskedKey = &masked
	}
	s.providers[input.ProviderID] = record
	return record.safe(), nil
}

func (s *Store) DeleteProvider(providerID string) (DeleteResult, *AppError) {
	if providerID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "missing providerId", "select a provider")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, providerID)
	return DeleteResult{OK: true, DeletedID: providerID}, nil
}

func (s *Store) TestProvider(providerID string) (TestResult, *AppError) {
	if providerID == "" {
		return TestResult{}, BadRequest("BAD_REQUEST", "missing providerId", "select a provider")
	}
	s.mu.Lock()
	provider, ok := s.providers[providerID]
	if !ok {
		s.mu.Unlock()
		return TestResult{}, NotFound("PROVIDER_NOT_FOUND", "provider not found", "refresh resource center")
	}
	s.mu.Unlock()
	message := "provider health check completed"
	status := "connected"
	lastError := (*string)(nil)
	if !provider.Enabled {
		message = "provider is disabled"
		status = "unknown"
	}
	if provider.ProviderType != ProviderOllama && provider.ProviderID != "provider_local_stub" && provider.APIKey == "" {
		message = "provider api key is missing"
		status = "error"
		errText := message
		lastError = &errText
	}
	health := s.modelGateway.HealthCheck(context.Background(), toChatModelProvider(provider))
	if health.Message != "" {
		message = health.Message
	}
	if health.Status != "" {
		status = health.Status
	}
	if health.ErrorCode != "" {
		errText := health.Message
		lastError = &errText
	}
	now := s.now()
	provider.Status = status
	provider.HealthStatus = status
	provider.LastTestedAt = &now
	provider.LastError = lastError
	provider.LastErrorCode = nil
	if health.ErrorCode != "" {
		provider.LastErrorCode = &health.ErrorCode
	}
	provider.LatencyMS = health.LatencyMS
	provider.StreamingVerified = health.StreamingVerified
	provider.UpdatedAt = now
	s.mu.Lock()
	s.providers[providerID] = provider
	s.mu.Unlock()
	return TestResult{
		OK:        health.OK,
		TargetID:  providerID,
		Message:   message,
		LatencyMS: health.LatencyMS,
		TraceID:   s.traceID(),
	}, nil
}

func (s *Store) RefreshProviderModels(providerID string) (SafeModelProvider, *AppError) {
	if providerID == "" {
		return SafeModelProvider{}, BadRequest("BAD_REQUEST", "missing providerId", "select a provider")
	}
	s.mu.Lock()
	provider, ok := s.providers[providerID]
	if !ok {
		s.mu.Unlock()
		return SafeModelProvider{}, NotFound("PROVIDER_NOT_FOUND", "provider not found", "refresh resource center")
	}
	s.mu.Unlock()
	discovery := s.modelGateway.DiscoverModels(context.Background(), toChatModelProvider(provider))
	if discovery.Discovered && len(discovery.Models) > 0 {
		provider.AvailableModels = discovery.Models
		if provider.DefaultModel == "" || !containsString(discovery.Models, provider.DefaultModel) {
			provider.DefaultModel = discovery.Models[0]
		}
		provider.Status = "connected"
		provider.HealthStatus = "connected"
		provider.LastError = nil
		provider.LastErrorCode = nil
	} else {
		if len(provider.AvailableModels) == 0 {
			provider.AvailableModels = defaultProviderModels(provider.ProviderType)
		}
		if provider.DefaultModel == "" && len(provider.AvailableModels) > 0 {
			provider.DefaultModel = provider.AvailableModels[0]
		}
		if discovery.LastError != "" {
			errText := discovery.LastError
			provider.Status = "error"
			provider.HealthStatus = "error"
			provider.LastError = &errText
			if discovery.ErrorCode != "" {
				provider.LastErrorCode = &discovery.ErrorCode
			}
		}
	}
	provider.Capabilities = normalizeCapabilities(provider.Capabilities)
	if len(provider.Capabilities) == 0 {
		provider.Capabilities = defaultProviderCapabilities(provider.ProviderType)
	}
	now := s.now()
	provider.LastTestedAt = &now
	provider.LastDiscoveryAt = &now
	provider.LatencyMS = discovery.LatencyMS
	provider.UpdatedAt = now
	provider.SupportsStream = providerSupportsStreaming(provider.ProviderType)
	provider.ModelCount = len(provider.AvailableModels)
	s.mu.Lock()
	s.providers[providerID] = provider
	s.mu.Unlock()
	return provider.safe(), nil
}

func (s *Store) ListProfiles() []ModelProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedValues(s.profiles, func(item ModelProfile) string { return item.DisplayName })
}

func (s *Store) SaveProfile(input ModelProfile) (ModelProfile, *AppError) {
	if input.ProfileID == "" {
		return ModelProfile{}, BadRequest("BAD_REQUEST", "invalid model profile", "check profileId, provider and model")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	existing, exists := s.profiles[input.ProfileID]
	if !exists {
		input.CreatedAt = now
	} else {
		input.CreatedAt = existing.CreatedAt
	}
	input = ensureModelProfileDefaults(input)
	input.UpdatedAt = now
	s.profiles[input.ProfileID] = input
	return input, nil
}

func (s *Store) DeleteProfile(profileID string) (DeleteResult, *AppError) {
	if profileID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 profileId。", "请选择要删除的模型配置。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profiles, profileID)
	return DeleteResult{OK: true, DeletedID: profileID}, nil
}

func (p ModelProviderRecord) safe() SafeModelProvider {
	safe := p.SafeModelProvider
	if safe.Status == "" {
		safe.Status = "unknown"
	}
	safe.Capabilities = normalizeCapabilities(safe.Capabilities)
	if len(safe.Capabilities) == 0 {
		safe.Capabilities = defaultProviderCapabilities(safe.ProviderType)
	}
	safe.SupportsStream = providerSupportsStreaming(safe.ProviderType)
	if safe.HealthStatus == "" {
		safe.HealthStatus = safe.Status
	}
	safe.ModelCount = len(safe.AvailableModels)
	safe.HasAPIKey = p.APIKey != ""
	safe.MaskedKey = nil
	if safe.HasAPIKey {
		masked := maskSecret(p.APIKey)
		safe.MaskedKey = &masked
	}
	return safe
}

func ensureModelProfileDefaults(profile ModelProfile) ModelProfile {
	if profile.ContextWindow == 0 {
		profile.ContextWindow = 128000
	}
	if profile.ResponseFormat == "" {
		profile.ResponseFormat = "text"
	}
	if profile.ToolMode == "" {
		profile.ToolMode = "auto"
	}
	if profile.TimeoutMS == 0 {
		profile.TimeoutMS = 120000
	}
	return profile
}

func defaultProviderModels(providerType ProviderType) []string {
	switch providerType {
	case ProviderOpenAI:
		return []string{"gpt-5.2", "gpt-5-mini", "gpt-4.1"}
	case ProviderAnthropic:
		return []string{"claude-sonnet-4-5", "claude-opus-4-1", "claude-haiku-4-5"}
	case ProviderDeepSeek:
		return []string{"deepseek-v4-flash", "deepseek-v4-pro", "deepseek-chat", "deepseek-reasoner"}
	case ProviderGLM:
		return []string{"glm-5.2", "glm-5.1", "glm-5", "glm-5-turbo", "glm-4.7", "glm-4.7-flashx", "glm-4.6"}
	case ProviderVolcano:
		return []string{"doubao-seed-1.6", "doubao-seed-1.6-thinking", "doubao-1.5-pro"}
	case ProviderSiliconFlow:
		return []string{"deepseek-ai/DeepSeek-V4-Flash", "deepseek-ai/DeepSeek-V4-Pro", "zai-org/GLM-5.2", "Qwen/Qwen3.5-4B"}
	case ProviderOllama:
		return []string{"llama3.1", "qwen3", "deepseek-r1"}
	default:
		return []string{"model-name"}
	}
}

func defaultProviderCapabilities(providerType ProviderType) []string {
	switch providerType {
	case ProviderOllama:
		return []string{"chat", "tools"}
	case ProviderAnthropic:
		return []string{"chat", "tools", "vision", "json_schema"}
	default:
		return []string{"chat", "tools", "vision", "json_schema"}
	}
}

func providerSupportsStreaming(providerType ProviderType) bool {
	switch providerType {
	case ProviderOpenAI, ProviderOpenAICompatible, ProviderDeepSeek, ProviderGLM, ProviderVolcano, ProviderSiliconFlow, ProviderAnthropic, ProviderOllama:
		return true
	default:
		return false
	}
}

func normalizeCapabilities(values []string) []string {
	allowed := map[string]bool{"chat": true, "tools": true, "vision": true, "json_schema": true}
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		item := strings.TrimSpace(value)
		if !allowed[item] || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sortedValuesFromSlice[T any](values []T, key func(T) string) []T {
	result := append([]T{}, values...)
	return sortSlice(result, key)
}
