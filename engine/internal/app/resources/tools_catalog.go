package resources

func (s *Store) ListTools() []ToolConfig {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return sortedValues(s.Tools, func(item ToolConfig) string { return item.DisplayName })
}

func (s *Store) GetTool(toolID string) (ToolConfig, *AppError) {
	if toolID == "" {
		return ToolConfig{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要查看的工具。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	tool, ok := s.Tools[toolID]
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
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if existing, ok := s.Tools[input.ToolID]; ok && existing.BuiltIn {
		input.BuiltIn = true
	}
	s.Tools[input.ToolID] = input
	if appErr := s.persistWorkspaceLocked(); appErr != nil {
		return ToolConfig{}, appErr
	}
	return input, nil
}

func (s *Store) SetToolEnabled(toolID string, enabled bool) (ToolConfig, *AppError) {
	if toolID == "" {
		return ToolConfig{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要切换的工具。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	tool, ok := s.Tools[toolID]
	if !ok {
		return ToolConfig{}, NotFound("TOOL_NOT_FOUND", "工具不存在。", "请刷新工具列表。")
	}
	tool.Enabled = enabled
	s.Tools[toolID] = tool
	if appErr := s.persistWorkspaceLocked(); appErr != nil {
		return ToolConfig{}, appErr
	}
	return tool, nil
}

func (s *Store) DeleteTool(toolID string) (DeleteResult, *AppError) {
	if toolID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 toolId。", "请选择要删除的工具。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Tools, toolID)
	if appErr := s.persistWorkspaceLocked(); appErr != nil {
		return DeleteResult{}, appErr
	}
	return DeleteResult{OK: true, DeletedID: toolID}, nil
}
