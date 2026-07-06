import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'

const require = createRequire(import.meta.url)

await import('@anthropic-ai/claude-agent-sdk')
await import('@openai/codex-sdk')
await import('@opencode-ai/sdk')
await import('@ai-sdk/openai-compatible')

const cliPackageDir = dirname(require.resolve('@opencode-ai/cli/package.json'))
const command = resolveOpenCodeCommand(cliPackageDir)
const result = spawnSync(command.command, [...command.argsPrefix, '--version'], {
  encoding: 'utf8',
  timeout: 10000,
  windowsHide: true
})

if (result.status !== 0) {
  throw new Error(
    result.stderr.trim() || result.stdout.trim() || 'OpenCode CLI failed to report version'
  )
}

const nineRouterPackageDir = dirname(require.resolve('9router/package.json'))
const nineRouterCliPath = join(nineRouterPackageDir, 'cli.js')
const nineRouterServerPath = resolveNineRouterServer(nineRouterPackageDir)
const nineRouterRuntime = require(join(nineRouterPackageDir, 'hooks', 'sqliteRuntime.js'))
nineRouterRuntime.ensureSqliteRuntime?.({ silent: true })
const nineRouterEnv = nineRouterRuntime.buildEnvWithRuntime?.(process.env) ?? process.env
const nineRouterCommand = {
  command: process.execPath,
  args: [nineRouterCliPath, '--version'],
  displayPath: nineRouterCliPath
}
const nineRouterResult = spawnSync(nineRouterCommand.command, nineRouterCommand.args, {
  encoding: 'utf8',
  timeout: 10000,
  windowsHide: true
})

if (nineRouterResult.status !== 0) {
  throw new Error(
    nineRouterResult.stderr.trim() ||
      nineRouterResult.stdout.trim() ||
      '9Router CLI failed to report version'
  )
}

console.log(
  JSON.stringify(
    {
      ok: true,
      runtime: 'desktop-main-runtime',
      sdkImports: ['claude-agent', 'codex', 'opencode', 'openai-compatible', '9router'],
      opencode: {
        command: command.displayPath,
        version: result.stdout.trim()
      },
      nineRouter: {
        cliCommand: nineRouterCommand.displayPath,
        serverCommand: nineRouterServerPath,
        version: nineRouterResult.stdout.trim(),
        runtimeNodePath: nineRouterEnv.NODE_PATH ?? ''
      }
    },
    null,
    2
  )
)

function resolveNineRouterServer(packageDir) {
  const appDir = join(packageDir, 'app')
  const customServerPath = join(appDir, 'custom-server.js')
  if (existsSync(customServerPath)) {
    return customServerPath
  }
  const serverPath = join(appDir, 'server.js')
  if (existsSync(serverPath)) {
    return serverPath
  }
  throw new Error('9Router bundled server was not found in desktop main package dependencies')
}

function resolveOpenCodeCommand(cliPackageDir) {
  const executable = resolveOpenCodeExecutable(cliPackageDir)
  if (executable) {
    return { command: executable, argsPrefix: [], displayPath: executable }
  }
  const cliPath = join(cliPackageDir, 'bin', 'lildax')
  if (existsSync(cliPath)) {
    return { command: process.execPath, argsPrefix: [cliPath], displayPath: cliPath }
  }
  throw new Error('OpenCode CLI binary was not found in desktop main package dependencies')
}

function resolveOpenCodeExecutable(cliPackageDir) {
  const nodeModulesDir = findAncestorNodeModules(cliPackageDir)
  const platform = process.platform === 'win32' ? 'windows' : process.platform
  const arch = process.arch === 'arm64' ? 'arm64' : process.arch === 'arm' ? 'arm' : 'x64'
  const binary = process.platform === 'win32' ? 'lildax.exe' : 'lildax'
  const packageNames = [
    `cli-${platform}-${arch}`,
    `cli-${platform}-${arch}-baseline`,
    `cli-${platform}-${arch}-musl`,
    `cli-${platform}-${arch}-baseline-musl`
  ]
  for (const packageName of packageNames) {
    const candidate = join(nodeModulesDir, '@opencode-ai', packageName, 'bin', binary)
    if (existsSync(candidate)) {
      return candidate
    }
  }
  return ''
}

function findAncestorNodeModules(path) {
  let current = path
  for (;;) {
    if (current.endsWith(`${process.platform === 'win32' ? '\\' : '/'}node_modules`)) {
      return current
    }
    const parent = dirname(current)
    if (parent === current) {
      throw new Error(`node_modules ancestor not found for ${path}`)
    }
    current = parent
  }
}
