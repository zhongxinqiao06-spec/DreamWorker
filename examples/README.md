# DreamWorker Examples

`examples/` 是后续 SDK、adapter、E2E demo 和 conformance case 的入口。当前主产品代码已经在 `apps/desktop/` 与 `engine/` 落地，本目录暂不承载运行时代码，也不把计划中的示例写成已完成能力。

## 当前状态

- 暂无可直接运行的 example package。
- README 先固定示例边界、验收标准和未来目录规划。
- 示例必须以真实 `specs/` schema、generated contracts、typed preload API 或 Go Engine Runtime API 为事实源。
- `.codex/tmp` 下的外部项目只允许作为参考缓存，不能复制为 DreamWorker 示例。

## 计划中的示例

- Provider adapter：OpenAI-compatible 服务商如何接入 ModelGateway，并完成 health check、model discovery、streaming verification。
- Extension provider bridge：以 9Router 类本地 Node 扩展为例，演示 detect/install/start/stop/logs/provider bridge 的最小闭环。
- MCP stdio server：最小 `initialize`、`tools/list`、`tools/call` server，并通过 Resource Center 刷新工具。
- Skill package：生成一个 `.agent/skills/<name>/SKILL.md`，被 Engine 扫描后出现在 Resource Center。
- Chat stream client：消费 `ChatStreamEvent`，聚合 token delta、usage、completed/failed/cancelled 和 final assistant message。
- Project incubation flow：从 idea 到 blueprint/artifact/eval 的端到端流程。
- Contract conformance：对 `specs/fixtures`、generated TypeScript contracts 和 Go runtime contract subset 做兼容性验证。

## 添加示例的标准

- 示例必须可用单条 npm 或 Go 命令运行。
- 示例必须说明输入、输出、依赖环境变量和失败状态。
- 示例必须使用真实 typed contracts，不复制内部私有结构。
- 示例不能要求 Renderer 直接访问 secret、Engine token、Provider key、文件系统或 raw IPC。
- 示例中出现的密钥、MCP env、provider raw error body 必须脱敏。
- 示例要能被 CI、smoke 或 conformance script 验证。
- 示例文案使用中文，协议名、字段名和 Provider 名称保留原文。

## 建议目录形态

```text
examples/
  |-- README.md
  |-- provider-adapter/
  |-- extension-provider-bridge/
  |-- mcp-stdio-server/
  |-- skill-package/
  |-- chat-stream-client/
  |-- project-incubation-flow/
  `-- contract-conformance/
```

这些目录只有在对应示例真正落地时再创建，避免空目录或伪完成文档扩散。
