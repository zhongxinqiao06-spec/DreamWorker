import { existsSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'
import { spawnSync } from 'node:child_process'

const rootDir = process.cwd()
const engineDir = join(rootDir, 'engine')
const goModPath = join(engineDir, 'go.mod')

if (!existsSync(goModPath)) {
  console.error('engine/go.mod not found; cannot package DreamWorker Go Engine.')
  process.exit(1)
}

const targetOS = process.env.GOOS || process.platform
const binaryName =
  targetOS === 'win32' || targetOS === 'windows' ? 'dreamworker-engine.exe' : 'dreamworker-engine'
const outputDir = join(engineDir, 'bin')
const outputPath = join(outputDir, binaryName)

mkdirSync(outputDir, { recursive: true })

const result = spawnSync('go', ['build', '-o', outputPath, './cmd/dreamworker-engine'], {
  cwd: engineDir,
  stdio: 'inherit'
})

if (result.status !== 0) {
  process.exit(result.status ?? 1)
}

console.log(`DreamWorker Go Engine packaged: ${outputPath}`)
