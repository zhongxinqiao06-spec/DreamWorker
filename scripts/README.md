# DreamWorker Scripts

`scripts/` 放仓库级工程脚本，主要服务 contract generation、CI gate、安全 smoke 和 Windows 打包。

## 当前脚本

- `generate-contracts.mjs`：从 `specs/*.schema.json` 生成 TypeScript/Go contracts。
- `validate-specs.mjs`：校验 JSON Schema 和 valid/invalid fixtures。
- `format-check.mjs`：Prettier 格式检查。
- `go-fmt-check.mjs`：Go 文件 gofmt 检查。
- `security-smoke.mjs`：检查 Renderer/Main/Preload 安全边界，防止 Node、secret、raw IPC 暴露。
- `build-engine.mjs`：构建 `engine/bin/dreamworker-engine.exe`，供 Electron 打包。
- `deepseek-smoke.mjs`：真实模型 smoke，可用于本地 Provider 验证。

## 根目录命令

- `npm run dev`：启动 Electron 开发环境。
- `npm run lint`：ESLint。
- `npm run format:check`：格式检查。
- `npm run specs:check`：contracts 生成检查 + specs validate。
- `npm run typecheck`：Vue/TypeScript 类型检查。
- `npm test`：desktop Vitest。
- `npm run go:test`：Go tests。
- `npm run go:vet`：Go vet。
- `npm run go:build`：Go 编译校验，不产出 exe。
- `npm run go:build:exe`：生成 Engine exe。
- `npm run build`：完整构建。
- `npm run package:win`：Windows unpacked package。
- `npm run ci`：完整门禁。

## 约束

- 生成类脚本必须支持 `--check`，CI 中不能悄悄改文件。
- 脚本输出错误要短、明确、可定位。
- 涉及 secret 的 smoke 不打印原始值，只检查泄露模式。
- Windows 打包默认包含 `engine/bin/dreamworker-engine.exe` 和根目录 `.agent`。
