# 00 Product Positioning

## 目标

把 DreamWorker 明确定义为开放式项目孵化器操作系统，而不是 Agent 编排工具、工作流画布或一次性报告生成器。产品主线是从 Mission 出发，持续推进 Discover、Validate、Shape、Build、Launch、Learn 六个孵化阶段，并在每个阶段通过 Evidence Graph 和 Decision Gate 控制下一步行动。

## 非目标

- 不做通用个人 Agent OS。
- 不把首版做成可视化 DAG workflow builder。
- 不追求第一版连接所有 MCP、A2A、Skill 和 SaaS。
- 不把多 Agent 数量当作产品卖点。
- 不把输出停留在静态商业计划书。

## 核心对象

- Mission：用户想推进的项目使命。
- Stage：孵化阶段，固定为 Discover、Validate、Shape、Build、Launch、Learn。
- Hypothesis：需要验证的假设。
- Evidence：支持或反驳假设的证据。
- Experiment：验证假设的实验。
- Decision：阶段门决策。
- Blueprint：可执行项目蓝图。
- Run：一次 Agent/工具/人工协作运行。
- Artifact：阶段或运行产生的交付物。

## 数据结构示例

```yaml
mission:
  id: msn_ai_launch_001
  title: "面向独立开发者的 AI 营销素材生成工具"
  audience: ["indie hacker", "AI startup founder"]
  success_metric:
    activation: "10 个目标用户愿意留下邮箱或试用"
    delivery: "生成 MVP blueprint 和发布计划"
  current_stage: discover
  principles:
    - evidence_first
    - blueprint_first
    - human_in_command
```

## 关键流程

1. 用户输入 idea，系统创建 Mission。
2. Orchestrator 生成 Discover 阶段任务和初始 Hypothesis。
3. Agent 和 capability 收集 Evidence，生成 Artifact。
4. Evaluator 为结论标注 confidence、risk、next_best_action。
5. 阶段结束进入 Decision Gate，输出 continue、pivot、pause、kill 或 ask_user。
6. Decision 写入 EventStore，并驱动下一阶段 Blueprint 或 Run。

## MVP 做法

- 聚焦单用户本地桌面工作台。
- 首版只覆盖 Discover、Validate、Shape 的闭环，Build、Launch、Learn 输出计划和 artifact，不做全自动执行。
- 内置 Product Analyst、Competitor Analyst、Tech Architect、Growth Agent、Evaluator。
- 每个结论必须绑定 evidence、confidence、risk、next_best_action。
- 每个阶段必须生成 Decision。

## 后续扩展

- 支持团队 Mission 和多人审批。
- 支持公共 Blueprint 模板库。
- 支持 DreamWorker as MCP/A2A Server，让外部平台推进 Mission。
- 支持跨项目 memory 和行业知识包。

## 风险

- 定位过大导致 MVP 发散。
- Evidence-first 如果做得粗糙，会退化成普通 ChatGPT 报告。
- Decision Gate 过重会降低速度；过轻又无法建立信任。
- 如果用户不能编辑 Mission 和 Blueprint，会失去 Human-in-command 差异化。
