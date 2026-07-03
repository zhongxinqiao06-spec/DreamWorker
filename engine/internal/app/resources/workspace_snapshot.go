package resources

type WorkspaceSnapshot struct {
	SchemaVersion    string                              `json:"schemaVersion"`
	Sequence         int                                 `json:"sequence"`
	Providers        map[string]ModelProviderRecord      `json:"providers"`
	ProviderSecrets  map[string]string                   `json:"providerSecrets"`
	Profiles         map[string]ModelProfile             `json:"profiles"`
	Agents           map[string]AgentConfig              `json:"agents"`
	Skills           map[string]SkillConfig              `json:"skills"`
	Tools            map[string]ToolConfig               `json:"tools"`
	Servers          map[string]MCPServerRecord          `json:"mcpServers"`
	MCPServerSecrets map[string]map[string]string        `json:"mcpServerSecrets"`
	MCPTools         map[string]MCPToolBinding           `json:"mcpTools"`
	Projects         map[string]Project                  `json:"projects"`
	Modules          map[string]map[string]ProjectModule `json:"modules"`
	Sessions         map[string]ChatSession              `json:"sessions"`
	Messages         map[string][]ChatMessage            `json:"messages"`
	ContextSummaries map[string][]ChatContextSummary     `json:"contextSummaries"`
}

func (s *Store) CaptureWorkspaceSnapshotLocked() WorkspaceSnapshot {
	return WorkspaceSnapshot{
		SchemaVersion:    "dreamworker.workspace.snapshot.v1",
		Sequence:         s.Sequence,
		Providers:        cloneMap(s.Providers),
		ProviderSecrets:  providerSecrets(s.Providers),
		Profiles:         cloneMap(s.Profiles),
		Agents:           cloneMap(s.Agents),
		Skills:           cloneMap(s.Skills),
		Tools:            cloneMap(s.Tools),
		Servers:          cloneMap(s.Servers),
		MCPServerSecrets: mcpServerSecrets(s.Servers),
		MCPTools:         cloneMap(s.MCPTools),
		Projects:         cloneMap(s.Projects),
		Modules:          cloneNestedMap(s.Modules),
		Sessions:         cloneMap(s.Sessions),
		Messages:         cloneSliceMap(s.Messages),
		ContextSummaries: cloneSliceMap(s.ContextSummaries),
	}
}

func (s *Store) ApplyWorkspaceSnapshotLocked(snapshot WorkspaceSnapshot) {
	s.Sequence = snapshot.Sequence
	s.Providers = cloneMapOrEmpty(snapshot.Providers)
	applyProviderSecrets(s.Providers, snapshot.ProviderSecrets)
	s.Profiles = cloneMapOrEmpty(snapshot.Profiles)
	s.Agents = cloneMapOrEmpty(snapshot.Agents)
	s.Skills = cloneMapOrEmpty(snapshot.Skills)
	s.Tools = cloneMapOrEmpty(snapshot.Tools)
	s.Servers = cloneMapOrEmpty(snapshot.Servers)
	applyMCPServerSecrets(s.Servers, snapshot.MCPServerSecrets)
	s.MCPTools = cloneMapOrEmpty(snapshot.MCPTools)
	s.Projects = cloneMapOrEmpty(snapshot.Projects)
	s.Modules = cloneNestedMap(snapshot.Modules)
	s.Sessions = cloneMapOrEmpty(snapshot.Sessions)
	s.Messages = cloneSliceMap(snapshot.Messages)
	s.ContextSummaries = cloneSliceMap(snapshot.ContextSummaries)
	s.Streams = make(map[string]contextCancel)
	if s.Providers == nil {
		s.Providers = map[string]ModelProviderRecord{}
	}
}

func cloneMap[T any](input map[string]T) map[string]T {
	result := make(map[string]T, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func cloneMapOrEmpty[T any](input map[string]T) map[string]T {
	if input == nil {
		return map[string]T{}
	}
	return cloneMap(input)
}

func cloneNestedMap[T any](input map[string]map[string]T) map[string]map[string]T {
	result := make(map[string]map[string]T, len(input))
	for key, value := range input {
		result[key] = cloneMap(value)
	}
	return result
}

func cloneSliceMap[T any](input map[string][]T) map[string][]T {
	result := make(map[string][]T, len(input))
	for key, values := range input {
		result[key] = append([]T{}, values...)
	}
	return result
}

func providerSecrets(providers map[string]ModelProviderRecord) map[string]string {
	result := make(map[string]string)
	for providerID, provider := range providers {
		if provider.APIKey != "" {
			result[providerID] = provider.APIKey
		}
	}
	return result
}

func applyProviderSecrets(providers map[string]ModelProviderRecord, secrets map[string]string) {
	now := ""
	for providerID, secret := range secrets {
		provider, ok := providers[providerID]
		if !ok {
			continue
		}
		provider.APIKey = secret
		providers[providerID] = normalizeProviderRecord(provider, now)
	}
}

func mcpServerSecrets(servers map[string]MCPServerRecord) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for serverID, server := range servers {
		if len(server.Secrets) > 0 {
			result[serverID] = cloneStringMap(server.Secrets)
		}
	}
	return result
}

func applyMCPServerSecrets(servers map[string]MCPServerRecord, secrets map[string]map[string]string) {
	for serverID, serverSecrets := range secrets {
		server, ok := servers[serverID]
		if !ok {
			continue
		}
		server.Secrets = cloneStringMap(serverSecrets)
		server.EnvKeys, server.MaskedSecrets = secretSummaries(server.Secrets)
		server.HasSecrets = len(server.Secrets) > 0
		servers[serverID] = server
	}
}
