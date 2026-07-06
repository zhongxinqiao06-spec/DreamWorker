import {
  existsSync,
  mkdirSync,
  realpathSync,
  readdirSync,
  statSync,
  unlinkSync,
  writeFileSync
} from 'node:fs'
import { dirname, isAbsolute, join, normalize, relative, resolve, sep } from 'node:path'
import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { ProjectRepository } from '../../store/repositories/project-repository'
import { projectDirectoryLayout, projectDocumentStubs } from '../../store/defaults'

export class ProjectDirectoryService {
  constructor(private readonly projects: ProjectRepository) {}

  codeRoot(projectId: string): string {
    const project = this.getProject(projectId)
    const root = this.projectRoot(project)
    const codeRoot = join(root, 'workspace', 'code')
    mkdirSync(codeRoot, { recursive: true })
    const resolvedRoot = safeRealPath(root)
    const resolvedCode = safeRealPath(codeRoot)
    assertInside(resolvedRoot, resolvedCode, 'LOCAL_CODE_DIRECTORY_INVALID')
    if (!directoryWritable(resolvedCode)) {
      throw badRequest(
        'LOCAL_DIRECTORY_NOT_WRITABLE',
        'project localRootPath is not writable',
        'check directory permissions'
      )
    }
    return resolvedCode
  }

  root(projectId: string): string {
    return this.projectRoot(this.getProject(projectId))
  }

  validate(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const check = this.inspectProjectDirectory(project)
    this.projects.save(projectId, {
      ...project,
      localRootPath: check.localRootPath,
      localDirectoryStatus: check.status,
      localDirectoryLastCheckedAt: check.lastCheckedAt,
      updatedAt: nowISO()
    })
    return check
  }

  initialize(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const root = this.projectRoot(project)
    for (const item of projectDirectoryLayout) {
      mkdirSync(join(root, ...item.split('/')), { recursive: true })
    }
    for (const [relativePath, content] of Object.entries(projectDocumentStubs)) {
      const fullPath = join(root, ...relativePath.split('/'))
      if (!existsSync(fullPath)) {
        mkdirSync(dirname(fullPath), { recursive: true })
        writeFileSync(fullPath, content)
      }
    }
    this.writeProjectManifest(project)
    return this.validate(projectId)
  }

  exportManifest(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const manifest = this.projectManifest(project)
    const localRootPath = typeof project.localRootPath === 'string' ? project.localRootPath : null
    if (!localRootPath) {
      return { projectId, localRootPath: null, manifestPath: null, manifest }
    }
    const manifestPath = this.writeProjectManifest(project)
    return { projectId, localRootPath, manifestPath, manifest }
  }

  private getProject(projectId: string): JsonRecord {
    const project = this.projects.get(projectId)
    if (!project) {
      throw notFound('PROJECT_NOT_FOUND', 'project not found', 'refresh project list')
    }
    return project
  }

  private projectRoot(project: JsonRecord): string {
    const root = asString(project.localRootPath)
    if (!root) {
      throw badRequest(
        'LOCAL_DIRECTORY_NOT_SET',
        'project has no localRootPath',
        'bind a local project directory first'
      )
    }
    const resolved = resolve(root)
    if (!existsSync(resolved) || !statSync(resolved).isDirectory()) {
      throw badRequest(
        'LOCAL_DIRECTORY_INVALID',
        'project localRootPath is not an available directory',
        'check project settings'
      )
    }
    return resolved
  }

  private inspectProjectDirectory(project: JsonRecord): JsonRecord {
    const projectId = asString(project.projectId)
    const localRootPath =
      typeof project.localRootPath === 'string' ? normalize(project.localRootPath) : null
    const check: JsonRecord = {
      projectId,
      localRootPath,
      status: 'not_set',
      lastCheckedAt: nowISO(),
      exists: false,
      readable: false,
      writable: false,
      dreamworkerInitialized: false,
      requiredDirectories: [],
      message: '项目尚未绑定本地目录。'
    }
    if (!localRootPath) {
      return check
    }
    if (!existsSync(localRootPath)) {
      return { ...check, status: 'missing', message: '本地目录不存在。' }
    }
    if (!statSync(localRootPath).isDirectory()) {
      return { ...check, status: 'invalid', exists: true, message: '本地路径不是目录。' }
    }
    const requiredDirectories = projectDirectoryLayout.map((entry) => ({
      path: entry,
      exists: existsSync(join(localRootPath, ...entry.split('/')))
    }))
    const readable = directoryReadable(localRootPath)
    const writable = directoryWritable(localRootPath)
    const dreamworkerInitialized = existsSync(join(localRootPath, '.dreamworker'))
    const complete = requiredDirectories.every((entry) => entry.exists)
    let status = 'valid'
    let message = '本地目录可用，项目结构完整。'
    if (!readable || !writable) {
      status = 'permission_denied'
      message = '本地目录读写权限不足。'
    } else if (!dreamworkerInitialized || !complete) {
      status = 'invalid'
      message = '本地目录尚未初始化 DreamWorker 项目结构。'
    }
    return {
      ...check,
      status,
      exists: true,
      readable,
      writable,
      dreamworkerInitialized,
      requiredDirectories,
      message
    }
  }

  private projectManifest(project: JsonRecord): JsonRecord {
    return {
      schemaVersion: 'dreamworker.project.v1',
      exportedAt: nowISO(),
      project,
      directories: projectDirectoryLayout
    }
  }

  private writeProjectManifest(project: JsonRecord): string {
    const root = this.projectRoot(project)
    const metadataDir = join(root, '.dreamworker')
    mkdirSync(metadataDir, { recursive: true })
    writeFileSync(join(metadataDir, 'project.json'), `${JSON.stringify(project, null, 2)}\n`)
    const manifestPath = join(metadataDir, 'manifest.json')
    writeFileSync(manifestPath, `${JSON.stringify(this.projectManifest(project), null, 2)}\n`)
    return manifestPath
  }
}

function directoryReadable(path: string): boolean {
  try {
    readdirSync(path)
    return true
  } catch {
    return false
  }
}

function directoryWritable(path: string): boolean {
  try {
    mkdirSync(path, { recursive: true })
    const probe = join(path, `.dreamworker-write-test-${process.pid}-${Date.now()}`)
    writeFileSync(probe, '')
    unlinkSync(probe)
    return true
  } catch {
    return false
  }
}

function safeRealPath(path: string): string {
  try {
    return realpathSync(resolve(statSync(path).isDirectory() ? path : dirname(path)))
  } catch {
    return resolve(path)
  }
}

function assertInside(root: string, target: string, code: string): void {
  const rel = relative(root, target)
  if (rel === '..' || rel.startsWith(`..${sep}`) || isAbsolute(rel)) {
    throw badRequest(
      code,
      'path resolves outside project root',
      'choose a path inside the project workspace'
    )
  }
}
