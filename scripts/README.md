# DreamWorker Scripts

`scripts/` 放仓库级工程脚本，主要服务 contract generation、CI gate、安全 smoke、真实模型 smoke 和 Windows 打包。脚本默认从仓库根目录执行，优先保持可定位、可复现、不会在 `--check` 模式悄悄改文件。

## 当前脚本

- `generate-contracts.mjs`：从 `specs/*.schema.json` 生成 TypeScript contracts 和 Go runtime contract subset；支持 `--check`。
- `validate-specs.mjs`：校验 JSON Schema 与 `specs/fixtures/valid|invalid` 样例。
- `format-check.mjs`：收集仓库内受控的 `.md/.json/.ts/.vue/.css/.html/.mjs/.yaml` 等文件并运行 Prettier check。
- `go-fmt-check.mjs`：检查 `engine/` 下 Go 文件是否符合 gofmt。
- `security-smoke.mjs`：检查 Renderer/Main/Preload 安全边界，防止 Node、secret、raw IPC 暴露到 Renderer。
- `build-engine.mjs`：构建 `engine/bin/dreamworker-engine.exe` 或当前平台对应的 Engine 可执行文件，供 Electron 打包。
- `deepseek-smoke.mjs`：真实模型 smoke，可读取 `DEEPSEEK_API_KEY`、`DEEPSEEK_BASE_URL`、`DEEPSEEK_MODEL` 或 `.env.local`。

## 根目录命令

- `npm run dev`：启动 Electron 开发环境。
- `npm run lint`：ESLint 检查 `apps/desktop`、`scripts` 和 `eslint.config.js`。
- `npm run format:check`：Prettier 格式检查。
- `npm run specs:validate`：校验 schemas 与 fixtures。
- `npm run specs:generate`：生成 contracts。
- `npm run specs:check`：先检查 generated contracts 是否最新，再运行 specs validate。
- `npm run typecheck`：运行 desktop web/node TypeScript 类型检查。
- `npm test`：运行 desktop Vitest。
- `npm run go:fmt`：对 `engine/` 执行 gofmt 写入。
- `npm run go:fmt:check`：只检查 gofmt，不写文件。
- `npm run go:test`：运行 `go test ./...`。
- `npm run go:vet`：运行 `go vet ./...`。
- `npm run go:build`：运行 `go build ./...`，不产出 packaged exe。
- `npm run go:build:exe`：生成 Engine exe。
- `npm run package:engine`：同 `go:build:exe`，供打包语义使用。
- `npm run package:win`：完整 build 后输出 Windows unpacked package。
- `npm run security:smoke`：安全边界 smoke。
- `npm run llm:smoke`：DeepSeek 最小连通 smoke。
- `npm run llm:long-task`：DeepSeek 长任务 QA smoke。
- `npm run build`：typecheck + Electron build + Go build + Engine exe。
- `npm run ci`：完整门禁：lint、format、specs、typecheck、Vitest、gofmt、Go test/vet、安全 smoke。

## 环境变量

`deepseek-smoke.mjs` 支持直接读取环境变量，也会读取根目录 `.env.local`：

```text
DEEPSEEK_API_KEY=...
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
```

`.env.local` 不应提交；脚本错误输出会脱敏 provider error，只保留 message/type/code。

## 约束

- 生成类脚本必须支持 `--check`，CI 中不能悄悄改文件。
- 脚本输出错误要短、明确、可定位。
- 涉及 secret 的 smoke 不打印原始值，只检查泄露模式或输出脱敏错误。
- Windows 打包默认包含 `engine/bin/dreamworker-engine.exe` 和根目录 `.agent`。
- 脚本应优先使用标准库和根目录 `devDependencies`，避免为一次性校验引入新 runtime。
