# DreamWorker

DreamWorker 是一个开放式 Agent 项目孵化台：把一个想法变成可验证方案、可执行蓝图、多 Agent 协作交付和发布增长计划。

它不是又一个固定 DAG 的 workflow builder，也不是单纯的 Agent builder。DreamWorker 的核心是 Mission / Blueprint / Run 模式：用户提出目标，系统先生成可审阅、可修改、可追踪的任务蓝图，再在运行中根据证据、失败、反馈、预算和用户指令动态重排任务。

英文定位可以概括为：

> DreamWorker turns any idea into an executable agent-powered launch plan.

## 产品定位

DreamWorker 面向独立开发者、AI 创业者和小型产品团队，解决从 idea 到 MVP 的早期混乱问题：

- 有想法，但不知道是否值得做。
- 不会系统做需求、竞品、痛点和成本分析。
- 不会把想法拆成 MVP、PRD、roadmap、issue 和发布计划。
- 会用 AI，但缺少一个能组织多 Agent、工具和交付物的项目工作台。

第一版聚焦一个清晰场景：

输入一个产品想法，在 30 分钟内生成 Dream Brief、Research Pack、可执行 Blueprint、PRD、GitHub issues、落地页文案和发布计划。

## 核心原则

- **Evidence-first**：每个关键结论都要绑定证据、来源、置信度、假设和待验证实验，避免变成普通聊天式报告。
- **Blueprint-first**：输出不是一堆静态文档，而是可执行、可修改、可回滚的项目蓝图。
- **Open Agent Bus**：MCP、A2A、Skills、HTTP/OpenAPI、Browser、Human Task 都统一抽象成 capability。
- **Human-in-command**：用户可以随时改目标、换 Agent、接工具、调预算、暂停、分支、回滚或把任务交给人。
- **Idea-to-launch**：链路从想法、验证、规划、执行延伸到宣发、反馈和迭代。

## MVP 能力

首版只做最小闭环：

- Idea Intake：把用户的一句话想法结构化成 Dream Brief。
- Research Pack：生成需求、用户、竞品、痛点、商业模式、MVP 范围和成本分析。
- Blueprint Generator：生成产品模块、页面结构、数据模型、API、Agent 分工、审批点和交付物清单。
- Agent Runtime：轻量调度内置 Agent，先保证过程可见、结果可追踪。
- Capability Registry：把内置工具、MCP、Skill、API、外部 Agent 和人工任务注册为统一能力。
- Artifact Hub：保存 PRD、roadmap、issue、schema、文案、发布 checklist 等交付物。
- Run Console：展示任务进度、工具调用、证据链、失败原因、成本和审批请求。
- Human Approval：对写文件、发外部请求、发布内容、调用付费 API、执行代码等高风险动作要求确认。

首版暂不做复杂 marketplace、团队权限、企业私有部署和全量 Agent 生态接入。这些能力进入后续阶段。

## 技术主线

DreamWorker 采用 Electron 桌面工作台 + Go 本地 Agent 引擎：

```text
DreamWorker Desktop
  |-- Electron Workspace
  |   |-- Idea Chat
  |   |-- Blueprint Canvas
  |   |-- Agent Run Console
  |   |-- Capability Panel
  |   |-- Artifact Hub
  |   `-- Approval Center
  |
  `-- Go Agent Kernel
      |-- Engine API
      |-- Event Bus / AG-UI Adapter
      |-- Project Event Store
      |-- Blueprint Compiler
      |-- Orchestrator
      |-- Multi-Agent Runtime
      |-- Capability Registry
      |-- MCP Gateway
      |-- A2A Gateway
      |-- Skill Runner
      |-- Model Gateway
      |-- Artifact Store
      |-- Memory / Search
      |-- Policy / Approval
      |-- Sandbox
      |-- Cost Meter
      `-- Observability / Eval
```

关键边界：

- Electron Renderer 只负责 UI、交互和运行状态镜像，不直接访问密钥、文件系统、Git、终端、SQLite 或 Go 进程。
- Electron Main 负责桌面生命周期、本地权限入口、Go 引擎启动、系统对话框、secrets broker 和 typed IPC。
- Go Engine 是 DreamWorker 的 Agent Kernel，负责蓝图编译、任务图、Agent 调度、能力路由、审批、沙箱、事件持久化和审计。
- Capability Registry 是统一抽象层，禁止在业务逻辑中写死 `callGithub`、`callNotion`、`callMCP` 这类点对点调用。
- Policy Engine 是安全边界，默认最小权限，高风险动作必须审批和记录。

## 协议兼容

DreamWorker 的扩展策略不是 Go in-process plugin，而是进程外、协议化接入：

- MCP：工具、数据源、文件、数据库、GitHub、Notion、Browser、企业系统。
- A2A：外部 Agent 发现、协作、状态回传和 artifact 交换。
- AG-UI-like Event Stream：Go Engine 与 Electron UI 的实时状态、工具调用、审批和用户 steering。
- Skill Package：说明文档、脚本、资源和权限声明组成的可审计能力包。
- HTTP/OpenAPI：传统 SaaS 或内部服务转成 capability。
- Browser Sandbox：处理没有 API 的网页任务。
- Human Task：把人类协作也作为 capability。

后续可以让 DreamWorker 同时作为 MCP Client、MCP Server、A2A Client 和 A2A Server，成为其他 Agent 平台可调用的项目操作系统节点。

## 路线

- Phase 0：产品定义。完成用户访谈、样例 idea 测试、Dream Brief 模板、Blueprint schema 和内置 Agent 定义。
- Phase 1：MVP。完成 idea 输入、自动追问、研究包、蓝图生成、文档导出、轻量 Run Console 和基础 capability。
- Phase 2：Alpha。加入 Capability Registry、MCP Client、审批中心、成本统计、失败重试、任务分支和 artifact versioning。
- Phase 3：Beta。加入 A2A、Skill Runner、项目记忆、模板库、反馈 Agent、宣发 Agent 和支付。
- Phase 4：商业化。加入团队空间、权限、审计、计费、私有部署和公共案例库。
- Phase 5：平台化。开放 DreamWorker MCP/A2A Server、SDK、marketplace 和第三方能力生态。

## 仓库说明

根目录是 DreamWorker 新项目的事实源。

- `.codex/`：后续 Codex/Agent 开发的计划、规则和项目内 skill 记忆。
- `code-q/`：只作为参考资产，可参考其中 Electron + Vue + Go sidecar、typed IPC、Agent runtime、MCP、skills 和文档组织风格；不要直接迁移、重命名或修改它来代表 DreamWorker。

## 文档入口

- [.codex/README.md](.codex/README.md)
- [.codex/plans/00-product-positioning.md](.codex/plans/00-product-positioning.md)
- [.codex/plans/01-incubator-domain.md](.codex/plans/01-incubator-domain.md)
- [.codex/plans/02-mvp-scope.md](.codex/plans/02-mvp-scope.md)
- [.codex/plans/03-architecture-blueprint.md](.codex/plans/03-architecture-blueprint.md)
- [.codex/plans/04-engine-code-skeleton.md](.codex/plans/04-engine-code-skeleton.md)
- [.codex/plans/05-capability-bus.md](.codex/plans/05-capability-bus.md)
- [.codex/plans/06-open-source-accessibility.md](.codex/plans/06-open-source-accessibility.md)
- [.codex/plans/07-uiux-interaction-spec.md](.codex/plans/07-uiux-interaction-spec.md)
- [.codex/plans/08-performance-observability.md](.codex/plans/08-performance-observability.md)
- [.codex/plans/09-security-policy.md](.codex/plans/09-security-policy.md)
- [.codex/plans/10-eval-quality-system.md](.codex/plans/10-eval-quality-system.md)
- [.codex/plans/11-roadmap.md](.codex/plans/11-roadmap.md)
