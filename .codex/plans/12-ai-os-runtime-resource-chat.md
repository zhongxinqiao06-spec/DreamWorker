# 12 AI OS Runtime Resource Chat Plan

## 目标

Resource Center 和 Chat Workspace 是 DreamWorker 的第一主入口。它们必须先成为真实可用的 AI OS 控制台，再承载 Explore、Product、Development、Sales 项目孵化闭环。

本阶段不是做“设置页 + 聊天框”，而是做：

- Multi-Model Provider System
- Agent Runtime entry
- Skill / Tool / MCP registry
- Project-aware Chat Workspace
- Runtime trace and tool-call visualization

## 强约束

- Provider 是一等系统，不是普通设置项。
- Chat 是 Agent Runtime 入口，不是单轮 prompt UI。
- Skill、Tool、MCP 必须严格分层。
- 所有外部能力必须抽象为 Capability。
- 所有 Agent 执行必须能被 trace、policy、approval、projectId 约束。
- 开发阶段不能用“先 MVP”省掉 runtime、schema、测试和 UX。

## Cherry Studio UX 拆解

参考 Cherry Studio 的方向，不复制代码：

- Provider 管理要开箱友好：列表、启用状态、密钥脱敏、连接测试、模型拉取、默认模型。
- 模型能力要可见：chat、tools、vision、json_schema。
- MCP 配置要可理解：server、transport、command/url、env keys、trust level、tool discovery。
- Agent 配置要跟模型和工具连起来：Agent 选择模型、绑定工具/MCP/Skill，不允许孤立配置。
- 用户流程要短：配置 key -> 测试 -> 获取模型 -> 设默认模型 -> 进入聊天验证。

## 标准对齐

- MCP：按 JSON-RPC 2.0、tools/resources/prompts、transport、capability negotiation 设计。
- Anthropic Skills：支持 `SKILL.md` + instructions/scripts/resources 的导入模型，内部转换为 DreamWorker Skill。
- OpenAI Agents SDK：对齐 tools、handoffs、guardrails、tracing、runner loop，内部实现仍由 Go Agent Runtime 管。
- OpenAI-compatible providers：优先支持 `/models` discovery 和 chat/tool calling 兼容路径。

## Resource Center

必须包含：

- Providers：OpenAI、Anthropic、DeepSeek、GLM、Volcano、SiliconFlow、OpenAI Compatible、Ollama。
- Model Profiles：model、temperature、maxTokens、contextWindow、purpose、enabled。
- Agents：systemPrompt、modelProfile、skills、tools、mcpServers、runtimeConfig、planner、executor、memoryScope。
- Skills：prompt injection、allowed tools、allowed MCP、parameter schema、output artifacts。
- Tools：input schema、output schema、risk level、approval required、enabled。
- MCP：transport、command/url、args、env keys、secrets masked、trust level、tool discovery。

## Chat Workspace

必须包含：

- 左侧：chat sessions、agent quick switch。
- 中间：messages、streaming state、runtime steps、tool calls。
- 右侧：agent config、model profile、skills/tools/MCP、project binding、memory scope。
- Composer：支持当前 session 的 agent/model/project 绑定，不隐藏关键上下文。

每轮消息必须返回：

- trace_id
- assistant message
- execution steps
- tool call preview
- runtime summary
- session update

## Engine Stub 要求

stub 也必须像真实 runtime：

- Provider refresh 返回 deterministic model catalog。
- Chat turn 返回 `PLAN -> GRAPH -> EXECUTE -> OBSERVE -> REPLAN`。
- Tool call preview 走 Tool registry。
- Agent runtime config 从 Engine 返回，不由 Renderer 编造。
- projectId 为空或明确绑定，不允许隐式串项目。

## 分阶段执行

1. Contract：Provider、Agent、Chat、ToolCall、ExecutionStep。
2. Engine：workspace store、provider refresh、chat getMessages、runtime turn。
3. Desktop：ResourceCenter 和 ChatWorkspace 完整交互。
4. Tests：Go、preload、store、security smoke。
5. Real adapters：OpenAI-compatible、Anthropic、Ollama、MCP discovery、Skill import。

## 验收

- 资源中心能完成 key 配置、测试、模型获取、默认模型选择、能力展示。
- 聊天能选择 Agent、模型、项目，发送后显示 runtime steps 和 tool calls。
- 所有能力经 typed API 和 Go Engine，不在 Renderer 硬编码业务状态。
- CI 和 build 通过。
