package agentruntime

import "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"

type TurnBinding struct {
	Session          resources.ChatSession
	Agent            resources.AgentConfig
	Profile          resources.ModelProfile
	Provider         resources.ModelProviderRecord
	ProviderFallback string
	ContextPack      resources.ChatContextPack
}
