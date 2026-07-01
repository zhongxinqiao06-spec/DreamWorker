# DreamWorker

DreamWorker 是一个本地优先的 AI OS + Agent Runtime + 项目孵化桌面工作台。当前实现已经从静态壳推进到可运行的 Electron + Vue + Go Engine 架构：资源中心负责模型、Agent、Skill、Tool、MCP；聊天工作区走真实流式模型闭环；项目模块负责资源绑定；探索、产品、开发、销售模块承载项目孵化流程。

## UI 文案规则

UI 层所有面向用户可见的文字必须使用中文，包括导航、按钮、表单、占位、空态、加载态、错误态、审批提示、状态面板、toast、tooltip、窗口标题、`aria-label` 和 `title`。`DreamWorker`、版本号、协议名、schema 字段名、Provider 名称等不可翻译标识可以保留原文，但必须搭配中文语境；不得新增纯英文占位文案。

## 当前完成度

已落地：

- Electron Main 启动 Go Engine sidecar，本地随机 token 鉴权，Renderer 永远拿不到 Engine URL、token、API Key。
- Vue + Pinia Renderer 工作台：聊天、项目、资源、探索、产品、开发、销售、设置、诊断。
- Typed Preload API：`window.dreamworker.*` 白名单接口，Renderer 不直接访问 Node、文件系统或 secret。
- Go Engine Workspace API：Provider、Profile、Agent、Skill、Tool、MCP、Project、Chat Session、Runtime Diagnostic。
- 真实流式模型闭环：Renderer -> Preload -> Main IPC/SSE proxy -> Go Engine -> ModelGateway -> Provider stream -> normalized events -> incremental UI -> final persisted message。
- ModelGateway adapters：OpenAI Responses、OpenAI-compatible Chat Completions、DeepSeek/GLM/Volcano/SiliconFlow/Gemini/custom compatible、Anthropic Messages、Ollama Chat。
- Provider 资源中心：保存密钥、脱敏展示、Test Connection、Refresh Models、Verify Streaming、health/model count/default model/streamingVerified。
- Chat Runtime：assistant attempt、retry 不重复创建 user message、cancel 保留 partial、runtime inspector、usage/latency/finish reason、自动下滚、模型思考默认收起。
- Context Manager：`ChatContextPack`、token budget、summary reuse、超预算压缩事件、deterministic fallback summary、secret redaction。
- Skill/Tool Runtime：`.agent/skills/<name>/SKILL.md` 自动扫描，内置 `skillcreator`，低风险工具可执行，高风险工具 policy block，MCP stdio `initialize` / `tools/list` / `tools/call` 最小闭环。
- SQLite adapters：EventStore、ArtifactStore、CapabilityRegistry 基础持久化。
- Windows packaging：Go Engine exe 打包进 Electron `extraResources`，`.agent` 同步进入安装产物。
- 工程门禁：lint、format、spec generation/check、typecheck、Vitest、Go test/vet/fmt check、security smoke、build。

仍在推进：

- HTTP/SSE MCP、远程 MCP 权限细化、工具审批 UI 完整闭环。
- 项目孵化四大模块从配置态升级为真实 artifact 生成与评估闭环。
- Installer、签名、自动更新、发布渠道。
- Cloud/team workspace、同步和多人协作。

## 产品定位

DreamWorker 不是普通 workflow 工具，不是普通 Agent Builder，也不是单纯聊天应用。它把模型、Agent、Skill、工具、MCP、项目空间和项目闭环模块组织到一个开放式桌面工作台里，让用户从普通 Agent 聊天开始，逐步进入项目的探索、产品、开发和销售闭环。

项目孵化链路：

```text
Idea
-> Mission
-> Hypothesis
-> Evidence
-> Experiment
-> Decision Gate
-> Blueprint
-> Multi-Agent Run
-> Artifact
-> Launch
-> Feedback
-> Next Iteration
```

英文定位：

> DreamWorker turns any idea into an executable agent-powered launch plan.

## 架构

```text
Electron Desktop
  |-- Renderer: Vue / Pinia / Canvas UX，只负责界面状态和交互
  |-- Preload: typed window.dreamworker API，只做安全桥接
  `-- Main: 窗口生命周期、Go Engine daemon、IPC/SSE proxy、本地 token

Go Engine
  |-- Runtime API: HTTP routes, SSE, cancel registry
  |-- Workspace Store: providers, profiles, agents, skills, tools, MCP, projects, chat
  |-- Chat Runtime: context pack, model stream, tool loop, audit summary
  |-- Model Gateway: OpenAI, compatible, Anthropic, Ollama adapters
  |-- Context Manager: budget, compaction, summaries, secret redaction
  |-- Capability Runtime: built-in tools, MCP stdio, policy gates
  |-- SQLite Adapters: events, artifacts, capability registry
  |-- Domain Contracts: versioned errors, events, artifacts, policies
  `-- Security / Diagnostics / Eval foundations
```

关键边界：

- Renderer 不直接访问 Node、Go、SQLite、文件系统、secrets。
- Main 只代理 typed API 和本地流，不写业务逻辑。
- Go Engine 必须能独立运行，未来支持 desktop local daemon、CLI、cloud server、team workspace 和第三方集成。
- Provider 原始事件不透传 UI；Engine 只发 DreamWorker typed stream events。
- token delta 不进入 EventStore；只持久化 final message、usage、tool calls、runtime steps 和 audit summary。
- 所有 secret、masked secret、MCP env、provider raw error body 不允许进入 prompt、event、message、log。
- 所有高风险 Tool/MCP/Skill 必须经过 Policy/Approval。
- 所有 schema、event、manifest 必须 versioned。

## 桌面信息架构

```text
DreamWorker Desktop
  |-- 聊天
  |   |-- Agent 对话
  |   |-- 会话列表
  |   |-- Agent / 模型 / 项目绑定
  |   `-- Runtime Inspector
  |
  |-- 项目
  |   |-- 项目列表 / 创建项目
  |   |-- 项目基础信息
  |   |-- 默认模型配置
  |   |-- 项目级 Agent / Skill / Tool / MCP 绑定
  |   `-- 删除项目
  |
  |-- 资源
  |   |-- 模型服务商
  |   |-- 模型配置
  |   |-- Agent
  |   |-- Skill
  |   |-- Tool
  |   `-- MCP
  |
  |-- 探索
  |-- 产品
  |-- 开发
  |-- 销售
  |-- 设置
  `-- 诊断
```

项目不是四大闭环模块的容器，只负责新增、修改、删除和基础资源绑定。探索、产品、开发、销售是左侧一级主模块，每个主模块用子模块卡片承载可配置能力组合。

## Chat Runtime

```text
validate session
-> create/reuse user message
-> create assistant attempt
-> emit started
-> build ChatContextPack
-> stream provider tokens
-> capture reasoning/tool deltas
-> policy check low/high risk tools
-> optionally execute low-risk tool
-> persist assistant final/partial result
-> emit completed/failed/cancelled
```

当前支持：

- `started | step | context_compacted | reasoning_delta | token_delta | tool_call_delta | tool_started | tool_result | tool_blocked | skill_used | usage | completed | failed | cancelled`。
- Main 到 Renderer 只发 typed IPC event；Renderer 不知道 Provider URL、Engine token 或 API Key。
- Stop 后保留 partial assistant attempt。
- Retry 使用同一个 user message 创建新的 assistant attempt。
- 发送期间允许浏览其他 session，stream event 只更新所属 session。

## Skill 与 Tool

项目内技能源在 `.agent/skills/<skill-name>/SKILL.md`，兼容 Anthropic/Claude Code 风格：

```text
.agent/
  |-- README.md
  `-- skills/
      |-- blueprint/SKILL.md
      |-- competitor-map/SKILL.md
      |-- evaluator/SKILL.md
      |-- launch-plan/SKILL.md
      |-- opportunity-scan/SKILL.md
      |-- prd-draft/SKILL.md
      `-- skillcreator/SKILL.md
```

Engine 启动时扫描 `.agent/skills`，把 frontmatter 和 markdown instructions 载入内存。后续 Skill 生成和安装也写入根目录 `.agent/skills`，不再走固定 seed 作为唯一来源。

## Go Engine 启动与打包

桌面端通过 Electron Main 启动本地 Go Engine HTTP daemon：

1. Renderer 调用 `window.dreamworker.*`。
2. Preload 把白名单方法转为 IPC。
3. Main 生成本地随机 token，启动 Go Engine 子进程。
4. Go Engine 执行 `serve --token <token>`，监听 `127.0.0.1:0`。
5. Go Engine 输出 `engine.ready` JSON，Main 读取 `baseUrl`。
6. Main 使用 `Authorization: Bearer <token>` 代理到 Engine。
7. Renderer 永远拿不到端口、token、Go 进程句柄或 API Key。

启动优先级：

```text
1. DREAMWORKER_ENGINE_PATH 指向的 engine 可执行文件。
2. engine/bin/dreamworker-engine.exe。
3. 开发态回退到 go run ./cmd/dreamworker-engine serve --token <token>。
```

## 快速开始

要求：

- Node.js 20+。
- npm workspaces。
- Go 1.22+，仅开发态 `go run` 或构建 engine 时需要。

安装依赖：

```powershell
npm install
```

开发启动：

```powershell
npm run dev
```

构建 Go Engine exe：

```powershell
npm run go:build:exe
```

完整构建：

```powershell
npm run build
```

Windows unpacked package：

```powershell
npm run package:win
```

完整门禁：

```powershell
npm run ci
```

## 常用脚本

- `npm run lint`：ESLint。
- `npm run format:check`：Prettier 检查。
- `npm run specs:check`：schema 生成产物和 fixtures 校验。
- `npm run typecheck`：Vue/TS 类型检查。
- `npm test`：desktop Vitest。
- `npm run go:test`：Go tests。
- `npm run go:vet`：Go vet。
- `npm run security:smoke`：Renderer/Main/Preload 安全 smoke。
- `npm run build`：类型、Electron build、Go build、engine exe。
- `npm run package:win`：完整 Windows dir package。

## 仓库说明

- `.agent/`：项目内 Skill 源，Engine 自动扫描。
- `.codex/`：Codex/Agent 开发计划、规则和阶段记忆。
- `apps/desktop/`：Electron + Vue + Pinia 桌面工作台。
- `engine/`：Go Engine、runtime API、workspace store、model gateway、capability runtime。
- `examples/`：后续 SDK、adapter、E2E 示例入口。
- `scripts/`：工程脚本。
- `specs/`：versioned JSON schemas、fixtures 和生成 contracts。

## 文档入口

- [.agent/README.md](.agent/README.md)
- [.codex/README.md](.codex/README.md)
- [.codex/dev/README.md](.codex/dev/README.md)
- [specs/README.md](specs/README.md)
- [scripts/README.md](scripts/README.md)
- [examples/README.md](examples/README.md)
