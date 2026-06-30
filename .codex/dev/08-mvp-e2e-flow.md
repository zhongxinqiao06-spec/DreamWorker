# 08 MVP E2E Flow

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | MVP Integration |
| Priority | P0 |
| DependsOn | 04, 05, 06, 07 |
| ExitGate | Seed idea produces required artifacts, eval report and UI timeline |
| PR Range | PR-08-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

定义并实现端到端 MVP demo flow：用户输入“我想做一个面向独立开发者的 AI 项目孵化工具。”，系统完成 Mission、Discover、Validate、Shape、Artifact、Evaluator、Decision Gate 和 UI 展示。

## 非目标

- 不自动写生产代码。
- 不真实创建 GitHub issue。
- 不自动发布社媒。
- 不调用未知远程 MCP。
- 不安装未验证 Skill。
- 不做多人协作。
- 不做 marketplace。

## 输入文档

- `.codex/plans/02-mvp-scope.md`
- `.codex/dev/04-incubator-domain-runtime.md`
- `.codex/dev/06-model-agent-runtime.md`
- `.codex/dev/07-desktop-workspace-uiux.md`

## 依赖阶段

依赖 `04`、`05`、`06`、`07`。

## 核心产物

- MVP demo seed idea。
- End-to-end run script 或 fixture。
- MVP artifact set。
- Eval report。
- UI demo path。

## 工程任务

完整 demo flow：

1. Create Mission。
2. Discover Stage。
3. 生成 Dream Brief。
4. 生成核心 Hypotheses。
5. Validate Stage。
6. 生成 Evidence。
7. 生成 Research Pack。
8. Shape Stage。
9. 生成 MVP Scope。
10. 生成 Blueprint。
11. 生成 PRD draft。
12. 生成 Launch Checklist。
13. Evaluator 打分。
14. Decision Gate 输出 continue / pivot / pause / kill / ask_user。
15. UI 展示 Run Timeline、Artifact、Evidence、Cost、Risk。

产物：

- `dream_brief.md`
- `hypotheses.yaml`
- `evidence_graph.yaml`
- `research_pack.md`
- `mvp_scope.md`
- `blueprint.yaml`
- `prd.md`
- `launch_checklist.md`
- `eval_report.yaml`

Demo script：

- Start app in stub model mode.
- Create Mission with seed idea.
- Run Discover / Validate / Shape.
- Approve one medium-risk artifact write if required.
- Open Evidence Drawer for one Hypothesis.
- Open each required artifact.
- Open Eval report.
- Verify Decision Gate output.

Seed data checklist：

- Seed idea text is fixed.
- Stub model outputs are deterministic.
- Evidence fixtures include at least one supporting and one weak evidence item.
- Cost/risk fixtures produce visible Cost/Risk Panel state.
- Approval fixture includes one approval card.

Golden output review gate：

- Required artifacts exist.
- Artifact schemas validate.
- Evidence refs resolve.
- Evaluator scores are present.
- No unsupported external action occurred.
- Output diff against baseline is explainable.

## 数据结构 / 接口 / schema 影响

E2E fixture：

```yaml
seed:
  idea: "我想做一个面向独立开发者的 AI 项目孵化工具。"
  expected_stages: [discover, validate, shape]
  expected_artifacts:
    - dream_brief.md
    - hypotheses.yaml
    - evidence_graph.yaml
    - research_pack.md
    - mvp_scope.md
    - blueprint.yaml
    - prd.md
    - launch_checklist.md
    - eval_report.yaml
```

## 测试要求

- E2E tests：
  - create mission。
  - run Discover / Validate / Shape。
  - approve capability。
  - view artifact。
  - evaluator produces score。
- Golden tasks：
  - 5 sample ideas。
  - stable artifact outputs。
  - eval regression。
- Security smoke：
  - high-risk action requires approval。
  - secret not present in event stream。

## 验收标准

- Seed idea 能完整跑通。
- 三阶段都有 Hypothesis、Evidence、Decision。
- 9 个 artifact 至少生成草稿。
- Eval report 有 artifact score、evidence quality score、hallucination risk、actionability score。
- UI 能展示 Run Timeline、Artifact、Evidence、Cost、Risk。
- trace_id 贯穿全链路。
- Demo script can be run by a fresh engineer without hidden setup.
- Golden output review gate passes with deterministic stub mode.

## Codex PR 拆分建议

- PR-08-01: 添加 MVP seed data 和 E2E fixture。
- PR-08-02: 实现 Create Mission -> Discover Stage E2E。
- PR-08-03: 实现 Hypotheses + Evidence + Validate Stage。
- PR-08-04: 实现 MVP Scope + Blueprint + Shape Stage。
- PR-08-05: 生成 PRD 和 Launch Checklist artifact。
- PR-08-06: 接入 Evaluator report。
- PR-08-07: UI 展示完整 demo path。
- PR-08-08: 添加 5 个 golden tasks 和 regression baseline。

## 风险

- Demo flow 过度拟合一个 idea，需要 golden tasks 扩展。
- Artifact 草稿质量不稳定，需要 Eval gate。
- E2E 如果依赖真实模型，测试会不稳定；MVP 应支持 stub mode。

## 暂不做

- 不接真实 GitHub issue 写入。
- 不做社媒发布。
- 不调用远程 untrusted MCP。
