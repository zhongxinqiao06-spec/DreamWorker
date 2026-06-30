# 04 Engine Code Skeleton

## 目标

定义 Go Engine 技术骨架，采用 domain / app / ports / adapters / runtime 分层。domain 不依赖 adapters，orchestrator 只依赖 ports，所有外部能力通过 CapabilityInvoker，所有状态变化写入 EventStore，所有高风险调用经过 PolicyEngine。

## 非目标

- 不在 domain 层引入 SQLite、HTTP、MCP、A2A、OpenAI SDK 或 Electron 类型。
- 不让 orchestrator 直接调用具体 adapter。
- 不允许工具调用绕过 CapabilityInvoker。
- 不允许状态变化只存在内存中。
- 不在 MVP 中实现完整插件市场。

## 核心对象

- domain：Mission、Stage、Hypothesis、Evidence、Experiment、Decision、Blueprint、Run、Artifact。
- app：Orchestrator、StageService、DecisionGate、EvidenceService、BlueprintCompiler。
- ports：EventStore、CapabilityInvoker、PolicyEngine、ArtifactRepository、ModelGateway、Clock、IdGenerator。
- adapters：SQLiteEventStore、MCPAdapter、A2AAdapter、SkillAdapter、OpenAICompatibleModel、FileArtifactStore。
- runtime：scheduler、eventbus、approval、retry、trace、worker pool。

## 数据结构示例

```text
engine/
  cmd/dreamworker-engine/
  internal/
    domain/
      mission.go
      stage.go
      evidence.go
      decision.go
      blueprint.go
      run.go
      artifact.go
    app/
      orchestrator/
      incubator/
      decisiongate/
      blueprint/
    ports/
      event_store.go
      capability_invoker.go
      policy_engine.go
      artifact_repository.go
      model_gateway.go
    adapters/
      sqlite/
      mcp/
      a2a/
      skills/
      models/
      filesystem/
    runtime/
      scheduler/
      events/
      approval/
      retry/
      tracing/
```

接口示例：

```go
type CapabilityInvoker interface {
    Invoke(ctx context.Context, req CapabilityInvocation) (CapabilityResult, error)
}

type EventStore interface {
    Append(ctx context.Context, events []DomainEvent) error
    LoadMission(ctx context.Context, missionID string) ([]DomainEvent, error)
}

type PolicyEngine interface {
    Evaluate(ctx context.Context, req PolicyRequest) (PolicyDecision, error)
}
```

## 关键流程

1. Command handler 接收用户意图。
2. app service 加载 Mission 事件并重建状态。
3. Orchestrator 生成下一批 Task。
4. Task 执行前调用 PolicyEngine。
5. 通过 CapabilityInvoker 调外部能力。
6. 结果转成 DomainEvent 写入 EventStore。
7. runtime eventbus 推送 UI event。
8. ArtifactRepository 写入交付物 metadata 和文件。

## MVP 做法

- 先实现内存 domain projection + SQLite append-only EventStore。
- CapabilityInvoker 先支持 builtin、model、artifact、web_search mock/adapter。
- PolicyEngine 先实现规则表和 risk-based approval。
- runtime 先支持单 Mission、单 Run、顺序 + 少量并行 task。
- trace_id 在 command 入口创建并透传到 run、task、tool call。

## 后续扩展

- 增加 outbox pattern 保证事件与 artifact 写入一致性。
- 增加 plugin sidecar runner 和 WASM sandbox。
- 增加 distributed engine 和 remote worker。
- 增加 event replay、time travel debug 和 deterministic eval。

## 风险

- 分层过度会拖慢 MVP，接口数量要克制。
- 如果 EventStore schema 过早固化，后续迁移成本高。
- CapabilityInvoker 抽象过薄会泄漏协议细节。
- PolicyEngine 如果只是 UI 判断，安全边界会失效。
