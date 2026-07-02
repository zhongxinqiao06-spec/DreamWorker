package workspace

import (
	"sync"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/chat"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/projects"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type StoreOption = resources.StoreOption
type ModelGateway = resources.ModelGateway

var WithClock = resources.WithClock
var WithTraceID = resources.WithTraceID
var WithAgentDir = resources.WithAgentDir
var WithModelGateway = resources.WithModelGateway
var NewLocalModelGateway = resources.NewLocalModelGateway

type Store struct {
	*resources.Store

	projectStore *projects.Store
	chatStore    *chat.Store

	mu       *sync.Mutex
	sessions map[string]ChatSession
	messages map[string][]ChatMessage
}

func NewStore(options ...StoreOption) *Store {
	state := resources.NewStore(options...)
	store := &Store{
		Store:        state,
		projectStore: projects.NewStore(state),
		chatStore:    chat.NewStore(state),
		mu:           &state.Mu,
		sessions:     state.Sessions,
		messages:     state.Messages,
	}
	store.projectStore.SeedDefaults(state.Now())
	store.seedDefaultChat()
	return store
}

func (s *Store) seedDefaultChat() {
	s.Mu.Lock()
	if len(s.Sessions) > 0 {
		s.Mu.Unlock()
		return
	}
	projectID := ""
	if _, ok := s.Projects["project_001"]; ok {
		projectID = "project_001"
	}
	s.Mu.Unlock()

	input := CreateChatSessionInput{
		Title:          "通用 Agent 工作台",
		AgentID:        "agent_general_assistant",
		ModelProfileID: "profile_fast",
	}
	if projectID != "" {
		input.ProjectID = &projectID
	}
	_, _ = s.chatStore.CreateChatSession(input)
}
