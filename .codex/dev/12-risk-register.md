# 12 Risk Register

| Field | Value |
| --- | --- |
| Status | Active |
| Owner | Tech Lead |
| Priority | P0 |
| DependsOn | All phases |
| ExitGate | Critical/high risks map to tests, owners and release gates |
| PR Range | PR-12-* |
| Risk Level | Critical |
| Last Review | 2026-06-30 |

## 目标

建立工程风险登记表，覆盖架构、安全、性能、开放接入、质量和商业定位风险，并要求每阶段更新风险状态。

## 非目标

- 不把风险登记当成泛泛讨论。
- 不登记无法行动的抽象担忧。
- 不替代具体测试和发布门禁。

## 输入文档

- `.codex/plans/*.md`
- `.codex/dev/*.md`

## 依赖阶段

从 `00-development-roadmap.md` 开始，贯穿所有阶段。

## 核心产物

- 风险登记表。
- 风险 owner。
- 风险状态。
- 预警信号。
- 缓解措施。

## 工程任务

每个风险必须包含：

- 风险 ID。
- 风险描述。
- 影响范围。
- 可能性。
- 严重程度。
- 预警信号。
- 缓解措施。
- owner。
- 状态。

Risk review cadence：

- Every PR updates risk entries it affects.
- Every phase exit review checks critical/high open risks.
- MVP release cannot proceed with unowned critical risks.
- Risk status values: `open`, `mitigating`, `accepted`, `closed`.

Risk escalation levels：

| Level | Meaning | Required Action |
| --- | --- | --- |
| Critical | Can leak data, corrupt state or break release gate | release blocker |
| High | Can break MVP user flow or trust | owner mitigation before phase exit |
| Medium | Can cause rework or quality drop | tracked mitigation |
| Low | Localized issue | backlog acceptable |

Risk-to-test mapping (`risk-to-test mapping`):

| Risk Area | Required Test/Gate |
| --- | --- |
| Electron boundary | renderer boundary security smoke |
| EventStore migration | migration rollback smoke |
| Capability permission | policy + revoked capability smoke |
| Secret handling | no secret in renderer events/logs |
| Performance | SLO smoke |
| Artifact quality | golden tasks + eval report |
| Provider lock-in | ModelGateway adapter contract |

## 数据结构 / 接口 / schema 影响

```yaml
risk:
  id: RISK-001
  description: "Electron 安全边界失效"
  impact: "Renderer 可访问 Node 或 secrets"
  likelihood: medium
  severity: critical
  warning_signals:
    - "Renderer test can access process"
  mitigation:
    - "contextIsolation + sandbox + nodeIntegration=false"
    - "renderer boundary smoke in CI"
  owner: "desktop"
  status: open
```

## 测试要求

- 每个 critical/high 风险必须至少有一个测试、smoke 或 release gate。
- 每次阶段完成时更新风险状态。
- 风险 owner 缺失时不得进入 release。

## 验收标准

- 至少覆盖指定 14 个风险。
- 每个风险有 owner、状态、缓解措施和预警信号。
- Critical 风险有明确发布门禁。
- Risk review cadence and escalation levels are defined.
- Critical/high risks map to concrete tests or release gates.

## Codex PR 拆分建议

- PR-12-01: 建立风险登记表初版。
- PR-12-02: 将 critical/high 风险映射到测试或 release gate。
- PR-12-03: 在每个阶段完成后更新风险状态。
- PR-12-04: 发布前生成风险摘要。

## 风险

风险登记自身可能失效：如果没有 owner 和门禁，它会变成摆设。

## 暂不做

- 不接外部风险管理系统。
- 不做复杂风险评分自动化。

## 风险登记表

| ID | 风险描述 | 影响范围 | 可能性 | 严重程度 | 预警信号 | 缓解措施 | owner | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RISK-001 | Electron 安全边界失效 | Renderer 访问 Node、文件系统或 secret | medium | critical | boundary smoke 失败 | sandbox、contextIsolation、nodeIntegration=false、CI smoke | desktop | open |
| RISK-002 | EventStore schema 迁移困难 | 历史 Mission 无法 replay | medium | high | migration 测试失败 | versioned event、migration runner、fixtures | engine | open |
| RISK-003 | Agent 输出不稳定 | Artifact 质量波动 | high | high | golden task regression | schema normalization、Evaluator、golden tasks | agent | open |
| RISK-004 | UI event reducer 复杂化 | UI 状态错乱 | medium | high | reducer 测试难维护 | typed events、state machine、fixtures | desktop | open |
| RISK-005 | Capability 权限过宽 | 外部能力越权 | medium | critical | high-risk bypass | PolicyEngine gate、least privilege、approval smoke | security | open |
| RISK-006 | Secret 泄露 | 日志、event、Renderer 泄露 token | medium | critical | secret scan 命中 | SecretStore、redaction、no secret in renderer events | security | open |
| RISK-007 | 外部 MCP prompt injection | Agent 被工具描述或网页诱导 | high | high | untrusted MCP 输出异常指令 | trust level、approval、content isolation | capability | open |
| RISK-008 | 多 Agent 运行成本过高 | 用户成本失控、体验变慢 | high | medium | token/cost 超预算 | budget、timeout、cancellation、cost panel | agent | open |
| RISK-009 | 竞品覆盖核心能力 | 产品差异化下降 | medium | high | 通用平台复制 idea-to-blueprint | evidence-first、domain workflow、vertical templates | product | open |
| RISK-010 | UI 过早复杂化 | MVP 延迟 | high | medium | Canvas 占用过多周期 | placeholder、PR scope gate | desktop | open |
| RISK-011 | 开源接口频繁破坏兼容 | 第三方接入失败 | medium | medium | conformance fixtures 频繁变化 | versioned specs、compat policy | platform | open |
| RISK-012 | 性能劣化 | Run Timeline 卡顿、启动慢 | medium | high | SLO smoke 失败 | batching、backpressure、virtualized list、WAL | platform | open |
| RISK-013 | Artifact 质量不可控 | 用户不信任输出 | high | high | eval score 下降 | artifact score、human spot check、Evaluator | eval | open |
| RISK-014 | Model provider lock-in | 成本和可用性受单厂商影响 | medium | medium | provider outage blocks run | ModelGateway、profiles、fallback | model | open |
