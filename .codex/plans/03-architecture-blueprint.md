# 03 Architecture Blueprint

## 目标

保留 Electron + Go Engine 架构原则，并把系统升级为项目孵化器操作系统：Electron 管理体验和桌面生命周期，Go Engine 承载 Agent Runtime、Capability Bus、Policy、EventStore、Artifact、MCP/A2A/Skill/Model Gateway。

## 非目标

- Renderer 不直接访问 Node、文件系统、Git、SQLite、secrets 或 Go 进程。
- Electron Main 不承载 Agent 编排逻辑。
- Go Engine 不依赖 Electron UI。
- 不使用 Go in-process plugin 做扩展生态。
- 不让外部能力绕过 PolicyEngine 和 EventStore。

## 核心对象

- Electron Renderer：Incubator Workspace UI。
- Electron Main：窗口、托盘、更新、系统入口、Go Engine 生命周期。
- Go Engine：Agent Kernel 和孵化器 runtime。
- Engine API：命令接口。
- Event Stream：AG-UI-like UI event stream。
- EventStore：append-only 状态变化。
- ArtifactStore：交付物存储。
- Capability Bus：外部能力统一路由。
- PolicyEngine：权限和审批边界。

## 数据结构示例

```yaml
engine_boundary:
  renderer:
    allowed: ["render_state", "send_user_intent", "subscribe_events"]
    denied: ["fs", "shell", "sqlite", "secrets", "direct_go_process"]
  main:
    responsibilities: ["window_lifecycle", "engine_lifecycle", "ipc_proxy", "secret_broker"]
  go_engine:
    responsibilities:
      - agent_runtime
      - capability_bus
      - policy_engine
      - event_store
      - artifact_store
      - mcp_gateway
      - a2a_gateway
      - skill_runner
      - model_gateway
```

## 关键流程

1. Main 启动 Go Engine，本地随机端口或命名管道通信。
2. Renderer 通过 Preload 调用 typed API。
3. Main 代理命令到 Go Engine，并隐藏 engine token。
4. Go Engine 处理 Mission、Run、Task、CapabilityInvocation。
5. 所有状态变化写入 EventStore。
6. Go Engine 输出 AG-UI-like event stream。
7. Renderer reducer 根据事件重建 UI 状态。

## MVP 做法

- 使用 HTTP + SSE 或 WebSocket 作为本地 Command API / Event Stream。
- Main 管理一次性 engine token，Renderer 不持有明文 token。
- SQLite 存 EventStore 和 metadata，artifact 文件存 project artifacts 目录。
- Go Engine 模块按 domain / app / ports / adapters / runtime 分层。
- `code-q/` 仅作为架构和文档风格参考，不迁移代码。

## 后续扩展

- Command API 升级到 Connect-RPC 或 gRPC。
- 增加 CLI 和 headless engine。
- 增加 DreamWorker as MCP Server / A2A Server。
- 支持云端 Engine 和团队 workspace。

## 风险

- Renderer 权限泄漏会破坏安全模型。
- Main 如果塞入业务逻辑，会导致桌面壳难维护。
- Go Engine 如果直接依赖 UI 类型，会阻碍未来 Web/CLI。
- 本地通信 token 管理不当会被恶意页面或插件滥用。
