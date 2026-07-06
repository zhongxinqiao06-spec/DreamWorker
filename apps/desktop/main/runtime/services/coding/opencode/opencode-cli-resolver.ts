import { existsSync } from 'node:fs'
import { dirname, join, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

export type OpenCodeCommand = {
  command: string
  argsPrefix: string[]
  displayPath: string
}

export function resolveOpenCodeCli(): string {
  return resolveOpenCodeCommand()?.displayPath ?? ''
}

export function resolveOpenCodeCommand(): OpenCodeCommand | null {
  const executable = resolveOpenCodeExecutable()
  if (executable) {
    return { command: executable, argsPrefix: [], displayPath: executable }
  }
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  if (!packageDir) {
    return null
  }
  const cliPath = join(packageDir, 'bin', 'lildax')
  const unpackedCliPath = existingAsarAwarePath(cliPath)
  return unpackedCliPath
    ? { command: process.execPath, argsPrefix: [unpackedCliPath], displayPath: unpackedCliPath }
    : null
}

export function resolveOpenCodeExecutable(): string {
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  if (!packageDir) {
    return ''
  }
  const nodeModulesDir = findAncestorNodeModules(packageDir)
  if (!nodeModulesDir) {
    return ''
  }
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
    const executable = existingAsarAwarePath(candidate)
    if (executable) {
      return executable
    }
  }
  return ''
}

export function nodePathForOpenCode(): string {
  const packageDir = findNodePackageDir('@opencode-ai/cli')
  const nodeModulesDir = packageDir ? findAncestorNodeModules(packageDir) : ''
  return [nodeModulesDir, process.env.NODE_PATH]
    .filter(Boolean)
    .join(process.platform === 'win32' ? ';' : ':')
}

export function runtimeRoot(): string {
  return resolve(dirname(fileURLToPath(import.meta.url)), '..', '..', '..')
}

export function findNodePackageDir(packageName: string): string {
  const segments = packageName.split('/')
  const starts = [runtimeRoot(), process.cwd(), dirname(fileURLToPath(import.meta.url))]
  for (const start of starts) {
    let current = resolve(start)
    for (;;) {
      const candidate = join(current, 'node_modules', ...segments)
      if (existsSync(join(candidate, 'package.json'))) {
        return candidate
      }
      const parent = dirname(current)
      if (parent === current) {
        break
      }
      current = parent
    }
  }
  return ''
}

export function findAncestorNodeModules(path: string): string {
  let current = resolve(path)
  for (;;) {
    if (current.endsWith(`${sep}node_modules`)) {
      return current
    }
    const parent = dirname(current)
    if (parent === current) {
      return ''
    }
    current = parent
  }
}

function existingAsarAwarePath(path: string): string {
  if (existsSync(path)) {
    return path
  }
  const unpacked = path.replace(`${sep}app.asar${sep}`, `${sep}app.asar.unpacked${sep}`)
  return unpacked !== path && existsSync(unpacked) ? unpacked : ''
}
