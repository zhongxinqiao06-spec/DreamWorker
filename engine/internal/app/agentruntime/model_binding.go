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

func ProviderHasUsableCredential(provider resources.ModelProviderRecord) bool {
	if provider.ProviderID == "provider_local_stub" || provider.DefaultModel == "model_generate_stub" {
		return true
	}
	if provider.ProviderType == resources.ProviderOllama {
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
