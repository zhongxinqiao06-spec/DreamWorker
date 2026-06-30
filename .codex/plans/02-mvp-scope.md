# 02 MVP Scope

## 目标

定义 DreamWorker MVP 的可交付边界：完成单用户本地项目孵化闭环，从 Mission 创建到 Discover、Validate、Shape 阶段输出可信 Blueprint 和首批 Artifact，并保留 Build、Launch、Learn 的计划型输出。

## 非目标

- 不实现团队协作和权限体系。
- 不做 marketplace。
- 不做任意外部 Agent 的自动执行。
- 不做云端多租户。
- 不做真实社媒发布、邮件发送或付费 API 的无审批调用。
- 不实现完整代码生成和部署流水线。

## 核心对象

- Mission：MVP 的根工作单元。
- Stage：MVP 主跑 Discover、Validate、Shape。
- Hypothesis：由 Agent 从 idea 中抽取。
- Evidence：来自 web research、用户输入、文件和只读浏览。
- Experiment：主要是访谈问题、落地页测试、竞品验证和技术 Spike 计划。
- Decision：每阶段必须产生。
- Blueprint：MVP 主要交付物。
- Run：一次孵化执行。
- Artifact：PRD、roadmap、issues、landing copy、launch checklist。

## 数据结构示例

```yaml
mvp_run:
  mission_id: msn_001
  enabled_stages: [discover, validate, shape]
  planned_artifacts:
    - dream_brief.md
    - research_pack.md
    - blueprint.yaml
    - prd.md
    - github_issues.json
    - landing_page_copy.md
    - launch_checklist.md
  capability_allowlist:
    - web_search
    - browser_readonly
    - artifact_write
    - github_issues_draft
```

## 关键流程

1. 用户输入 idea，系统生成 Mission 和 Dream Brief 草稿。
2. Product Analyst 抽取目标用户、场景、价值假设。
3. Competitor Analyst 收集竞品和替代方案 Evidence。
4. Tech Architect 生成技术可行性、成本和模块拆解。
5. Growth Agent 生成发布、渠道和冷启动建议。
6. Evaluator 评分 Evidence 和 Artifact，生成阶段 Decision。
7. Orchestrator 汇总为 Blueprint，并生成导出 Artifact。

## MVP 做法

- UI 提供 Idea Chat、Incubator Board、Blueprint Canvas、Run Timeline、Artifact Studio、Evidence Drawer。
- Go Engine 提供本地 EventStore、PolicyEngine、CapabilityInvoker、ModelGateway、ArtifactStore。
- Capability 首批内置 web_search、browser_readonly、artifact_read/write、filesystem_project_read、github_issues_draft。
- 高风险能力默认生成 Approval Diff Card，不直接执行。
- 所有 run、task、tool call 带 trace_id。

## 后续扩展

- Build 阶段接入代码 Agent 和仓库写入。
- Launch 阶段接入邮件、社媒、Product Hunt、广告和 analytics。
- Learn 阶段接入用户反馈、指标追踪和 roadmap 更新。
- 支持 Mission 模板和行业包。

## 风险

- MVP 如果覆盖六阶段全执行，会失去焦点。
- 竞品和市场 Evidence 可能需要付费数据源，首版要明确置信度边界。
- Artifact 多但不可执行会削弱价值；Blueprint 必须能转任务。
- 过多审批会让体验变慢，需要 risk-based approval。
