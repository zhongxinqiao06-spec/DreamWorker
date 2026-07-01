//go:build dreamworker_split_experiment

package projects

func createProjectModules(projectID string) map[string]ProjectModule {
	return map[string]ProjectModule{
		"explore": {
			ProjectID:       projectID,
			ModuleID:        "explore",
			DisplayName:     "探索模块",
			Status:          "ready",
			Summary:         "负责机会扫描、用户细分、竞品地图和证据收集。",
			DefaultAgents:   []string{"agent_opportunity_scout", "agent_competitor_analyst", "agent_customer_segment"},
			EnabledSkills:   []string{"skill_opportunity_scan", "skill_competitor_map"},
			EnabledTools:    []string{"tool_web_search_stub", "tool_model_generate_stub", "tool_artifact_write"},
			OutputArtifacts: []string{"dream_brief.md", "hypotheses.yaml", "research_pack.md"},
			NextBestAction:  "先跑机会雷达，再补用户画像、竞品地图和证据图谱。",
			Submodules: []ProjectSubmodule{
				moduleCard(projectID, "explore", "opportunity_radar", "机会雷达", "ready", "扫描用户痛点、市场窗口和可验证机会。", []string{"agent_opportunity_scout"}, []string{"skill_opportunity_scan"}, []string{"tool_web_search_stub", "tool_model_generate_stub"}, []string{"dream_brief.md", "hypotheses.yaml"}, "先生成机会清单，再挑选高置信假设。", "Discover"),
				moduleCard(projectID, "explore", "user_persona", "用户画像", "idle", "把目标用户、场景、付费动机和反对理由结构化。", []string{"agent_customer_segment"}, []string{"skill_opportunity_scan"}, []string{"tool_model_generate_stub"}, []string{"persona_map.md"}, "基于机会雷达结果补齐 ICP 和痛点证据。", "Discover"),
				moduleCard(projectID, "explore", "competitor_map", "竞品地图", "idle", "整理替代方案、差异化空间和进入壁垒。", []string{"agent_competitor_analyst"}, []string{"skill_competitor_map"}, []string{"tool_web_search_stub", "tool_artifact_write"}, []string{"competitor_map.md"}, "先确认竞品范围，再输出差异化判断。", "Validate"),
				moduleCard(projectID, "explore", "evidence_graph", "证据图谱", "idle", "把假设、证据、风险和下一步动作连成可审计图谱。", []string{"agent_evaluator"}, []string{"skill_opportunity_scan"}, []string{"tool_artifact_write"}, []string{"evidence_graph.yaml"}, "证据不足时返回 ask_user，不静默推进。", "Validate"),
			},
			Config: map[string]any{"stage": "Discover", "evidenceRequired": true},
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
			Submodules: []ProjectSubmodule{
				moduleCard(projectID, "product", "mvp_scope", "MVP 收敛", "idle", "把目标用户、核心场景和必须交付物压缩到首版范围。", []string{"agent_product_designer"}, []string{"skill_prd_draft"}, []string{"tool_model_generate_stub"}, []string{"mvp_scope.md"}, "先锁定不可省略的核心闭环。", "Shape"),
				moduleCard(projectID, "product", "prd_draft", "PRD 草案", "idle", "输出用户故事、功能边界、状态和验收条件。", []string{"agent_product_designer", "agent_evaluator"}, []string{"skill_prd_draft"}, []string{"tool_artifact_write"}, []string{"prd.md"}, "等待 MVP 范围确认后生成 PRD 草案。", "Shape"),
				moduleCard(projectID, "product", "prototype_notes", "原型说明", "idle", "描述关键界面、交互状态和用户路径。", []string{"agent_prototype_designer"}, []string{"skill_prd_draft"}, []string{"tool_model_generate_stub"}, []string{"prototype_notes.md"}, "先补齐核心路径，再进入视觉稿。", "Shape"),
				moduleCard(projectID, "product", "blueprint_canvas", "蓝图画布", "idle", "把产品对象、事件、能力和风险整理成工程蓝图输入。", []string{"agent_product_designer", "agent_evaluator"}, []string{"skill_blueprint"}, []string{"tool_artifact_write"}, []string{"blueprint.yaml"}, "PRD 稳定后同步 Blueprint Canvas。", "Shape"),
			},
			Config: map[string]any{"stage": "Shape", "requiresDecisionGate": true},
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
			Submodules: []ProjectSubmodule{
				moduleCard(projectID, "development", "architecture", "技术架构", "idle", "定义桌面、Engine、能力总线和数据边界。", []string{"agent_system_architect"}, []string{"skill_blueprint"}, []string{"tool_model_generate_stub"}, []string{"architecture.md"}, "先读取 Blueprint，再拆系统边界。", "Build"),
				moduleCard(projectID, "development", "tech_stack_cost", "技术栈与成本", "idle", "评估依赖、模型成本、运行成本和替代方案。", []string{"agent_tech_stack_advisor"}, []string{"skill_blueprint"}, []string{"tool_model_generate_stub"}, []string{"tech_stack.md", "cost_estimate.md"}, "等待架构约束明确后评估成本。", "Build"),
				moduleCard(projectID, "development", "pr_breakdown", "PR 拆分", "idle", "把蓝图拆成可独立验证、可回滚的 PR 序列。", []string{"agent_dev_orchestrator"}, []string{"skill_blueprint"}, []string{"tool_artifact_write"}, []string{"issue_plan.md"}, "先锁定验收门，再切 PR。", "Build"),
				moduleCard(projectID, "development", "test_gates", "测试门禁", "idle", "定义单测、契约测试、安全 smoke 和 E2E 验收。", []string{"agent_dev_orchestrator", "agent_evaluator"}, []string{"skill_blueprint"}, []string{"tool_artifact_write"}, []string{"test_plan.md"}, "PR 拆分完成后补测试矩阵。", "Build"),
			},
			Config: map[string]any{"stage": "Build", "writeCodeAutomatically": false},
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
			Submodules: []ProjectSubmodule{
				moduleCard(projectID, "sales", "positioning_copy", "定位文案", "idle", "把目标人群、痛点、承诺和差异化压成一句话。", []string{"agent_sales_strategist"}, []string{"skill_launch_plan"}, []string{"tool_model_generate_stub"}, []string{"positioning.md"}, "先确认 ICP，再写定位。", "Launch"),
				moduleCard(projectID, "sales", "landing_page", "落地页", "idle", "生成首屏、功能区、FAQ 和转化 CTA 文案。", []string{"agent_sales_strategist"}, []string{"skill_launch_plan"}, []string{"tool_artifact_write"}, []string{"landing_copy.md"}, "定位确认后生成落地页草案。", "Launch"),
				moduleCard(projectID, "sales", "launch_plan", "发布计划", "idle", "安排渠道、节奏、素材和审批点。", []string{"agent_sales_strategist", "agent_demo_designer"}, []string{"skill_launch_plan"}, []string{"tool_artifact_write"}, []string{"launch_checklist.md"}, "产品 demo 稳定后进入发布计划。", "Launch"),
				moduleCard(projectID, "sales", "feedback_loop", "反馈循环", "idle", "收集首批用户反馈，回写假设和下一轮迭代。", []string{"agent_evaluator"}, []string{"skill_launch_plan"}, []string{"tool_human_input", "tool_artifact_write"}, []string{"feedback_log.md"}, "发布后把反馈转成下一轮验证任务。", "Learn"),
			},
			Config: map[string]any{"stage": "Launch", "publishRequiresApproval": true},
		},
	}
}

func moduleCard(
	projectID string,
	moduleID string,
	submoduleID string,
	displayName string,
	status string,
	summary string,
	agents []string,
	skills []string,
	tools []string,
	artifacts []string,
	nextBestAction string,
	stage string,
) ProjectSubmodule {
	return ProjectSubmodule{
		ProjectID:       projectID,
		ModuleID:        moduleID,
		SubmoduleID:     submoduleID,
		DisplayName:     displayName,
		Status:          status,
		Summary:         summary,
		DefaultAgents:   agents,
		EnabledSkills:   skills,
		EnabledTools:    tools,
		OutputArtifacts: artifacts,
		NextBestAction:  nextBestAction,
		Config:          map[string]any{"stage": stage},
	}
}
