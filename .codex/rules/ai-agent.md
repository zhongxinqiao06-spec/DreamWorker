# AI Agent Rules

DreamWorker 的 Agent 是可执行单元，不是随意聊天角色。

## 编排原则

- 默认采用 Orchestrator-mediated multi-agent runtime。
- Agent 之间不自由互聊；协作通过任务图、事件和 Orchestrator。
- Planner 只规划，不执行工具。
- Executor 只执行当前任务节点，不私自改任务图。
- Evaluator 负责质量检查、证据检查和重规划建议。

## Agent Spec

每个 Agent 必须声明：

- id、name、description。
- 负责的任务类型。
- 可用 capability。
- memory scope。
- model profile。
- budget。
- approval policy。
- 输出 artifact 类型。
- 质量标准。

## 任务运行

- 每次 tool call 必须经过 schema 校验、policy 检查和事件记录。
- 证据不足、工具失败、预算超限或用户修改目标时进入 replan。
- 高风险动作必须暂停并请求 approval。
- Agent 输出不得直接作为事实源，必须区分结论、证据、假设和建议。

## 用户控制

用户可以随时：

- 修改目标。
- 调整预算。
- 替换 Agent。
- 启用或禁用 capability。
- 暂停、继续、分支、回滚 run。
- 把任务转成人工任务。
