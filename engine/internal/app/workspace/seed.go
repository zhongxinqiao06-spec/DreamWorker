package workspace

import (
	"os"
	"strings"
)

func (s *Store) seed() {
	timestamp := "2026-07-01T00:00:00Z"
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	deepseekBaseURL := envDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
	deepseekModel := envDefault("DEEPSEEK_MODEL", "deepseek-v4-flash")
	var deepseekMasked *string
	deepseekStatus := "unknown"
	if deepseekKey != "" {
		masked := maskSecret(deepseekKey)
		deepseekMasked = &masked
		deepseekStatus = "unknown"
	}
	s.providers["provider_deepseek"] = ModelProviderRecord{
		SafeModelProvider: SafeModelProvider{
			ProviderID:      "provider_deepseek",
			ProviderType:    ProviderDeepSeek,
			DisplayName:     "DeepSeek 兼容服务",
			BaseURL:         deepseekBaseURL,
			DefaultModel:    deepseekModel,
			AvailableModels: prependUnique(deepseekModel, []string{"deepseek-v4-flash", "deepseek-v4-pro", "deepseek-chat", "deepseek-reasoner"}),
			Enabled:         true,
			Status:          deepseekStatus,
			HealthStatus:    deepseekStatus,
			Capabilities:    []string{"chat", "tools", "json_schema"},
			SupportsStream:  true,
			HasAPIKey:       deepseekKey != "",
			MaskedKey:       deepseekMasked,
			CreatedAt:       timestamp,
			UpdatedAt:       timestamp,
		},
		APIKey: deepseekKey,
	}
	s.providers["provider_local_stub"] = ModelProviderRecord{
		SafeModelProvider: SafeModelProvider{
			ProviderID:        "provider_local_stub",
			ProviderType:      ProviderOpenAICompatible,
			DisplayName:       "本地 Stub 模型",
			BaseURL:           "http://127.0.0.1/model-stub",
			DefaultModel:      "model_generate_stub",
			AvailableModels:   []string{"model_generate_stub"},
			Enabled:           true,
			Status:            "connected",
			Capabilities:      []string{"chat", "tools", "json_schema"},
			SupportsStream:    true,
			HealthStatus:      "connected",
			StreamingVerified: true,
			CreatedAt:         timestamp,
			UpdatedAt:         timestamp,
		},
	}
	stubFallback := "profile_stub"
	s.profiles["profile_fast"] = ModelProfile{
		ProfileID:         "profile_fast",
		DisplayName:       "快速真实模型",
		ProviderID:        "provider_deepseek",
		Model:             deepseekModel,
		Temperature:       0.4,
		MaxTokens:         4096,
		ContextWindow:     128000,
		ResponseFormat:    "text",
		ToolMode:          "auto",
		FallbackProfileID: &stubFallback,
		TimeoutMS:         120000,
		Purpose:           "聊天、探索、短产物生成",
		Enabled:           true,
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
	}
	s.profiles["profile_stub"] = ModelProfile{
		ProfileID:      "profile_stub",
		DisplayName:    "离线确定性模型",
		ProviderID:     "provider_local_stub",
		Model:          "model_generate_stub",
		Temperature:    0,
		MaxTokens:      2048,
		ContextWindow:  32000,
		ResponseFormat: "text",
		ToolMode:       "none",
		TimeoutMS:      30000,
		Purpose:        "测试、CI、无网络演示",
		Enabled:        true,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
	}
	s.bootstrapModelGatewayFromEnv(timestamp)
	s.seedTools()
	s.seedAgents(timestamp)
	s.servers["mcp_local_files"] = MCPServerRecord{
		MCPServerConfig: MCPServerConfig{
			ServerID:      "mcp_local_files",
			DisplayName:   "本地文件 MCP",
			Command:       "dreamworker-mcp-files",
			Args:          []string{"--project-root", "."},
			EnvKeys:       []string{},
			TrustLevel:    "trusted_builtin",
			Enabled:       false,
			HasSecrets:    false,
			MaskedSecrets: []string{},
			CreatedAt:     timestamp,
			UpdatedAt:     timestamp,
		},
		Secrets: map[string]string{},
	}
	s.createProjectLocked(CreateProjectInput{
		Title:       "独立开发者 AI 项目孵化器",
		Description: "围绕资源配置、普通 Agent 聊天、项目空间和四大项目闭环模块验证产品骨架。",
	}, timestamp)
	session := ChatSession{
		SessionID:      "chat_general_001",
		ProjectID:      ptr("project_001"),
		Title:          "普通 Agent 工作台",
		AgentID:        "agent_general_assistant",
		ModelProfileID: "profile_fast",
		MessageCount:   2,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
	}
	s.sessions[session.SessionID] = session
	s.messages[session.SessionID] = []ChatMessage{
		{
			MessageID: "msg_001",
			SessionID: session.SessionID,
			Role:      "system",
			Content:   "当前对话运行在 DreamWorker 普通 Agent 工作台，资源与项目上下文由 Go Engine 提供。",
			TraceID:   "tr_seed",
			CreatedAt: timestamp,
		},
		{
			MessageID: "msg_002",
			SessionID: session.SessionID,
			Role:      "assistant",
			Content:   "可以直接提问、切换 Agent，或把当前对话绑定到项目空间中的探索、产品、开发、销售模块。",
			TraceID:   "tr_seed",
			CreatedAt: timestamp,
		},
	}
}

func (s *Store) seedTools() {
	tools := []ToolConfig{
		{"tool_artifact_read", "读取产物", "读取项目空间内的 Artifact 元数据和内容。", "artifact", "low", true, true},
		{"tool_artifact_write", "写入产物", "只允许写入当前项目目录内的 Artifact。", "artifact", "medium", true, true},
		{"tool_web_search_stub", "网页搜索 Stub", "离线返回确定性搜索线索，后续可接真实搜索。", "search", "medium", true, true},
		{"tool_browser_readonly_stub", "只读浏览 Stub", "只读网页摘要能力占位，不执行浏览器写操作。", "browser", "medium", true, true},
		{"tool_model_generate_stub", "模型生成 Stub", "确定性模型生成能力，用于 CI 和本地演示。", "model", "low", true, true},
		{"tool_human_input", "人工输入", "把 ask_user、审批和 steering 交还给用户。", "human", "low", true, true},
	}
	for _, tool := range tools {
		s.tools[tool.ToolID] = tool
	}
}

func (s *Store) seedAgents(timestamp string) {
	agents := []AgentConfig{
		seedAgent("agent_general_assistant", "通用助手", "普通 Agent 聊天入口", "处理日常问答、上下文整理和轻量任务拆解。", "你是 DreamWorker 的通用助手，优先用中文清晰回答。", "profile_fast", []string{"skill_opportunity_scan"}, []string{"tool_model_generate_stub", "tool_human_input"}, []string{}, timestamp),
		seedAgent("agent_opportunity_scout", "机会侦察员", "探索机会", "分析目标人群、痛点、替代方案和机会窗口。", "你负责把想法拆成可验证机会。", "profile_fast", []string{"skill_opportunity_scan", "skill_competitor_map"}, []string{"tool_web_search_stub", "tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_competitor_analyst", "竞品分析师", "竞品分析", "比较竞品定位、功能、价格、渠道和差异化。", "你负责生成证据优先的竞品分析。", "profile_fast", []string{"skill_competitor_map"}, []string{"tool_web_search_stub", "tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_customer_segment", "客群分析师", "客群细分", "拆解 ICP、购买动机、预算和触达方式。", "你负责判断谁最可能先买单。", "profile_fast", []string{"skill_opportunity_scan"}, []string{"tool_model_generate_stub"}, []string{}, timestamp),
		seedAgent("agent_product_designer", "产品设计师", "产品定义", "把验证后的机会转成信息架构、用户路径和 PRD。", "你负责把证据转成产品结构。", "profile_fast", []string{"skill_prd_draft"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_prototype_designer", "原型设计师", "原型方案", "输出关键页面、状态和交互说明。", "你负责快速定义可验证原型。", "profile_fast", []string{"skill_prd_draft"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_system_architect", "系统架构师", "系统设计", "设计边界、数据流、契约和安全策略。", "你负责工程边界和可演进架构。", "profile_fast", []string{"skill_blueprint"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_tech_stack_advisor", "技术栈顾问", "技术选型", "评估框架、存储、模型、部署和成本。", "你负责技术路线的取舍。", "profile_fast", []string{"skill_blueprint"}, []string{"tool_model_generate_stub"}, []string{}, timestamp),
		seedAgent("agent_dev_orchestrator", "开发编排员", "开发计划", "把蓝图拆成可验收 PR 和测试门禁。", "你负责把方案变成工程计划。", "profile_fast", []string{"skill_blueprint"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_sales_strategist", "销售策略师", "销售策略", "生成定位、渠道、线索和定价验证动作。", "你负责把产品推到市场。", "profile_fast", []string{"skill_launch_plan"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_demo_designer", "演示设计师", "演示方案", "设计 demo flow、话术和发布素材。", "你负责把价值讲清楚。", "profile_fast", []string{"skill_launch_plan"}, []string{"tool_artifact_write"}, []string{}, timestamp),
		seedAgent("agent_evaluator", "评估员", "质量评估", "检查产物完整度、证据质量、幻觉风险和下一步行动。", "你负责判断输出是否可以进入下一阶段。", "profile_stub", []string{"skill_opportunity_scan", "skill_blueprint"}, []string{"tool_model_generate_stub"}, []string{}, timestamp),
	}
	for _, agent := range agents {
		s.agents[agent.AgentID] = agent
	}
}

func seedAgent(agentID string, displayName string, role string, description string, systemPrompt string, modelProfileID string, skills []string, tools []string, mcpServers []string, timestamp string) AgentConfig {
	return AgentConfig{
		AgentID:           agentID,
		DisplayName:       displayName,
		Role:              role,
		Description:       description,
		SystemPrompt:      systemPrompt,
		ModelProfileID:    modelProfileID,
		EnabledSkills:     skills,
		EnabledTools:      tools,
		EnabledMCPServers: mcpServers,
		RuntimeConfig:     AgentRuntimeConfig{ContextWindow: 128000, Temperature: 0.4, MaxTokens: 4096},
		Planner:           AgentPlannerConfig{Enabled: true, Strategy: "plan-execute"},
		Executor:          AgentExecutorConfig{TimeoutMS: 120000, RetryPolicy: "retry_twice_then_ask"},
		MemoryScope:       "project",
		Enabled:           true,
		BuiltIn:           true,
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
	}
}

func createDefaultModules(projectID string) map[string]ProjectModule {
	return map[string]ProjectModule{
		"explore": {
			ProjectID:       projectID,
			ModuleID:        "explore",
			DisplayName:     "探索模块",
			Status:          "ready",
			Summary:         "负责机会扫描、客群细分、竞品地图和证据收集。",
			DefaultAgents:   []string{"agent_opportunity_scout", "agent_competitor_analyst", "agent_customer_segment"},
			EnabledSkills:   []string{"skill_opportunity_scan", "skill_competitor_map"},
			EnabledTools:    []string{"tool_web_search_stub", "tool_model_generate_stub", "tool_artifact_write"},
			OutputArtifacts: []string{"dream_brief.md", "hypotheses.yaml", "research_pack.md"},
			NextBestAction:  "先跑机会扫描，再补竞品和客群证据。",
			Config:          map[string]any{"stage": "Discover", "evidenceRequired": true},
		},
		"product": {
			ProjectID:       projectID,
			ModuleID:        "product",
			DisplayName:     "产品模块",
			Status:          "idle",
			Summary:         "负责 MVP 收敛、PRD、原型说明和 Blueprint Canvas 输入。",
			DefaultAgents:   []string{"agent_product_designer", "agent_prototype_designer", "agent_evaluator"},
			EnabledSkills:   []string{"skill_prd_draft"},
			EnabledTools:    []string{"tool_model_generate_stub", "tool_artifact_write"},
			OutputArtifacts: []string{"mvp_scope.md", "prd.md", "prototype_notes.md"},
			NextBestAction:  "确认验证结论后再生成 MVP 范围。",
			Config:          map[string]any{"stage": "Shape", "requiresDecisionGate": true},
		},
		"development": {
			ProjectID:       projectID,
			ModuleID:        "development",
			DisplayName:     "开发模块",
			Status:          "idle",
			Summary:         "负责系统架构、技术栈、PR 拆分、测试门禁和运行计划。",
			DefaultAgents:   []string{"agent_system_architect", "agent_tech_stack_advisor", "agent_dev_orchestrator"},
			EnabledSkills:   []string{"skill_blueprint"},
			EnabledTools:    []string{"tool_model_generate_stub", "tool_artifact_write"},
			OutputArtifacts: []string{"blueprint.yaml", "dev_plan.md", "issue_plan.md"},
			NextBestAction:  "等 PRD 草稿稳定后输出工程蓝图。",
			Config:          map[string]any{"stage": "Build", "writeCodeAutomatically": false},
		},
		"sales": {
			ProjectID:       projectID,
			ModuleID:        "sales",
			DisplayName:     "销售模块",
			Status:          "idle",
			Summary:         "负责定位、落地页文案、发布计划、Demo 和反馈循环。",
			DefaultAgents:   []string{"agent_sales_strategist", "agent_demo_designer", "agent_evaluator"},
			EnabledSkills:   []string{"skill_launch_plan"},
			EnabledTools:    []string{"tool_model_generate_stub", "tool_artifact_write", "tool_human_input"},
			OutputArtifacts: []string{"launch_checklist.md", "landing_copy.md", "demo_script.md"},
			NextBestAction:  "等产品范围确认后准备发布清单。",
			Config:          map[string]any{"stage": "Launch", "publishRequiresApproval": true},
		},
	}
}

func (s *Store) bootstrapModelGatewayFromEnv(timestamp string) {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		s.addEnvProviderAndProfile(timestamp, "provider_openai", ProviderOpenAI, "OpenAI", "https://api.openai.com", "gpt-5-mini", []string{"gpt-5-mini", "gpt-5.2"}, key)
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		s.addEnvProviderAndProfile(timestamp, "provider_anthropic", ProviderAnthropic, "Anthropic", "https://api.anthropic.com", "claude-sonnet-4-5", []string{"claude-sonnet-4-5", "claude-haiku-4-5"}, key)
	}
	if key := os.Getenv("SILICONFLOW_API_KEY"); key != "" {
		model := envDefault("SILICONFLOW_MODEL", "deepseek-ai/DeepSeek-V4-Flash")
		models := prependUnique(model, []string{"deepseek-ai/DeepSeek-V4-Flash", "deepseek-ai/DeepSeek-V4-Pro", "zai-org/GLM-5.2", "Qwen/Qwen3.5-4B"})
		s.addEnvProviderAndProfile(timestamp, "provider_siliconflow", ProviderSiliconFlow, "SiliconFlow 硅基流动", envDefault("SILICONFLOW_BASE_URL", "https://api.siliconflow.cn/v1"), model, models, key)
	}
	if key := firstEnv("GLM_API_KEY", "ZAI_API_KEY", "BIGMODEL_API_KEY"); key != "" {
		model := envDefault("GLM_MODEL", "glm-5.2")
		models := prependUnique(model, defaultProviderModels(ProviderGLM))
		s.addEnvProviderAndProfile(timestamp, "provider_glm", ProviderGLM, "GLM 智谱", envDefault("GLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"), model, models, key)
	}
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		s.providers["provider_ollama"] = ModelProviderRecord{
			SafeModelProvider: SafeModelProvider{
				ProviderID:        "provider_ollama",
				ProviderType:      ProviderOllama,
				DisplayName:       "Ollama Local",
				BaseURL:           host,
				DefaultModel:      "llama3.1",
				AvailableModels:   []string{"llama3.1"},
				Enabled:           true,
				Status:            "unknown",
				HealthStatus:      "unknown",
				Capabilities:      []string{"chat", "tools"},
				SupportsStream:    true,
				StreamingVerified: false,
				CreatedAt:         timestamp,
				UpdatedAt:         timestamp,
			},
		}
		s.profiles["profile_ollama"] = ModelProfile{
			ProfileID:      "profile_ollama",
			DisplayName:    "Ollama local",
			ProviderID:     "provider_ollama",
			Model:          "llama3.1",
			Temperature:    0.4,
			MaxTokens:      4096,
			ContextWindow:  32000,
			ResponseFormat: "text",
			ToolMode:       "auto",
			TimeoutMS:      120000,
			Purpose:        "Local Ollama chat profile",
			Enabled:        true,
			CreatedAt:      timestamp,
			UpdatedAt:      timestamp,
		}
	}
}

func (s *Store) addEnvProviderAndProfile(
	timestamp string,
	providerID string,
	providerType ProviderType,
	displayName string,
	baseURL string,
	defaultModel string,
	models []string,
	apiKey string,
) {
	masked := maskSecret(apiKey)
	s.providers[providerID] = ModelProviderRecord{
		SafeModelProvider: SafeModelProvider{
			ProviderID:      providerID,
			ProviderType:    providerType,
			DisplayName:     displayName,
			BaseURL:         baseURL,
			DefaultModel:    defaultModel,
			AvailableModels: append([]string{}, models...),
			Enabled:         true,
			Status:          "unknown",
			HealthStatus:    "unknown",
			Capabilities:    defaultProviderCapabilities(providerType),
			SupportsStream:  true,
			HasAPIKey:       true,
			MaskedKey:       &masked,
			CreatedAt:       timestamp,
			UpdatedAt:       timestamp,
		},
		APIKey: apiKey,
	}
	profileID := "profile_" + sanitizeID(strings.TrimPrefix(providerID, "provider_"))
	stubFallback := "profile_stub"
	s.profiles[profileID] = ModelProfile{
		ProfileID:         profileID,
		DisplayName:       displayName + " 默认配置",
		ProviderID:        providerID,
		Model:             defaultModel,
		Temperature:       0.4,
		MaxTokens:         4096,
		ContextWindow:     128000,
		ResponseFormat:    "text",
		ToolMode:          "auto",
		FallbackProfileID: &stubFallback,
		TimeoutMS:         120000,
		Purpose:           displayName + " 环境变量自动注册的默认配置",
		Enabled:           true,
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
	}
}

func envDefault(key string, fallbackValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallbackValue
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func prependUnique(first string, values []string) []string {
	result := []string{}
	seen := map[string]bool{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		result = append(result, value)
	}
	add(first)
	for _, value := range values {
		add(value)
	}
	return result
}
