# 04 Incubator Domain Runtime

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Incubator Engine |
| Priority | P0 |
| DependsOn | 03 |
| ExitGate | Discover/Validate/Shape domain flow replays with Decision Gate |
| PR Range | PR-04-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

实现孵化器核心 Runtime：Mission、Stage、Hypothesis、Evidence、Experiment、Decision、Blueprint、Run、Artifact，以及 Discover / Validate / Shape / Build / Launch / Learn 六阶段和 Decision Gate。

## 非目标

- MVP 不自动执行 Build / Launch / Learn。
- 不引入图数据库。
- 不做多人协作阶段审批。
- 不允许证据不足时静默 continue。

## 输入文档

- `.codex/plans/01-incubator-domain.md`
- `.codex/plans/02-mvp-scope.md`
- `.codex/dev/03-engine-foundation.md`

## 依赖阶段

依赖 `03-engine-foundation.md`。

## 核心产物

- Incubator domain reducer。
- Stage runtime。
- Evidence Graph 最小实现。
- Decision Gate。
- Build / Launch / Learn placeholder。
- EventStore-backed state transitions。

## 工程任务

- 定义六阶段：
  - Discover
  - Validate
  - Shape
  - Build
  - Launch
  - Learn
- MVP 自动跑：
  - Discover
  - Validate
  - Shape
- Build / Launch / Learn：
  - 只建结构。
  - 输出 placeholder next actions。
- Decision Gate 支持：
  - continue
  - pivot
  - pause
  - kill
  - ask_user
- Decision 必须包含：
  - decision_type
  - confidence
  - reason
  - evidence_refs
  - risks
  - next_best_action
- Evidence 支持绑定：
  - Hypothesis
  - Artifact
  - Run
  - CapabilityInvocation
  - Decision

Domain invariant table：

| Invariant | Enforcement | Test |
| --- | --- | --- |
| Mission must have current stage | domain reducer | create mission replay |
| Stage transition must append event | app service | stage transition integration |
| Decision requires evidence_refs or ask_user reason | DecisionGate | decision gate unit |
| Evidence must reference at least one target | domain validation | evidence binding unit |
| Build/Launch/Learn cannot auto-run in MVP | runtime guard | placeholder guard test |
| Stage cannot silently continue with low confidence | DecisionGate | insufficient evidence test |

Decision Gate failure modes：

- Missing evidence -> `ask_user` or `pause`。
- Conflicting evidence -> `ask_user` with comparison summary。
- Low confidence + high risk -> `pause`。
- Invalid stage transition -> domain error with recoverable user_action。
- User-requested pivot -> `pivot` with new Hypothesis seed。
- User-requested stop -> `kill` with audit event。

## 数据结构 / 接口 / schema 影响

Decision 示例：

```json
{
  "decision_type": "ask_user",
  "confidence": 0.58,
  "reason": "核心付费意愿证据不足。",
  "evidence_refs": ["ev_001", "ev_002"],
  "risks": ["market_uncertainty"],
  "next_best_action": "继续访谈 5 个独立开发者。"
}
```

Stage transition event：

```json
{
  "type": "stage.decision_recorded",
  "payload": {
    "stage": "validate",
    "decision": {}
  }
}
```

## 测试要求

- Go unit：
  - domain reducer。
  - decision gate。
  - evidence binding。
  - stage transition。
- Go integration：
  - stage events append/replay。
- Contract tests：
  - incubation schema。
  - event schema。
- Golden fixture：
  - one sample Mission through Discover/Validate/Shape.

## 验收标准

- 用户可创建 Mission。
- Discover / Validate / Shape 可按事件推进。
- 每阶段有 Hypothesis、Evidence、Decision。
- 证据不足时返回 ask_user 或 pause。
- 所有阶段变化写入 EventStore。
- Build / Launch / Learn placeholder 可显示但不自动执行。
- Domain invariants have unit tests.
- Decision Gate failure modes are represented in fixtures.

## Codex PR 拆分建议

- PR-04-01: 定义 Mission/Stage/Hypothesis/Evidence/Experiment/Decision/Blueprint/Run/Artifact domain。
- PR-04-02: 实现 Incubator reducer 和 EventStore replay。
- PR-04-03: 实现 Stage runtime 和六阶段枚举。
- PR-04-04: 实现 Evidence Graph 最小绑定。
- PR-04-05: 实现 Decision Gate。
- PR-04-06: 添加 Build/Launch/Learn placeholder。
- PR-04-07: 添加 Discover/Validate/Shape integration fixture。

## 风险

- 阶段模型过硬会变成固定 workflow。
- Evidence Graph 初版如果不够简单，会拖慢 MVP。
- Decision Gate 如果没有测试，容易被 Agent 输出绕过。

## 暂不做

- 不做真实 Build 代码写入。
- 不做真实 Launch 发布。
- 不做 Learn 指标接入。
