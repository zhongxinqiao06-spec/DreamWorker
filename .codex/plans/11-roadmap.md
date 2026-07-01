# 11 Roadmap

## 目标

把 DreamWorker 的路线统一为 AI OS + Agent Runtime + 项目孵化系统，而不是“先做一个聊天壳，再慢慢加能力”。第一阶段必须先把资源中心、聊天工作区、Agent Runtime 契约和项目隔离打穿；后续 Explore、Product、Development、Sales 才有真实执行底座。

这里的 MVP 只表示“第一条完整可验收闭环”，不表示展示型、缩水型、糊弄型实现。

## 强约束

- 不做只有 UI 的 demo。
- 不做没有 Agent Runtime 的聊天工具。
- 不做 Skill、Tool、MCP 混在一起的配置页。
- 不让 Renderer 保存 secret、项目数据或聊天历史。
- 不让模型调用绕过 Model Gateway。
- 不让工具调用绕过 Tool Router、Policy 和 Approval。
- 不让项目数据缺少 `projectId` 隔离。

## 参考基准

- Cherry Studio：参考多模型供应商、模型列表、MCP 配置、Agent 配置和用户友好的资源管理体验，不复制代码和视觉资产。
- MCP：对齐 JSON-RPC、tools、resources、prompts、capability negotiation、stdio/HTTP/SSE 传输扩展。
- Anthropic Agent Skills：兼容 `SKILL.md`、instructions、scripts、resources 的 skill bundle 形态。
- OpenAI Agents SDK：对齐 tools、handoffs、guardrails、tracing、runner loop、structured output，但 DreamWorker 自己管理 project isolation、approval 和 state。

## 阶段路线

### Phase 0：计划和契约校准

- 重构 `.codex/plans` 与 `.codex/dev`，统一 AI OS / Runtime / Resource / Chat 主线。
- 明确 Resource Center + Chat Workspace 是 Phase 1 主入口。
- 所有阶段输出必须包含 schema、UI、runtime stub、test case、example。

ExitGate：

- `plans/11`、`plans/12`、`dev/00`、`dev/07`、`dev/13` 不再互相冲突。
- PR 路线能直接交给 Codex 执行。

### Phase 1：Resource + Chat Runtime Core

- Provider system：OpenAI、Anthropic、DeepSeek、GLM、Volcano、SiliconFlow、OpenAI Compatible、Ollama。
- Provider UX：masked key、test connection、auto fetch models、default model、capability/status。
- Agent config：system prompt、model profile、skills、tools、MCP、runtime config、planner、executor、memory scope。
- Chat runtime：session history、agent/model/project binding、execution steps、tool call preview、trace_id。
- Engine stub：所有前端功能通过 typed API 进入 Go Engine，不走本地硬编码。

ExitGate：

- 资源中心和聊天页不是静态展示，能读写 Engine 数据。
- 发送聊天消息返回 `PLAN -> GRAPH -> EXECUTE -> OBSERVE -> REPLAN` 结构。
- CI、build、安全 smoke 全过。

### Phase 2：Capability + MCP + Skill Alpha

- MCP registry：stdio、HTTP、SSE、JSON-RPC 2.0 抽象。
- Tool discovery：MCP tools 映射为 DreamWorker Tool。
- Skill import：Anthropic-style skill bundle 导入为 DreamWorker Skill。
- Policy：MCP、Skill、Tool 都有 trust level、risk level、approval required。

ExitGate：

- 未验证 MCP/Skill 默认不可静默执行。
- Tool Router 和 PolicyEngine 对所有外部能力生效。

### Phase 3：Project Incubation Graph

- Explore、Product、Development、Sales 四个项目模块接入 Agent Runtime。
- 每个模块有 default agent、default skill、default tool、MCP recommendation、input schema、output artifact。
- 项目级 memory、artifact、event、decision gate 形成闭环。

ExitGate：

- 单个 idea 能进入项目空间并生成可追踪 artifact。
- 每个 artifact 可追溯到 agent、tool、trace、projectId。

### Phase 4：Real Model Gateway

- OpenAI-compatible `/models` 和 chat/tool calling。
- Anthropic model mapping 与 tool/skill 兼容。
- Ollama local model discovery。
- DeepSeek、GLM、Volcano、SiliconFlow provider adapter。
- Streaming response、structured output、cost budget、fallback resolution。

ExitGate：

- 至少两个云端 provider 和一个本地 provider 可切换。
- Provider 错误和成本状态能在 UI 中清晰呈现。

### Phase 5：Execution Hardening

- Task Graph scheduler、replanner、memory injection、context compression。
- Run Timeline、trace store、diagnostics、golden tasks。
- Cost/Risk panel、approval diff、rollback。

ExitGate：

- 长任务可观察、可中断、可恢复、可评估。
- 高风险动作没有 approval 不能执行。

### Phase 6：Open Platform

- DreamWorker as MCP Server。
- A2A external agent adapter。
- SDK/examples/conformance tests。
- Marketplace 只在 security、policy、sandbox 成熟后进入。

ExitGate：

- 外部接入有稳定 manifest、conformance 和安全边界。

## 优先级

- P0：Resource Center、Chat Workspace、Agent Runtime contract、Provider system、project isolation、secret isolation。
- P1：真实 Model Gateway、MCP discovery、Skill import、streaming、memory。
- P2：SDK、marketplace、team workspace、cloud sync。

## 风险

- 如果 Phase 1 只做 UI，后续 Agent 会变成 prompt 拼接。
- 如果 Provider 只是设置项，模型切换、Agent 配置、成本控制都会失真。
- 如果 Skill/Tool/MCP 不分层，安全策略无法落地。
- 如果聊天不是 Runtime 入口，产品会滑回普通 chat app。
