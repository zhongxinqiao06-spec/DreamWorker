import { readdirSync, statSync } from 'node:fs'
import { basename, join, relative, sep } from 'node:path'
import { asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'
import type { CodeRootService } from './code-root-service'
import type { GitStatusService } from './git-status-service'

export class FileTreeService {
  constructor(
    private readonly roots: CodeRootService,
    private readonly git: GitStatusService
  ) {}

  list(input: JsonRecord): JsonRecord[] {
    const root = this.roots.resolve(asString(input.projectId))
    const limitValue = typeof input.limit === 'number' ? input.limit : 500
    const limit = limitValue > 0 && limitValue <= 1000 ? limitValue : 500
    const query = asString(input.query).toLowerCase()
    const gitStatus = this.git.statusMap(root)
    const entries: JsonRecord[] = []

    walkProject(root, (path, isDir) => {
      if (entries.length >= limit) {
        return false
      }
      const rel = relative(root, path).replaceAll(sep, '/')
      if (query && !rel.toLowerCase().includes(query)) {
        return true
      }
      const stats = statSync(path)
      entries.push({
        path: rel,
        name: basename(path) || rel,
        isDir,
        size: Number(stats.size),
        modifiedAt: stats.mtime.toISOString(),
        gitStatus: gitStatus.get(rel) ?? ''
      })
      return true
    })

    return entries.sort((left, right) => {
      if (left.isDir !== right.isDir) {
        return left.isDir ? -1 : 1
      }
      return asString(left.path).localeCompare(asString(right.path))
    })
  }
}

function walkProject(root: string, visit: (path: string, isDir: boolean) => boolean): void {
  const stack = [root]
  while (stack.length > 0) {
    const current = stack.pop()
    if (!current) {
      continue
    }
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      if (entry.isDirectory() && shouldSkipDir(entry.name)) {
        continue
      }
      const fullPath = join(current, entry.name)
      if (!visit(fullPath, entry.isDirectory())) {
        return
      }
      if (entry.isDirectory()) {
        stack.push(fullPath)
      }
    }
  }
}

function shouldSkipDir(name: string): boolean {
  return new Set([
    '.git',
    'node_modules',
    'dist',
    'out',
    'release',
    'coverage',
    '.cache',
    '.vite',
    'tmp'
  ]).has(name)
}
