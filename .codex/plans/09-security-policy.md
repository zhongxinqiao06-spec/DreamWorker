# 09 Security Policy

## 目标

定义 DreamWorker 的权限、审批、Secret、沙箱、审计、成本上限和外部能力风险策略。所有高风险 capability 调用必须经过 PolicyEngine，所有决策和调用必须写入 EventStore。

## 非目标

- 不默认信任用户安装的 Skill、MCP server 或远程 Agent。
- 不把安全判断放在 Renderer。
- 不允许 secrets 进入普通 artifact、日志或 event payload 明文。
- 不在 MVP 中承诺完整企业合规认证。

## 核心对象

- PolicyEngine。
- PolicyDecision。
- ApprovalRequest。
- Approval Diff Card。
- SecretBroker。
- SandboxPolicy。
- TrustLevel。
- RiskLevel。
- AuditEvent。
- CostBudget。

## 数据结构示例

```yaml
policy:
  default:
    network: deny
    filesystem: project_only
    shell: deny
    paid_api: ask
    external_post: ask
    credential_access: deny
  approvals:
    requiredFor:
      - write_file_outside_project
      - publish_content
      - send_email
      - call_paid_api
      - install_skill
      - execute_code
      - connect_untrusted_mcp

policy_decision:
  id: pol_001
  trace_id: tr_001
  action: invoke_capability
  capability_id: cap_github_issues_draft
  result: requires_approval
  risk: medium
  reason: "external_side_effect"
```

## 关键流程

1. CapabilityInvoker 调用前创建 PolicyRequest。
2. PolicyEngine 根据 capability manifest、trust level、runtime context、预算和用户设置评估。
3. low risk 可直接 allow 并写 AuditEvent。
4. medium/high risk 生成 ApprovalRequest。
5. UI 展示 Approval Diff Card。
6. 用户 approve/reject/edit/ask_user。
7. ApprovalResolved 写入 EventStore。
8. Invoke 执行或中止。

## MVP 做法

- secrets 存系统 Secret Store，Renderer 不接触明文。
- 文件系统默认 project-only。
- shell 默认 deny。
- 不可信 MCP 和 Skill 默认 ask。
- paid API、external post、write operation 默认 ask。
- 所有日志和错误 detail 使用 sanitizer。

## 后续扩展

- Skill 签名和来源验证。
- 远程 capability 安全评分。
- 企业 policy profile。
- 审计导出。
- 数据分级和 DLP 检查。

## 风险

- 审批过多造成用户疲劳。
- trust level 展示不清晰导致误授权。
- 外部网页和 tool result 可能携带 prompt injection。
- 成本限制如果只在 UI 层实现，无法阻止失控调用。
