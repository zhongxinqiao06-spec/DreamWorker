package workspace

import "strings"

var projectModuleOrder = map[string]int{
	"explore":     0,
	"product":     1,
	"development": 2,
	"sales":       3,
}

func (s *Store) ListProjects() []Project {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedValues(s.projects, func(item Project) string { return item.Title })
}

func (s *Store) CreateProject(input CreateProjectInput) (Project, *AppError) {
	if strings.TrimSpace(input.Title) == "" {
		return Project{}, BadRequest("BAD_REQUEST", "项目名称不能为空。", "请填写项目名称后重试。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createProjectLocked(input, s.now()), nil
}

func (s *Store) GetProject(projectID string) (Project, *AppError) {
	if projectID == "" {
		return Project{}, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择要查看的项目。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[projectID]
	if !ok {
		return Project{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	return project, nil
}

func (s *Store) UpdateProject(input UpdateProjectInput) (Project, *AppError) {
	if input.ProjectID == "" {
		return Project{}, BadRequest("BAD_REQUEST", "项目更新格式无效。", "请检查 projectId。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[input.ProjectID]
	if !ok {
		return Project{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	if input.Title != nil {
		project.Title = *input.Title
	}
	if input.Description != nil {
		project.Description = *input.Description
	}
	if input.Status != nil {
		project.Status = *input.Status
	}
	if input.DefaultModelProfileID != nil {
		project.DefaultModelProfileID = *input.DefaultModelProfileID
	}
	if input.EnabledAgents != nil {
		project.EnabledAgents = append([]string{}, *input.EnabledAgents...)
	}
	if input.EnabledSkills != nil {
		project.EnabledSkills = append([]string{}, *input.EnabledSkills...)
	}
	if input.EnabledTools != nil {
		project.EnabledTools = append([]string{}, *input.EnabledTools...)
	}
	if input.EnabledMCPServers != nil {
		project.EnabledMCPServers = append([]string{}, *input.EnabledMCPServers...)
	}
	project.UpdatedAt = s.now()
	s.projects[input.ProjectID] = project
	return project, nil
}

func (s *Store) DeleteProject(projectID string) (DeleteResult, *AppError) {
	if projectID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择要删除的项目。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[projectID]; !ok {
		return DeleteResult{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	delete(s.projects, projectID)
	delete(s.modules, projectID)
	now := s.now()
	for sessionID, session := range s.sessions {
		if session.ProjectID != nil && *session.ProjectID == projectID {
			session.ProjectID = nil
			session.UpdatedAt = now
			s.sessions[sessionID] = session
		}
	}
	return DeleteResult{OK: true, DeletedID: projectID}, nil
}

func (s *Store) ListProjectModules(projectID string) ([]ProjectModule, *AppError) {
	if projectID == "" {
		return nil, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择项目后查看模块。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	modules, ok := s.modules[projectID]
	if !ok {
		return nil, NotFound("PROJECT_NOT_FOUND", "项目不存在或模块未初始化。", "请刷新项目列表。")
	}
	values := make([]ProjectModule, 0, len(modules))
	for _, module := range modules {
		values = append(values, module)
	}
	return sortSlice(values, func(item ProjectModule) string {
		index, ok := projectModuleOrder[item.ModuleID]
		if !ok {
			return "99-" + item.ModuleID
		}
		return string(rune('0'+index)) + "-" + item.ModuleID
	}), nil
}

func (s *Store) GetProjectModule(input ModuleRequest) (ProjectModule, *AppError) {
	if input.ProjectID == "" || input.ModuleID == "" {
		return ProjectModule{}, BadRequest("BAD_REQUEST", "缺少 projectId 或 moduleId。", "请选择项目模块。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	module, ok := s.modules[input.ProjectID][input.ModuleID]
	if !ok {
		return ProjectModule{}, NotFound("MODULE_NOT_FOUND", "项目模块不存在。", "请刷新项目空间。")
	}
	return module, nil
}

func (s *Store) UpdateProjectModuleConfig(input UpdateModuleConfigInput) (ProjectModule, *AppError) {
	if input.ProjectID == "" || input.ModuleID == "" {
		return ProjectModule{}, BadRequest("BAD_REQUEST", "模块配置格式无效。", "请检查 projectId 和 moduleId。")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	module, ok := s.modules[input.ProjectID][input.ModuleID]
	if !ok {
		return ProjectModule{}, NotFound("MODULE_NOT_FOUND", "项目模块不存在。", "请刷新项目空间。")
	}
	module.Config = cloneAnyMap(input.Config)
	s.modules[input.ProjectID][input.ModuleID] = module
	return module, nil
}

func (s *Store) createProjectLocked(input CreateProjectInput, timestamp string) Project {
	projectID := s.nextIDLocked("project")
	if len(s.projects) == 0 {
		projectID = "project_001"
	}
	project := Project{
		ProjectID:             projectID,
		Title:                 input.Title,
		Description:           input.Description,
		Status:                "active",
		DefaultModelProfileID: "profile_fast",
		EnabledAgents:         []string{"agent_general_assistant", "agent_opportunity_scout", "agent_product_designer", "agent_system_architect", "agent_sales_strategist"},
		EnabledSkills:         []string{"skill_opportunity_scan", "skill_competitor_map", "skill_prd_draft", "skill_blueprint", "skill_launch_plan"},
		EnabledTools:          []string{"tool_model_generate_stub", "tool_artifact_write", "tool_web_search_stub", "tool_human_input"},
		EnabledMCPServers:     []string{},
		CreatedAt:             timestamp,
		UpdatedAt:             timestamp,
	}
	s.projects[project.ProjectID] = project
	s.modules[project.ProjectID] = createProjectModules(project.ProjectID)
	return project
}
