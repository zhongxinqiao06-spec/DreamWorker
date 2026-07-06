# DreamWorker Scripts

`scripts/` 存放仓库级工程脚本，用于契约生成、质量检查、smoke test 和 Windows 打包。

## 当前脚本

- `generate-contracts.mjs`：从 `specs/*.schema.json` 生成 TypeScript contracts。
- `validate-specs.mjs`：校验 JSON Schema fixtures。
- `format-check.mjs`：对仓库文本资产执行 Prettier 检查。
- `security-smoke.mjs`：校验 Renderer/Main/Preload 安全边界。
- `deepseek-smoke.mjs`：通过 `DEEPSEEK_API_KEY`、`DEEPSEEK_BASE_URL`、`DEEPSEEK_MODEL` 或 `.env.local` 运行真实模型 smoke。
- `check-desktop-runtime.mjs`：校验 desktop 主包内置的 Claude Agent SDK、Codex SDK、OpenCode SDK/CLI 和 OpenAI-compatible adapter。

## 根命令

- `npm run dev`：启动 Electron dev，Main 在进程内创建 Runtime。
- `npm run lint`：检查 desktop 和 script 代码。
- `npm run format:check`：执行 Prettier 检查。
- `npm run specs:check`：检查 generated contracts 是否最新，并校验 specs。
- `npm run runtime:check`：校验 Main Runtime SDK 依赖和 OpenCode CLI。
- `npm run typecheck`：检查 desktop TypeScript。
- `npm test`：运行 desktop Vitest。
- `DREAMWORKER_OPENCODE_SMOKE=1 npm --workspace @dreamworker/desktop run test -- main/runtime/opencode-smoke.test.ts`：开启 OpenCode Main Runtime 端到端 smoke。
- `npm run build`：执行 typecheck，并构建 Electron。
- `npm run package:win`：构建并打包 Windows 安装包。
- `npm run ci`：运行 lint、format、specs、typecheck、tests、runtime check 和 security smoke。

## Runtime 打包

生产应用不再携带独立 Runtime 包。Main Runtime 随 `apps/desktop/main/runtime` 编译进 Electron Main，桌面进程直接调用 Node/TypeScript 服务对象，不经过 Go Engine 或本机 HTTP。

Runtime 目录按 `bootstrap`、`router`、`kernel`、`services`、`store/repositories` 分层：装配、路由、生命周期/取消/追踪、业务能力、领域仓储和 SQLite snapshot 持久化分别演进。Router 只面向 service，service 通过 repository 访问 Workspace snapshot。

安装包重点校验：

- desktop production dependencies 中包含三家编码 SDK。
- OpenCode CLI native assets 通过 `asarUnpack` 解包后可执行。
- `.agent` 能力资源通过 `extraResources` 分发。

编码 Agent 依赖的 Claude Agent SDK、Codex SDK、OpenCode SDK/CLI 和 OpenAI-compatible adapter 必须在打包阶段安装并校验成功，用户运行时不再下载 npm 包。
