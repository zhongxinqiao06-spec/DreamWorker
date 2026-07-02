package agentruntime

import (
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

func (r Runtime) ResolveChatModelBinding(
	session resources.ChatSession,
	agent resources.AgentConfig,
) (resources.ModelProfile, resources.ModelProviderRecord, string, *resources.AppError) {
	return ResolveChatModelBinding(r.State, session, agent)
}

func ResolveChatModelBinding(
	state *resources.Store,
	session resources.ChatSession,
	agent resources.AgentConfig,
) (resources.ModelProfile, resources.ModelProviderRecord, string, *resources.AppError) {
	if shouldBindProviderModel(session.ModelProfileID, session.ProviderID, session.Model) {
		if profile, provider, ok := bindProviderModel(state, session.ProviderID, session.Model); ok {
			return profile, provider, "", nil
		}
	}
	if shouldBindProviderModel(agent.ModelProfileID, agent.ProviderID, agent.Model) {
		if profile, provider, ok := bindProviderModel(state, agent.ProviderID, agent.Model); ok {
			return profile, provider, "", nil
		}
	}

	seen := map[string]bool{}
	var skipped []string
	var candidates []string
	appendCandidate := func(profileID string) {
		profileID = strings.TrimSpace(profileID)
		if profileID == "" || seen[profileID] {
			return
		}
		seen[profileID] = true
		candidates = append(candidates, profileID)
	}
	appendCandidate(session.ModelProfileID)
	appendCandidate(agent.ModelProfileID)
	if session.ProjectID != nil {
		if project, ok := state.Projects[*session.ProjectID]; ok {
			appendCandidate(project.DefaultModelProfileID)
		}
	}
	appendCandidate("profile_fast")
	appendCandidate("profile_openai")
	appendCandidate("profile_anthropic")
	appendCandidate("profile_deepseek")
	appendCandidate("profile_ollama")
	appendCandidate("profile_stub")

	for index := 0; index < len(candidates); index++ {
		profileID := candidates[index]
		profile, ok := state.Profiles[profileID]
		if !ok || !profile.Enabled {
			skipped = append(skipped, profileID+":profile_unavailable")
			continue
		}
		profile = resources.EnsureModelProfileDefaults(profile)
		if profile.FallbackProfileID != nil {
			appendCandidate(*profile.FallbackProfileID)
		}
		provider, ok := state.Providers[profile.ProviderID]
		if !ok || !provider.Enabled {
			skipped = append(skipped, profileID+":provider_unavailable")
			continue
		}
		if !ProviderHasUsableCredential(provider) {
			skipped = append(skipped, profileID+":credential_missing")
			continue
		}
		fallbackPath := ""
		if len(skipped) > 0 || profile.ProfileID != session.ModelProfileID {
			fallbackPath = strings.Join(append(skipped, profile.ProfileID+":selected"), " -> ")
		}
		return profile, provider, fallbackPath, nil
	}
	return resources.ModelProfile{}, resources.ModelProviderRecord{}, "", resources.BadRequest(
		"MODEL_PROVIDER_UNAVAILABLE",
		"没有可用的模型配置",
		"请配置供应商 API Key，或切换到本地 Stub 模型",
	)
}

func shouldBindProviderModel(profileID string, providerID string, model string) bool {
	providerID = strings.TrimSpace(providerID)
	model = strings.TrimSpace(model)
	if providerID == "" {
		return false
	}
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return true
	}
	return profileID == resources.ProfileIDForProviderModel(providerID, model)
}

func bindProviderModel(
	state *resources.Store,
	providerID string,
	model string,
) (resources.ModelProfile, resources.ModelProviderRecord, bool) {
	providerID = strings.TrimSpace(providerID)
	model = strings.TrimSpace(model)
	if providerID == "" {
		return resources.ModelProfile{}, resources.ModelProviderRecord{}, false
	}
	provider, ok := state.Providers[providerID]
	if !ok || !provider.Enabled || !ProviderHasUsableCredential(provider) {
		return resources.ModelProfile{}, resources.ModelProviderRecord{}, false
	}
	if model == "" {
		model = provider.DefaultModel
	}
	profileID := resources.ProfileIDForProviderModel(providerID, model)
	profile, ok := state.Profiles[profileID]
	if !ok {
		profile = resources.ProfileFromProviderModel(provider, model, state.Now())
		state.Profiles[profile.ProfileID] = profile
	}
	if !profile.Enabled {
		return resources.ModelProfile{}, resources.ModelProviderRecord{}, false
	}
	return resources.EnsureModelProfileDefaults(profile), provider, true
}

func ProviderHasUsableCredential(provider resources.ModelProviderRecord) bool {
	if resources.ProviderAllowsMissingAPIKey(provider) {
		return true
	}
	if provider.APIKey == "" || provider.APIKey == "sk-local-demo" {
		return false
	}
	return true
}

func ToChatModelProvider(provider resources.ModelProviderRecord) ports.ChatModelProvider {
	return resources.ToChatModelProvider(provider)
}

func ToChatModelProfile(profile resources.ModelProfile) ports.ChatModelProfile {
	return resources.ToChatModelProfile(profile)
}
