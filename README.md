# DreamWorker

DreamWorker 是本地优先的 AI OS、Agent Runtime 和项目孵化桌面工作台。桌面壳使用 Electron + Vue，本地运行时已经收敛为 Electron Main 内嵌 Runtime。

> DreamWorker 把一个想法推进成可执行、可追踪、可协作的 Agent 项目计划。

## 当前运行时

- Electron Main 在同一进程内创建 `apps/desktop/main/runtime`，不再启动独立 Runtime 子进程。
- Renderer 只通过 typed `window.dreamworker.*` preload API 调用能力，不直接访问 token、文件系统或原始 IPC。
- Main 负责桌面生命周期、内嵌 Runtime、IPC bridge、流式事件转发和本地安全边界。
- Main Runtime 负责 providers、profiles、settings、extensions、agents、skills、tools、MCP、projects、requirements、chat、coding agents、runtime diagnostics 和本地 SQLite 持久化。
- 旧 `workspace.db` 继续按 `workspace_state.payload` snapshot 读取，provider、project、chat、module 数据可跨运行时迁移保留。
- UI 层所有面向用户可见的文字必须使用中文，协议名、字段名和 Provider 名称保留原文。

## 编码 Agent

开发模块内置「编码 Agent」工作台：

- 左栏文件树固定基于当前项目的 `localRootPath/workspace/code`；新建项目时会补齐 `workspace/code` 目录。
- 中栏是编码对话、工具调用、命令输出、composer 和停止按钮。
- 右栏展示 SDK runtime 状态、OpenCode server/session、文件变更、最近命令和错误详情。
- 内置三类引擎：Claude Agent SDK、Codex SDK、OpenCode SDK/CLI。
- OpenCode 由 Main Runtime 托管：启动本地 server、创建/恢复 session、通过 CLI authenticated API 发送 prompt、轮询 session messages/diff 并归一化为 DreamWorker stream events。
- OpenCode 配置写入项目根目录 `opencode.json`，只包含 provider/env 占位与模型映射，不写入密钥；真实密钥通过 Runtime env 注入。
- SDK 包作为 DreamWorker 固有能力随安装包分发，运行时不执行 `npm install`。
- 文件读取和写入强制限制在当前项目 `workspace/code` 内。

## 架构

```text
Electron Desktop
  |-- Renderer: Vue / Pinia UI only
  |-- Preload: typed window.dreamworker API
  `-- Main: lifecycle, embedded Runtime, IPC bridge, stream event proxy

Main Runtime
  |-- In-memory route dispatch and stream generators
  |-- Workspace Store backed by SQLite workspace_state snapshot
  |-- Model provider/profile resources
  |-- Project/module/local directory management
  |-- Requirements, chat, and coding services
  |-- Coding SDK runtime: Claude Agent, Codex, OpenCode
  |-- Extension and MCP management
  `-- Runtime diagnostics
```

## 常用命令

```powershell
npm install
npm run dev
npm run typecheck
npm test
npm run build
npm run runtime:check
npm run package:win
```

常用脚本：

- `npm run runtime:check`：校验 Main Runtime 的编码 SDK import 和 OpenCode CLI。
- `npm run security:smoke`：检查桌面安全边界。
- `npm run llm:smoke`：运行 DeepSeek 真实模型 smoke。
- `DREAMWORKER_OPENCODE_SMOKE=1 npm --workspace @dreamworker/desktop run test -- main/runtime/opencode-smoke.test.ts`：用本地 fake OpenAI-compatible provider 验证 OpenCode server/session/prompt/event/diff 链路。

## 打包

Windows 安装包包含：

- Electron Main 内嵌 Runtime 编译产物。
- Claude Agent SDK、Codex SDK、OpenCode SDK/CLI、OpenAI-compatible adapter 及 optional runtime assets 的 desktop production dependencies。
- OpenCode CLI native assets 通过 `asarUnpack` 解包，确保安装包内可执行。
- 根目录 `.agent` 能力资源。

`npm run runtime:check` 会校验三家编码 SDK import 和 OpenCode CLI 启动能力；任一能力缺失会直接失败。

## 仓库结构

- `.agent/`：skill source 和 capability resources。
- `apps/desktop/`：Electron + Vue 桌面应用。
- `apps/desktop/main/runtime/`：Main 内嵌 Runtime，按 coding、store、shared 模块分层。
- `scripts/`：仓库脚本和打包辅助。
- `specs/`：版本化 JSON schemas、fixtures 和 generated contracts。
