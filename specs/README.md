# DreamWorker Specs

`specs/` 是 DreamWorker 跨进程、跨运行时契约的事实源。当前 schema 版本统一为 `0.1`，用于 Go Engine、Electron typed API、fixtures、contract tests 和后续 SDK。

## 当前覆盖

- `event.schema.json`：append-only domain event envelope。
- `error.schema.json`：用户可见、可恢复、可追踪的错误 envelope。
- `incubation.schema.json`：Mission / Stage / Hypothesis / Evidence / Experiment / Decision。
- `capability.schema.json`：Capability Manifest v1。
- `policy.schema.json`：PolicyRequest / PolicyDecision。
- `approval.schema.json`：ApprovalRequest / ApprovalResolution。
- `artifact.schema.json`：Artifact metadata。
- `blueprint.schema.json`：Blueprint 可执行计划。
- `agent.schema.json`：Agent 配置、runtimeConfig、planner、executor、memoryScope。
- `task.schema.json`：Agent/项目执行任务契约。

## 生成产物

`npm run specs:generate` 从 JSON Schema 生成：

- TypeScript contracts：供 Electron shared/renderer/preload/main 使用。
- Go contracts：供 Engine domain/runtime 使用。

生成产物不得手写；schema 变更必须同步 fixtures、generated contracts 和 tests。

## Versioning

- JSON Schema 使用 draft 2020-12。
- `$id` 固定为 `https://schemas.dreamworker.dev/<name>/v0.1/schema.json`。
- `schema_version` 使用 `major.minor`，当前统一为 `0.1`。
- additive change 升 minor；breaking change 升 major。
- 删除字段必须先在一个 minor 版本标记 deprecated，再在 major 版本移除。
- EventStore 保留原始 event；migration 只追加新版本视图，不覆盖原始 payload。

## Fixtures

每个 schema 至少有一组 valid 和 invalid fixture：

```text
specs/fixtures/valid/<name>.json
specs/fixtures/invalid/<name>.json
```

这些 fixtures 是 CI 的 contract smoke，新增字段必须更新对应样例。

## Validation

```powershell
npm run specs:validate
npm run specs:generate
npm run specs:check
```

`specs:check` 会先检查 generated contracts 是否最新，再跑 schema/fixture 校验。

## Ownership

| Schema               | Owner            | Reviewers        | Gate             |
| -------------------- | ---------------- | ---------------- | ---------------- |
| incubation           | Product + Engine | Engine, UI       | 孵化闭环变更     |
| event                | Platform         | Engine, UI, Eval | EventStore 变更  |
| capability           | Capability       | Security, Engine | Tool/MCP 变更    |
| policy / approval    | Security         | Capability, UI   | 高风险动作变更   |
| artifact / blueprint | Product + Engine | UI, Eval         | 产物生成变更     |
| agent / task         | Runtime          | Engine, UI       | Agent loop 变更  |
| error                | Platform         | UI, Diagnostics  | 用户错误展示变更 |
