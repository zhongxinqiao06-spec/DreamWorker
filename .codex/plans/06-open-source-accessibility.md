# 06 Open Source Accessibility

## 目标

定义开放接入体系，让第三方能够以 specs、SDK、adapter examples、conformance tests、capability registry、trust level 和 versioned manifest 接入 DreamWorker，同时不破坏安全边界。

## 非目标

- 不承诺 MVP 建立完整 marketplace。
- 不允许第三方 adapter 绕过 manifest、PolicyEngine 和 EventStore。
- 不把开放接入等同于默认信任。
- 不要求所有 adapter 必须运行在 Go 进程内。

## 核心对象

- specs/：开放协议和 manifest 规范。
- SDK：JS、Python、Go。
- adapter examples：MCP、A2A、HTTP/OpenAPI、Skill、Human Task。
- conformance tests：接入兼容性测试。
- capability registry：本地和未来远程注册中心。
- trust level：信任等级。
- versioned manifest：带 apiVersion、kind、metadata.version 的能力声明。

## 数据结构示例

```text
specs/
  capability-manifest-v1.md
  event-stream-v1.md
  approval-card-v1.md
  artifact-v1.md
sdks/
  js/
  python/
  go/
adapters/
  examples/
    mcp-github/
    a2a-research-agent/
    openapi-crm/
    skill-market-research/
conformance/
  capability_manifest_test_suite.json
  adapter_invocation_test_suite.json
```

TrustLevel：

```yaml
trustLevel:
  builtin: "由 DreamWorker 发布并随应用打包"
  verified_local: "本地安装且通过签名或校验"
  user_added: "用户手动添加"
  remote_untrusted: "远程发现且未验证"
```

## 关键流程

1. 第三方根据 specs 编写 manifest 和 adapter。
2. 使用 SDK 实现 describeCapability、invokeCapability、streamEvents、healthCheck。
3. 运行 conformance tests。
4. 用户导入 adapter。
5. Registry 校验 manifest、版本、schema 和 trust level。
6. PolicyEngine 根据 trust level 限制权限。
7. 所有调用写入 EventStore。

## MVP 做法

- 先在仓库文档中定义 specs 草案，不发布正式 SDK。
- 提供一个 MCP adapter example 和一个 Skill package example。
- conformance tests 先覆盖 manifest 必填字段、schema 校验、approval 标记、observability 字段。
- trust level 只影响默认权限和是否需要审批。

## 后续扩展

- 发布 `dreamworker-sdk-js`、`dreamworker-sdk-python`、`dreamworker-sdk-go`。
- 建立远程 capability registry。
- 支持 adapter 签名、版本锁定和漏洞公告。
- 支持 marketplace 评分、审核和分发。

## 风险

- 开放过早会把安全问题前置到 MVP。
- SDK 抽象不稳定会导致第三方生态迁移成本。
- conformance tests 只能测协议兼容，不能证明业务安全。
- trust level 展示不清晰会让用户误授权。
