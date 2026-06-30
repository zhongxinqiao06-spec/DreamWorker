# 08 Performance Observability

## 目标

定义桌面 MVP 的性能 SLO 和可观测性要求，保证 Incubator Workspace 在长 Run、多事件、多 artifact 场景下仍可用、可诊断、可回放。

## 非目标

- 不在 MVP 中承诺云端横向扩展。
- 不把所有 trace 上传到第三方服务。
- 不为首版实现复杂分布式追踪后端。
- 不牺牲安全脱敏来换取调试便利。

## 核心对象

- cold_start。
- engine_ready。
- event_write。
- UI_render。
- Run Console。
- artifact_search。
- idle_memory。
- trace_id。
- pprof。
- OpenTelemetry。
- SQLite WAL。
- batching。
- backpressure。
- virtualized list。

## 数据结构示例

```yaml
slo:
  cold_start_ms_p95: 3000
  engine_ready_ms_p95: 1500
  event_write_ms_p95: 20
  ui_render_ms_p95: 100
  run_console_append_ms_p95: 50
  artifact_search_ms_p95: 300
  idle_memory_mb_p95: 600

trace_context:
  trace_id: tr_01
  run_id: run_001
  task_id: task_003
  tool_call_id: call_009
```

## 关键流程

1. Main 启动 Go Engine，记录 cold_start 和 engine_ready。
2. 每个 command 创建或继承 trace_id。
3. Run、Task、ToolCall、PolicyDecision、ArtifactWrite 都带 trace_id。
4. EventStore 写入记录 latency 和 error。
5. Event Stream 对 UI 做 batching，拥塞时 backpressure。
6. Run Console 用 virtualized list 渲染事件。
7. Artifact search 使用 SQLite FTS。
8. Go Engine 暴露本地受控 pprof。
9. OpenTelemetry span 可导出到本地文件或开发模式 collector。

## MVP 做法

- SQLite 开启 WAL。
- EventStore append 批量写入。
- Event Stream 每 50-100ms batch UI event，审批和错误事件立即 flush。
- Run Console 只渲染可视区域。
- idle memory 做 smoke measurement。
- pprof 仅开发模式启用，默认绑定本地。
- trace_id 进入日志、事件和错误 detail。

## 后续扩展

- 引入本地 trace viewer。
- 增加 replay profiler。
- 支持 OpenTelemetry OTLP 导出。
- 支持性能回归 CI。
- 支持 artifact index 增量构建和向量检索。

## 风险

- 事件过细会拖慢 UI 和 SQLite。
- batching 太大导致 Run Console 不实时。
- pprof 暴露不当会产生本地安全风险。
- trace 中写入敏感数据会造成泄露，必须脱敏。
