# DreamWorker Dev Plan

这里是 Codex 执行 DreamWorker 的工程入口。开发按 AI OS + Agent Runtime + 项目孵化系统推进，当前重点是把 Resource Center、Chat Workspace、Model Gateway、Context Manager、Skill/Tool/MCP Runtime、Extension Runtime、AIOS UI 和 Windows packaging 继续夯实为工业级底座。

## 当前完成度

- Repo/Electron/Go bootstrap：已落地。
- Specs/contracts：已落地 schema、fixtures、generation/check。
- Engine foundation：已落地 Go daemon、runtime API、SQLite adapters、EventStore/ArtifactStore/CapabilityRegistry。
- Resource Center：已落地 Provider/Profile/Agent/Skill/Tool/MCP 管理和 provider health/model discovery/stream verification。
- Extension Runtime：已落地 Node-managed extension manager 与 9Router provider bridge 的基础闭环。
- Chat Runtime：已落地 SSE stream、cancel、retry、assistant attempt、context pack、tool preview/execution、audit summary、typed events。
- Model Gateway：已落地 OpenAI Responses、OpenAI-compatible、Anthropic、Ollama streaming adapters。
- Skill：已迁移到根目录 `.agent/skills/<name>/SKILL.md` 自动扫描，内置 `skillcreator`。
- MCP：已落地 stdio 最小闭环；HTTP/SSE MCP 放后续阶段。
- UX：已落地 AI OS 2.0 视觉资产、开屏 Canvas 粒子漩涡、聊天自动下滚、模型思考默认收起、Runtime Inspector。
- Packaging：已落地 Windows unpacked package，包含 Engine exe 和 `.agent`。

## 入口顺序

1. `00-development-roadmap.md`：总路线和当前 P0 主线。
2. `13-resource-chat-runtime-ux.md`：Resource/Chat/Runtime 当前主战场。
3. `06-model-agent-runtime.md`：Model Gateway、Agent loop、Context、Tool runtime。
4. `05-capability-policy-runtime.md`：Tool/MCP/Skill/Policy。
5. `07-desktop-workspace-uiux.md`：桌面信息架构和 UX。
6. `04-incubator-domain-runtime.md`：项目孵化四大模块。
7. `11-release-packaging.md`：发布、打包、safe mode、migration backup 后续门禁。

## 当前 P0 标准

- Resource Center 必须真实驱动 Engine 数据，不允许硬编码假状态。
- Chat Workspace 必须走真实 streaming runtime，不允许 stub 假装成功。
- Agent 必须有 runtimeConfig、planner、executor、memoryScope、enabled skills/tools/MCP。
- Provider 必须支持 masked key、health、model discovery、streaming verification、capability/status。
- Extension 必须通过 typed API 暴露状态、动作、日志和 provider bridge，不允许 Renderer 直接操作进程或 secret。
- Context 必须通过 Context Manager 构建，secret 不进入 prompt/event/log。
- Tool/MCP 必须经过 policy gate；低风险可执行，高风险阻断或审批。
- Retry/cancel/partial/failure 必须有清晰状态和可恢复路径。
- AIOS UI 资产、主题 token 和中文文案要保持一致，不新增纯英文占位。

## 阶段索引

| Dev | 作用                         | 当前定位                         |
| --- | ---------------------------- | -------------------------------- |
| 00  | 总路线                       | P0 入口                          |
| 01  | Repo/Electron/Go bootstrap   | 已完成基础设施                   |
| 02  | Specs/contracts              | 已完成 v0.1 契约门禁             |
| 03  | Engine foundation            | 已完成 daemon/API/store          |
| 04  | Incubator domain runtime     | 下一阶段深化 artifact/eval       |
| 05  | Capability policy runtime    | 已有基础，继续补审批 UI          |
| 06  | Model agent runtime          | 当前持续加固                     |
| 07  | Desktop workspace UIUX       | 当前持续打磨，AIOS 视觉已接入    |
| 08  | E2E flow                     | 待扩展端到端孵化闭环             |
| 09  | Open source accessibility    | SDK/example/conformance 待落地   |
| 10  | Observability eval hardening | trace、eval、diagnostics 待深化  |
| 11  | Release packaging            | Windows dir 已验证，发布门禁待补 |
| 12  | Risk register                | 持续维护                         |
| 13  | Resource chat runtime UX     | 当前主战场                       |

## Definition of Done

- schema、typed API、Go Engine、Renderer、tests 同步。
- UI 通过 Engine 数据驱动，不用描述性小字掩盖缺失交互。
- Renderer 不接触 secret、Engine token、Provider key 或 raw provider response。
- 所有 high-risk Tool/MCP/Skill/Extension 动作经过 Policy/Approval 或明确阻断。
- 流式事件只走统一 `ChatStreamEvent`，Provider 原始事件不透传。
- 生成产物、fixtures 和 README 描述保持一致。
- CI、build、安全 smoke 通过。

## 绝不接受

- 用“MVP”当借口省略核心功能。
- 用 prompt 拼接冒充 Agent Runtime。
- 用配置页冒充 Model Provider System。
- 用硬编码列表冒充 Skill/Tool/MCP/Extension 体系。
- 用成功 toast 掩盖 provider/key/model/network/timeout/rate-limit 真实失败。
- 把 `.codex/tmp` 外部参考或过期计划写成 DreamWorker 已落地事实。
