import { spawnSync } from 'node:child_process'

const result = spawnSync(
  'npm',
  ['--workspace', '@dreamworker/coding-agent-adapter', 'run', 'build'],
  {
    stdio: 'inherit',
    shell: process.platform === 'win32'
  }
)

process.exit(result.status ?? 1)
