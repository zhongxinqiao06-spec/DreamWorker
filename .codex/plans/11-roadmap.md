# 11 Roadmap

## 目标

给出 DreamWorker 从本地 MVP 到开放式项目孵化器操作系统的阶段路线，确保产品、架构、安全、开放接入、可观测性和 eval 同步演进。

## 非目标

- 不把路线写成承诺日期。
- 不在 Phase 1 做平台化。
- 不为了生态开放牺牲本地 MVP 的可信闭环。
- 不把 marketplace 放在安全和 capability 基础之前。

## 核心对象

- Phase。
- Milestone。
- AcceptanceCriteria。
- RiskRegister。
- CapabilitySet。
- EvalGate。
- ReleaseDecision。

## 数据结构示例

```yaml
phase:
  id: phase_1_mvp
  name: "Local Incubator MVP"
  goals:
    - "完成 Mission 到 Blueprint 的闭环"
    - "实现 Evidence Graph 和 Decision Gate"
  acceptance:
    - "Discover/Validate/Shape 可跑通"
    - "每阶段输出 Decision"
    - "所有 tool call 有 trace_id"
    - "高风险动作走 PolicyEngine"
```

## 关键流程

1. 每个 Phase 进入前定义 acceptance。
2. 开发过程中用 plans 和 rules 控制范围。
3. Phase 结束跑 golden tasks 和核心 SLO smoke。
4. ReleaseDecision 使用 continue、pivot、pause、kill、ask_user。
5. 风险进入下一阶段 backlog。

## MVP 做法

Phase 0：Domain and Docs

- 完成孵化器域模型、架构、Capability Bus、安全、UIUX、性能和 eval 计划。
- 明确 `code-q/` 只作参考。

Phase 1：Local Incubator MVP

- Electron workspace + Go Engine skeleton。
- Mission、Stage、Hypothesis、Evidence、Decision、Blueprint、Run、Artifact。
- Discover、Validate、Shape 主流程。
- EventStore、ArtifactStore、PolicyEngine、CapabilityInvoker。

Phase 2：Capability Alpha

- MCP Client。
- Capability lifecycle。
- TrustLevel。
- Approval Diff Card。
- Basic conformance tests。

Phase 3：Execution Alpha

- Build/Launch/Learn 半自动执行。
- A2A external agent。
- Skill Runner。
- Cost/Risk Panel 和 Run Timeline hardening。

Phase 4：Open Platform Beta

- SDK 草案。
- Adapter examples。
- DreamWorker as MCP Server / A2A Server。
- Marketplace 雏形。

Phase 5：Team and Commercial

- Team workspace。
- 权限、审计、计费。
- 私有部署。
- 行业模板和公共案例库。

## 后续扩展

- 云端 Engine。
- 多租户 project workspace。
- 企业 policy profile。
- 行业孵化包。
- 第三方 capability marketplace。
- 跨项目 knowledge graph。

## 风险

- Phase 1 如果不守边界，会滑向大而全 Agent OS。
- Phase 2 如果没有安全基础，开放接入会引入供应链风险。
- Phase 3 自动执行过度会造成用户不信任。
- Phase 4 SDK 过早稳定会锁死错误抽象。
- Phase 5 商业化前如果 eval 不足，质量波动会影响留存。
