# 05 Capability Policy Runtime

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Security/Capability |
| Priority | P0 |
| DependsOn | 03, 04 |
| ExitGate | Capability lifecycle, policy and approval smoke tests pass |
| PR Range | PR-05-* |
| Risk Level | Critical |
| Last Review | 2026-06-30 |

## 目标

实现 Capability Manifest v1、Capability Registry、Capability lifecycle、TrustLevel、CapabilityInvoker、PolicyEngine 和 Approval runtime，确保所有外部能力受控调用。

## 非目标

- 不实现完整 marketplace。
- 不接未知远程 MCP。
- 不安装未验证 Skill。
- 不让 approval gate 只存在 UI。

## 输入文档

- `.codex/plans/05-capability-bus.md`
- `.codex/plans/09-security-policy.md`
- `.codex/dev/02-specs-contracts.md`
- `.codex/dev/03-engine-foundation.md`

## 依赖阶段

依赖 `03-engine-foundation.md`，可与 `04-incubator-domain-runtime.md` 后半并行。

## 核心产物

- Capability Manifest v1。
- Capability lifecycle。
- TrustLevel。
- PolicyEngine。
- Approval events。
- MVP builtin capabilities。

## 工程任务

- Manifest 字段：
  - apiVersion
  - kind
  - metadata
  - protocol
  - inputSchema
  - outputSchema
  - permissions
  - risk
  - approval
  - runtime
  - observability
- kind：
  - builtin
  - mcp_tool
  - a2a_agent
  - skill
  - openapi
  - browser
  - human
  - webhook
  - model
- lifecycle：
  - discovered
  - registered
  - schema_validated
  - risk_classified
  - authorized
  - enabled
  - disabled
  - revoked
  - deprecated
- trust level：
  - trusted_builtin
  - verified_partner
  - community
  - local_unverified
  - remote_untrusted
- PolicyEngine result：
  - allow
  - deny
  - requires_approval
- 高风险动作：
  - external_write
  - file_write_outside_project
  - secret_access
  - network_untrusted
  - paid_call
  - code_execution
  - browser_action
  - publish_content
  - send_email
  - install_skill
  - connect_remote_mcp
- MVP builtin:
  - artifact_read
  - artifact_write
  - web_search_stub
  - browser_readonly_stub
  - model_generate_stub
  - human_input

Threat model：

| Threat | Attack Surface | Required Control |
| --- | --- | --- |
| Prompt injection from tool data | MCP/browser/model outputs | untrusted content boundary + evaluator |
| Secret exfiltration | capability input/log/event | SecretStore refs + redaction |
| Unauthorized external write | webhook/openapi/mcp write | PolicyEngine approval |
| Cost runaway | model/browser/remote agent | budget + timeout + cancellation |
| Malicious skill | skill install/runtime | trust level + sandbox + approval |
| Revoked capability reuse | cached task graph | invocation-time lifecycle check |

Least-privilege matrix：

| Trust Level | Network | Filesystem | Secret | Default Approval |
| --- | --- | --- | --- | --- |
| trusted_builtin | allowlisted | project-only | reference-only | risk-based |
| verified_partner | allowlisted | project-only | reference-only | medium+ |
| community | restricted | artifact-only | denied | all external side effects |
| local_unverified | deny by default | artifact-only | denied | all tool calls |
| remote_untrusted | deny by default | none | denied | cannot enable without authorization |

Capability abuse cases：

- capability claims read-only but produces external side effect。
- manifest schema valid but protocol endpoint changes behavior。
- revoked capability remains referenced by an existing task。
- model-generated args include hidden secret or path traversal。
- browser action tries to submit form or click purchase button。

## 数据结构 / 接口 / schema 影响

Capability invocation event：

```json
{
  "type": "capability.invocation_requested",
  "payload": {
    "capability_id": "artifact_write",
    "lifecycle_state": "enabled",
    "trust_level": "trusted_builtin",
    "risk_actions": ["external_write"],
    "policy_result": "requires_approval"
  }
}
```

Policy request:

```json
{
  "capability_id": "cap_browser_readonly",
  "actor": "competitor_analyst",
  "risk_actions": ["network_untrusted"],
  "trace_id": "tr_001"
}
```

## 测试要求

- Go unit:
  - capability lifecycle transition。
  - policy decision。
  - trust level default。
- Go integration:
  - CapabilityRegistry SQLite。
  - CapabilityInvoker stub。
  - Approval pause/resume。
- Contract tests:
  - capability manifest schema。
  - approval behavior。
- Security smoke:
  - revoked capability cannot run。
  - high-risk action requires approval。
  - artifact write stays inside project。

## 验收标准

- 未注册 capability 不能调用。
- 未 schema_validated capability 不能 enabled。
- revoked capability 不能调用。
- 所有 invocation 写 EventStore。
- approval gate 在 Go Engine 生效。
- Orchestrator 只能通过 CapabilityInvoker 调用能力。
- Threat model controls mapped to policy tests.
- Least-privilege defaults applied by trust level.
- Abuse cases covered by fixtures or smoke tests.

## Codex PR 拆分建议

- PR-05-01: 实现 Capability Manifest v1 schema 和 fixtures。
- PR-05-02: 实现 CapabilityRegistry 和 lifecycle。
- PR-05-03: 实现 TrustLevel 和 risk classification。
- PR-05-04: 实现 PolicyEngine allow/deny/requires_approval。
- PR-05-05: 实现 ApprovalRequest/ApprovalResolved events。
- PR-05-06: 实现 MVP builtin capabilities。
- PR-05-07: 添加 revoked/high-risk/artifact boundary security smoke。

## 风险

- 权限默认过宽会破坏安全基础。
- trust level 语义不清会让用户误授权。
- UI approval 如果不和 Engine gate 绑定会形成假安全。

## 暂不做

- 不实现真实 MCP client。
- 不实现 A2A。
- 不实现 Skill sandbox。
- 不实现 OpenAPI adapter。
