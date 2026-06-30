# DreamWorker Codex Memory

`.codex/` 是 DreamWorker 后续开发的项目记忆入口，用来放稳定计划、工程规则和项目内 Agent 能力说明。

当前阶段只落文档和决策，不初始化 Electron / Go 代码骨架。

## 目录

- `plans/`：产品路线、MVP 范围、架构蓝图、能力总线、安全策略。
- `rules/`：后续开发必须遵守的产品、工程、Agent、UI/UX 规则。
- `skills/`：项目内 Agent 能力说明，只定义职责、输入、输出、可用 capability、审批点和质量标准。

## 参考边界

`code-q/` 是参考资产，不是 DreamWorker 的实现目录。可以参考它的文档组织、Electron + Vue + Go sidecar、typed IPC、Agent runtime、MCP/skills 等风格和能力，但不要把 `code-q/` 当成当前产品事实源。

DreamWorker 的事实源优先级：

1. 根目录 `README.md`
2. `.codex/plans/`
3. `.codex/rules/`
4. `.codex/skills/`
5. `code-q/` 参考资料

## 命名约定

项目名统一写作 `DreamWorker`。如历史资料中出现少字母或错字母的变体，视为拼写漂移，新增文档和代码不得继续使用。
