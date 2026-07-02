package resources

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const providerConfigFileName = "model-providers.json"

type persistedProviderConfig struct {
	Providers []persistedModelProvider `json:"providers"`
}

type persistedModelProvider struct {
	SafeModelProvider
	APIKey string `json:"apiKey,omitempty"`
}

func (s *Store) providerConfigPath() string {
	if strings.TrimSpace(s.ConfigDir) == "" {
		return ""
	}
	return filepath.Join(s.ConfigDir, providerConfigFileName)
}

func (s *Store) loadProviderConfig() {
	path := s.providerConfigPath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var config persistedProviderConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}
	now := s.Now()
	for _, item := range config.Providers {
		record := ModelProviderRecord{
			SafeModelProvider: item.SafeModelProvider,
			APIKey:            strings.TrimSpace(item.APIKey),
		}
		record.ProviderID = strings.TrimSpace(record.ProviderID)
		if record.ProviderID == "" || record.DisplayName == "" {
			continue
		}
		if existing, ok := s.Providers[record.ProviderID]; ok {
			if record.APIKey == "" {
				record.APIKey = existing.APIKey
			}
			if record.CreatedAt == "" {
				record.CreatedAt = existing.CreatedAt
			}
		}
		s.Providers[record.ProviderID] = normalizeProviderRecord(record, now)
	}
}

func (s *Store) PersistProviders() *AppError {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.persistProvidersLocked()
}

func (s *Store) persistProvidersLocked() *AppError {
	path := s.providerConfigPath()
	if path == "" {
		return nil
	}
	providers := sortedValues(s.Providers, func(item ModelProviderRecord) string {
		return item.DisplayName + item.ProviderID
	})
	config := persistedProviderConfig{
		Providers: make([]persistedModelProvider, 0, len(providers)),
	}
	now := s.Now()
	for _, provider := range providers {
		provider = normalizeProviderRecord(provider, now)
		config.Providers = append(config.Providers, persistedModelProvider{
			SafeModelProvider: provider.safe(),
			APIKey:            provider.APIKey,
		})
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return Internal("PROVIDER_CONFIG_ENCODE_FAILED", "无法序列化模型服务商配置。", "检查服务商配置字段后重试。")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return Internal("PROVIDER_CONFIG_DIR_FAILED", "无法创建模型服务商配置目录。", "检查 DreamWorker 配置目录权限后重试。")
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return Internal("PROVIDER_CONFIG_WRITE_FAILED", "无法写入模型服务商配置文件。", "检查 DreamWorker 配置目录权限后重试。")
	}
	return nil
}

func normalizeProviderRecord(record ModelProviderRecord, now string) ModelProviderRecord {
	record.ProviderID = strings.TrimSpace(record.ProviderID)
	record.DisplayName = strings.TrimSpace(record.DisplayName)
	record.BaseURL = strings.TrimSpace(record.BaseURL)
	record.DefaultModel = NormalizeProviderModelID(record.ProviderID, record.DefaultModel)
	record.AvailableModels = normalizeProviderModelList(record.ProviderID, record.AvailableModels)
	record.Capabilities = normalizeCapabilities(record.Capabilities)
	if len(record.Capabilities) == 0 {
		record.Capabilities = defaultProviderCapabilities(record.ProviderType)
	}
	if record.Status == "" {
		record.Status = "unknown"
	}
	if record.HealthStatus == "" {
		record.HealthStatus = record.Status
	}
	record.SupportsStream = providerSupportsStreaming(record.ProviderType)
	record.ModelCount = len(record.AvailableModels)
	if record.CreatedAt == "" {
		record.CreatedAt = now
	}
	if record.UpdatedAt == "" {
		record.UpdatedAt = now
	}
	record.APIKey = strings.TrimSpace(record.APIKey)
	record.HasAPIKey = record.APIKey != ""
	record.MaskedKey = nil
	if record.HasAPIKey {
		masked := maskSecret(record.APIKey)
		record.MaskedKey = &masked
	}
	return record
}

func normalizeStringList(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}
