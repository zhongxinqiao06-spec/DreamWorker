# DreamWorker Codex Memory

`.codex/` 是 DreamWorker 的项目记忆入口，用来存放稳定计划、工程规则、阶段开发重点和 Agent 协作约束。它不是运行时代码目录；运行时代码以 `apps/desktop/`、`engine/`、`.agent/`、`specs/` 为事实源。

## 当前状态

项目已经进入可运行工程阶段：

- Electron + Vue 桌面工作台已接入 Go Engine sidecar。
- Resource Center 已管理 Provider、Profile、Agent、Skill、Tool、MCP。
- Chat Workspace 已打通真实流式模型闭环、context pack、tool loop、retry/cancel、runtime inspector。
- `.agent/skills` 已成为 Skill 源，Engine 启动自动扫描。
- Windows dir package 已能包含 Engine exe 和 `.agent`。

## 目录

- `plans/`：产品定位、架构蓝图、能力总线、安全策略、路线图。
- `dev/`：可执行开发阶段说明，当前重点是 Resource/Chat/Runtime 工业化。
- `rules/`：开发必须遵守的产品、工程、Agent、UI/UX 规则。
- `skills/`：历史项目内 Agent 能力说明；新的运行时 Skill 以根目录 `.agent/skills` 为准。
- `tmp/`：外部参考缓存，不是 DreamWorker 源码，不进入提交。

## 事实源优先级

1. 根目录 `README.md`。
2. 当前代码：`apps/desktop/`、`engine/`、`.agent/`、`specs/`。
3. `.codex/dev/` 当前阶段开发计划。
4. `.codex/plans/` 长期规划。
5. `.codex/rules/` 工程和 UX 规则。

## 参考边界

Cherry Studio、code-q 或其他外部项目只能用于 UX/架构参考，不能直接复制代码，不能把参考项目的 README 当成 DreamWorker 事实源。

## 命名约定

项目名统一写作 `DreamWorker`。历史资料中出现的拼写漂移不得继续扩散到新增文档、代码或 UI。
