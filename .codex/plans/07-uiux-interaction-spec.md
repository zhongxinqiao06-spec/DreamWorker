# 07 UIUX Interaction Spec

## 目标

定义 DreamWorker Incubator Workspace 的交互规格，让用户在一个工作台里完成 idea 输入、孵化阶段推进、证据查看、蓝图编辑、运行观察、交付物编辑、审批和下一步行动。

## 非目标

- 不做营销式 landing page。
- 不把聊天框作为唯一入口。
- 不隐藏 Agent 行为、证据、成本和风险。
- 不让用户只能旁观自动运行。

## 核心对象

- Incubator Workspace。
- Left Rail。
- Center Workspace。
- Right Inspector。
- Idea Chat。
- Incubator Board。
- Blueprint Canvas。
- Run Timeline。
- Artifact Studio。
- Evidence Drawer。
- Approval Diff Card。
- Cost/Risk Panel。
- Command-K。
- Steering Chips。
- Next Best Action。

## 数据结构示例

```yaml
workspace_state:
  selected_mission: msn_001
  active_stage: validate
  center_view: incubator_board
  right_inspector:
    mode: evidence_drawer
    target_id: hyp_001
  next_best_action:
    type: ask_user
    label: "确认是否继续验证独立开发者细分人群"
    risk: low
```

Approval Diff Card：

```yaml
approval_card:
  id: appr_001
  action: "create_github_issue_drafts"
  before: null
  after:
    count: 12
    repository: "user/project"
  data_shared: ["issue title", "issue body"]
  risk: medium
  cost_estimate: 0
  choices: ["approve", "reject", "edit", "ask_user"]
```

## 关键流程

1. 用户在 Idea Chat 输入想法。
2. Center Workspace 显示 Incubator Board，按六阶段展示进度。
3. 用户点击某个 Hypothesis，Right Inspector 打开 Evidence Drawer。
4. Run Timeline 展示 Agent、tool call、artifact 和 approval 事件。
5. Blueprint Canvas 支持编辑任务、依赖、Agent 和 capability。
6. Artifact Studio 编辑 PRD、roadmap、issues、copy。
7. Cost/Risk Panel 常驻展示预算、风险和高风险待审批项。
8. Command-K 支持快速创建任务、切换阶段、注册 capability、导出 artifact。
9. Steering Chips 提供“加深竞品分析”“降低预算”“先做落地页”等快捷 steering。
10. Next Best Action 明确下一步建议。

## MVP 做法

- Left Rail：Mission 列表、阶段导航、Artifact 入口、Capability 入口。
- Center Workspace：Idea Chat、Incubator Board、Blueprint Canvas、Run Timeline、Artifact Studio 多视图切换。
- Right Inspector：Evidence Drawer、Cost/Risk Panel、Approval Diff Card。
- Run Timeline 使用 virtualized list。
- UI 状态由 event stream reducer 重建。

## 后续扩展

- 多人协作 presence 和评论。
- Blueprint Canvas 图形化编辑。
- Evidence Graph 可视化。
- Artifact Studio 支持 diff、版本、批注和发布。
- Command-K 支持自定义命令和插件命令。

## 风险

- 视图过多会让 MVP 复杂；首版要用统一布局承载多状态。
- 如果 Evidence Drawer 难用，Evidence-first 会变成口号。
- Approval Diff Card 如果信息不足，用户无法安全决策。
- Cost/Risk Panel 如果太吵，会造成审批疲劳。
