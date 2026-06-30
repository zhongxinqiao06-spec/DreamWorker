# 10 Observability Eval Hardening

| Field | Value |
| --- | --- |
| Status | Planned |
| Owner | Platform Quality |
| Priority | P1 early, P0 before release |
| DependsOn | 08 |
| ExitGate | SLO, security smoke and golden tasks pass release gate |
| PR Range | PR-10-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

完成 MVP 的性能 SLO、可观测性、eval、diagnostics 和安全 hardening，确保系统可诊断、可回放、可评估、可发布。

## 非目标

- 不建设云端 observability 平台。
- 不做完整 OTLP 后端。
- 不做企业合规认证。
- 不牺牲脱敏换调试便利。

## 输入文档

- `.codex/plans/08-performance-observability.md`
- `.codex/plans/10-eval-quality-system.md`
- `.codex/plans/09-security-policy.md`

## 依赖阶段

依赖 `08-mvp-e2e-flow.md`。

## 核心产物

- SLO smoke。
- trace_id / run_id / task_id / tool_call_id / approval_id / artifact_id。
- pprof dev mode。
- OpenTelemetry basic traces。
- structured logs。
- diagnostics export。
- golden tasks。
- eval scoring。
- security hardening。

## 工程任务

SLO：

- cold start 到 shell 可见 p95 <= 3000ms。
- Go Engine ready p95 <= 1500ms。
- 打开最近项目 p95 <= 1000ms。
- EventStore append p95 <= 20ms。
- UI event render p95 <= 100ms。
- Run Timeline 支持 10000 events。
- Artifact search p95 <= 300ms。
- idle memory p95 <= 600MB。
- agent event 丢失率 0。
- 高风险动作误执行次数 0。

性能策略：

- event stream batching。
- backpressure。
- virtualized Run Timeline。
- SQLite WAL。
- append-only event store。
- artifact lazy loading。
- markdown lazy rendering。
- model call cancellation。
- task timeout。
- capability concurrency limit。
- structured logs。
- pprof dev-only。
- OpenTelemetry traces。
- diagnostics export。

Eval：

- golden tasks 至少 5 个。
- artifact score。
- evidence quality score。
- hallucination risk。
- actionability score。
- regression tests。
- human spot check。

Security hardening：

- secret redaction。
- sanitizer。
- approval smoke tests。
- policy smoke tests。
- capability revoke smoke。
- renderer boundary smoke。

SLO measurement scripts：

| SLO | Measurement |
| --- | --- |
| cold start | launch app and record time to shell visible |
| engine ready | Main starts engine and receives ping |
| EventStore append | benchmark append batch and p95 |
| UI event render | feed fixture events and measure render commit |
| Run Timeline events | render 10k event fixture |
| artifact search | query SQLite FTS fixture |
| idle memory | sample process memory after idle window |

Dashboard expectations：

- MVP local diagnostics page shows app version, engine version, active run, event lag, recent errors.
- Dev mode can expose pprof link and trace export path.
- Release mode hides pprof and only exposes redacted diagnostics export.

Incident diagnostics checklist：

- User can export redacted diagnostics.
- Diagnostics include trace IDs for recent failures.
- No secret values appear in diagnostics.
- Engine startup failure, migration failure and policy denial are distinguishable by error code.
- Support can replay event summary without artifact body by default.

## 数据结构 / 接口 / schema 影响

Trace context：

```json
{
  "trace_id": "tr_001",
  "run_id": "run_001",
  "task_id": "task_001",
  "tool_call_id": "call_001",
  "approval_id": "appr_001",
  "artifact_id": "art_001"
}
```

Eval report：

```yaml
artifact_score: 0.82
evidence_quality_score: 0.76
hallucination_risk: medium
actionability_score: 0.79
regression: false
```

## 测试要求

- Go integration：
  - EventStore append latency smoke。
  - diagnostics export。
- Renderer performance:
  - 10k event virtualized Run Timeline。
- E2E:
  - full MVP flow trace complete。
- Golden tasks:
  - 5 sample ideas stable outputs。
- Security smoke:
  - Renderer cannot access Node。
  - secret not present in event stream。
  - revoked capability cannot run。
  - high-risk action requires approval。
  - markdown sanitizer works。

## 验收标准

- SLO smoke 有脚本或明确手动步骤。
- 每个 run、task、tool call、approval、artifact 有 trace_id。
- pprof 只在 dev mode 暴露。
- diagnostics export 不含 secret。
- golden tasks 可重复运行。
- eval report 输出四项核心分数。
- 高风险动作误执行次数为 0。
- SLO measurement scripts or manual procedures are documented.
- Diagnostics checklist passes on a seeded failure run.

## Codex PR 拆分建议

- PR-10-01: 添加 trace context 贯穿事件和日志。
- PR-10-02: 添加 structured logs 和 diagnostics export。
- PR-10-03: 添加 pprof dev mode 和 OpenTelemetry basic traces。
- PR-10-04: 实现 event batching/backpressure smoke。
- PR-10-05: 添加 Run Timeline 10k events performance test。
- PR-10-06: 添加 golden tasks 和 eval report。
- PR-10-07: 添加 security smoke suite。
- PR-10-08: 添加 SLO release gate。

## 风险

- trace 写入敏感数据会泄露，需要 sanitizer。
- 过早微优化会影响 MVP 速度。
- LLM judge eval 有偏差，需要 human spot check。

## 暂不做

- 不接 hosted eval。
- 不接远程 telemetry 默认上传。
- 不做大规模 benchmark。
