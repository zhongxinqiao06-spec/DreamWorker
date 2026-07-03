package projects

import "strings"

var defaultModuleOutputDirs = map[string]string{
	"explore":     "artifacts/explore",
	"product":     "artifacts/product",
	"development": "artifacts/development",
	"sales":       "artifacts/sales",
}

func normalizeProject(project Project) Project {
	if strings.TrimSpace(project.Status) == "" {
		project.Status = "active"
	}
	if strings.TrimSpace(project.DefaultModelProfileID) == "" {
		project.DefaultModelProfileID = "profile_fast"
	}
	if strings.TrimSpace(project.LocalDirectoryStatus) == "" {
		if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
			project.LocalDirectoryStatus = "not_set"
			project.LocalRootPath = nil
		} else {
			project.LocalDirectoryStatus = "invalid"
		}
	}
	project.EnabledAgents = cloneStringSlice(project.EnabledAgents)
	project.EnabledSkills = cloneStringSlice(project.EnabledSkills)
	project.EnabledTools = cloneStringSlice(project.EnabledTools)
	project.EnabledMCPServers = cloneStringSlice(project.EnabledMCPServers)
	project.ModuleConfigs = mergeDefaultModuleConfigs(project.ModuleConfigs)
	if project.MemoryConfig.MaxContextTokens <= 0 {
		project.MemoryConfig = defaultProjectMemoryConfig()
	}
	if strings.TrimSpace(project.RunPolicy.PlannerMode) == "" {
		project.RunPolicy = defaultProjectRunPolicy()
	}
	if strings.TrimSpace(project.SecurityPolicy.FileAccessScope) == "" {
		project.SecurityPolicy = defaultProjectSecurityPolicy()
	}
	return project
}

func mergeDefaultModuleConfigs(configs map[string]ProjectModuleConfig) map[string]ProjectModuleConfig {
	defaults := defaultProjectModuleConfigs()
	result := make(map[string]ProjectModuleConfig, len(defaults))
	for moduleID, config := range defaults {
		result[moduleID] = cloneProjectModuleConfig(config)
	}
	for moduleID, config := range configs {
		defaultConfig, ok := defaults[moduleID]
		if ok {
			result[moduleID] = mergeProjectModuleConfig(defaultConfig, config)
			continue
		}
		result[moduleID] = cloneProjectModuleConfig(config)
	}
	return result
}

func mergeProjectModuleConfig(defaultConfig ProjectModuleConfig, config ProjectModuleConfig) ProjectModuleConfig {
	result := cloneProjectModuleConfig(config)
	if result.OutputDir == "" {
		result.OutputDir = defaultConfig.OutputDir
	}
	if result.DefaultAgentIDs == nil {
		result.DefaultAgentIDs = cloneStringSlice(defaultConfig.DefaultAgentIDs)
	}
	if result.EnabledSkillIDs == nil {
		result.EnabledSkillIDs = cloneStringSlice(defaultConfig.EnabledSkillIDs)
	}
	if result.EnabledToolIDs == nil {
		result.EnabledToolIDs = cloneStringSlice(defaultConfig.EnabledToolIDs)
	}
	if result.EnabledMCPServerIDs == nil {
		result.EnabledMCPServerIDs = cloneStringSlice(defaultConfig.EnabledMCPServerIDs)
	}
	if result.InputSchema == nil {
		result.InputSchema = cloneAnyMap(defaultConfig.InputSchema)
	}
	if result.Parameters == nil {
		result.Parameters = cloneAnyMap(defaultConfig.Parameters)
	}
	return result
}

func defaultProjectModuleConfigs() map[string]ProjectModuleConfig {
	modules := createProjectModules("__default__")
	result := make(map[string]ProjectModuleConfig, len(modules))
	for moduleID, module := range modules {
		result[moduleID] = ProjectModuleConfig{
			Enabled:             true,
			DefaultAgentIDs:     cloneStringSlice(module.DefaultAgents),
			EnabledSkillIDs:     cloneStringSlice(module.EnabledSkills),
			EnabledToolIDs:      cloneStringSlice(module.EnabledTools),
			EnabledMCPServerIDs: cloneStringSlice(module.EnabledMCPServers),
			OutputDir:           fallback(defaultModuleOutputDirs[moduleID], "artifacts/"+moduleID),
			InputSchema:         map[string]any{},
			Parameters:          cloneAnyMap(module.Config),
		}
	}
	return result
}

func defaultProjectMemoryConfig() ProjectMemoryConfig {
	return ProjectMemoryConfig{
		ProjectMemoryEnabled:  true,
		ArtifactIndexEnabled:  true,
		LocalFileIndexEnabled: false,
		MaxContextTokens:      64000,
	}
}

func defaultProjectRunPolicy() ProjectRunPolicy {
	return ProjectRunPolicy{
		PlannerMode:                     "plan_execute",
		ExecutorMode:                    "safe",
		MaxRunCostUSD:                   5,
		MaxRunMinutes:                   30,
		RequireApprovalForHighRiskTools: true,
	}
}

func defaultProjectSecurityPolicy() ProjectSecurityPolicy {
	return ProjectSecurityPolicy{
		FileAccessScope:     "project_directory_only",
		AllowWriteArtifacts: true,
		AllowWriteSource:    false,
		AllowShellExecution: false,
		AllowNetworkTools:   true,
	}
}

func cloneProjectModuleConfig(config ProjectModuleConfig) ProjectModuleConfig {
	return ProjectModuleConfig{
		Enabled:             config.Enabled,
		DefaultAgentIDs:     cloneStringSlice(config.DefaultAgentIDs),
		EnabledSkillIDs:     cloneStringSlice(config.EnabledSkillIDs),
		EnabledToolIDs:      cloneStringSlice(config.EnabledToolIDs),
		EnabledMCPServerIDs: cloneStringSlice(config.EnabledMCPServerIDs),
		OutputDir:           config.OutputDir,
		InputSchema:         cloneAnyMap(config.InputSchema),
		Parameters:          cloneAnyMap(config.Parameters),
	}
}

func cloneModuleConfigs(configs map[string]ProjectModuleConfig) map[string]ProjectModuleConfig {
	result := make(map[string]ProjectModuleConfig, len(configs))
	for key, config := range configs {
		result[key] = cloneProjectModuleConfig(config)
	}
	return result
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string{}, values...)
}
