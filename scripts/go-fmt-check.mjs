import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { spawnSync } from 'node:child_process'
import { join, relative } from 'node:path'

if (!existsSync('engine/go.mod')) {
  console.log('engine/go.mod not found; skipping go:fmt:check until PR-01-04')
  process.exit(0)
}

const unformattedFiles = []
for (const filePath of collectGoFiles('engine')) {
  const result = spawnSync('gofmt', [filePath], {
    encoding: 'utf8'
  })

  if (result.error) {
    console.error(result.error.message)
    process.exit(1)
  }

  if (result.status !== 0) {
    console.error(result.stderr.trim())
    process.exit(result.status ?? 1)
  }

  const current = readFileSync(filePath, 'utf8')
  if (normalizeLineEndings(current) !== normalizeLineEndings(result.stdout)) {
    unformattedFiles.push(relative('engine', filePath))
  }
}

if (unformattedFiles.length > 0) {
  console.error('Go files need formatting:')
  console.error(unformattedFiles.join('\n'))
  process.exit(1)
}

process.exit(0)

function collectGoFiles(rootDir) {
  const entries = []
  for (const entry of readdirSync(rootDir)) {
    const entryPath = join(rootDir, entry)
    const stats = statSync(entryPath)
    if (stats.isDirectory()) {
      entries.push(...collectGoFiles(entryPath))
      continue
    }
    if (stats.isFile() && entry.endsWith('.go')) {
      entries.push(entryPath)
    }
  }
  return entries
}

function normalizeLineEndings(content) {
  return content.replace(/\r\n/g, '\n')
}
