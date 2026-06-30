# DreamWorker Dev Plan

| Field | Value |
| --- | --- |
| Status | Active |
| Owner | Tech Lead |
| Priority | P0 |
| DependsOn | `.codex/plans/*` |
| ExitGate | Dev plan can drive independently verifiable PRs |
| PR Range | PR-README-* |
| Risk Level | Medium |
| Last Review | 2026-06-30 |

## 目标

把 `.codex/plans/` 的产品、架构、孵化器域模型、Capability Bus、UIUX、安全、性能、eval 和开放接入规格，转成 Codex 可以按 PR 分阶段执行的工程开发计划。

## 非目标

- 本目录不实现业务代码。
- 不替代 `.codex/plans/` 的高层规格。
- 不把 DreamWorker 降级为 workflow 工具、普通 Agent Builder 或聊天应用。

## 输入文档

- `.codex/plans/*.md`
- `.codex/rules/*.md`
- `.codex/skills/*.md`

## 依赖阶段

本 README 是开发执行入口，无前置阶段。

## 核心产物

- 13 份阶段开发计划。
- 每阶段可单独验收。
- 每阶段包含 Codex PR 拆分建议。

## 工程任务

- 按阶段顺序执行：bootstrap -> specs -> engine -> runtime -> capability -> model/agent -> UI -> E2E -> open access -> hardening -> release -> risk.
- 每个 PR 必须可运行测试或 smoke check。
- 所有实现必须遵守 Renderer/Main/Go Engine 边界。

PR 命名规范：

- 格式：`PR-<phase>-<number>: <verb> <bounded outcome>`。
- 示例：`PR-03-02: implement SQLite EventStore adapter`。
- 一个 PR 只能关闭一个清晰工程结果，不混合无关 UI、Engine、文档和测试改动。
- PR 描述必须引用输入文档、影响的 schema/API、测试命令和 rollback 方式。

分支规范：

- 默认前缀：`codex/`。
- 阶段分支：`codex/dev-03-engine-foundation`。
- PR 分支：`codex/pr-03-02-eventstore-adapter`。
- 不在同一分支混做不同阶段的 P0 能力。

Definition of Ready：

- 输入文档已列明。
- 依赖阶段已完成或有 mock/fake 替代。
- schema/API 影响已标出。
- 验收标准可测试。
- 风险和回滚方式已写入 PR 描述。

Definition of Done：

- 实现满足阶段验收标准。
- 对应测试或 smoke check 通过。
- high-risk action 没有绕过 PolicyEngine。
- 状态变化写入 EventStore。
- PR 可回滚，且不会破坏上一阶段 demo path。
- 文档、schema fixture、测试 fixture 与实现同步。

Review checklist：

- 架构边界：Renderer / Main / Go Engine 是否守边界。
- 契约：event、schema、manifest、error 是否 versioned。
- 安全：secret、文件系统、外部网络、执行代码是否受控。
- 可观测性：trace_id 是否贯穿新增 run/task/tool/approval/artifact。
- 测试：是否覆盖 unit、integration、contract、renderer、E2E 或 security smoke 中至少一种。
- 回滚：是否能禁用 feature flag、撤销 migration 或恢复 backup。

## 数据结构 / 接口 / schema 影响

本目录只定义计划，不直接定义运行时 schema；schema 影响在 `02-specs-contracts.md` 开始落地。

## 测试要求

- 文档结构测试：每个阶段文档必须包含统一 12 个章节。
- 计划一致性测试：阶段编号、依赖和 PR 建议不得互相冲突。
- 拼写测试：项目名统一写作 `DreamWorker`。

## 验收标准

- `dev/` 与 `plans/`、`rules/`、`skills/` 平级。
- 文件结构与路线图一致。
- 每份阶段文档可直接交给 Codex 拆 PR 执行。

## Codex PR 拆分建议

- PR-README-01: 建立 `dev/` 总入口和阶段索引。
- PR-README-02: 校验所有阶段文档都包含统一章节。
- PR-README-03: 把 README 的阶段入口与文件结构保持同步。
- PR-README-04: 增加 PR template、ADR、RFC、release checklist 和 phase exit review 模板。
- PR-README-05: 建立 traceability matrix 维护规则。

## 风险

- 开发计划过细可能显得重，但它能减少后续 PR 模糊度。
- 计划与 `plans/` 脱节会导致实现偏航。

## 暂不做

- 不创建应用代码。
- 不创建 CI。
- 不创建 SDK。
