package workspace

import (
	"context"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/extensions"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func (s *Store) ListExtensions() []ExtensionSpec {
	return s.extensionManager.ListExtensions()
}

func (s *Store) GetExtensionStatus(input ExtensionIDRequest) (ExtensionStatus, *AppError) {
	status, err := s.extensionManager.GetExtensionStatus(input.ExtensionID)
	if err != nil {
		return ExtensionStatus{}, extensionError(err)
	}
	return status, nil
}

func (s *Store) DetectExtension(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.DetectExtension(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) InstallExtension(input InstallExtensionInput) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.InstallExtension(context.Background(), input)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) StartExtension(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.StartExtension(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) StopExtension(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.StopExtension(input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) RestartExtension(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.RestartExtension(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) TestExtension(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.TestExtension(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) RefreshExtensionModels(input ExtensionIDRequest) (ExtensionModelRefreshResult, *AppError) {
	result, err := s.extensionManager.RefreshModels(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) VerifyExtensionStreaming(input ExtensionIDRequest) (ExtensionStreamingResult, *AppError) {
	result, err := s.extensionManager.VerifyStreaming(context.Background(), input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) TailExtensionLogs(input TailLogsInput) ([]ExtensionLogLine, *AppError) {
	result, err := s.extensionManager.TailLogs(input)
	if err != nil {
		return nil, extensionError(err)
	}
	return result, nil
}

func (s *Store) ClearExtensionLogs(input ExtensionIDRequest) (ExtensionActionResult, *AppError) {
	result, err := s.extensionManager.ClearLogs(input.ExtensionID)
	if err != nil {
		return result, extensionError(err)
	}
	return result, nil
}

func (s *Store) GetSettings() AppSettings {
	return s.extensionManager.GetSettings()
}

func (s *Store) UpdateSettings(input UpdateSettingsInput) (AppSettings, *AppError) {
	settings, err := s.extensionManager.UpdateSettings(input)
	s.syncExtensionProviders()
	if err != nil {
		return settings, extensionError(err)
	}
	return settings, nil
}

func (s *Store) ResetExtensionSettings(input ExtensionIDRequest) (AppSettings, *AppError) {
	settings, err := s.extensionManager.ResetExtensionSettings(input.ExtensionID)
	s.syncExtensionProviders()
	if err != nil {
		return settings, extensionError(err)
	}
	return settings, nil
}

func (s *Store) ListProviders() []SafeModelProvider {
	s.syncExtensionProviders()
	providers := s.Store.ListProviders()
	return moveProviderToEnd(providers, extensions.NineRouterProviderID)
}

func (s *Store) SaveProvider(input SaveModelProviderInput) (SafeModelProvider, *AppError) {
	if input.ProviderID == extensions.NineRouterProviderID {
		settings := s.extensionManager.GetSettings()
		update := UpdateSettingsInput{
			NineRouterBaseURL:      stringPtr(input.BaseURL),
			NineRouterDefaultModel: stringPtr(input.DefaultModel),
		}
		if strings.TrimSpace(input.APIKey) != "" {
			if err := s.extensionManager.SetSecret(extensions.NineRouterExtensionID, input.APIKey); err != nil {
				return SafeModelProvider{}, extensionError(err)
			}
		}
		enabled := input.Enabled
		update.EnableNineRouterIntegration = &enabled
		if !settings.AllowAgentsUseNineRouter && enabled {
			allow := true
			update.AllowAgentsUseNineRouter = &allow
		}
		if _, err := s.extensionManager.UpdateSettings(update); err != nil {
			return SafeModelProvider{}, extensionError(err)
		}
		s.syncExtensionProviders()
		if appErr := s.Store.PersistProviders(); appErr != nil {
			return SafeModelProvider{}, appErr
		}
		if appErr := s.Store.PersistWorkspaceSnapshot(); appErr != nil {
			return SafeModelProvider{}, appErr
		}
		provider, ok := s.providerByID(extensions.NineRouterProviderID)
		if !ok {
			return SafeModelProvider{}, NotFound("PROVIDER_NOT_FOUND", "未找到 9Router 服务商。", "刷新资源中心后重试。")
		}
		return provider.Safe(), nil
	}
	return s.Store.SaveProvider(input)
}

func (s *Store) DeleteProvider(providerID string) (DeleteResult, *AppError) {
	if providerID == extensions.NineRouterProviderID {
		return DeleteResult{}, BadRequest("SYSTEM_PROVIDER_NOT_DELETABLE", "9Router 是系统预置拓展服务商，不能删除。", "如需停用，请关闭 9Router 集成或禁用服务商。")
	}
	return s.Store.DeleteProvider(providerID)
}

func (s *Store) TestProvider(providerID string) (TestResult, *AppError) {
	if providerID == extensions.NineRouterProviderID {
		result, appErr := s.TestExtension(ExtensionIDRequest{ExtensionID: extensions.NineRouterExtensionID})
		if appErr == nil {
			if persistErr := s.Store.PersistWorkspaceSnapshot(); persistErr != nil {
				appErr = persistErr
			}
		}
		return TestResult{
			OK:        result.OK,
			TargetID:  providerID,
			Message:   result.Message,
			LatencyMS: 0,
			TraceID:   s.TraceID(),
		}, appErr
	}
	return s.Store.TestProvider(providerID)
}

func (s *Store) RefreshProviderModels(providerID string) (SafeModelProvider, *AppError) {
	if providerID == extensions.NineRouterProviderID {
		_, appErr := s.RefreshExtensionModels(ExtensionIDRequest{ExtensionID: extensions.NineRouterExtensionID})
		if appErr != nil {
			return SafeModelProvider{}, appErr
		}
		if appErr := s.Store.PersistWorkspaceSnapshot(); appErr != nil {
			return SafeModelProvider{}, appErr
		}
		provider, ok := s.providerByID(providerID)
		if !ok {
			return SafeModelProvider{}, NotFound("PROVIDER_NOT_FOUND", "未找到 9Router 服务商。", "刷新资源中心后重试。")
		}
		return provider.Safe(), nil
	}
	return s.Store.RefreshProviderModels(providerID)
}

func (s *Store) syncExtensionProviders() {
	settings := s.extensionManager.GetSettings()
	status, _ := s.extensionManager.GetExtensionStatus(extensions.NineRouterExtensionID)
	apiKey := s.extensionManager.Secret(extensions.NineRouterExtensionID)
	s.Mu.Lock()
	existing, exists := s.Providers[extensions.NineRouterProviderID]
	s.Mu.Unlock()
	if apiKey == "" && exists && strings.TrimSpace(existing.APIKey) != "" {
		candidate := existing.APIKey
		if err := s.extensionManager.SetSecret(extensions.NineRouterExtensionID, candidate); err == nil {
			apiKey = candidate
		}
	}
	now := s.Now()
	models := append([]string{}, status.Models...)
	if len(models) == 0 {
		models = []string{fallback(settings.NineRouterDefaultModel, "kr/claude-sonnet-4.5")}
	}
	enabled := settings.EnableNineRouterIntegration && settings.AllowAgentsUseNineRouter
	healthStatus := status.HealthStatus
	if healthStatus == "" {
		healthStatus = "unknown"
	}
	record := resources.ModelProviderRecord{
		SafeModelProvider: resources.SafeModelProvider{
			ProviderID:        extensions.NineRouterProviderID,
			ProviderType:      resources.ProviderOpenAICompatible,
			DisplayName:       "9Router 免费模型路由",
			BaseURL:           fallback(settings.NineRouterBaseURL, "http://localhost:20128/v1"),
			DefaultModel:      fallback(settings.NineRouterDefaultModel, "kr/claude-sonnet-4.5"),
			AvailableModels:   models,
			Enabled:           enabled,
			Status:            healthStatus,
			HealthStatus:      healthStatus,
			Capabilities:      []string{"chat", "tools", "json_schema"},
			SupportsStream:    true,
			ModelCount:        len(models),
			LatencyMS:         0,
			StreamingVerified: status.StreamingVerified,
			HasAPIKey:         apiKey != "",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		APIKey: apiKey,
	}
	if status.LastErrorCode != "" {
		record.LastErrorCode = &status.LastErrorCode
	}
	if status.LastErrorMessage != "" {
		record.LastError = &status.LastErrorMessage
	}
	if record.HasAPIKey {
		masked := maskSecret(apiKey)
		record.MaskedKey = &masked
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if exists {
		record.CreatedAt = existing.CreatedAt
		if len(status.Models) == 0 && len(existing.AvailableModels) > 0 {
			record.AvailableModels = existing.AvailableModels
			record.ModelCount = len(existing.AvailableModels)
		}
	}
	s.Providers[extensions.NineRouterProviderID] = record
	profileID := resources.ProfileIDForProviderModel(record.ProviderID, record.DefaultModel)
	if _, ok := s.Profiles[profileID]; !ok {
		s.Profiles[profileID] = resources.ProfileFromProviderModel(record, record.DefaultModel, now)
	}
}

func (s *Store) providerByID(providerID string) (resources.ModelProviderRecord, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	provider, ok := s.Providers[providerID]
	return provider, ok
}

func extensionError(err *extensions.Error) *AppError {
	if err == nil {
		return nil
	}
	return BadRequest(err.Code, err.Message, err.UserAction)
}

func moveProviderToEnd(providers []SafeModelProvider, providerID string) []SafeModelProvider {
	result := make([]SafeModelProvider, 0, len(providers))
	var target *SafeModelProvider
	for _, provider := range providers {
		if provider.ProviderID == providerID {
			copy := provider
			target = &copy
			continue
		}
		result = append(result, provider)
	}
	if target != nil {
		result = append(result, *target)
	}
	return result
}

func stringPtr(value string) *string {
	return &value
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

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return strings.TrimSpace(value)
}
