# 05 Capability Bus

## 目标

完善 Capability Bus 规格：所有外部工具、数据、Agent、Skill、模型、浏览器和人工任务都通过版本化 Capability Manifest 注册、校验、授权、启用和审计。MCP 用于 tool/data，A2A 用于 external agent，AG-UI 用于 UI event stream，Skill 用于 packaged capability。

## 非目标

- 不让业务代码直接调用 GitHub、Notion、MCP、A2A 或模型 SDK。
- 不允许未注册 capability 被 Orchestrator 使用。
- 不把 Skill 当成可信代码。
- 不把 AG-UI 用作工具协议。
- 不在 MVP 中支持所有第三方协议细节。

## 核心对象

- CapabilityManifest。
- CapabilityRegistry。
- CapabilityInvoker。
- CapabilityLifecycle。
- TrustLevel。
- RiskClassification。
- ApprovalPolicy。
- RuntimePolicy。
- ObservabilitySpec。

## 数据结构示例

```yaml
apiVersion: capability.dreamworker.dev/v1
kind: Capability
metadata:
  id: cap_github_issues_draft
  name: GitHub Issues Draft
  version: 0.1.0
  provider: builtin
protocol:
  type: mcp
  transport: stdio
inputSchema:
  type: object
  required: ["title", "body"]
  properties:
    title: { type: string }
    body: { type: string }
outputSchema:
  type: object
  properties:
    draft_id: { type: string }
permissions:
  scopes: ["github:issues:write"]
  filesystem: "none"
  network:
    allow: ["github.com"]
risk:
  level: medium
  reasons: ["external_side_effect"]
approval:
  requiredWhen: ["write_operation", "external_side_effect"]
runtime:
  timeoutMs: 60000
  retry:
    maxAttempts: 2
observability:
  logInputs: "summary"
  logOutputs: "summary"
  metrics: ["latency_ms", "cost_usd", "error_code"]
```

Lifecycle：

```text
discovered -> registered -> schema_validated -> risk_classified -> authorized -> enabled -> revoked
```

## 关键流程

1. Adapter discovery 发现外部能力。
2. Registry 创建 discovered 记录。
3. Manifest schema validation。
4. Risk classifier 标注 risk 和 trust level。
5. 用户或 policy 授权。
6. Capability enabled 后进入 Orchestrator 候选。
7. 调用前 PolicyEngine 再次评估。
8. Invocation 结果写入 EventStore 和 observability。
9. 用户、系统或策略可 revoke capability。

## MVP 做法

- Manifest 使用 YAML/JSON，支持 `capability.dreamworker.dev/v1`。
- Registry 存 SQLite。
- 内置 builtin、MCP、model、artifact、browser_readonly。
- TrustLevel 先分 builtin、verified_local、user_added、remote_untrusted。
- 未完成 schema_validated 和 risk_classified 的 capability 不可启用。

## 后续扩展

- 支持签名 manifest。
- 支持远程 registry 和 marketplace。
- 支持 conformance tests 自动认证。
- 支持 OpenAPI 自动转 capability。
- 支持 capability version pinning 和迁移。

## 风险

- Manifest 太复杂会阻碍第三方接入。
- TrustLevel 如果只靠用户选择，会造成误信任。
- schema_validated 不代表语义安全，仍需 PolicyEngine。
- revoked 后已有 Run 的 replay 和 audit 要保持可读。
