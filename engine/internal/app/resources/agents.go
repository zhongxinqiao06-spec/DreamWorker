package resources

import "strings"

func (s *Store) ListAgents() []AgentConfig {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return sortedValues(s.Agents, func(item AgentConfig) string { return item.DisplayName })
}

func (s *Store) GetAgent(agentID string) (AgentConfig, *AppError) {
	if agentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要查看的 Agent。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	agent, ok := s.Agents[agentID]
	if !ok {
		return AgentConfig{}, NotFound("AGENT_NOT_FOUND", "Agent 不存在。", "请刷新 Agent 列表。")
	}
	return agent, nil
}

func (s *Store) SaveAgent(input AgentConfig) (AgentConfig, *AppError) {
	if input.AgentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "Agent 配置格式无效。", "请检查 agentId 和模型配置。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	now := s.Now()
	existing, exists := s.Agents[input.AgentID]
	if !exists {
		input.CreatedAt = now
	} else {
		input.CreatedAt = existing.CreatedAt
		input.BuiltIn = existing.BuiltIn
	}
	input = ensureAgentRuntimeDefaults(input)
	input = s.ensureAgentModelDefaultsLocked(input)
	input.UpdatedAt = now
	s.Agents[input.AgentID] = input
	return input, nil
}

func (s *Store) ensureAgentModelDefaultsLocked(agent AgentConfig) AgentConfig {
	providerID := strings.TrimSpace(agent.ProviderID)
	model := strings.TrimSpace(agent.Model)
	if providerID != "" {
		if provider, ok := s.Providers[providerID]; ok {
			model = NormalizeProviderModelID(providerID, model)
			if model == "" {
				model = provider.DefaultModel
			}
			profileID := ProfileIDForProviderModel(providerID, model)
			if _, ok := s.Profiles[profileID]; !ok {
				s.Profiles[profileID] = ProfileFromProviderModel(provider, model, s.Now())
			}
			agent.ModelProfileID = profileID
			agent.ProviderID = providerID
			agent.Model = model
			return agent
		}
	}
	if profile, ok := s.Profiles[agent.ModelProfileID]; ok {
		agent.ProviderID = profile.ProviderID
		agent.Model = profile.Model
		return agent
	}
	if provider, ok := s.Providers["provider_deepseek"]; ok {
		agent.ProviderID = provider.ProviderID
		agent.Model = provider.DefaultModel
		agent.ModelProfileID = ProfileIDForProviderModel(provider.ProviderID, provider.DefaultModel)
	}
	return agent
}

func (s *Store) DuplicateAgent(agentID string) (AgentConfig, *AppError) {
	if agentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要复制的 Agent。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	agent, ok := s.Agents[agentID]
	if !ok {
		return AgentConfig{}, NotFound("AGENT_NOT_FOUND", "Agent 不存在。", "请刷新 Agent 列表。")
	}
	now := s.Now()
	agent.AgentID = s.nextIDLocked("agent_custom")
	agent.DisplayName += " 副本"
	agent.BuiltIn = false
	agent.CreatedAt = now
	agent.UpdatedAt = now
	s.Agents[agent.AgentID] = agent
	return agent, nil
}

func (s *Store) DeleteAgent(agentID string) (DeleteResult, *AppError) {
	if agentID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要删除的 Agent。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Agents, agentID)
	return DeleteResult{OK: true, DeletedID: agentID}, nil
}

func ensureAgentRuntimeDefaults(agent AgentConfig) AgentConfig {
	if agent.RuntimeConfig.ContextWindow == 0 {
		agent.RuntimeConfig.ContextWindow = 128000
	}
	if agent.RuntimeConfig.MaxTokens == 0 {
		agent.RuntimeConfig.MaxTokens = 4096
	}
	if !agent.Planner.Enabled {
		agent.Planner.Enabled = true
	}
	if agent.Planner.Strategy == "" {
		agent.Planner.Strategy = "plan-execute"
	}
	if agent.Executor.TimeoutMS == 0 {
		agent.Executor.TimeoutMS = 120000
	}
	if agent.Executor.RetryPolicy == "" {
		agent.Executor.RetryPolicy = "retry_twice_then_ask"
	}
	if agent.MemoryScope == "" {
		agent.MemoryScope = "project"
	}
	return agent
}
