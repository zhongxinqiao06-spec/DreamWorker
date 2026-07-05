package projects

import "strings"

var projectModuleOrder = map[string]int{
	"explore":     0,
	"product":     1,
	"development": 2,
	"sales":       3,
}

func (s *Store) ListProjects() []Project {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	for projectID, project := range s.Projects {
		s.Projects[projectID] = normalizeProject(project)
	}
	return sortedValues(s.Projects, func(item Project) string { return item.Title })
}

func (s *Store) CreateProject(input CreateProjectInput) (Project, *AppError) {
	if strings.TrimSpace(input.Title) == "" {
		return Project{}, BadRequest("BAD_REQUEST", "项目名称不能为空。", "请填写项目名称后重试。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	project := s.createProjectLocked(input, s.Now())
	if appErr := s.PersistWorkspaceSnapshotLocked(); appErr != nil {
		return Project{}, appErr
	}
	return project, nil
}

func (s *Store) GetProject(projectID string) (Project, *AppError) {
	if projectID == "" {
		return Project{}, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择要查看的项目。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	project, ok := s.Projects[projectID]
	if !ok {
		return Project{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	project = normalizeProject(project)
	s.Projects[projectID] = project
	return project, nil
}

func (s *Store) UpdateProject(input UpdateProjectInput) (Project, *AppError) {
	if input.ProjectID == "" {
		return Project{}, BadRequest("BAD_REQUEST", "项目更新格式无效。", "请检查 projectId。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	project, ok := s.Projects[input.ProjectID]
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
	if input.LocalRootPath != nil {
		previousPath := optionalPathValue(project.LocalRootPath)
		cleaned := strings.TrimSpace(*input.LocalRootPath)
		if cleaned == "" {
			project.LocalRootPath = nil
			project.LocalDirectoryStatus = "not_set"
			project.LocalDirectoryLastCheckedAt = nil
		} else {
			project.LocalRootPath = &cleaned
			if previousPath != cleaned || project.LocalDirectoryStatus == "not_set" || project.LocalDirectoryStatus == "" {
				project.LocalDirectoryStatus = "invalid"
				project.LocalDirectoryLastCheckedAt = nil
			}
		}
	}
	if input.DefaultModelProfileID != nil {
		project.DefaultModelProfileID = *input.DefaultModelProfileID
	}
	if input.DefaultRouteProfileID != nil {
		cleaned := strings.TrimSpace(*input.DefaultRouteProfileID)
		if cleaned == "" {
			project.DefaultRouteProfileID = nil
		} else {
			project.DefaultRouteProfileID = &cleaned
		}
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
	if input.ModuleConfigs != nil {
		project.ModuleConfigs = cloneModuleConfigs(*input.ModuleConfigs)
	}
	if input.MemoryConfig != nil {
		project.MemoryConfig = *input.MemoryConfig
	}
	if input.RunPolicy != nil {
		project.RunPolicy = *input.RunPolicy
	}
	if input.SecurityPolicy != nil {
		project.SecurityPolicy = *input.SecurityPolicy
	}
	project.UpdatedAt = s.Now()
	project = normalizeProject(project)
	if appErr := s.writeProjectManifestFilesIfInitialized(project); appErr != nil {
		return Project{}, appErr
	}
	s.Projects[input.ProjectID] = project
	if appErr := s.PersistWorkspaceSnapshotLocked(); appErr != nil {
		return Project{}, appErr
	}
	return project, nil
}

func (s *Store) DeleteProject(projectID string) (DeleteResult, *AppError) {
	if projectID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择要删除的项目。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, ok := s.Projects[projectID]; !ok {
		return DeleteResult{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	delete(s.Projects, projectID)
	delete(s.Modules, projectID)
	now := s.Now()
	for sessionID, session := range s.Sessions {
		if session.ProjectID != nil && *session.ProjectID == projectID {
			session.ProjectID = nil
			session.UpdatedAt = now
			s.Sessions[sessionID] = session
		}
	}
	if appErr := s.PersistWorkspaceSnapshotLocked(); appErr != nil {
		return DeleteResult{}, appErr
	}
	return DeleteResult{OK: true, DeletedID: projectID}, nil
}

func (s *Store) ListProjectModules(projectID string) ([]ProjectModule, *AppError) {
	if projectID == "" {
		return nil, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择项目后查看模块。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	modules, ok := s.Modules[projectID]
	if !ok {
		return nil, NotFound("PROJECT_NOT_FOUND", "项目不存在或模块未初始化。", "请刷新项目列表。")
	}
	modules = normalizeProjectModuleSet(projectID, modules)
	s.Modules[projectID] = modules
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
	s.Mu.Lock()
	defer s.Mu.Unlock()
	module, ok := s.Modules[input.ProjectID][input.ModuleID]
	if !ok {
		return ProjectModule{}, NotFound("MODULE_NOT_FOUND", "项目模块不存在。", "请刷新项目空间。")
	}
	return module, nil
}

func (s *Store) UpdateProjectModuleConfig(input UpdateModuleConfigInput) (ProjectModule, *AppError) {
	if input.ProjectID == "" || input.ModuleID == "" {
		return ProjectModule{}, BadRequest("BAD_REQUEST", "模块配置格式无效。", "请检查 projectId 和 moduleId。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	module, ok := s.Modules[input.ProjectID][input.ModuleID]
	if !ok {
		return ProjectModule{}, NotFound("MODULE_NOT_FOUND", "项目模块不存在。", "请刷新项目空间。")
	}
	module.Config = cloneAnyMap(input.Config)
	s.Modules[input.ProjectID][input.ModuleID] = module
	if appErr := s.PersistWorkspaceSnapshotLocked(); appErr != nil {
		return ProjectModule{}, appErr
	}
	return module, nil
}

func (s *Store) createProjectLocked(input CreateProjectInput, timestamp string) Project {
	projectID := s.NextIDLocked("project")
	if len(s.Projects) == 0 {
		projectID = "project_001"
	}
	project := Project{
		ProjectID:             projectID,
		Title:                 input.Title,
		Description:           input.Description,
		Status:                "active",
		LocalRootPath:         cleanOptionalPath(input.LocalRootPath),
		LocalDirectoryStatus:  "not_set",
		DefaultModelProfileID: "profile_fast",
		EnabledAgents:         []string{"agent_general_assistant", "agent_opportunity_scout", "agent_product_designer", "agent_system_architect", "agent_sales_strategist"},
		EnabledSkills:         []string{"skill_opportunity_scan", "skill_competitor_map", "skill_prd_draft", "skill_blueprint", "skill_launch_plan"},
		EnabledTools:          []string{"tool_model_generate_stub", "tool_artifact_write", "tool_web_search_stub", "tool_human_input"},
		EnabledMCPServers:     []string{},
		ModuleConfigs:         defaultProjectModuleConfigs(),
		MemoryConfig:          defaultProjectMemoryConfig(),
		RunPolicy:             defaultProjectRunPolicy(),
		SecurityPolicy:        defaultProjectSecurityPolicy(),
		CreatedAt:             timestamp,
		UpdatedAt:             timestamp,
	}
	if project.LocalRootPath != nil {
		project.LocalDirectoryStatus = "invalid"
	}
	project = normalizeProject(project)
	s.Projects[project.ProjectID] = project
	s.Modules[project.ProjectID] = normalizeProjectModuleSet(project.ProjectID, createProjectModules(project.ProjectID))
	return project
}

func cleanOptionalPath(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func optionalPathValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
