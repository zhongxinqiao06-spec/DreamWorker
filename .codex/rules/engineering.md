# Engineering Rules

工程实现优先保持清晰边界、local-first、安全默认和可测试。

## 进程边界

- Renderer 不直接访问 Node、SQLite、Git、终端、文件系统、secrets 或 Go 进程。
- Preload 只暴露最小化、类型化 API，不暴露 `ipcRenderer`。
- Electron Main 负责本地高权限入口和 Go Engine 生命周期。
- Go Engine 是唯一 Agent Runtime。

## 契约

- 跨进程能力必须先定义共享契约，再实现 Main、Preload、Renderer。
- 事件 payload、错误码、领域对象要稳定可序列化。
- 用户可见错误用中文 message，技术细节进入 detail，并在写入前脱敏。

## 存储

- 桌面 MVP 默认 local-first。
- 项目状态、run、task、approval、tool call、artifact metadata 使用 SQLite。
- 交付物文件进入 project artifact store。
- secrets 进入系统 Secret Store，不写普通 JSON。

## 模块化

- Capability 调用不得散落在业务模块中，必须走 Capability Registry。
- 权限判断不得散落在 UI 中，必须走 Policy Engine。
- Agent 执行不得绕过 Runtime、Event Store 和 Approval。

## 测试

- 领域 schema、IPC contract、policy、capability routing、event replay 必须有单元测试。
- 高风险工具调用必须有 approval 测试。
- 文档变更至少检查 README、`.codex` 结构和项目名一致性。
