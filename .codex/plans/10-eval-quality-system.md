# 10 Eval Quality System

## 目标

建立 DreamWorker 的质量评估系统，用 golden tasks、artifact score、evidence quality score、hallucination risk 和 regression tests 持续验证孵化结果是否可信、可执行、可复现。

## 非目标

- 不把人工主观满意度作为唯一指标。
- 不用单次模型自评替代系统 eval。
- 不在 MVP 中构建复杂在线实验平台。
- 不把 eval 结果写成不可追踪的临时日志。

## 核心对象

- GoldenTask。
- EvalRun。
- ArtifactScore。
- EvidenceQualityScore。
- HallucinationRisk。
- RegressionTest。
- Rubric。
- FailureCase。

## 数据结构示例

```yaml
golden_task:
  id: gt_indie_marketing_tool
  input_idea: "面向独立开发者的 AI 营销素材生成工具"
  expected_artifacts:
    - dream_brief.md
    - research_pack.md
    - blueprint.yaml
  rubric:
    evidence_quality: 0.35
    blueprint_actionability: 0.30
    artifact_completeness: 0.20
    risk_awareness: 0.15

eval_result:
  golden_task_id: gt_indie_marketing_tool
  artifact_score: 0.81
  evidence_quality_score: 0.74
  hallucination_risk: medium
  regression: false
```

## 关键流程

1. 维护一组 golden tasks 覆盖常见 idea 类型。
2. 每次关键改动后跑 EvalRun。
3. 检查 artifact 是否完整、结构正确、可执行。
4. 检查 evidence 是否有来源、相关性、置信度和反例。
5. 检查 hallucination risk：无来源断言、过度承诺、虚构竞品、假数据。
6. 与 baseline 对比，标记 regression。
7. FailureCase 进入 backlog。

## MVP 做法

- 建立 5-10 个 golden tasks。
- Artifact score 先用结构完整性 + rubric LLM judge + 人工 spot check。
- Evidence quality score 检查 source、confidence、risk、next_best_action。
- Hallucination risk 用规则 + Evaluator Agent 双层检查。
- Regression tests 在文档和 Engine 核心流程实现后进入 CI。

## 后续扩展

- 增加真实用户反馈作为 eval 信号。
- 增加行业分层 golden set。
- 增加 artifact diff quality。
- 增加 cost-quality Pareto 分析。
- 增加 replay-based deterministic eval。

## 风险

- LLM judge 自身可能偏置，必须保留人工抽检。
- golden tasks 太少会过拟合。
- 只评 artifact 不评 evidence 会鼓励漂亮但不可信的输出。
- eval 不进入开发流程会迅速失效。
