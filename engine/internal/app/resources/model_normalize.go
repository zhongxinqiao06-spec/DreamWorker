package resources

import "strings"

const nineRouterProviderID = "provider_9router_local"

func NormalizeProviderModelID(providerID string, model string) string {
	model = strings.TrimSpace(model)
	if strings.TrimSpace(providerID) != nineRouterProviderID {
		return model
	}
	lower := strings.ToLower(model)
	if strings.HasPrefix(lower, "kiro/") {
		return "kr/" + strings.TrimSpace(model[len("kiro/"):])
	}
	return model
}

func normalizeProviderModelList(providerID string, values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		model := NormalizeProviderModelID(providerID, value)
		if model != "" {
			normalized = append(normalized, model)
		}
	}
	return normalizeStringList(normalized)
}
