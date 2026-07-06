import { createHash } from 'node:crypto'
import {
  existsSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  realpathSync,
  statSync,
  unlinkSync,
  writeFileSync
} from 'node:fs'
import { DatabaseSync } from 'node:sqlite'
import { homedir } from 'node:os'
import { basename, dirname, isAbsolute, join, normalize, relative, resolve, sep } from 'node:path'
import {
  createDefaultSnapshot,
  createProject,
  createProjectModules,
  defaultSettings,
  ensureSnapshotShape,
  projectDirectoryLayout,
  projectDocumentStubs,
  touch
} from './defaults'
import {
  badRequest,
  internalError,
  notFound,
  type DeleteResult,
  type JsonRecord,
  type WorkspaceSnapshot
} from '../types'
import {
  asRecord,
  asString,
  asStringArray,
  maskSecret,
  newTraceId,
  nowISO,
  sortedValues
} from '../shared/util'

const workspaceSnapshotKey = 'workspace'

export class WorkspaceStore {
  readonly configDir: string
  private readonly db: DatabaseSync
  snapshot: WorkspaceSnapshot

  constructor(configDir = defaultConfigDir()) {
    this.configDir = configDir
    mkdirSync(this.configDir, { recursive: true })
    this.db = new DatabaseSync(join(this.configDir, 'workspace.db'))
    bootstrapDatabase(this.db)
    this.snapshot = this.loadSnapshot()
  }

  close(): void {
    this.db.close()
  }

  nextId(prefix: string): string {
    this.snapshot.sequence += 1
    return `${prefix}_${String(this.snapshot.sequence).padStart(3, '0')}`
  }

  save(): void {
    const payload = JSON.stringify(this.snapshot)
    this.db
      .prepare(
        `INSERT INTO workspace_state (key, payload, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET payload = excluded.payload, updated_at = excluded.updated_at`
      )
      .run(workspaceSnapshotKey, payload, nowISO())
  }

  listProviders(): JsonRecord[] {
    return sortedValues(this.snapshot.providers, 'providerId').map((provider) =>
      this.safeProvider(provider)
    )
  }

  saveProvider(input: JsonRecord): JsonRecord {
    const providerId = asString(input.providerId) || this.nextId('provider')
    const previous = this.snapshot.providers[providerId] ?? {}
    const existingSecret = this.snapshot.providerSecrets[providerId] ?? ''
    const nextSecret = asString(input.apiKey) || existingSecret
    const availableModels = asStringArray(input.availableModels)
    const now = nowISO()
    const provider: JsonRecord = {
      ...previous,
      ...input,
      providerId,
      providerType:
        asString(input.providerType) || asString(previous.providerType) || 'openai_compatible',
      displayName: asString(input.displayName) || asString(previous.displayName) || providerId,
      baseURL: asString(input.baseURL),
      organization: input.organization ?? null,
      project: input.project ?? null,
      defaultModel:
        asString(input.defaultModel) || availableModels[0] || asString(previous.defaultModel),
      availableModels,
      enabled: input.enabled !== false,
      capabilities: asStringArray(input.capabilities),
      supportsStreaming: true,
      status: 'unknown',
      healthStatus: 'unknown',
      modelCount: availableModels.length,
      hasApiKey: nextSecret !== '',
      maskedKey: nextSecret ? maskSecret(nextSecret) : null,
      updatedAt: now,
      createdAt: asString(previous.createdAt) || now
    }
    delete provider.apiKey
    this.snapshot.providers[providerId] = provider
    if (nextSecret) {
      this.snapshot.providerSecrets[providerId] = nextSecret
    }
    this.save()
    return this.safeProvider(provider)
  }

  deleteProvider(providerId: string): DeleteResult {
    if (!providerId) {
      throw badRequest('BAD_REQUEST', 'missing providerId', 'select a provider')
    }
    const provider = this.snapshot.providers[providerId]
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    if (provider.systemPreset === true || provider.allowDeletion === false) {
      throw badRequest(
        'SYSTEM_PROVIDER_NOT_DELETABLE',
        'system provider cannot be deleted',
        'disable it instead'
      )
    }
    delete this.snapshot.providers[providerId]
    delete this.snapshot.providerSecrets[providerId]
    this.save()
    return { ok: true, deletedId: providerId }
  }

  testProvider(providerId: string): JsonRecord {
    if (!this.snapshot.providers[providerId]) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    const provider = touch({
      ...this.snapshot.providers[providerId],
      status: 'connected',
      healthStatus: 'connected',
      lastTestedAt: nowISO()
    })
    this.snapshot.providers[providerId] = provider
    this.save()
    return {
      ok: true,
      targetId: providerId,
      message: 'provider configuration is reachable by Main Runtime',
      latencyMs: 0,
      trace_id: newTraceId()
    }
  }

  refreshProviderModels(providerId: string): JsonRecord {
    const provider = this.snapshot.providers[providerId]
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    const models = asStringArray(provider.availableModels)
    provider.modelCount = models.length
    provider.lastDiscoveryAt = nowISO()
    provider.updatedAt = nowISO()
    this.save()
    return this.safeProvider(provider)
  }

  listProfiles(): JsonRecord[] {
    return sortedValues(this.snapshot.profiles, 'profileId')
  }

  saveProfile(input: JsonRecord): JsonRecord {
    const profileId = asString(input.profileId) || this.nextId('profile')
    const previous = this.snapshot.profiles[profileId] ?? {}
    const now = nowISO()
    const profile = {
      ...previous,
      ...input,
      profileId,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    this.snapshot.profiles[profileId] = profile
    this.save()
    return profile
  }

  deleteProfile(profileId: string): DeleteResult {
    return this.deleteFromMap(this.snapshot.profiles, profileId, 'profileId')
  }

  getSettings(): JsonRecord {
    this.snapshot.settings = { ...defaultSettings(), ...(this.snapshot.settings ?? {}) }
    return this.snapshot.settings
  }

  updateSettings(input: JsonRecord): JsonRecord {
    this.snapshot.settings = { ...this.getSettings(), ...input }
    this.save()
    return this.snapshot.settings
  }

  resetExtensionSettings(): JsonRecord {
    this.snapshot.settings = defaultSettings()
    this.save()
    return this.snapshot.settings
  }

  listExtensions(): JsonRecord[] {
    const settings = this.getSettings()
    return [
      {
        extensionId: '9router',
        name: '9Router',
        kind: 'model_router',
        runtimeKind: 'node',
        description: '本地 OpenAI-compatible 模型路由。',
        install: {
          packageName: settings.nineRouterManagedPackageName,
          packageVersion: settings.nineRouterManagedInstallVersion,
          runtimeDir: '',
          logDir: '',
          configDir: this.configDir
        },
        process: {
          defaultCommand: settings.nineRouterManagedCommand,
          defaultArgs: [],
          port: 9399,
          env: []
        },
        health: {
          dashboardURL: settings.nineRouterDashboardURL,
          baseURL: settings.nineRouterBaseURL,
          modelsPath: '/models',
          chatPath: '/chat/completions'
        },
        providerBridge: {
          providerId: 'provider_9router_local',
          providerType: 'openai_compatible',
          displayName: '9Router 本地模型路由',
          baseURL: settings.nineRouterBaseURL,
          defaultModel: settings.nineRouterDefaultModel,
          sortOrder: 10,
          systemPreset: true,
          allowDeletion: false
        },
        capabilities: ['model_routing', 'openai_compatible'],
        security: {
          riskLevel: 'medium',
          allowedHosts: ['127.0.0.1', 'localhost'],
          secretKeys: [],
          envAllowList: [],
          managedRequiresExplicitEnable: true
        },
        systemPreset: true,
        enabled: settings.enableNineRouterIntegration
      }
    ]
  }

  extensionStatus(extensionId = '9router'): JsonRecord {
    const settings = this.getSettings()
    const nodeVersion = process.version
    return {
      extensionId,
      installed: true,
      installSource: 'node-engine',
      nodeAvailable: true,
      npmAvailable: true,
      nodeVersion,
      npmVersion: '',
      command: asString(settings.nineRouterManagedCommand) || '9router',
      runMode: settings.nineRouterRunMode ?? 'managed',
      processState: 'external_or_idle',
      startedByDreamWorker: false,
      baseURL: settings.nineRouterBaseURL,
      dashboardURL: settings.nineRouterDashboardURL,
      healthStatus: 'unknown',
      modelCount: 0,
      models: [],
      streamingVerified: false,
      hasApiKey: false,
      logDir: '',
      workDir: '',
      lastCheckedAt: nowISO(),
      runtime: {
        nodeAvailable: true,
        npmAvailable: true,
        nodeVersion,
        npmVersion: '',
        commandAvailable: true,
        command: settings.nineRouterManagedCommand,
        installSource: 'node-engine'
      }
    }
  }

  extensionAction(input: JsonRecord, verb: string): JsonRecord {
    const extensionId = asString(input.extensionId) || '9router'
    return {
      ok: true,
      extensionId,
      message: `Main Runtime handled ${verb}.`,
      status: this.extensionStatus(extensionId)
    }
  }

  listAgents(): JsonRecord[] {
    return sortedValues(this.snapshot.agents, 'agentId')
  }

  getAgent(agentId: string): JsonRecord {
    return this.getFromMap(this.snapshot.agents, agentId, 'AGENT_NOT_FOUND', 'agent not found')
  }

  saveAgent(input: JsonRecord): JsonRecord {
    return this.saveMapRecord(this.snapshot.agents, input, 'agentId', 'agent')
  }

  duplicateAgent(agentId: string): JsonRecord {
    const agent = this.getAgent(agentId)
    const duplicated = {
      ...agent,
      agentId: this.nextId('agent'),
      displayName: `${asString(agent.displayName) || agentId} Copy`,
      builtIn: false,
      createdAt: nowISO(),
      updatedAt: nowISO()
    }
    this.snapshot.agents[asString(duplicated.agentId)] = duplicated
    this.save()
    return duplicated
  }

  deleteAgent(agentId: string): DeleteResult {
    return this.deleteFromMap(this.snapshot.agents, agentId, 'agentId')
  }

  listSkills(): JsonRecord[] {
    return sortedValues(this.snapshot.skills, 'skillId')
  }

  getSkill(skillId: string): JsonRecord {
    return this.getFromMap(this.snapshot.skills, skillId, 'SKILL_NOT_FOUND', 'skill not found')
  }

  saveSkill(input: JsonRecord): JsonRecord {
    return this.saveMapRecord(this.snapshot.skills, input, 'skillId', 'skill')
  }

  deleteSkill(skillId: string): DeleteResult {
    return this.deleteFromMap(this.snapshot.skills, skillId, 'skillId')
  }

  listTools(): JsonRecord[] {
    return sortedValues(this.snapshot.tools, 'toolId')
  }

  getTool(toolId: string): JsonRecord {
    return this.getFromMap(this.snapshot.tools, toolId, 'TOOL_NOT_FOUND', 'tool not found')
  }

  saveTool(input: JsonRecord): JsonRecord {
    return this.saveMapRecord(this.snapshot.tools, input, 'toolId', 'tool')
  }

  setToolEnabled(toolId: string, enabled: boolean): JsonRecord {
    const tool = this.getTool(toolId)
    const updated = { ...tool, enabled, updatedAt: nowISO() }
    this.snapshot.tools[toolId] = updated
    this.save()
    return updated
  }

  deleteTool(toolId: string): DeleteResult {
    return this.deleteFromMap(this.snapshot.tools, toolId, 'toolId')
  }

  listMcpServers(): JsonRecord[] {
    return sortedValues(this.snapshot.mcpServers, 'serverId').map((server) =>
      this.safeMcpServer(server)
    )
  }

  saveMcpServer(input: JsonRecord): JsonRecord {
    const serverId = asString(input.serverId) || this.nextId('mcp')
    const secrets = asRecord(input.secrets)
    const secretMap = Object.fromEntries(
      Object.entries(secrets).filter(
        (entry): entry is [string, string] => typeof entry[1] === 'string'
      )
    )
    const previous = this.snapshot.mcpServers[serverId] ?? {}
    const now = nowISO()
    const server: JsonRecord = {
      ...previous,
      ...input,
      serverId,
      args: Array.isArray(input.args) ? input.args : [],
      envKeys: Object.keys(secretMap),
      hasSecrets: Object.keys(secretMap).length > 0,
      maskedSecrets: Object.entries(secretMap).map(([key, value]) => `${key}=${maskSecret(value)}`),
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    delete server.secrets
    this.snapshot.mcpServers[serverId] = server
    if (Object.keys(secretMap).length > 0) {
      this.snapshot.mcpServerSecrets[serverId] = secretMap
    }
    this.save()
    return this.safeMcpServer(server)
  }

  deleteMcpServer(serverId: string): DeleteResult {
    delete this.snapshot.mcpServerSecrets[serverId]
    return this.deleteFromMap(this.snapshot.mcpServers, serverId, 'serverId')
  }

  testMcpServer(serverId: string): JsonRecord {
    if (!this.snapshot.mcpServers[serverId]) {
      throw notFound('MCP_SERVER_NOT_FOUND', 'MCP server not found', 'select another MCP server')
    }
    return {
      ok: true,
      targetId: serverId,
      message: 'MCP config accepted by Main Runtime',
      latencyMs: 0,
      trace_id: newTraceId()
    }
  }

  refreshMcpTools(serverId: string): JsonRecord[] {
    if (!this.snapshot.mcpServers[serverId]) {
      throw notFound('MCP_SERVER_NOT_FOUND', 'MCP server not found', 'select another MCP server')
    }
    return []
  }

  listProjects(): JsonRecord[] {
    return sortedValues(this.snapshot.projects, 'projectId')
  }

  createProject(input: JsonRecord): JsonRecord {
    const now = nowISO()
    const projectId = this.nextId('project')
    const project = createProject(projectId, now, {
      title: asString(input.title) || '新项目',
      description: asString(input.description),
      localRootPath: typeof input.localRootPath === 'string' ? input.localRootPath : null
    })
    this.snapshot.projects[projectId] = project
    this.snapshot.modules[projectId] = createProjectModules(projectId)
    this.save()
    if (project.localRootPath) {
      this.initializeLocalDirectory(projectId)
    }
    return this.snapshot.projects[projectId] ?? project
  }

  getProject(projectId: string): JsonRecord {
    return this.getFromMap(
      this.snapshot.projects,
      projectId,
      'PROJECT_NOT_FOUND',
      'project not found'
    )
  }

  updateProject(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const previous = this.getProject(projectId)
    const project = touch({ ...previous, ...input, projectId })
    this.snapshot.projects[projectId] = project
    if (!this.snapshot.modules[projectId]) {
      this.snapshot.modules[projectId] = createProjectModules(projectId)
    }
    this.save()
    return project
  }

  deleteProject(projectId: string): DeleteResult {
    if (!this.snapshot.projects[projectId]) {
      throw notFound('PROJECT_NOT_FOUND', 'project not found', 'refresh project list')
    }
    delete this.snapshot.projects[projectId]
    delete this.snapshot.modules[projectId]
    this.save()
    return { ok: true, deletedId: projectId }
  }

  validateLocalDirectory(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const check = this.inspectProjectDirectory(project)
    this.snapshot.projects[projectId] = {
      ...project,
      localRootPath: check.localRootPath,
      localDirectoryStatus: check.status,
      localDirectoryLastCheckedAt: check.lastCheckedAt,
      updatedAt: nowISO()
    }
    this.save()
    return check
  }

  initializeLocalDirectory(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const root = asString(project.localRootPath)
    if (!root) {
      throw badRequest(
        'LOCAL_DIRECTORY_NOT_SET',
        'project has no localRootPath',
        'bind a local project directory first'
      )
    }
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
    return this.validateLocalDirectory(projectId)
  }

  exportProjectManifest(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const manifest = this.projectManifest(project)
    const localRootPath = typeof project.localRootPath === 'string' ? project.localRootPath : null
    if (!localRootPath) {
      return { projectId, localRootPath: null, manifestPath: null, manifest }
    }
    const manifestPath = this.writeProjectManifest(project)
    return { projectId, localRootPath, manifestPath, manifest }
  }

  listProjectModules(projectId: string): JsonRecord[] {
    this.getProject(projectId)
    if (!this.snapshot.modules[projectId]) {
      this.snapshot.modules[projectId] = createProjectModules(projectId)
      this.save()
    }
    return sortedValues(this.snapshot.modules[projectId] ?? {}, 'moduleId')
  }

  getProjectModule(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const moduleId = asString(input.moduleId)
    const modules = this.snapshot.modules[projectId] ?? createProjectModules(projectId)
    const module = modules[moduleId]
    if (!module) {
      throw notFound('MODULE_NOT_FOUND', 'module not found', 'select another module')
    }
    return module
  }

  updateProjectModuleConfig(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const moduleId = asString(input.moduleId)
    const module = this.getProjectModule({ projectId, moduleId })
    const updated = { ...module, config: { ...asRecord(module.config), ...asRecord(input.config) } }
    if (!this.snapshot.modules[projectId]) {
      this.snapshot.modules[projectId] = createProjectModules(projectId)
    }
    this.snapshot.modules[projectId][moduleId] = updated
    this.save()
    return updated
  }

  importRequirementFiles(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const project = this.getProject(projectId)
    const root = this.projectRoot(project)
    const runId = this.nextId('requirements')
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
    this.save()
    return { projectId, runId, sources, message: `imported ${sources.length} requirement file(s)` }
  }

  listRequirementSources(projectId: string): JsonRecord {
    const project = this.getProject(projectId)
    const root = this.projectRoot(project)
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

  previewRequirementSource(input: JsonRecord): JsonRecord {
    const projectId = asString(input.projectId)
    const sources = this.listRequirementSources(projectId).sources
    const source = Array.isArray(sources)
      ? (sources.find((item) => asRecord(item).sourceId === input.sourceId) as
          JsonRecord | undefined)
      : undefined
    if (!source) {
      throw notFound(
        'REQUIREMENT_SOURCE_NOT_FOUND',
        'requirement source not found',
        'select another source'
      )
    }
    const absolutePath = asString(source.absolutePath)
    const raw = readFileSync(absolutePath)
    const text = raw.toString('utf8')
    return {
      projectId,
      source,
      parser: 'node-inline',
      content: text.slice(0, 20000),
      charCount: text.length,
      truncated: text.length > 20000,
      traceId: newTraceId(),
      createdAt: nowISO()
    }
  }

  runRequirementAnalysis(input: JsonRecord): JsonRecord {
    const project = this.getProject(asString(input.projectId))
    const projectTitle = asString(project.title)
    const runId = this.nextId('requirements_run')
    const root = this.projectRoot(project)
    const outputDir = join(root, 'artifacts', 'product')
    mkdirSync(outputDir, { recursive: true })
    const analysis = {
      projectTitle,
      summary: asString(input.prompt) || 'Main Runtime generated requirement analysis placeholder.',
      sources: asStringArray(input.sourceIds),
      roles: ['用户'],
      features: [],
      nonFunctionalRequirements: [],
      risks: [],
      openQuestions: []
    }
    const analysisPath = join(outputDir, 'requirements_analysis.json')
    writeFileSync(analysisPath, `${JSON.stringify(analysis, null, 2)}\n`)
    return {
      runId,
      projectId: asString(project.projectId),
      status: 'completed',
      sources: [],
      featureCount: 0,
      outputFiles: [
        {
          kind: 'analysis_json',
          fileName: 'requirements_analysis.json',
          relativePath: 'artifacts/product/requirements_analysis.json',
          absolutePath: analysisPath
        }
      ],
      warnings: ['Main Runtime has not yet connected the document parser pipeline.'],
      traceId: newTraceId(),
      createdAt: nowISO(),
      analysis
    }
  }

  listChatSessions(): JsonRecord[] {
    return sortedValues(this.snapshot.sessions, 'updatedAt').reverse()
  }

  createChatSession(input: JsonRecord): JsonRecord {
    const sessionId = this.nextId('chat')
    const now = nowISO()
    const session = {
      sessionId,
      title: asString(input.title) || '新会话',
      projectId: input.projectId ?? null,
      agentId: asString(input.agentId) || 'agent_general_assistant',
      modelProfileId: asString(input.modelProfileId) || 'profile_fast',
      status: 'active',
      lastMessageAt: null,
      createdAt: now,
      updatedAt: now
    }
    this.snapshot.sessions[sessionId] = session
    this.snapshot.messages[sessionId] = []
    this.save()
    return session
  }

  updateChatSession(input: JsonRecord): JsonRecord {
    const sessionId = asString(input.sessionId)
    const session = touch({ ...this.getChatSession(sessionId), ...input })
    this.snapshot.sessions[sessionId] = session
    this.save()
    return session
  }

  getChatSession(sessionId: string): JsonRecord {
    return this.getFromMap(
      this.snapshot.sessions,
      sessionId,
      'CHAT_SESSION_NOT_FOUND',
      'chat session not found'
    )
  }

  listChatMessages(sessionId: string): JsonRecord[] {
    this.getChatSession(sessionId)
    return this.snapshot.messages[sessionId] ?? []
  }

  sendChatMessage(input: JsonRecord): JsonRecord {
    const sessionId = asString(input.sessionId)
    const session = this.getChatSession(sessionId)
    const now = nowISO()
    const messages = this.snapshot.messages[sessionId] ?? []
    const userMessage = {
      messageId: this.nextId('msg'),
      sessionId,
      role: 'user',
      content: asString(input.content) || asString(input.prompt),
      createdAt: now,
      updatedAt: now
    }
    const assistantMessage = {
      messageId: this.nextId('msg'),
      sessionId,
      role: 'assistant',
      content: 'Main Runtime 已接管会话链路；真实模型流式调用将在模型网关迁移步骤接入。',
      createdAt: nowISO(),
      updatedAt: nowISO()
    }
    this.snapshot.messages[sessionId] = [...messages, userMessage, assistantMessage]
    this.snapshot.sessions[sessionId] = { ...session, updatedAt: nowISO(), lastMessageAt: nowISO() }
    this.save()
    return {
      sessionId,
      userMessage,
      assistantMessage,
      traceId: newTraceId(),
      usage: { promptTokens: 0, completionTokens: 0, totalTokens: 0 }
    }
  }

  deleteChatSession(sessionId: string): DeleteResult {
    if (!this.snapshot.sessions[sessionId]) {
      throw notFound('CHAT_SESSION_NOT_FOUND', 'chat session not found', 'refresh sessions')
    }
    delete this.snapshot.sessions[sessionId]
    delete this.snapshot.messages[sessionId]
    this.save()
    return { ok: true, deletedId: sessionId }
  }

  safeProvider(provider: JsonRecord): JsonRecord {
    const providerId = asString(provider.providerId)
    const secret = this.snapshot.providerSecrets[providerId] ?? ''
    const safe = { ...provider }
    delete safe.apiKey
    safe.hasApiKey = secret !== '' || safe.hasApiKey === true
    safe.maskedKey = secret ? maskSecret(secret) : (safe.maskedKey ?? null)
    safe.modelCount = asStringArray(safe.availableModels).length
    return safe
  }

  providerForCoding(providerId: string, model: string): JsonRecord {
    if (!providerId) {
      const enabled = Object.values(this.snapshot.providers).find(
        (provider) =>
          provider.enabled !== false &&
          (!model ||
            asStringArray(provider.availableModels).includes(model) ||
            provider.defaultModel === model)
      )
      if (enabled) {
        return {
          ...enabled,
          apiKey: this.snapshot.providerSecrets[asString(enabled.providerId)] ?? ''
        }
      }
      throw notFound(
        'PROVIDER_NOT_FOUND',
        'no enabled model provider found',
        'configure a provider'
      )
    }
    const provider = this.snapshot.providers[providerId]
    if (!provider) {
      throw notFound('PROVIDER_NOT_FOUND', 'provider not found', 'select another provider')
    }
    if (provider.enabled === false) {
      throw badRequest('PROVIDER_DISABLED', 'provider is disabled', 'enable the provider')
    }
    return { ...provider, apiKey: this.snapshot.providerSecrets[providerId] ?? '' }
  }

  projectCodeRoot(projectId: string): string {
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

  private loadSnapshot(): WorkspaceSnapshot {
    const row = this.db
      .prepare('SELECT payload FROM workspace_state WHERE key = ?')
      .get(workspaceSnapshotKey) as { payload?: string } | undefined
    if (!row?.payload) {
      const seeded = createDefaultSnapshot()
      this.snapshot = seeded
      this.save()
      return seeded
    }
    try {
      return ensureSnapshotShape(JSON.parse(row.payload) as WorkspaceSnapshot)
    } catch {
      throw internalError(
        'WORKSPACE_PERSIST_DECODE_FAILED',
        'failed to decode workspace state',
        'check workspace database'
      )
    }
  }

  private deleteFromMap(
    map: Record<string, JsonRecord>,
    id: string,
    idField: string
  ): DeleteResult {
    if (!id) {
      throw badRequest('BAD_REQUEST', `missing ${idField}`, 'select an item')
    }
    if (!map[id]) {
      throw notFound('RESOURCE_NOT_FOUND', 'resource not found', 'refresh list')
    }
    delete map[id]
    this.save()
    return { ok: true, deletedId: id }
  }

  private getFromMap(
    map: Record<string, JsonRecord>,
    id: string,
    code: string,
    message: string
  ): JsonRecord {
    if (!id) {
      throw badRequest('BAD_REQUEST', 'missing id', 'select an item')
    }
    const value = map[id]
    if (!value) {
      throw notFound(code, message, 'refresh list')
    }
    return value
  }

  private saveMapRecord(
    map: Record<string, JsonRecord>,
    input: JsonRecord,
    idField: string,
    prefix: string
  ): JsonRecord {
    const id = asString(input[idField]) || this.nextId(prefix)
    const previous = map[id] ?? {}
    const now = nowISO()
    const record = {
      ...previous,
      ...input,
      [idField]: id,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    map[id] = record
    this.save()
    return record
  }

  private safeMcpServer(server: JsonRecord): JsonRecord {
    const safe = { ...server }
    delete safe.secrets
    return safe
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

function bootstrapDatabase(db: DatabaseSync): void {
  db.exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  non_destructive INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  sequence INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id TEXT NOT NULL UNIQUE,
  schema_version TEXT NOT NULL,
  trace_id TEXT NOT NULL,
  mission_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  actor TEXT NOT NULL,
  timestamp TEXT NOT NULL,
  type TEXT NOT NULL,
  payload TEXT NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_events_mission_id_sequence ON events (mission_id, sequence);
CREATE INDEX IF NOT EXISTS idx_events_run_id_sequence ON events (run_id, sequence);
CREATE INDEX IF NOT EXISTS idx_events_trace_id ON events (trace_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events (type);
CREATE TABLE IF NOT EXISTS artifacts (
  artifact_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  schema_version TEXT NOT NULL,
  mission_id TEXT NOT NULL,
  run_id TEXT,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  uri TEXT NOT NULL,
  content_type TEXT,
  path TEXT NOT NULL,
  trace_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (artifact_id, version)
);
CREATE INDEX IF NOT EXISTS idx_artifacts_mission_id ON artifacts (mission_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_run_id ON artifacts (run_id);
CREATE TABLE IF NOT EXISTS capabilities (
  capability_id TEXT PRIMARY KEY,
  manifest TEXT NOT NULL,
  lifecycle TEXT NOT NULL,
  trust_level TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  risk_actions TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_transition TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_capabilities_lifecycle ON capabilities (lifecycle);
CREATE INDEX IF NOT EXISTS idx_capabilities_trust_level ON capabilities (trust_level);
CREATE TABLE IF NOT EXISTS workspace_state (
  key TEXT PRIMARY KEY,
  payload TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
`)
}

export function defaultConfigDir(): string {
  const configured = process.env.DREAMWORKER_CONFIG_DIR?.trim()
  if (configured) {
    return resolve(configured)
  }
  if (process.platform === 'win32') {
    return join(process.env.APPDATA || join(homedir(), 'AppData', 'Roaming'), 'DreamWorker')
  }
  if (process.platform === 'darwin') {
    return join(homedir(), 'Library', 'Application Support', 'DreamWorker')
  }
  return join(process.env.XDG_CONFIG_HOME || join(homedir(), '.config'), 'DreamWorker')
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
