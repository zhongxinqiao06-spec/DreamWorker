import { existsSync } from 'node:fs'
import { spawnSync } from 'node:child_process'

if (!existsSync('engine/go.mod')) {
  console.log('engine/go.mod not found; skipping go:fmt:check until PR-01-04')
  process.exit(0)
}

const result = spawnSync('gofmt', ['-l', '.'], {
  cwd: 'engine',
  encoding: 'utf8'
})

if (result.error) {
  console.error(result.error.message)
  process.exit(1)
}

const output = result.stdout.trim()
if (output.length > 0) {
  console.error('Go files need formatting:')
  console.error(output)
  process.exit(1)
}

process.exit(0)
