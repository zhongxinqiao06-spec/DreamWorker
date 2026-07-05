import { spawnSync } from 'node:child_process'
import { chmodSync, cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const root = resolve(__dirname, '..')
const adapterRoot = join(root, 'apps', 'coding-agent-adapter')
const adapterDist = join(adapterRoot, 'dist')
const outRoot = join(root, 'out', 'coding-agent-runtime')
const runtimeAdapterRoot = join(outRoot, 'adapter')
const runtimeNodeRoot = join(outRoot, 'node')

if (!existsSync(adapterDist)) {
  run('npm', ['run', 'coding:build'], root)
}

rmSync(outRoot, { recursive: true, force: true })
mkdirSync(runtimeAdapterRoot, { recursive: true })
mkdirSync(runtimeNodeRoot, { recursive: true })

cpSync(adapterDist, join(runtimeAdapterRoot, 'dist'), { recursive: true })

const adapterPackage = {
  name: '@dreamworker/coding-agent-runtime-adapter',
  version: '0.1.0',
  private: true,
  type: 'module',
  main: './dist/index.js',
  dependencies: {
    '@anthropic-ai/claude-agent-sdk': '0.3.201',
    '@openai/codex-sdk': '0.142.5',
    '@opencode-ai/sdk': '1.17.13'
  }
}
writeFileSync(join(runtimeAdapterRoot, 'package.json'), `${JSON.stringify(adapterPackage, null, 2)}\n`)

const nodeTarget = process.platform === 'win32' ? 'node.exe' : 'node'
const nodeBin = join(runtimeNodeRoot, nodeTarget)
cpSync(process.execPath, nodeBin)
if (process.platform !== 'win32') {
  chmodSync(nodeBin, 0o755)
}

run('npm', ['install', '--omit=dev', '--include=optional', '--no-audit', '--no-fund'], runtimeAdapterRoot)

const health = run(nodeBin, [join(runtimeAdapterRoot, 'dist', 'index.js'), '--health-check'], runtimeAdapterRoot, {
  capture: true
})
let healthPayload
try {
  healthPayload = JSON.parse(health.stdout.trim())
} catch {
  throw new Error(`coding runtime health check returned invalid JSON:\n${health.stdout}`)
}
if (!healthPayload.ok) {
  throw new Error(`coding runtime health check failed:\n${JSON.stringify(healthPayload, null, 2)}`)
}

writeFileSync(
  join(outRoot, 'runtime-manifest.json'),
  `${JSON.stringify(
    {
      schemaVersion: 'dreamworker.coding-agent-runtime.v1',
      generatedAt: new Date().toISOString(),
      node: `node/${nodeTarget}`,
      adapter: 'adapter/dist/index.js',
      checks: healthPayload.checks
    },
    null,
    2
  )}\n`
)

console.log(`Packaged coding agent runtime: ${outRoot}`)

function run(command, args, cwd, options = {}) {
  const result = spawnSync(command, args, {
    cwd,
    stdio: options.capture ? 'pipe' : 'inherit',
    encoding: options.capture ? 'utf8' : undefined,
    shell: process.platform === 'win32'
  })
  if (result.status !== 0) {
    const stdout = options.capture ? result.stdout ?? '' : ''
    const stderr = options.capture ? result.stderr ?? '' : ''
    throw new Error(
      `${command} ${args.join(' ')} failed with ${result.status ?? 'unknown'}\n${stdout}${stderr}`
    )
  }
  return {
    stdout: options.capture ? result.stdout ?? '' : '',
    stderr: options.capture ? result.stderr ?? '' : ''
  }
}
