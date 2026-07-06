import { readdirSync, statSync } from 'node:fs'
import { spawnSync } from 'node:child_process'
import { join } from 'node:path'

const roots = [
  'README.md',
  'package.json',
  'tsconfig.base.json',
  'eslint.config.js',
  '.prettierrc.json',
  'apps',
  'examples',
  'scripts',
  'specs',
  '.github'
]

const ignoredDirectories = new Set([
  '.git',
  'code-q',
  'node_modules',
  'out',
  'dist',
  'release',
  'coverage',
  'bin'
])
const supportedExtensions = new Set([
  '.css',
  '.html',
  '.js',
  '.json',
  '.md',
  '.mjs',
  '.ts',
  '.vue',
  '.yaml',
  '.yml'
])

function collectFiles(path, files) {
  let stat
  try {
    stat = statSync(path)
  } catch {
    return
  }

  if (stat.isDirectory()) {
    const name = path.split(/[\\/]/).pop()
    if (name && ignoredDirectories.has(name)) {
      return
    }
    for (const entry of readdirSync(path)) {
      collectFiles(join(path, entry), files)
    }
    return
  }

  const extension = path.includes('.') ? path.slice(path.lastIndexOf('.')) : ''
  if (supportedExtensions.has(extension)) {
    files.push(path)
  }
}

const files = []
for (const root of roots) {
  collectFiles(root, files)
}

if (files.length === 0) {
  console.log('No files found for format check.')
  process.exit(0)
}

for (const batch of batchFiles(files)) {
  const result = spawnSync(
    'npx',
    ['prettier', '--check', '--ignore-path', '.prettierignore', ...batch],
    {
      stdio: 'inherit',
      shell: process.platform === 'win32'
    }
  )
  if ((result.status ?? 1) !== 0) {
    process.exit(result.status ?? 1)
  }
}

process.exit(0)

function batchFiles(paths) {
  const maxCommandChars = process.platform === 'win32' ? 7000 : 120000
  const baseLength = 'npx prettier --check --ignore-path .prettierignore '.length
  const batches = []
  let batch = []
  let length = baseLength
  for (const path of paths) {
    const nextLength = length + path.length + 3
    if (batch.length > 0 && nextLength > maxCommandChars) {
      batches.push(batch)
      batch = []
      length = baseLength
    }
    batch.push(path)
    length += path.length + 3
  }
  if (batch.length > 0) {
    batches.push(batch)
  }
  return batches
}
