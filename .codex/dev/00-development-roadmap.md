# 00 Development Roadmap

| Field       | Value                                                                          |
| ----------- | ------------------------------------------------------------------------------ |
| Status      | Active                                                                         |
| Owner       | Tech Lead                                                                      |
| Priority    | P0                                                                             |
| DependsOn   | `.codex/plans/11-roadmap.md`, `.codex/plans/12-ai-os-runtime-resource-chat.md` |
| ExitGate    | 每个阶段都能映射到 PR、测试和 release gate                                     |
| PR Range    | PR-00-*                                                                        |
| Risk Level  | High                                                                           |
| Last Review | 2026-07-01                                                                     |

## 目标

把开发路线从“孵化器 MVP”重构为“AI OS + Agent Runtime + 项目孵化系统”的工程执行计划。当前阶段优先完成 Resource Center、Chat Workspace、Provider System、Agent Runtime contract 和 Project Isolation。

MVP 在本文里只表示第一条完整闭环，不允许用半成品 UI 或省 token 的浅实现冒充。

## 输入文档

- `.codex/plans/11-roadmap.md`
- `.codex/plans/12-ai-os-runtime-resource-chat.md`
- `.codex/plans/03-architecture-blueprint.md`
- `.codex/plans/05-capability-bus.md`
- `.codex/plans/07-uiux-interaction-spec.md`
- `.codex/plans/09-security-policy.md`
- `.codex/dev/13-resource-chat-runtime-ux.md`

## 阶段依赖

```text
00 roadmap
  -> 01 repo bootstrap
  -> 02 specs/contracts
  -> 03 engine foundation
  -> 05 capability policy runtime
  -> 06 model agent runtime
  -> 07 desktop workspace UIUX
  -> 13 resource chat runtime UX
  -> 04 incubator domain runtime
  -> 08 project incubation E2E
  -> 10 observability eval hardening
  -> 11 release packaging

09 open source accessibility waits until capability boundaries are stable.
12 risk register starts immediately and is updated every phase.
```

## P0 当前主线

1. Resource Center
   - Provider CRUD、masked key、status、capabilities、auto fetch models。
   - Model profiles、Agent、Skill、Tool、MCP 统一工作台。
   - 不暴露 raw key，不在 Renderer 落状态。

2. Chat Workspace
   - 会话历史、Agent 选择、模型选择、项目绑定。
   - 发送消息进入 Agent Runtime stub。
   - 显示 runtime steps、tool call preview、trace_id、memory scope。

3. Agent Runtime Contract
   - Agent 必须包含 `runtimeConfig`、`planner`、`executor`、`memoryScope`。
   - 所有复杂任务必须经过 `PLAN -> GRAPH -> EXECUTE -> OBSERVE -> REPLAN`。

4. Capability Boundary
   - Skill = 思考策略。
   - Tool = 执行能力。
   - MCP = 外部系统能力。
   - Policy/Approval 控制高风险动作。

## 里程碑

| Milestone        | 目标                             | ExitGate                                | Dev   |
| ---------------- | -------------------------------- | --------------------------------------- | ----- |
| M0 Plan Lock     | 计划重构完成                     | 旧 roadmap/dev 不再冲突                 | 00    |
| M1 Runtime Shell | Electron + Go Engine + typed API | Renderer 只访问 `window.dreamworker`    | 01/03 |
| M2 Contracts     | Provider/Agent/Chat contract     | TS/Go 类型一致                          | 02/13 |
| M3 Resource Core | 资源中心可读写                   | Provider/Agent/Skill/Tool/MCP 经 Engine | 07/13 |
| M4 Chat Core     | 聊天接入 Runtime                 | execution steps + tool preview          | 06/13 |
| M5 Project Core  | 四大模块接 Runtime               | projectId 隔离和 artifact 输出          | 04/08 |
| M6 Hardening     | 可诊断可发布                     | CI/build/security/eval 通过             | 10/11 |

## PR 拆分

- PR-00-01：重构 plans/dev roadmap 和入口索引。
- PR-02-01：补 Provider、Agent、Chat runtime contract。
- PR-06-01：实现 Agent Runtime stub 和 task graph event。
- PR-07-01：整理桌面信息架构。
- PR-13-01：实现 Provider status/capability/model discovery。
- PR-13-02：实现 Chat history/agent/model/project binding。
- PR-13-03：实现 execution steps/tool call preview UI。
- PR-13-04：补齐 store/preload/main/go tests。

## 发布门禁

- `npm run lint`
- `npm run format:check`
- `npm run specs:check`
- `npm run typecheck`
- `npm test`
- `npm run go:fmt:check`
- `npm run go:test`
- `npm run go:vet`
- `npm run security:smoke`
- `npm run build`

## 禁止事项

- 禁止只交 UI 截图。
- 禁止跳过 typed API。
- 禁止让 Renderer 直接访问外部模型或 MCP。
- 禁止把 unknown Skill/MCP 默认为 trusted。
- 禁止把项目、会话、artifact 跨 projectId 混用。

## 验收

- 开发路线能从本文直接映射到 dev 01-13。
- P0 主线先 Resource + Chat + Runtime，再进入项目孵化深水区。
- 每个阶段有明确 PR、测试、风险和回滚边界。
