import { readFileSync, statSync } from 'node:fs'
import { relative, sep } from 'node:path'
import { badRequest } from '../../../kernel/errors'
import { asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'
import type { CodeRootService } from './code-root-service'
import { safeProjectPath } from './path-sandbox'

const maxReadFileBytes = 512 * 1024

export class FileReadService {
  constructor(private readonly roots: CodeRootService) {}

  read(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const root = this.roots.resolve(projectId)
    const path = safeProjectPath(root, asString(input.path))
    const stats = statSync(path)
    if (stats.isDirectory()) {
      throw badRequest('FILE_IS_DIRECTORY', 'path is a directory', 'select a file')
    }

    const raw = readFileSync(path)
    const truncated = raw.length > maxReadFileBytes
    const payload = truncated ? raw.subarray(0, maxReadFileBytes) : raw
    const rel = relative(root, path).replaceAll(sep, '/')
    return {
      projectId,
      path: rel,
      content: payload.toString('utf8'),
      size: Number(stats.size),
      truncated,
      mimeType: 'text/plain'
    }
  }
}
