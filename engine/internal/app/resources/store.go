package resources

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
	Mu               sync.Mutex
	Now              func() string
	TraceID          func() string
	PersistWorkspace func(*Store) *AppError
	CloseWorkspace   func() error
	InitialSnapshot  *WorkspaceSnapshot
	SnapshotLoaded   bool
	Sequence         int
	Streams          map[string]contextCancel
	ModelGateway     ModelGateway
	AgentDir         string
	ConfigDir        string
	Providers        map[string]ModelProviderRecord
	Profiles         map[string]ModelProfile
	Agents           map[string]AgentConfig
	Skills           map[string]SkillConfig
	Tools            map[string]ToolConfig
	Servers          map[string]MCPServerRecord
	MCPTools         map[string]MCPToolBinding
	Projects         map[string]Project
	Modules          map[string]map[string]ProjectModule
	Sessions         map[string]ChatSession
	Messages         map[string][]ChatMessage
	ContextSummaries map[string][]ChatContextSummary
}

type contextCancel = func()

type StoreOption func(*Store)

func WithClock(now func() string) StoreOption {
	return func(store *Store) {
		store.Now = now
	}
}

func WithTraceID(traceID func() string) StoreOption {
	return func(store *Store) {
		store.TraceID = traceID
	}
}

func WithWorkspacePersistence(persist func(*Store) *AppError) StoreOption {
	return func(store *Store) {
		store.PersistWorkspace = persist
	}
}

func WithWorkspacePersistenceClose(close func() error) StoreOption {
	return func(store *Store) {
		store.CloseWorkspace = close
	}
}

func WithInitialWorkspaceSnapshot(snapshot WorkspaceSnapshot) StoreOption {
	return func(store *Store) {
		copied := snapshot
		store.InitialSnapshot = &copied
	}
}

func WithAgentDir(agentDir string) StoreOption {
	return func(store *Store) {
		store.AgentDir = agentDir
	}
}

func WithConfigDir(configDir string) StoreOption {
	return func(store *Store) {
		store.ConfigDir = strings.TrimSpace(configDir)
		if store.ConfigDir != "" {
			store.ConfigDir = filepath.Clean(store.ConfigDir)
		}
	}
}

func NewStore(options ...StoreOption) *Store {
	store := &Store{
		Now: func() string {
			return time.Now().UTC().Format(time.RFC3339)
		},
		TraceID: func() string {
			return "tr_workspace_stub"
		},
		Streams:          make(map[string]contextCancel),
		ModelGateway:     NewLocalModelGateway(),
		AgentDir:         defaultAgentDir(),
		ConfigDir:        "",
		Providers:        make(map[string]ModelProviderRecord),
		Profiles:         make(map[string]ModelProfile),
		Agents:           make(map[string]AgentConfig),
		Skills:           make(map[string]SkillConfig),
		Tools:            make(map[string]ToolConfig),
		Servers:          make(map[string]MCPServerRecord),
		MCPTools:         make(map[string]MCPToolBinding),
		Projects:         make(map[string]Project),
		Modules:          make(map[string]map[string]ProjectModule),
		Sessions:         make(map[string]ChatSession),
		Messages:         make(map[string][]ChatMessage),
		ContextSummaries: make(map[string][]ChatContextSummary),
	}
	for _, option := range options {
		option(store)
	}
	store.seed()
	store.loadProviderConfig()
	store.loadAgentSkills()
	if store.InitialSnapshot != nil {
		store.ApplyWorkspaceSnapshotLocked(*store.InitialSnapshot)
		store.SnapshotLoaded = true
	}
	return store
}

func DefaultConfigDir() string {
	if configured := strings.TrimSpace(os.Getenv("DREAMWORKER_CONFIG_DIR")); configured != "" {
		return filepath.Clean(configured)
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "DreamWorker")
	}
	return filepath.Join(os.TempDir(), "DreamWorker")
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
	s.Sequence++
	return fmt.Sprintf("%s_%03d", prefix, s.Sequence)
}

func (s *Store) NextIDLocked(prefix string) string {
	return s.nextIDLocked(prefix)
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

func (s *Store) PersistWorkspaceSnapshot() *AppError {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.persistWorkspaceLocked()
}

func (s *Store) PersistWorkspaceSnapshotLocked() *AppError {
	return s.persistWorkspaceLocked()
}

func (s *Store) persistWorkspaceLocked() *AppError {
	if s.PersistWorkspace == nil {
		return nil
	}
	return s.PersistWorkspace(s)
}

func (s *Store) Close() error {
	if s.CloseWorkspace == nil {
		return nil
	}
	return s.CloseWorkspace()
}
