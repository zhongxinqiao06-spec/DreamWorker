# 09 Open Source Accessibility

| Field | Value |
| --- | --- |
| Status | Planned |
| Owner | Platform Ecosystem |
| Priority | P2 for MVP, P1 for Alpha |
| DependsOn | 02, 05 |
| ExitGate | Minimal examples and conformance skeleton run locally |
| PR Range | PR-09-* |
| Risk Level | Medium |
| Last Review | 2026-06-30 |

## 目标

定义 DreamWorker 的开放接入工程计划。重点不是全量开源商业功能，而是开放 schema、capability manifest、event protocol、adapter examples、SDK、conformance tests 和 extension docs。

## 非目标

- MVP 不实现完整 SDK。
- MVP 不实现 marketplace。
- 不开放团队协作、云同步、企业策略等商业功能。
- 不允许第三方 adapter 绕过 CapabilityInvoker、PolicyEngine 和 EventStore。

## 输入文档

- `.codex/plans/06-open-source-accessibility.md`
- `.codex/plans/05-capability-bus.md`
- `.codex/dev/02-specs-contracts.md`
- `.codex/dev/05-capability-policy-runtime.md`

## 依赖阶段

依赖 `02-specs-contracts.md`。examples 依赖 `05-capability-policy-runtime.md`。

## 核心产物

- `specs/`
- SDK skeleton。
- examples。
- conformance tests。
- adapter kit docs。
- capability registry docs。
- trust badge docs。
- Open Core 策略。

## 工程任务

Open Core 策略：

开源 / 可开放：

- specs。
- SDK。
- examples。
- base adapters。
- conformance tests。
- basic engine interfaces。
- built-in demo skills。

商业化 / 后续：

- marketplace。
- team workspace。
- cloud sync。
- enterprise policy。
- advanced templates。
- hosted eval。
- managed agent runtime。

examples：

- hello-capability。
- hello-mcp-tool。
- hello-a2a-agent。
- hello-skill。
- hello-openapi-adapter。
- idea-to-blueprint-demo。

SDK：

- dreamworker-sdk-js。
- dreamworker-sdk-python。
- dreamworker-sdk-go。

MVP 做法：

- 只建目录和最小 example。
- SDK 只提供 manifest validation 和 stub invocation 示例。
- conformance tests 先跑本地 fixtures。

Open Core boundary table：

| Area | Open / Accessible | Commercial / Later |
| --- | --- | --- |
| Specs | schema, event protocol, manifest | hosted governance |
| SDK | base JS/Python/Go helpers | managed cloud SDK features |
| Adapters | demo/base adapters | certified marketplace adapters |
| Eval | local golden task harness | hosted eval dashboards |
| Runtime | basic interfaces | managed agent runtime |
| Workspace | local single-user | team workspace/cloud sync |

SDK compatibility policy：

- SDK major version follows spec major version.
- Minor SDK releases can add helpers but cannot break manifest validation.
- Examples must pin spec version.
- Conformance tests define compatibility, not README claims.

Adapter certification levels：

- `example`: documentation-only, not trusted.
- `conformant`: passes conformance tests.
- `verified`: reviewed manifest, signed release, scoped permissions.
- `trusted_builtin`: shipped with DreamWorker.

## 数据结构 / 接口 / schema 影响

Conformance cases：

```yaml
tests:
  - capability_manifest_validation
  - event_schema_validation
  - policy_approval_behavior
  - revoked_capability_cannot_run
  - artifact_write_stays_inside_project
  - no_secret_in_renderer_events
```

trust badge：

```yaml
trustBadge:
  level: community
  manifestValidated: true
  conformancePassed: true
  signed: false
```

## 测试要求

- Contract tests：
  - capability manifest validation。
  - event schema validation。
- Security conformance：
  - revoked capability cannot run。
  - artifact write stays inside project。
  - no secret in renderer events。
- Example smoke：
  - hello-capability loads。
  - hello-skill manifest validates。

## 验收标准

- `specs/` 可作为第三方接入入口。
- examples 目录存在并有 README。
- conformance tests 可以本地运行。
- Open Core 边界写清楚。
- trust level / trust badge 语义明确。
- Open Core boundary table is explicit.
- SDK compatibility policy and adapter certification levels are documented.

## Codex PR 拆分建议

- PR-09-01: 建立 open access 目录和 Open Core 文档。
- PR-09-02: 添加 hello-capability 和 hello-skill example。
- PR-09-03: 添加 hello-mcp-tool / hello-a2a-agent / hello-openapi-adapter skeleton。
- PR-09-04: 建立 conformance test runner skeleton。
- PR-09-05: 添加 capability/event/policy conformance tests。
- PR-09-06: 建立 SDK JS/Python/Go skeleton README。
- PR-09-07: 添加 trust badge metadata 草案。

## 风险

- 开放接口过早稳定会锁死错误抽象。
- examples 如果不能运行，会伤害开发者信任。
- Open Core 边界不清会造成用户预期混乱。

## 暂不做

- 不发布 npm/pip/go module。
- 不做 marketplace。
- 不做远程 registry。
