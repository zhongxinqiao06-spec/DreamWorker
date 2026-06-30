# 00 Development Roadmap

| Field | Value |
| --- | --- |
| Status | Active |
| Owner | Tech Lead |
| Priority | P0 |
| DependsOn | Product and architecture plans |
| ExitGate | MVP release gates are explicit and traceable to PRs/tests |
| PR Range | PR-00-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

定义 DreamWorker 从文档到 MVP、Alpha、Beta 的工程路线图，并先补齐当前 dev 计划缺口。DreamWorker 的目标是开放式 AI 项目孵化器操作系统：Idea -> Mission -> Hypothesis -> Evidence -> Experiment -> Decision Gate -> Blueprint -> Multi-Agent Run -> Artifact -> Launch -> Feedback -> Next Iteration。

## 非目标

- 不做普通 workflow 工具。
- 不做普通 Agent Builder。
- 不做只围绕聊天窗口的应用。
- 不在 MVP 里实现完整 marketplace、团队协作或云端多租户。

## 输入文档

- `.codex/plans/00-product-positioning.md`
- `.codex/plans/01-incubator-domain.md`
- `.codex/plans/02-mvp-scope.md`
- `.codex/plans/03-architecture-blueprint.md`
- `.codex/plans/11-roadmap.md`

## 依赖阶段

无前置阶段。本文件是所有开发阶段的总入口。

## 核心产物

- 当前计划缺口登记。
- MVP / Alpha / Beta 目标。
- 阶段依赖图。
- 并行开发策略。
- P0 / P1 / P2 优先级。
- 统一验收门和发布门禁。

## 工程任务

当前 dev 计划缺口必须补齐：

- specs-first 开发阶段：新增 `02-specs-contracts.md`，先定义 JSON Schema、typed API、event protocol。
- typed contract / schema generation：明确 Go / TypeScript 类型生成策略。
- API/event 版本策略：所有 schema、event、manifest versioned。
- PR 粒度拆分：每份 dev 文档必须给出可单独验证的 PR。
- 测试金字塔：Go unit、Go integration、contract、renderer、E2E、golden tasks、security smoke。
- CI/CD 和 release packaging：新增 `11-release-packaging.md`。
- 开源 SDK / examples / conformance tests：新增 `09-open-source-accessibility.md`。
- UI 状态机和交互状态定义：升级 `07-desktop-workspace-uiux.md`。
- 性能 SLO 测量方式：升级 `10-observability-eval-hardening.md`。
- 错误码、日志、trace、diagnostics：在 specs、engine、observability 阶段落地。
- data migration / artifact versioning：在 Engine Foundation 和 Release 阶段预留。
- model gateway / schema normalization：新增 `06-model-agent-runtime.md`。
- 安全 threat model 工程落地：在 Capability Policy、Security Smoke、Risk Register 中落地。
- MVP end-to-end demo seed data：在 `08-mvp-e2e-flow.md` 落地。

阶段依赖图：

```text
01 repo bootstrap
  -> 02 specs/contracts
  -> 03 engine foundation
  -> 04 incubator domain runtime
  -> 05 capability policy runtime
  -> 06 model agent runtime
  -> 07 desktop workspace UIUX
  -> 08 MVP E2E flow
  -> 10 observability eval hardening
  -> 11 release packaging

09 open source accessibility starts after 02 and matures after 05.
12 risk register starts immediately and is updated every phase.
```

并行开发策略：

- `01` 和 `02` 可以小范围并行，但 Engine API 必须等 `02` contract 草案稳定。
- `03` 和 `07` 可以并行，UI 使用 mock event stream。
- `05` 和 `06` 可以并行，Agent runtime 只能依赖 CapabilityInvoker port。
- `09` 在 MVP 中只做 specs、examples、conformance skeleton。
- `10` 从 Phase 1 就接入 trace_id，最终作为发布门禁。

优先级：

- P0：repo bootstrap、specs/contracts、EventStore、domain runtime、Capability + Policy、MVP E2E、security smoke。
- P1：UI 完整体验、Model Gateway、golden tasks、diagnostics export、release packaging。
- P2：SDK、examples、conformance 扩展、A2A、Skill sandbox、marketplace 预留。

MVP 完成定义：

- 用户可创建 Mission。
- Discover / Validate / Shape 三阶段可跑通。
- 每阶段有 Hypothesis、Evidence、Decision。
- 可以生成 Dream Brief、Research Pack、Blueprint、PRD、Launch Checklist。
- Run Timeline 可观察 agent、task、tool call、approval、artifact event。
- Capability Manifest、lifecycle、trust level 有最小实现。
- PolicyEngine 可以阻止或审批高风险 capability。
- trace_id 贯穿 run、task、tool call、approval、artifact。
- 至少 5 个 golden tasks 可重复运行。

发布门禁：

- typecheck、go test、contract tests、renderer tests、E2E smoke、security smoke 全通过。
- high-risk action 不得绕过 Approval。
- secret 不得出现在 renderer event、日志、artifact metadata。
- revoked capability 不能运行。
- EventStore replay 能恢复 Mission 状态。

里程碑矩阵：

| Milestone | 目标 | P0 ExitGate | 主要阶段 |
| --- | --- | --- | --- |
| M0 Planning | dev/specs 可执行 | 每阶段有 PR 和测试映射 | 00 |
| M1 Bootstrap | 桌面和 Engine 可启动 | `runtime.ping` from Renderer via Main | 01 |
| M2 Contracts | contract-first 基础 | versioned schema + generated types | 02 |
| M3 Engine Core | 状态可回放 | EventStore + ArtifactStore + replay | 03 |
| M4 Incubator Core | 三阶段可跑 | Discover/Validate/Shape + Decision Gate | 04 |
| M5 Safe Capability | 外部能力受控 | policy/approval/revoke smoke pass | 05 |
| M6 Agent MVP | Agent 输出可控 | normalized outputs + evaluator | 06 |
| M7 Desktop MVP | 工作台可用 | UI state machine + Run Timeline | 07 |
| M8 E2E Demo | MVP 闭环 | seed idea produces required artifacts | 08 |
| M9 Hardening | 可发布 | SLO/security/eval gates pass | 10/11 |

关键路径：

- `01 -> 02 -> 03 -> 04 -> 08` 是 MVP 骨架关键路径。
- `05 -> 06 -> 08` 是 Agent 受控执行关键路径。
- `07 -> 08` 是用户体验关键路径。
- `10 -> 11` 是发布关键路径。

并行开发边界：

- UI 可用 mock event stream 并行，但不能自定义未登记 event shape。
- Model/Agent 可用 stub model 并行，但所有调用必须通过 ModelGateway 和 CapabilityInvoker。
- Open access 可先写 examples 和 conformance skeleton，但不得承诺 stable SDK。

技术债登记方式：

- 每个技术债必须写入 `12-risk-register.md` 或阶段文档的风险区。
- 技术债字段：`debt_id`、`introduced_by_pr`、`reason`、`payback_phase`、`owner`、`release_blocking`。
- P0 技术债不能跨 MVP 发布；P1 技术债必须有 payback phase；P2 可进入 backlog。

traceability matrix：

| Source | Dev Phase | PR | Tests | Release Gate |
| --- | --- | --- | --- | --- |
| `plans/01-incubator-domain.md` | 04 | PR-04-* | domain reducer, event replay | Mission stages replay |
| `plans/05-capability-bus.md` | 05 | PR-05-* | capability lifecycle, policy smoke | high-risk approval |
| `plans/07-uiux-interaction-spec.md` | 07 | PR-07-* | renderer state, interaction QA | UI E2E smoke |
| `plans/08-performance-observability.md` | 10 | PR-10-* | SLO smoke | SLO release gate |
| `plans/10-eval-quality-system.md` | 10 | PR-10-06 | golden tasks | eval gate |

## 数据结构 / 接口 / schema 影响

本阶段不新增运行时 schema，但要求后续所有 schema、event、manifest、error 均有版本字段，并在 `02-specs-contracts.md` 定义。

## 测试要求

- 文档一致性检查：确认 dev 文件名和路线图一致。
- 计划覆盖检查：每个用户要求都有对应阶段。
- 发布门禁 checklist 要能映射到测试金字塔。

## 验收标准

- 当前 dev 计划缺口已明确登记。
- 路线图包含 MVP、Alpha、Beta。
- 阶段依赖、并行策略、优先级、MVP 完成定义、发布门禁完整。
- 每个 P0 阶段都有明确 ExitGate。
- 每条 P0 能力都能追踪到 PR、测试和 release gate。

## Codex PR 拆分建议

- PR-00-01: 补齐 dev 路线图和缺口登记。
- PR-00-02: 建立阶段依赖图和 P0/P1/P2 优先级。
- PR-00-03: 建立统一验收门和发布门禁 checklist。
- PR-00-04: 将 roadmap 与 README 阶段索引同步。

## 风险

- 阶段太多可能导致执行者迷路，需要 README 做强入口。
- 如果 P0/P1/P2 不严格执行，MVP 会滑向平台化。
- 缺口登记不持续更新，会变成一次性文档。

## 暂不做

- 不创建代码。
- 不创建 CI。
- 不实现 schema。
