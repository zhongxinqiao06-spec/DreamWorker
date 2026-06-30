# 11 Release Packaging

| Field | Value |
| --- | --- |
| Status | Planned |
| Owner | Release Engineering |
| Priority | P1, P0 before MVP distribution |
| DependsOn | 10 |
| ExitGate | Packaged app passes startup, diagnostics, safe-mode and migration backup smoke |
| PR Range | PR-11-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

定义桌面发布计划和发布门禁，确保 MVP 可打包、可诊断、可恢复，并为 signing、auto update、crash report、safe mode 和 migration backup 预留架构。

## 非目标

- MVP 不做真实自动更新。
- MVP 不做真实签名。
- MVP 不接真实 crash report SaaS。
- 不发布云端版本。

## 输入文档

- `.codex/plans/03-architecture-blueprint.md`
- `.codex/plans/08-performance-observability.md`
- `.codex/dev/10-observability-eval-hardening.md`

## 依赖阶段

依赖 `10-observability-eval-hardening.md`。

## 核心产物

- dev build。
- production build。
- signing placeholder。
- auto update placeholder。
- crash report placeholder。
- diagnostics export。
- local data directory。
- migration backup。
- config reset。
- safe mode。
- first-run onboarding。
- release checklist。

## 工程任务

- 配置 dev build 和 production build。
- 打包 Go Engine 到 Electron resources。
- 定义用户数据路径和项目路径：
  - app data dir。
  - project data dir。
  - artifact dir。
  - diagnostics export dir。
- Engine 启动失败 UI fallback：
  - 显示错误码。
  - 提供 retry。
  - 提供 diagnostics export。
  - 提供 safe mode。
- migration 失败策略：
  - backup before migration。
  - restore option。
  - migration log。
- config reset：
  - reset UI settings。
  - reset engine local config。
  - 不删除用户 project artifact，除非用户明确确认。
- first-run onboarding：
  - 选择项目目录。
  - 创建示例 Mission。
  - 说明本地数据和安全边界。
- release checklist：
  - tests pass。
  - security smoke pass。
  - SLO smoke pass。
  - golden tasks pass。
  - no secret in logs/events。

Release train：

- `dev`: local developer build, unsigned, diagnostics verbose.
- `preview`: packaged build for internal smoke, pprof disabled by default.
- `mvp`: user-facing build, release checklist required.
- Hotfix releases must reference failing release gate or risk item.

Rollback plan：

- App config reset does not delete user artifacts.
- Engine migration backup can restore previous DB.
- Feature flags can disable non-critical adapters.
- Release notes must list known incompatible changes.

Safe-mode drill：

- Launch app with Engine disabled.
- Show diagnostics and repair actions.
- Allow config reset.
- Allow project selection without opening last broken run.

Migration backup drill：

- Create test project DB.
- Run migration.
- Simulate failure.
- Restore backup.
- Verify EventStore replay still works.

## 数据结构 / 接口 / schema 影响

Release diagnostics package：

```yaml
diagnostics:
  app_version: 0.1.0
  engine_version: 0.1.0
  platform: windows
  logs: redacted
  traces: redacted
  recent_errors: []
  migrations: []
```

## 测试要求

- Build tests：
  - dev build。
  - production build。
  - Go Engine bundled。
- Smoke:
  - first-run onboarding。
  - Engine startup failure fallback。
  - safe mode。
  - diagnostics export redaction。
  - migration backup/restore mock。
- Release gate:
  - all tests from `10` pass。

## 验收标准

- 可生成本地安装/运行包。
- Go Engine 随应用启动。
- Engine 启动失败时 UI 可恢复。
- migration 前有 backup。
- diagnostics export 可生成且脱敏。
- release checklist 完整。
- Release train states are documented.
- Safe-mode and migration backup drills have smoke steps.

## Codex PR 拆分建议

- PR-11-01: 配置 dev/prod build 和 Go Engine bundle。
- PR-11-02: 定义 local data directory 和 project directory。
- PR-11-03: 实现 Engine startup failure fallback。
- PR-11-04: 实现 diagnostics export。
- PR-11-05: 实现 migration backup/restore skeleton。
- PR-11-06: 实现 safe mode 和 config reset。
- PR-11-07: 实现 first-run onboarding。
- PR-11-08: 添加 release checklist 和 packaging smoke。

## 风险

- 打包路径和开发路径不一致会造成只在 release 崩溃。
- migration 无 backup 会导致用户数据风险。
- diagnostics 如果不脱敏会泄露隐私。

## 暂不做

- 不做真实签名。
- 不做真实自动更新。
- 不做 crash report SaaS。
