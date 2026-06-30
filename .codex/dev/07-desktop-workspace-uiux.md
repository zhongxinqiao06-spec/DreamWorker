# 07 Desktop Workspace UIUX

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Desktop/UI |
| Priority | P0 |
| DependsOn | 02, 03 mock events |
| ExitGate | Incubator Workspace handles mission, run, approval, evidence and artifact states |
| PR Range | PR-07-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

实现 Incubator Workspace 的效率型桌面 UI：用户能创建 Mission、观察阶段、查看证据、编辑 Artifact、处理审批、steer 当前 run，并始终知道系统在做什么、为什么做、用了什么证据、花了多少成本、有什么风险、下一步是什么。

## 非目标

- 不做营销落地页。
- 不做传统 workflow 画布。
- 不把聊天窗口作为唯一产品形态。
- 不让 Renderer 做重计算或直接访问本地能力。

## 输入文档

- `.codex/plans/07-uiux-interaction-spec.md`
- `.codex/rules/ui-ux.md`
- `.codex/dev/02-specs-contracts.md`

## 依赖阶段

依赖 `01-repo-bootstrap.md` 和 `02-specs-contracts.md`。可用 mock event stream 与 Engine 并行开发。

## 核心产物

- Incubator Workspace。
- Top Bar / Left Rail / Center Workspace / Right Inspector。
- Idea Chat。
- Incubator Board。
- Blueprint Canvas placeholder。
- Run Timeline。
- Artifact Studio。
- Evidence Drawer。
- Approval Diff Card。
- Cost / Risk Panel。
- Command-K。
- Steering Chips。
- Next Best Action。
- UI state machine。

## 工程任务

信息架构：

```text
Top Bar: mission switcher, run status, cost/risk summary, command-k
Left Rail: missions, stages, artifacts, capabilities
Center Workspace: Idea Chat | Incubator Board | Blueprint Canvas | Run Timeline | Artifact Studio
Right Inspector: Evidence Drawer | Approval Diff Card | Cost/Risk Panel | Next Best Action
```

交互状态：

- empty
- loading
- streaming
- waiting_approval
- error
- recoverable_error
- completed
- paused
- cancelled

主要用户路径：

1. Empty state 输入 idea。
2. 创建 Mission。
3. Run Timeline 开始 streaming。
4. Incubator Board 展示 Discover / Validate / Shape。
5. Evidence Drawer 查看证据。
6. Approval Diff Card 处理高风险动作。
7. Artifact Studio 查看和编辑 artifact。
8. Stage Gate 做 continue / pivot / pause / kill / ask_user。

高级交互：

- steer 当前 run。
- approve / reject / edit / ask_user。
- 查看 evidence。
- 查看 artifact version。
- 查看 cost / risk。
- Command-K 执行全局命令。
- Steering Chips 快速改变方向。
- Next Best Action 提供下一步建议。

性能和安全：

- Run Timeline 使用 virtualized list。
- event stream batching。
- 大 artifact lazy load。
- Canvas lazy load。
- markdown lazy rendering。
- Renderer 不做重计算。
- Renderer 不直接读文件系统和 secret。
- Markdown 渲染必须 sanitizer。
- 外部链接必须安全打开。

UI 风格：

- 高级、克制、效率型。
- 参考 Linear / Arc / Notion / Raycast 的质感。
- 信息密度清晰，不做装饰性重 UI。

UI state transition table：

| From | Event | To |
| --- | --- | --- |
| empty | mission.created | loading |
| loading | run.started | streaming |
| streaming | approval.requested | waiting_approval |
| waiting_approval | approval.resolved | streaming |
| streaming | run.paused | paused |
| streaming | run.completed | completed |
| streaming | error.recoverable | recoverable_error |
| streaming | error.fatal | error |
| any | run.cancelled | cancelled |

Accessibility checks：

- Keyboard access for Command-K, stage gate, approval actions and artifact tabs.
- Focus trap for approval modal/card.
- Visible focus ring on interactive controls.
- Color is not sole indicator for risk/confidence.
- Run Timeline events have accessible labels.

Interaction QA checklist：

- Empty state explains local-first workspace and first action.
- Approval Diff Card shows before/after, data shared, risk, cost and reversibility.
- Evidence Drawer links evidence to Hypothesis, Artifact and Decision.
- Next Best Action is always visible when run is paused, failed or waiting_user.
- Cost/Risk Panel updates from events, never from local estimation only.

## 数据结构 / 接口 / schema 影响

UI state：

```ts
type InteractionState =
  | 'empty'
  | 'loading'
  | 'streaming'
  | 'waiting_approval'
  | 'error'
  | 'recoverable_error'
  | 'completed'
  | 'paused'
  | 'cancelled'
```

Renderer event reducer 输入只接受 versioned event envelope。

## 测试要求

- Renderer tests：
  - event reducer。
  - UI state。
  - approval card interaction。
  - artifact view。
  - Command-K。
  - Steering Chips。
- Security smoke：
  - Renderer cannot access Node。
  - Markdown sanitizer works。
  - external link safe open。
- Performance smoke：
  - Run Timeline 10k events virtualized render。
  - large artifact lazy load。

## 验收标准

- 用户可从 empty state 创建 Mission。
- UI 能展示 streaming、waiting_approval、paused、completed、error。
- Approval Diff Card 支持 approve / reject / edit / ask_user。
- Evidence Drawer 能按 Hypothesis 展示 evidence。
- Artifact Studio 能展示 artifact 和 version placeholder。
- Cost/Risk Panel 显示当前预算、风险和待审批项。
- Renderer 只通过 typed API + event stream 获取状态。
- UI state transition table is implemented in reducer tests.
- Accessibility checks pass manual QA for keyboard-only flow.
- Interaction QA checklist is covered by screenshots or component tests.

## Codex PR 拆分建议

- PR-07-01: 实现 Incubator Workspace 布局。
- PR-07-02: 实现 event reducer 和 UI state machine。
- PR-07-03: 实现 Idea Chat 和 empty/loading/streaming 状态。
- PR-07-04: 实现 Incubator Board 和 Stage Gate 控件。
- PR-07-05: 实现 Run Timeline virtualized list。
- PR-07-06: 实现 Evidence Drawer 和 Cost/Risk Panel。
- PR-07-07: 实现 Approval Diff Card。
- PR-07-08: 实现 Artifact Studio lazy loading。
- PR-07-09: 实现 Command-K、Steering Chips、Next Best Action。
- PR-07-10: 添加 renderer/security/performance smoke tests。

## 风险

- UI 视图多，首版必须复用布局和状态组件。
- Event reducer 复杂化会导致 UI 难调试。
- Approval 信息不足会造成危险确认。
- 过早追求 Canvas 完整能力会拖慢 MVP。

## 暂不做

- 不做完整图形化 Blueprint 编辑。
- 不做多人协作 presence。
- 不做 marketplace UI。
- 不做真实发布操作。
