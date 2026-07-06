import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from 'node:fs'
import { basename, dirname, join, relative, sep } from 'node:path'
import { asString, asStringArray, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../../store/workspace-store'
import type { ProjectDirectoryService } from '../projects/project-directory-service'

export class RequirementImportService {
  constructor(
    private readonly store: WorkspaceStore,
    private readonly projectDirectory: ProjectDirectoryService
  ) {}

  importRequirementFiles(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const root = this.projectDirectory.root(projectId)
    const runId = this.store.nextId('requirements')
    const sources = asStringArray(input.filePaths).map((filePath, index) => {
      const stats = statSync(filePath)
      const name = basename(filePath)
      const targetRelative = `workspace/imports/${runId}/${name}`
      const target = join(root, ...targetRelative.split('/'))
      mkdirSync(dirname(target), { recursive: true })
      writeFileSync(target, readFileSync(filePath))
      return {
        sourceId: `${runId}_${index + 1}`,
        kind: 'imported_file',
        fileName: name,
        relativePath: targetRelative,
        absolutePath: target,
        mimeType: name.toLowerCase().endsWith('.pdf')
          ? 'application/pdf'
          : 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        charCount: Number(stats.size),
        importedAt: nowISO(),
        summary: 'Imported by Main Runtime'
      }
    })
    this.store.save()
    return { projectId, runId, sources, message: `imported ${sources.length} requirement file(s)` }
  }

  listRequirementSources(projectId: string): JsonRecord {
    const root = this.projectDirectory.root(projectId)
    const importsRoot = join(root, 'workspace', 'imports')
    const sources: JsonRecord[] = []
    if (existsSync(importsRoot)) {
      for (const filePath of walkFiles(importsRoot, 200)) {
        const rel = relative(root, filePath).replaceAll(sep, '/')
        const stats = statSync(filePath)
        sources.push({
          sourceId: createHash('sha1').update(rel).digest('hex').slice(0, 12),
          kind: 'imported_file',
          fileName: basename(filePath),
          relativePath: rel,
          absolutePath: filePath,
          mimeType: 'text/plain',
          charCount: Number(stats.size),
          importedAt: stats.mtime.toISOString(),
          summary: rel
        })
      }
    }
    return { projectId, sources }
  }
}

function walkFiles(root: string, limit: number): string[] {
  const result: string[] = []
  const stack = [root]
  while (stack.length > 0 && result.length < limit) {
    const current = stack.pop()
    if (!current) {
      continue
    }
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const fullPath = join(current, entry.name)
      if (entry.isDirectory()) {
        stack.push(fullPath)
      } else {
        result.push(fullPath)
        if (result.length >= limit) {
          break
        }
      }
    }
  }
  return result
}
