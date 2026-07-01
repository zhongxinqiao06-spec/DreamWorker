# 06 Model Agent Runtime

## Stream ModelGateway Addendum

本阶段 ModelGateway 必须支持真实流式闭环：

- `ModelGateway.Stream(ctx, request)` 是主路径，`Generate` 只能作为聚合兼容层。
- Provider registry 按 `ModelProfile.providerId -> ProviderType` 选择 adapter。
- OpenAI 使用 Responses API streaming；OpenAI Compatible 系列使用 Chat Completions SSE；Anthropic 使用 Messages streaming；Ollama 使用 `/api/chat` JSON line stream。
- Engine 对外只输出 DreamWorker normalized stream event，不透传 provider 原始事件。
- `model.requested`、`model.completed`、`model.failed` 写审计摘要；token delta 只走 transient UI stream。
- 无 adapter、Provider disabled、缺 key、模型不存在时返回明确错误，不回退 stub 假装成功。

| Field       | Value                                                                  |
| ----------- | ---------------------------------------------------------------------- |
| Status      | Ready for Implementation                                               |
| Owner       | Agent Runtime                                                          |
| Priority    | P0                                                                     |
| DependsOn   | 04, 05                                                                 |
| ExitGate    | Orchestrator-mediated run creates normalized artifacts and eval report |
| PR Range    | PR-06-*                                                                |
| Risk Level  | High                                                                   |
| Last Review | 2026-06-30                                                             |

## 目标

实现 Model Gateway、内置 Agent spec、Task 模型和 Orchestrator-mediated runtime，让 MVP 可以受控生成 Dream Brief、Research Pack、Blueprint 和 Artifact。

## 非目标

- 不允许 Agent 自由无限聊天。
- 不允许 Agent 直接绕过 CapabilityInvoker。
- 不在 MVP 中实现多模型复杂路由市场。
- 不自动写生产代码或发布内容。

## 输入文档

- `.codex/plans/02-mvp-scope.md`
- `.codex/plans/10-eval-quality-system.md`
- `.codex/skills/*.md`
- `.codex/dev/05-capability-policy-runtime.md`

## 依赖阶段

依赖 `04-incubator-domain-runtime.md` 和 `05-capability-policy-runtime.md`。

## 核心产物

- ModelGateway port。
- Provider abstraction。
- Schema normalization。
- AgentSpec。
- Task model。
- Orchestrator-mediated runtime。
- Evaluator checks。

## 工程任务

- Model Gateway 支持：
  - provider abstraction
  - streaming
  - structured output
  - schema normalization
  - token / cost usage
  - retry
  - fallback
  - model profile
  - budget
  - cancellation
- 内置 Agent：
  - Orchestrator
  - Product Analyst
  - Competitor Analyst
  - Tech Architect
  - Growth Agent
  - Evaluator
- AgentSpec 字段：
  - id
  - role
  - input_schema
  - output_schema
  - allowed_capabilities
  - default_model_profile
  - budget
  - timeout
  - approval_policy
  - expected_artifacts
- Task 字段：
  - task_id
  - stage
  - goal
  - assigned_agent
  - required_capabilities
  - expected_artifacts
  - depends_on
  - budget
  - status
  - trace_id
- Evaluator 检查：
  - artifact completeness
  - evidence quality
  - hallucination risk
  - next_best_action

Model provider lock-in mitigation：

- All providers implement ModelGateway port.
- AgentSpec references model_profile, never concrete provider.
- Model output normalization is provider-independent.
- Golden tasks run in stub mode and at least one real-provider profile before release.
- Provider-specific features require capability flags and fallback behavior.

Prompt/version tracking：

- Every built-in Agent prompt has `prompt_id`, `prompt_version`, `agent_id` and changelog.
- Run events include prompt version references, not full secret-bearing prompt context.
- Prompt changes that affect output structure require golden task regression.

Agent output normalization gates：

- Raw model output never directly becomes Artifact.
- Structured output must validate against output_schema.
- Failed normalization creates recoverable error and retry/fallback event.
- Evaluator receives normalized artifact plus evidence refs.
- Normalization errors are counted in eval regression.

## 数据结构 / 接口 / schema 影响

AgentSpec 示例：

```json
{
  "id": "product_analyst",
  "role": "Analyze users, needs and MVP scope",
  "allowed_capabilities": ["model_generate_stub", "artifact_write", "human_input"],
  "default_model_profile": "reasoning_light",
  "budget": { "max_tokens": 20000, "max_cost_usd": 1.5 },
  "timeout": "120s",
  "expected_artifacts": ["dream_brief.md", "mvp_scope.md"]
}
```

## 测试要求

- Go unit:
  - schema normalization。
  - task graph rules。
  - budget cancellation。
  - Evaluator scoring stubs。
- Go integration:
  - model_generate_stub。
  - Agent task execution。
  - cancellation。
- Golden tasks:
  - 5 sample ideas。
- Contract tests:
  - AgentSpec schema。
  - Task event schema。

## 验收标准

- Orchestrator 通过 task graph 调度 Agent。
- Agent 只能使用 allowed_capabilities。
- tool/capability 调用都经过 CapabilityInvoker。
- trace_id 贯穿 task、model call、tool call。
- Evaluator 能产生 artifact score、evidence quality、hallucination risk。
- Agent prompt/version references are written to run events.
- Normalization failures are recoverable and tested.
- ModelGateway does not expose concrete provider types to Orchestrator.

## Codex PR 拆分建议

- PR-06-01: 实现 ModelGateway port 和 model_generate_stub。
- PR-06-02: 定义 AgentSpec 和 Task schema。
- PR-06-03: 实现 Orchestrator-mediated task graph runtime。
- PR-06-04: 实现内置 Agent specs。
- PR-06-05: 实现 structured output normalization。
- PR-06-06: 实现 budget、timeout、cancellation。
- PR-06-07: 实现 Evaluator checks 和 golden task runner skeleton。

## 风险

- 模型输出不稳定会污染 EventStore，必须先 normalize。
- 多 Agent 成本可能过高，需要 budget gate。
- fallback 策略过复杂会推迟 MVP。

## 暂不做

- 不接真实付费模型作为默认路径。
- 不实现外部 A2A Agent。
- 不实现代码执行 Agent。
