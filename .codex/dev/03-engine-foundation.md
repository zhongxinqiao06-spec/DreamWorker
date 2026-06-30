# 03 Engine Foundation

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Engine |
| Priority | P0 |
| DependsOn | 02 |
| ExitGate | EventStore replay and architecture fitness tests pass |
| PR Range | PR-03-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

建立 Go Engine 基础层：domain / app / ports / runtime / adapters / platform，落地 EventStore、ArtifactStore、migration、projection、event replay 和核心 ports。

## 非目标

- 不实现完整 Incubator Runtime。
- 不实现真实模型调用。
- 不实现 MCP/A2A/Skill adapter。
- 不让 domain 依赖 adapters。

## 输入文档

- `.codex/plans/04-engine-code-skeleton.md`
- `.codex/plans/08-performance-observability.md`
- `.codex/dev/02-specs-contracts.md`

## 依赖阶段

依赖 `02-specs-contracts.md`。

## 核心产物

- Go Engine 目录结构。
- 核心 ports。
- SQLite EventStore + WAL。
- migration system skeleton。
- projection / event replay。
- Artifact metadata + file store。
- artifact versioning 预留。
- outbox / transaction consistency 预留。

## 工程任务

- 建立目录：
  - `domain`
  - `app`
  - `ports`
  - `runtime`
  - `adapters`
  - `platform`
  - `api`
- 定义 ports：
  - EventStore
  - ArtifactStore
  - CapabilityRegistry
  - CapabilityInvoker
  - PolicyEngine
  - ModelGateway
  - SecretStore
  - SearchIndex
  - Clock
  - IdGenerator
- 实现 SQLite EventStore：
  - append-only events。
  - WAL enabled。
  - event schema version stored。
  - trace_id indexed。
- 实现 migration table：
  - schema_migrations。
  - idempotent migration contract。
- 实现 projection：
  - Mission projection skeleton。
  - Run projection skeleton。
- 实现 ArtifactStore：
  - metadata table。
  - project-local file store。
  - version column 预留。
- api 层只做协议转换，不写业务逻辑。

Architecture fitness tests：

- `domain` cannot import `adapters`, `platform`, `api` or concrete SDK packages.
- `app` cannot import external SDK packages.
- `runtime` can depend only on `domain` and `ports`.
- `adapters` implement `ports` and do not leak adapter types into `domain`.
- `api` maps request/response only and does not mutate domain without app service.

Migration rollback gate：

- Every migration has backup strategy or is explicitly marked non-destructive.
- Migration runner records applied version and checksum.
- Failed migration leaves database in readable pre-migration state or restores backup.

Data retention policy:

- EventStore is append-only and retained for project lifetime.
- Artifact binary/file content may be versioned and pruned only by explicit user action.
- Diagnostics export redacts secrets and may omit large artifact bodies by default.
- Development logs can be rotated; audit events cannot be silently deleted in MVP.

## 数据结构 / 接口 / schema 影响

核心接口：

```go
type EventStore interface {
    Append(ctx context.Context, events []DomainEvent) error
    LoadMission(ctx context.Context, missionID string) ([]DomainEvent, error)
}

type ArtifactStore interface {
    Put(ctx context.Context, artifact ArtifactWrite) (ArtifactMeta, error)
    Get(ctx context.Context, artifactID string) (Artifact, error)
}
```

依赖规则：

- `domain` 不 import `adapters`。
- `app` 不直接调用外部 API。
- `runtime` 只依赖 `ports`。
- `adapters` 实现 `ports`。
- `api` 只做协议转换。

## 测试要求

- Go unit：
  - domain reducer。
  - event replay。
  - IdGenerator / Clock fake。
- Go integration：
  - SQLite EventStore append/load。
  - WAL enabled check。
  - migration idempotency。
  - ArtifactStore write/read。
- Contract tests：
  - DomainEvent matches event schema。

## 验收标准

- 可以 append event 并 replay Mission projection。
- SQLite WAL 开启。
- migration system 可运行且可重复执行。
- artifact metadata 和文件可写入项目目录。
- domain、app、runtime、adapters 依赖方向通过架构测试。
- Migration rollback gate documented and covered by smoke test.
- Data retention policy is documented for EventStore, artifacts and diagnostics.

## Codex PR 拆分建议

- PR-03-01: 建立 Go Engine 分层目录和架构测试。
- PR-03-02: 实现 DomainEvent、EventStore port 和 SQLite adapter。
- PR-03-03: 添加 WAL、migration table 和 migration runner skeleton。
- PR-03-04: 实现 Mission/Run projection replay。
- PR-03-05: 实现 ArtifactStore metadata + file store。
- PR-03-06: 添加 trace_id index 和 EventStore integration tests。
- PR-03-07: 添加 outbox/versioning TODO 接口和文档。

## 风险

- migration 太晚做会让 EventStore 难升级。
- ArtifactStore 和 EventStore 一致性需要后续 outbox 加固。
- 依赖规则不测试会被逐步破坏。

## 暂不做

- 不实现 Decision Gate。
- 不实现 Capability lifecycle。
- 不实现 Model Gateway。
