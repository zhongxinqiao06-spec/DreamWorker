# 07 UIUX Interaction Spec

## 目标

定义 DreamWorker 桌面端的产品信息架构：它首先是可配置、可扩展的 AI 工作台，其次才承载项目孵化器运行态。用户应先获得一个正常可用的 Agent 聊天工作台、资源配置中心和项目配置页，再通过左侧一级入口进入探索、产品、开发、销售闭环模块。

## 非目标

- 不把首页设计成运行态演示大屏。
- 不把 DreamWorker 做成传统 workflow builder。
- 不把聊天窗口做成唯一产品能力。
- 不在 Renderer 中保存密钥、项目数据或对话历史。
- 不复制第三方产品代码、切图、资产或品牌样式。

## 核心对象

- App Shell：侧栏、顶部状态栏、命令面板、主工作区。
- Chat Workspace：普通 Agent 对话、会话、Agent 摘要、项目上下文。
- Resource Center：模型服务商、模型配置、Agent、Skill、工具、MCP。
- Project Space：项目列表、创建项目、基础信息、项目级资源绑定、删除项目。
- Module Workspaces：Explore、Product、Development、Sales 四个一级主模块。
- Submodule Cards：每个主模块下的具体能力入口，以卡片展示状态、能力、产物和 next_best_action。
- Settings：本地优先、中文界面、安全边界和密钥策略。
- Diagnostics：runtime.ping、trace_id、Engine 状态、API 覆盖。

## 数据结构示例

```yaml
workspace_state:
  active_primary: chat
  active_project_id: project_001
  active_submodule_id: opportunity_radar
  active_agent_id: agent_general_assistant
  active_model_profile_id: profile_fast
  runtime:
    ping_status: ready
    trace_id: tr_001
```

```yaml
model_provider:
  providerId: provider_deepseek
  providerType: deepseek
  displayName: DeepSeek 兼容服务
  baseURL: https://api.deepseek.com
  defaultModel: deepseek-v4-flash
  availableModels:
    - deepseek-v4-flash
  enabled: true
  hasApiKey: true
  maskedKey: sk-b...4f3c
```

```yaml
project_module:
  projectId: project_001
  moduleId: explore
  displayName: 探索模块
  submodules:
    - submoduleId: opportunity_radar
      displayName: 机会雷达
      status: ready
      outputArtifacts:
        - dream_brief.md
        - hypotheses.yaml
  defaultAgents:
    - agent_opportunity_scout
    - agent_competitor_analyst
  enabledSkills:
    - skill_opportunity_scan
    - skill_competitor_map
  outputArtifacts:
    - dream_brief.md
    - research_pack.md
  nextBestAction: 先跑机会扫描，再补竞品和客群证据。
```

## 关键流程

1. 用户打开 DreamWorker，默认进入普通 Agent 聊天工作台。
2. 用户在聊天里提出问题、输入想法或让 Agent 做轻量分析。
3. 用户进入资源配置中心，配置模型服务商、模型配置、Agent、Skill、工具和 MCP。
4. 用户进入项目页，创建项目并绑定项目级资源。
5. 用户从左侧一级导航进入探索、产品、开发、销售模块。
6. 每个主模块以子模块卡片展示具体能力入口，例如机会雷达、MVP 收敛、PR 拆分、发布计划。
7. 模块运行态、Evidence、Decision Gate、Artifact Studio、Run Timeline 后续作为子模块能力展开。
8. runtime.ping、trace_id、引擎错误只在状态栏和诊断区展示。

## MVP 做法

- 左侧一级导航：聊天、项目、资源、探索、产品、开发、销售、设置、诊断。
- Chat Workspace：
  - 会话列表。
  - 对话消息区。
  - Agent / 项目上下文摘要。
  - 发送消息走 Go Engine chat stub。
- Resource Center：
  - 模型服务商保存 / 测试。
  - API Key 只发送到 Engine，Renderer 只显示 maskedKey。
  - Agent、Skill、Tool、MCP 使用 Engine seed data。
- Project Space：
  - 项目列表、创建项目、编辑基础信息、删除项目。
  - 绑定默认模型、Agent、Skill、Tool、MCP。
- Module Workspaces：
  - 探索：机会雷达、用户画像、竞品地图、证据图谱。
  - 产品：MVP 收敛、PRD 草案、原型说明、蓝图画布。
  - 开发：技术架构、技术栈与成本、PR 拆分、测试门禁。
  - 销售：定位文案、落地页、发布计划、反馈循环。
  - 所有模块和子模块必须携带 projectId。
- Settings / Diagnostics：
  - 说明安全边界。
  - 展示 runtime.ping 和 typed API 覆盖状态。

## 后续扩展

- Chat Workspace 接入 Model Gateway streaming。
- Resource Center 接入真实模型测试、MCP tool discovery、Skill manifest validation。
- Project Modules 接入真实 Run Timeline、Artifact Studio、Evidence Drawer、Decision Gate。
- Command-K 支持资源创建、项目跳转、Agent 切换、模块启动。
- 项目级数据接入 EventStore 和 ArtifactStore。

## 风险

- 过早突出运行态大屏会让产品看起来像演示，而不是可用工作台。
- 资源中心如果没有 typed API 和 Engine stub，后续 Agent 能力会被硬编码。
- Renderer 如果保存 secret，会破坏 Electron 安全边界。
- 项目模块如果没有 projectId，会导致多项目隔离失败。
- UI 如果继续堆在单文件，会降低后续迭代速度和测试覆盖。
