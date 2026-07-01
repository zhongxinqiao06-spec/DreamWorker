package workspace

func (s *Store) ListAgents() []AgentConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedValues(s.agents, func(item AgentConfig) string { return item.DisplayName })
}

func (s *Store) GetAgent(agentID string) (AgentConfig, *AppError) {
	if agentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要查看的 Agent。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentConfig{}, NotFound("AGENT_NOT_FOUND", "Agent 不存在。", "请刷新 Agent 列表。")
	}
	return agent, nil
}

func (s *Store) SaveAgent(input AgentConfig) (AgentConfig, *AppError) {
	if input.AgentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "Agent 配置格式无效。", "请检查 agentId 和模型配置。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	existing, exists := s.agents[input.AgentID]
	if !exists {
		input.CreatedAt = now
	} else {
		input.CreatedAt = existing.CreatedAt
		input.BuiltIn = existing.BuiltIn
	}
	input = ensureAgentRuntimeDefaults(input)
	input.UpdatedAt = now
	s.agents[input.AgentID] = input
	return input, nil
}

func (s *Store) DuplicateAgent(agentID string) (AgentConfig, *AppError) {
	if agentID == "" {
		return AgentConfig{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要复制的 Agent。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentConfig{}, NotFound("AGENT_NOT_FOUND", "Agent 不存在。", "请刷新 Agent 列表。")
	}
	now := s.now()
	agent.AgentID = s.nextIDLocked("agent_custom")
	agent.DisplayName += " 副本"
	agent.BuiltIn = false
	agent.CreatedAt = now
	agent.UpdatedAt = now
	s.agents[agent.AgentID] = agent
	return agent, nil
}

func (s *Store) DeleteAgent(agentID string) (DeleteResult, *AppError) {
	if agentID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 agentId。", "请选择要删除的 Agent。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, agentID)
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

func (s *Store) ListSkills() []SkillConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedValues(s.skills, func(item SkillConfig) string { return item.DisplayName })
}

func (s *Store) GetSkill(skillID string) (SkillConfig, *AppError) {
	if skillID == "" {
		return SkillConfig{}, BadRequest("BAD_REQUEST", "缺少 skillId。", "请选择要查看的 Skill。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	skill, ok := s.skills[skillID]
	if !ok {
		return SkillConfig{}, NotFound("SKILL_NOT_FOUND", "Skill 不存在。", "请刷新 Skill 列表。")
	}
	return skill, nil
}

func (s *Store) SaveSkill(input SkillConfig) (SkillConfig, *AppError) {
	if input.SkillID == "" {
		return SkillConfig{}, BadRequest("BAD_REQUEST", "Skill 配置格式无效。", "请检查 skillId 和输出产物。")
	}
	input = ensureSkillDefaults(input)
	s.mu.Lock()
	if existing, ok := s.skills[input.SkillID]; ok {
		input.BuiltIn = existing.BuiltIn
		if input.SourcePath == "" {
			input.SourcePath = existing.SourcePath
		}
	}
	s.mu.Unlock()
	written, err := s.writeAgentSkillFile(input)
	if err != nil {
		return SkillConfig{}, BadRequest("SKILL_WRITE_FAILED", "Skill 文件写入失败。", "请检查 .agent 目录权限。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	input = written
	s.skills[input.SkillID] = input
	return input, nil
}

func (s *Store) DeleteSkill(skillID string) (DeleteResult, *AppError) {
	if skillID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 skillId。", "请选择要删除的 Skill。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.skills, skillID)
	return DeleteResult{OK: true, DeletedID: skillID}, nil
}

func (s *Store) ListTools() []ToolConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedValues(s.tools, func(item ToolConfig) string { return item.DisplayName })
}

func (s *Store) GetTool(toolID string) (ToolConfig, *AppError) {
	if toolID == "" {
		return ToolConfig{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要查看的工具。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tool, ok := s.tools[toolID]
	if !ok {
		return ToolConfig{}, NotFound("TOOL_NOT_FOUND", "工具不存在。", "请刷新工具列表。")
	}
	return tool, nil
}

func (s *Store) SaveTool(input ToolConfig) (ToolConfig, *AppError) {
	if input.ToolID == "" || input.DisplayName == "" {
		return ToolConfig{}, BadRequest("BAD_REQUEST", "工具配置格式无效。", "请填写 toolId 和名称。")
	}
	if input.Category == "" {
		input.Category = "project"
	}
	if input.RiskLevel == "" {
		input.RiskLevel = "medium"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.tools[input.ToolID]; ok && existing.BuiltIn {
		input.BuiltIn = true
	}
	s.tools[input.ToolID] = input
	return input, nil
}

func (s *Store) SetToolEnabled(toolID string, enabled bool) (ToolConfig, *AppError) {
	if toolID == "" {
		return ToolConfig{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要切换的工具。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tool, ok := s.tools[toolID]
	if !ok {
		return ToolConfig{}, NotFound("TOOL_NOT_FOUND", "工具不存在。", "请刷新工具列表。")
	}
	tool.Enabled = enabled
	s.tools[toolID] = tool
	return tool, nil
}

func (s *Store) DeleteTool(toolID string) (DeleteResult, *AppError) {
	if toolID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要删除的工具。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tools, toolID)
	return DeleteResult{OK: true, DeletedID: toolID}, nil
}
