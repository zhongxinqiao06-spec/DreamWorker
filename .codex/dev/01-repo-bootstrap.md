# 01 Repo Bootstrap

| Field | Value |
| --- | --- |
| Status | Ready for Implementation |
| Owner | Platform/Desktop |
| Priority | P0 |
| DependsOn | 00 |
| ExitGate | Electron app calls Go `runtime.ping` through typed preload API |
| PR Range | PR-01-* |
| Risk Level | High |
| Last Review | 2026-06-30 |

## 目标

初始化 DreamWorker monorepo 和最小可运行 Electron + Go Engine 骨架，建立 Renderer / Preload / Main / Go Engine 边界，确保 Go Engine 可作为 desktop local daemon 和 CLI 独立启动。

## 非目标

- 不实现业务 Agent。
- 不实现真实 Capability adapter。
- 不迁移 `code-q/`。
- 不在 Main 写孵化器业务逻辑。

## 输入文档

- `.codex/plans/03-architecture-blueprint.md`
- `.codex/plans/04-engine-code-skeleton.md`
- `.codex/rules/engineering.md`

## 依赖阶段

依赖 `00-development-roadmap.md`。

## 核心产物

- monorepo 目录结构。
- Electron + Vite + Vue + Pinia app shell。
- Go module 和 `dreamworker-engine` CLI。
- typed preload API skeleton。
- `runtime.ping` smoke test。
- CI 基础任务草案。

## 工程任务

- 建立目录：
  - `apps/desktop/main`
  - `apps/desktop/preload`
  - `apps/desktop/renderer`
  - `engine/cmd/dreamworker-engine`
  - `engine/internal`
  - `specs`
  - `examples`
  - `scripts`
- 包管理建议：优先 npm，先降低工具链复杂度；后续需要 workspace 时再切 pnpm。
- TypeScript 开启 strict。
- 配置 ESLint、Prettier、Go fmt、Go vet。
- Electron 使用 Vite / electron-vite。
- Renderer 只用 Vue + Pinia 管理 UI 状态。
- Main 负责窗口、生命周期、启动 Go Engine、本地 RPC 代理。
- Go Engine 支持：
  - `dreamworker-engine serve`
  - `dreamworker-engine ping`
- Electron 安全默认值：
  - contextIsolation enabled
  - sandbox enabled
  - nodeIntegration disabled
  - remote disabled
  - strict CSP
- `code-q/` 只作为参考，不复制目录。

CI matrix：

| Job | Command | Required |
| --- | --- | --- |
| TypeScript typecheck | `npm run typecheck` | P0 |
| Renderer unit | `npm test` | P0 |
| Go unit | `npm run go:test` | P0 |
| Go vet | `npm run go:vet` | P1 initially, P0 before MVP |
| Build smoke | `npm run build` | P1 initially, P0 before release |
| Security smoke | renderer boundary test | P0 |

Dependency policy：

- 任何新 runtime dependency 必须在 PR 描述说明用途、替代方案和安全影响。
- Electron、Vite、Vue、Pinia、Go stdlib 优先；暂不引入重型框架。
- 新增 native dependency 需要说明打包影响。
- 锁文件必须随 dependency PR 一起提交。

Workspace bootstrap acceptance checklist：

- App shell 可启动。
- Go Engine 可独立 CLI 启动。
- Main 能启动/停止 Engine。
- Renderer 只能通过 preload 调 `runtime.ping`。
- Renderer boundary smoke 证明 `process`、`require`、`fs` 不可访问。

## 数据结构 / 接口 / schema 影响

最小 typed API：

```ts
type RuntimePingResult = {
  ok: true
  engineVersion: string
  traceId: string
}
```

Go CLI 输出：

```json
{"ok":true,"engineVersion":"0.1.0","traceId":"tr_bootstrap"}
```

## 测试要求

- Go unit：`runtime.ping` handler。
- Go integration：CLI `dreamworker-engine ping`。
- Renderer tests：app shell renders。
- Contract smoke：preload API exposes `runtime.ping` only.
- Security smoke：Renderer cannot access `process`, `require`, `fs`。
- CI smoke：typecheck、go test、go vet 可运行。

## 验收标准

- `npm run typecheck` 通过。
- `npm run test` 通过或有空测试配置。
- `npm run go:test` 通过。
- Electron 启动空工作台。
- Main 启动 Go Engine 并完成 `runtime.ping`。
- Renderer 不保存 secret，不直接访问 Node API。
- Go Engine 可脱离 Electron 通过 CLI 启动。
- CI matrix 中 P0 job 全部有脚本或明确 stub。
- dependency policy 已写入 README 或 dev docs。

## Codex PR 拆分建议

- PR-01-01: 初始化 monorepo 目录、package.json、tsconfig strict。
- PR-01-02: 初始化 Electron + Vite + Vue + Pinia 空壳。
- PR-01-03: 建立 preload typed API skeleton，禁用 Renderer Node 能力。
- PR-01-04: 初始化 Go module 和 `dreamworker-engine ping`。
- PR-01-05: Main 启动 Go Engine local daemon 并代理 `runtime.ping`。
- PR-01-06: 加入 ESLint、Prettier、gofmt、go vet、CI 基础任务。
- PR-01-07: 添加 Renderer boundary security smoke。

## 风险

- 复制 `code-q/` 会带入旧领域模型。
- Electron 安全默认值晚做会导致 Renderer 边界返工。
- Main 过早写业务逻辑会破坏 Go Engine 独立运行目标。

## 暂不做

- 不实现 Mission。
- 不实现 SQLite。
- 不实现 EventStore。
- 不实现 Agent Runtime。
