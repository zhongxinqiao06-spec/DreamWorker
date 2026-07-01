package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu               sync.Mutex
	now              func() string
	traceID          func() string
	sequence         int
	streams          map[string]contextCancel
	modelGateway     ModelGateway
	agentDir         string
	providers        map[string]ModelProviderRecord
	profiles         map[string]ModelProfile
	agents           map[string]AgentConfig
	skills           map[string]SkillConfig
	tools            map[string]ToolConfig
	servers          map[string]MCPServerRecord
	mcpTools         map[string]MCPToolBinding
	projects         map[string]Project
	modules          map[string]map[string]ProjectModule
	sessions         map[string]ChatSession
	messages         map[string][]ChatMessage
	contextSummaries map[string][]ChatContextSummary
}

type contextCancel = func()

type StoreOption func(*Store)

func WithClock(now func() string) StoreOption {
	return func(store *Store) {
		store.now = now
	}
}

func WithTraceID(traceID func() string) StoreOption {
	return func(store *Store) {
		store.traceID = traceID
	}
}

func WithAgentDir(agentDir string) StoreOption {
	return func(store *Store) {
		store.agentDir = agentDir
	}
}

func NewStore(options ...StoreOption) *Store {
	store := &Store{
		now: func() string {
			return time.Now().UTC().Format(time.RFC3339)
		},
		traceID: func() string {
			return "tr_workspace_stub"
		},
		streams:          make(map[string]contextCancel),
		modelGateway:     NewLocalModelGateway(),
		agentDir:         defaultAgentDir(),
		providers:        make(map[string]ModelProviderRecord),
		profiles:         make(map[string]ModelProfile),
		agents:           make(map[string]AgentConfig),
		skills:           make(map[string]SkillConfig),
		tools:            make(map[string]ToolConfig),
		servers:          make(map[string]MCPServerRecord),
		mcpTools:         make(map[string]MCPToolBinding),
		projects:         make(map[string]Project),
		modules:          make(map[string]map[string]ProjectModule),
		sessions:         make(map[string]ChatSession),
		messages:         make(map[string][]ChatMessage),
		contextSummaries: make(map[string][]ChatContextSummary),
	}
	for _, option := range options {
		option(store)
	}
	store.seed()
	store.loadAgentSkills()
	return store
}

func defaultAgentDir() string {
	if configured := strings.TrimSpace(os.Getenv("DREAMWORKER_AGENT_DIR")); configured != "" {
		return filepath.Clean(configured)
	}
	if found := findUpward(".agent"); found != "" {
		return found
	}
	if executable, err := os.Executable(); err == nil {
		if found := findUpwardFrom(filepath.Dir(executable), ".agent"); found != "" {
			return found
		}
	}
	return filepath.Clean(".agent")
}

func findUpward(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return findUpwardFrom(wd, name)
}

func findUpwardFrom(start string, name string) string {
	dir := filepath.Clean(start)
	for {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func (s *Store) nextIDLocked(prefix string) string {
	s.sequence++
	return fmt.Sprintf("%s_%03d", prefix, s.sequence)
}

func sortedValues[T any](items map[string]T, key func(T) string) []T {
	values := make([]T, 0, len(items))
	for _, value := range items {
		values = append(values, value)
	}
	return sortSlice(values, key)
}

func sortSlice[T any](values []T, key func(T) string) []T {
	sort.Slice(values, func(i, j int) bool {
		return key(values[i]) < key(values[j])
	})
	return values
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func secretSummaries(secrets map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(secrets))
	masked := make([]string, 0, len(secrets))
	for key, value := range secrets {
		keys = append(keys, key)
		masked = append(masked, key+"="+maskSecret(value))
	}
	sort.Strings(keys)
	sort.Strings(masked)
	return keys, masked
}

func cloneStringMap(value map[string]string) map[string]string {
	result := make(map[string]string, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func cloneAnyMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func ptr(value string) *string {
	return &value
}
