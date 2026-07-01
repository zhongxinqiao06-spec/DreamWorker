# 07 Desktop Workspace UIUX

| Field       | Value                                                          |
| ----------- | -------------------------------------------------------------- |
| Status      | In Implementation                                              |
| Owner       | Desktop/UI                                                     |
| Priority    | P0                                                             |
| DependsOn   | 01, 02, 03, 05, 06, 13                                         |
| ExitGate    | 桌面工作台能通过 typed API 驱动资源、聊天、项目和 runtime 状态 |
| PR Range    | PR-07-*                                                        |
| Risk Level  | High                                                           |
| Last Review | 2026-07-01                                                     |

## 目标

把桌面端做成高级、清晰、可操作的 AI OS 工作台。默认入口是 Chat Workspace，但聊天必须展示 Agent、模型、项目、runtime、tool calls 等真实上下文；Resource Center 是所有模型、Agent、Skill、Tool、MCP 的统一控制台。

## UX 原则

- 少描述性小字，多明确控件和状态。
- 工作台第一屏直接可用，不做营销页。
- 不用“概念大屏”替代用户工作流。
- 控件要专业：选择器、开关、状态、列表、测试按钮、刷新按钮、执行阶段。
- 文案中文清晰，术语保留原文时必须有上下文。

## 信息架构

```text
DreamWorker Desktop
  Sidebar
    - 聊天
    - 项目
    - 资源
    - 探索
    - 产品
    - 开发
    - 销售
    - 设置
    - 诊断
  Top Bar
    - 当前工作区
    - Engine 状态
    - Command-K
  Main
    - Chat Workspace
    - Projects Workspace
    - Resource Center
    - Module Workspaces
    - Settings
    - Diagnostics
```

## Chat Workspace

必须实现：

- 会话列表。
- Agent 切换。
- 模型配置切换。
- 项目绑定。
- 消息历史加载。
- 发送中状态。
- runtime steps：PLAN、GRAPH、EXECUTE、OBSERVE、REPLAN。
- tool call preview。
- Agent runtime 摘要：planner、executor、memory、context window。

## Resource Center

必须实现：

- Provider 列表和编辑器。
- Provider 类型：OpenAI、Anthropic、DeepSeek、GLM、Volcano、SiliconFlow、OpenAI Compatible、Ollama。
- API Key password input，Renderer 只显示 maskedKey。
- Test connection。
- Auto fetch models。
- Default model selection。
- Capability checkboxes。
- Model profiles。
- Agent runtime config 摘要。
- Skills、Tools、MCP 列表和启用状态。

## typed API

- `models.listProviders/saveProvider/deleteProvider/testProvider/refreshProviderModels/listModelProfiles/saveModelProfile`
- `agents.listAgents/getAgent/saveAgent/duplicateAgent/deleteAgent`
- `skills.listSkills/getSkill/saveSkill/deleteSkill`
- `tools.listTools/getTool/setToolEnabled`
- `mcp.listServers/saveServer/deleteServer/testServer/refreshTools`
- `projects.listProjects/createProject/getProject/updateProject/deleteProject/listProjectModules/getProjectModule/updateProjectModuleConfig`
- `chat.listSessions/createSession/getMessages/sendMessage/deleteSession`

## 安全边界

- Renderer 只调用 `window.dreamworker.*`。
- Preload 只暴露白名单。
- Main 只代理 Engine HTTP。
- Go Engine 管理 workspace state、secret masking、project isolation。
- UI 不存 localStorage，不存 raw key，不访问 Node/fs/process。

## 验收

- 打开应用默认进入 Chat Workspace。
- Chat 可以绑定 Agent、模型和项目，并显示 runtime steps/tool calls。
- Resource Center 可以配置 Provider、测试连接、刷新模型、选择默认模型。
- 所有数据来自 Engine typed API。
- `npm run ci` 和 `npm run build` 通过。

## PR 拆分

- PR-07-01：App Shell、主导航、顶部状态栏。
- PR-07-02：typed preload API 和 Main proxy。
- PR-07-03：Workspace store 和 Engine seed data。
- PR-07-04：Chat Workspace。
- PR-07-05：Resource Center。
- PR-07-06：Projects Workspace。
- PR-07-07：Module Workspaces。
- PR-07-08：Settings/Diagnostics。
- PR-07-09：Renderer/preload/Go/security tests。
