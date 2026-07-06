import { existsSync, realpathSync } from 'node:fs'
import { dirname, isAbsolute, relative, resolve, sep } from 'node:path'
import { badRequest, notFound } from '../../../kernel/errors'

export function safeProjectPath(root: string, raw: string): string {
  const normalized = raw.trim().replaceAll('\\', '/')
  if (!normalized || normalized.startsWith('/') || normalized.includes('\0')) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'file path must be relative to the project',
      'select a project file'
    )
  }

  const joined = resolve(root, normalized)
  const rel = relative(root, joined)
  if (isOutside(rel)) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'file path escapes the project root',
      'select a file inside the project'
    )
  }
  if (!existsSync(joined)) {
    throw notFound('FILE_NOT_FOUND', 'file not found', 'select another project file')
  }

  const realRoot = realpathSync(root)
  const realPath = realpathSync(joined)
  const realRel = relative(realRoot, realPath)
  if (isOutside(realRel)) {
    throw badRequest(
      'PATH_OUTSIDE_PROJECT',
      'resolved file path escapes the project root',
      'select a file inside the project'
    )
  }
  return realPath
}

export function isInside(root: string, target: string): boolean {
  return !isOutside(relative(root, target))
}

export function safeRealPath(path: string): string {
  try {
    return realpathSync(resolve(path))
  } catch {
    return realpathSync(resolve(dirname(path)))
  }
}

function isOutside(relativePath: string): boolean {
  return relativePath === '..' || relativePath.startsWith(`..${sep}`) || isAbsolute(relativePath)
}
