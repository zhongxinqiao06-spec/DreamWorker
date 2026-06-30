# 02 Specs Contracts

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Platform/API |
| Priority | P0 |
| DependsOn | 01 |
| ExitGate | Versioned schemas generate Go/TS types and pass contract tests |
| PR Range | PR-02-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

建立 specs-first 开发方式，先定义 schema、event protocol、typed API、错误模型和类型生成策略，再进入 Engine 和 UI 实现。

## 非目标

- 不在本阶段实现完整业务流程。
- 不绑定单一代码生成工具到不可替换状态。
- 不让 UI 和 Engine 使用未版本化的 ad hoc payload。

## 输入文档

- `.codex/plans/01-incubator-domain.md`
- `.codex/plans/05-capability-bus.md`
- `.codex/plans/06-open-source-accessibility.md`
- `.codex/plans/09-security-policy.md`

## 依赖阶段

依赖 `01-repo-bootstrap.md`。

## 核心产物

- `specs/blueprint.schema.json`
- `specs/capability.schema.json`
- `specs/event.schema.json`
- `specs/policy.schema.json`
- `specs/incubation.schema.json`
- `specs/artifact.schema.json`
- `specs/approval.schema.json`
- `specs/error.schema.json`
- typed API contract。
- Go / TypeScript 类型生成策略。
- contract test skeleton。

## 工程任务

- 定义 schema version 策略：
  - JSON Schema `$id` 包含版本。
  - `schema_version` 使用 `major.minor`。
  - breaking change 升 major。
  - additive change 升 minor。
- 定义 event version 策略：
  - event envelope versioned。
  - payload 按 event type versioned。
  - EventStore 永远保留原始 event。
- 定义 typed API：
  - command request/response versioned。
  - event stream 使用统一 envelope。
- 定义 backward compatibility：
  - 新 Engine 必须能读取上一 minor 版本 event。
  - migration 只追加，不覆盖原始 event。
- 定义 migration 规则：
  - schema migration 有 `from_version`、`to_version`、`idempotent`。
  - artifact metadata migration 与 event migration 分离。

Schema ownership：

| Schema | Owner | Reviewers | Freeze Gate |
| --- | --- | --- | --- |
| incubation | Engine + Product | Engine, UI | before PR-04 |
| event | Platform | Engine, UI, Eval | before PR-03 |
| capability | Capability | Security, Engine | before PR-05 |
| policy/approval | Security | Capability, UI | before PR-05 |
| artifact/blueprint | Product + Engine | UI, Eval | before PR-08 |
| error | Platform | UI, Diagnostics | before PR-03 |

Breaking-change policy：

- Breaking changes require major version bump.
- Existing fixtures must remain readable by migration or compatibility adapter.
- Removing a field requires deprecation in one minor version before removal.
- Event payload breaking changes require replay compatibility test.

Contract freeze gate：

- Freeze v0.1 contracts before Engine Foundation starts writing persistent events.
- After freeze, schema changes require RFC.
- P0 contract changes must update fixtures, generated types and contract tests in the same PR.

## 数据结构 / 接口 / schema 影响

所有 event 必须包含：

```json
{
  "event_id": "evt_001",
  "schema_version": "1.0",
  "trace_id": "tr_001",
  "mission_id": "msn_001",
  "run_id": "run_001",
  "actor": "orchestrator",
  "timestamp": "2026-06-30T00:00:00Z",
  "type": "mission.created",
  "payload": {}
}
```

所有错误必须包含：

```json
{
  "code": "CAPABILITY_REVOKED",
  "message": "该能力已被撤销，不能继续调用。",
  "recoverable": true,
  "user_action": "选择其他能力或重新授权。",
  "trace_id": "tr_001"
}
```

## 测试要求

- Contract tests：
  - event schema validation。
  - capability manifest schema validation。
  - approval schema validation。
  - typed API request/response validation。
  - error schema validation。
- Go unit：schema loader。
- TS unit：generated types compile。
- Compatibility tests：上一 minor 示例 event 可读取。

## 验收标准

- specs 文件齐全。
- 至少有 1 个 valid 和 1 个 invalid fixture。
- Go 和 TypeScript 类型生成命令可运行。
- UI 和 Engine 之间只允许 typed API + event stream。
- 未版本化 event 不允许进入 EventStore。
- Contract freeze gate 明确记录。
- Breaking-change policy 有 fixtures 验证路径。

## Codex PR 拆分建议

- PR-02-01: 新增 specs 目录和 schema version policy。
- PR-02-02: 添加 event/error/incubation/artifact/approval schema。
- PR-02-03: 添加 capability/policy/blueprint schema。
- PR-02-04: 建立 Go/TS 类型生成脚本。
- PR-02-05: 添加 contract test fixtures 和 schema validation。
- PR-02-06: 将 preload/main/engine ping contract 改为 versioned response。

## 风险

- schema 过度设计会阻塞实现，MVP 只定义必要字段。
- 类型生成工具选型不稳定会造成后续返工。
- backward compatibility 规则不清晰会导致 EventStore 无法长期演进。

## 暂不做

- 不实现完整 migration runner。
- 不发布 SDK。
- 不定义 marketplace schema。
