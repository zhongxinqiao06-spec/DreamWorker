# DreamWorker Examples

`examples/` 是后续 SDK、adapter、E2E demo 和 conformance case 的入口。当前主产品代码已经在 `apps/desktop/` 与 `engine/` 落地，示例目录暂不承载运行时代码。

## 计划中的示例

- Provider adapter 示例：OpenAI-compatible 服务商如何接入 ModelGateway。
- MCP stdio 示例：最小 `initialize`、`tools/list`、`tools/call` server。
- Skill 示例：生成一个 `.agent/skills/<name>/SKILL.md` 并被 Engine 扫描。
- Chat stream 示例：消费 `ChatStreamEvent` 并聚合 final assistant message。
- Project incubation 示例：从 idea 到 blueprint/artifact/eval 的端到端流程。

## 当前原则

- 示例必须使用真实 typed contracts，不复制内部私有结构。
- 示例不能要求 Renderer 直接访问 secret、Engine token 或 Provider key。
- 示例要能被 CI 或 conformance script 验证。
- 示例文案使用中文，协议名和字段名保留原文。
