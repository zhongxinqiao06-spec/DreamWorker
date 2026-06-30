# 01 Incubator Domain

## 目标

定义项目孵化器域模型，作为 Go Engine、UI、EventStore、Artifact 和 Eval 的共同语言。所有 Agent 运行、capability 调用和 UI 状态都围绕 Mission、Stage、Hypothesis、Evidence、Experiment、Decision、Blueprint、Run、Artifact 展开。

## 非目标

- 不在域模型中绑定具体模型厂商。
- 不让 domain 依赖 MCP、A2A、SQLite、HTTP 或 Electron。
- 不把 Stage 实现成不可变死流程；阶段可重入、可回滚、可并行补证据。
- 不把 Artifact 当作唯一事实源；事实以 EventStore 和 Evidence Graph 为准。

## 核心对象

- Mission：项目孵化根对象。
- Stage：孵化阶段枚举。
- Hypothesis：可证伪判断。
- Evidence：结构化证据节点。
- Experiment：验证假设的行动。
- Decision：阶段门输出。
- Blueprint：可执行计划。
- Run：一次执行上下文。
- Artifact：文档、schema、issue、文案、报告等交付物。

## 数据结构示例

```yaml
stage:
  id: stg_validate
  mission_id: msn_001
  name: validate
  status: running
  entry_criteria:
    - "Discover 产生至少 3 个核心假设"
  exit_criteria:
    - "每个核心假设至少有 1 条 evidence"
    - "生成阶段 decision"

hypothesis:
  id: hyp_001
  mission_id: msn_001
  stage: validate
  statement: "独立开发者愿意为自动生成营销素材每月付费"
  owner_agent: product_analyst
  status: testing

evidence:
  id: ev_001
  hypothesis_id: hyp_001
  source: "interview:user_003"
  summary: "用户每周花 3 小时制作发布素材"
  confidence: 0.72
  risk: medium
  next_best_action: "继续访谈 5 个同类用户"
```

## 关键流程

1. Mission 创建后进入 Discover。
2. 每个 Stage 生成或继承 Hypothesis。
3. Experiment 负责把 Hypothesis 转成可执行验证动作。
4. Evidence 绑定 Hypothesis，也可绑定 Artifact、Run 和 CapabilityInvocation。
5. Evaluator 对 Evidence Graph 做质量评估。
6. Decision Gate 读取 Stage 状态、Evidence Graph、风险和预算，生成 Decision。
7. Decision 触发下一阶段、pivot、pause、kill 或 ask_user。

## MVP 做法

- 使用固定六阶段枚举：Discover、Validate、Shape、Build、Launch、Learn。
- 每个 Mission 同一时间只有一个主 Stage running，但允许补充前一阶段 Evidence。
- Evidence Graph 先用 SQLite 邻接表或 JSON 列持久化，不引入外部图数据库。
- Decision 作为 append-only event 写入 EventStore，不允许静默覆盖。

## 后续扩展

- 支持多个 Stage 并行，例如 Validate 和 Shape 局部并行。
- 支持 Evidence Graph 可视化和图检索。
- 支持行业模板定义 Stage exit criteria。
- 支持团队角色对 Decision 共同审批。

## 风险

- 域模型过抽象会拖慢开发；MVP 必须围绕首个 idea-to-blueprint 场景落地。
- Evidence 数据质量不足会影响 Decision。
- 如果 Decision 可被随意覆盖，审计和信任会被破坏。
- Stage 规则太硬会让开放式孵化体验变成传统流程工具。
