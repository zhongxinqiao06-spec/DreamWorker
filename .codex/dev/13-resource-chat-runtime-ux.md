# 13 Resource Chat Runtime UX

| Field       | Value                                                     |
| ----------- | --------------------------------------------------------- |
| Status      | In Implementation                                         |
| Owner       | Runtime/Desktop                                           |
| Priority    | P0                                                        |
| DependsOn   | 02, 03, 05, 06, 07                                        |
| ExitGate    | Resource Center 和 Chat Workspace 形成可测试 runtime 闭环 |
| PR Range    | PR-13-*                                                   |
| Risk Level  | High                                                      |
| Last Review | 2026-07-01                                                |

## 开发目标

落实 `.codex/plans/12-ai-os-runtime-resource-chat.md`。这一阶段不接受静态 UI：所有资源、聊天、Agent、模型、工具调用和 runtime 状态都必须通过 typed API 和 Go Engine 驱动。

## 当前扫描结论

- 已有 Electron + Vue、typed preload、Main proxy、Go Engine daemon。
- 已有 workspace store、ResourceCenter、ChatWorkspace。
- 已有 Provider、Agent、Skill、Tool、MCP、Project、Chat 的基础契约。
- 已补 Provider status/capability、Agent runtime config、chat execution steps、tool call preview。
- 剩余重点是聊天里的 Agent/model/project binding、资源中心的配置闭环、测试覆盖和真实 adapter 预留。

## Contract 工作

- Provider：
  - `providerType`
  - `status`
  - `capabilities`
  - `availableModels`
  - `defaultModel`
  - `hasApiKey`
  - `maskedKey`
  - `lastTestedAt`
  - `lastError`

- Agent：
  - `systemPrompt`
  - `modelProfileId`
  - `enabledSkills`
  - `enabledTools`
  - `enabledMcpServers`
  - `runtimeConfig`
  - `planner`
  - `executor`
  - `memoryScope`

- Chat turn：
  - `session`
  - `messages`
  - `executionSteps`
  - `toolCalls`
  - `runtimeSummary`

## Engine 工作

- `RefreshProviderModels(providerId)`：stub 阶段按 provider type 返回 deterministic catalog。
- `ListChatMessages(sessionId)`：切换会话加载历史。
- `SendChatMessage`：返回 runtime steps、tool calls、trace_id。
- seed agents 全部包含 runtime config。
- 保存 provider/agent 时补默认 capability/runtime，不让前端兜底业务逻辑。

## Desktop 工作

- ResourceCenter：
  - Provider type selector。
  - masked API key。
  - test connection。
  - auto fetch models。
  - default model select。
  - capability picker。
  - provider status line。
  - agent runtime summary。

- ChatWorkspace：
  - session select loads messages。
  - agent selector。
  - model selector。
  - project selector。
  - sending state。
  - runtime steps panel。
  - tool call preview panel。
  - right panel shows planner、executor、memory、skills、tools、MCP。

## 测试

- Go：
  - raw API key 不泄露。
  - refresh models fallback 稳定。
  - seed agents 有 runtime config。
  - chat turn 有 PLAN 到 REPLAN。
  - tool call preview 存在。

- TypeScript：
  - shared contract 覆盖 Provider/Agent/Chat。
  - preload 暴露白名单 API。
  - store 加载资源、刷新模型、切换会话、发送消息。

- 安全：
  - Renderer 不出现 raw `ipcRenderer`、`process`、`fs`、`localStorage`。
  - Preload 只暴露 `window.dreamworker`。

## 后续真实接入

- OpenAI-compatible `/models` discovery。
- Anthropic model mapping。
- Ollama local discovery。
- MCP tool discovery。
- Anthropic-style Skill import。
- Streaming response。
- Cost budget。

## 验收

- `npm run ci` 通过。
- `npm run build` 通过。
- 用户可以在 UI 中完成 Provider 配置、模型刷新、默认模型选择。
- 用户可以在聊天中选择 Agent、模型、项目，并看到 runtime execution。
- `npm run ci` 和 `npm run build` 通过。

## Stream Runtime Addendum

当前阶段补齐真实流式模型调用闭环：

- Engine 新增 `POST /chat/messages/stream`，输出 DreamWorker 统一 SSE：`started`、`step`、`token_delta`、`tool_call_delta`、`usage`、`completed`、`failed`、`cancelled`。
- Engine 新增 `POST /chat/messages/cancel`，按 `streamId` 取消当前 provider request。
- Electron Main 持有 Engine token 并代理 SSE，Renderer 只接收 typed IPC event。
- Preload 暴露 `chat.streamMessage(input, onEvent)` 和 `chat.cancelStream(input)`。
- Chat Workspace 发送后进入 streaming 状态，assistant message 按 token 增量更新，Stop 可取消。
- final `completed` event 返回完整 `ChatTurnResult`，并持久化 assistant message、usage、provider、model、finish reason 和 runtime summary。
- token delta 不落 EventStore；只保留 final message 和审计摘要。

Provider matrix：

| Provider                               | Stream endpoint             | Model discovery              | Notes                             |
| -------------------------------------- | --------------------------- | ---------------------------- | --------------------------------- |
| OpenAI                                 | Responses API `stream=true` | `/v1/models`                 | 解析 `response.output_text.delta` |
| OpenAI Compatible                      | `/v1/chat/completions`      | `/v1/models`                 | 解析 `choices[].delta.content`    |
| DeepSeek / GLM / Volcano / SiliconFlow | Chat Completions compatible | provider catalog / `/models` | baseURL 由 Resource Center 控制   |
| Anthropic                              | `/v1/messages`              | `/v1/models` / mapping       | 解析 `content_block_delta`        |
| Ollama                                 | `/api/chat`                 | `/api/tags`                  | 解析逐行 JSON                     |

Stream UX gates：

- 发送期间禁止修改当前 session 的 Agent、模型和项目绑定。
- Provider disabled、缺 key、模型不存在、网络错误、超时、取消都必须有明确状态。
- raw API key、MCP secrets、provider 原始事件不得进入 Renderer。
