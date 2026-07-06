import { nowISO } from '../shared/util'
import type { JsonRecord, WorkspaceSnapshot } from '../types'

const moduleIds = ['explore', 'product', 'development', 'sales'] as const

export const projectDirectoryLayout = [
  '.dreamworker',
  '.dreamworker/runs',
  '.dreamworker/logs',
  '.dreamworker/cache',
  '.dreamworker/indexes',
  'docs',
  'artifacts',
  'artifacts/explore',
  'artifacts/product',
  'artifacts/development',
  'artifacts/sales',
  'workspace',
  'workspace/code',
  'workspace/imports',
  'workspace/exports',
  'workspace/temp',
  'source',
  'source/repo'
] as const

export const projectDocumentStubs: Record<string, string> = {
  'docs/dream_brief.md': '# Dream Brief\n\n',
  'docs/research_pack.md': '# Research Pack\n\n',
  'docs/prd.md': '# PRD\n\n',
  'docs/architecture_blueprint.md': '# Architecture Blueprint\n\n',
  'docs/launch_plan.md': '# Launch Plan\n\n'
}

export function createDefaultSnapshot(): WorkspaceSnapshot {
  const timestamp = '2026-07-01T00:00:00Z'
  const deepseekModel =
    process.env.DEEPSEEK_FAST_MODEL || process.env.DEEPSEEK_MODEL || 'deepseek-v4-flash'
  const deepseekProModel = process.env.DEEPSEEK_PRO_MODEL || 'deepseek-v4-pro'
  const deepseekKey = process.env.DEEPSEEK_API_KEY || ''
  const project = createProject('project_001', timestamp, {
    title: '独立开发者 AI 项目孵化器',
    description: '从机会探索、产品定义、工程开发到销售发布的默认项目空间。'
  })
  const projectId = String(project.projectId)
  const modules = createProjectModules(projectId)
  const session = createChatSession('chat_001', timestamp, projectId)
  const sessionId = String(session.sessionId)

  return {
    schemaVersion: 'dreamworker.workspace.snapshot.v1',
    sequence: 1,
    providers: {
      provider_deepseek: {
        providerId: 'provider_deepseek',
        providerType: 'deepseek',
        displayName: 'DeepSeek 兼容服务',
        baseURL: process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com',
        organization: null,
        project: null,
        defaultModel: deepseekModel,
        availableModels: unique([
          deepseekModel,
          deepseekProModel,
          'deepseek-v4-flash',
          'deepseek-v4-pro',
          'deepseek-chat',
          'deepseek-reasoner'
        ]),
        enabled: true,
        status: 'unknown',
        capabilities: ['chat', 'tools', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'unknown',
        modelCount: 6,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: false,
        hasApiKey: deepseekKey !== '',
        maskedKey: deepseekKey ? maskInline(deepseekKey) : null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp
      },
      provider_9router_local: {
        providerId: 'provider_9router_local',
        providerType: 'openai_compatible',
        displayName: '9Router 本地模型路由',
        baseURL: 'http://127.0.0.1:9399/v1',
        organization: null,
        project: null,
        defaultModel: process.env.NINE_ROUTER_DEFAULT_MODEL || deepseekModel,
        availableModels: [process.env.NINE_ROUTER_DEFAULT_MODEL || deepseekModel],
        enabled: true,
        status: 'unknown',
        capabilities: ['chat', 'tools', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'unknown',
        modelCount: 1,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: false,
        hasApiKey: false,
        maskedKey: null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp,
        systemPreset: true,
        allowDeletion: false
      },
      provider_local_stub: {
        providerId: 'provider_local_stub',
        providerType: 'openai_compatible',
        displayName: '本地 Stub 模型',
        baseURL: 'http://127.0.0.1/model-stub',
        organization: null,
        project: null,
        defaultModel: 'model_generate_stub',
        availableModels: ['model_generate_stub'],
        enabled: true,
        status: 'connected',
        capabilities: ['chat', 'tools', 'image_generation', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'connected',
        modelCount: 1,
        latencyMs: 0,
        lastDiscoveryAt: null,
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: true,
        hasApiKey: false,
        maskedKey: null,
        lastTestedAt: null,
        lastError: null,
        createdAt: timestamp,
        updatedAt: timestamp
      }
    },
    providerSecrets: deepseekKey ? { provider_deepseek: deepseekKey } : {},
    profiles: {
      profile_fast: createProfile(
        'profile_fast',
        'DeepSeek V4 Flash 快速',
        'provider_deepseek',
        deepseekModel,
        timestamp
      ),
      profile_pro: createProfile(
        'profile_pro',
        'DeepSeek V4 Pro 高级',
        'provider_deepseek',
        deepseekProModel,
        timestamp
      ),
      profile_stub: createProfile(
        'profile_stub',
        '离线确定性模型',
        'provider_local_stub',
        'model_generate_stub',
        timestamp,
        {
          temperature: 0,
          toolMode: 'none',
          fallbackProfileId: null,
          timeoutMs: 30000
        }
      )
    },
    agents: defaultAgents(timestamp),
    skills: defaultSkills(),
    tools: defaultTools(),
    mcpServers: {
      mcp_local_files: {
        serverId: 'mcp_local_files',
        displayName: '本地文件 MCP',
        command: 'dreamworker-mcp-files',
        args: ['--project-root', '.'],
        envKeys: [],
        url: null,
        trustLevel: 'trusted_builtin',
        enabled: false,
        hasSecrets: false,
        maskedSecrets: [],
        createdAt: timestamp,
        updatedAt: timestamp
      }
    },
    mcpServerSecrets: {},
    mcpTools: {},
    projects: { [projectId]: project },
    modules: { [projectId]: modules },
    sessions: { [sessionId]: session },
    messages: { [sessionId]: [] },
    contextSummaries: {},
    settings: defaultSettings()
  }
}

export function defaultSettings(): JsonRecord {
  return {
    enableNineRouterIntegration: true,
    nineRouterRunMode: 'managed',
    nineRouterBaseURL: 'http://127.0.0.1:9399/v1',
    nineRouterDashboardURL: 'http://127.0.0.1:9399',
    nineRouterDefaultModel: process.env.NINE_ROUTER_DEFAULT_MODEL || 'deepseek-v4-flash',
    nineRouterAutoDetectOnStart: true,
    nineRouterManagedAutoStart: true,
    nineRouterManagedAutoRestart: true,
    nineRouterManagedInstallVersion: 'latest',
    nineRouterManagedPackageName: '@9router/cli',
    nineRouterManagedCommand: '9router',
    nineRouterManagedWorkDir: '',
    nineRouterManagedLogDir: '',
    nineRouterManagedTimeoutMs: 15000,
    allowNineRouterAsFreeRoute: true,
    allowAgentsUseNineRouter: true
  }
}

export function createProject(
  projectId: string,
  timestamp: string,
  input: { title: string; description: string; localRootPath?: string | null }
): JsonRecord {
  return {
    projectId,
    title: input.title,
    description: input.description,
    status: 'active',
    localRootPath: input.localRootPath ?? null,
    localDirectoryStatus: input.localRootPath ? 'invalid' : 'not_set',
    localDirectoryLastCheckedAt: null,
    defaultModelProfileId: 'profile_fast',
    defaultRouteProfileId: null,
    enabledAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_opportunity_scan'],
    enabledTools: ['tool_model_generate_stub', 'tool_human_input'],
    enabledMcpServers: [],
    moduleConfigs: defaultModuleConfigs(),
    memoryConfig: {
      projectMemoryEnabled: true,
      artifactIndexEnabled: true,
      localFileIndexEnabled: false,
      maxContextTokens: 64000
    },
    runPolicy: {
      plannerMode: 'plan_execute',
      executorMode: 'safe',
      maxRunCostUsd: 5,
      maxRunMinutes: 30,
      requireApprovalForHighRiskTools: true
    },
    securityPolicy: {
      fileAccessScope: 'project_directory_only',
      allowWriteArtifacts: true,
      allowWriteSource: false,
      allowShellExecution: false,
      allowNetworkTools: true
    },
    createdAt: timestamp,
    updatedAt: timestamp
  }
}

export function createProjectModules(projectId: string): Record<string, JsonRecord> {
  return {
    explore: moduleCard(
      projectId,
      'explore',
      '探索模块',
      'ready',
      '负责机会扫描、用户细分、竞品地图和证据收集。',
      [
        submodule(projectId, 'explore', 'opportunity_radar', '机会雷达', 'ready', [
          'dream_brief.md',
          'hypotheses.yaml'
        ])
      ]
    ),
    product: moduleCard(
      projectId,
      'product',
      '产品模块',
      'idle',
      '负责需求分析、PRD、原型说明和 Blueprint Canvas 输入。',
      [
        submodule(projectId, 'product', 'requirement_analysis', '需求分析', 'ready', [
          'feature_list.xlsx',
          'requirements_spec.docx',
          'requirements_analysis.json'
        ])
      ]
    ),
    development: moduleCard(
      projectId,
      'development',
      '开发模块',
      'idle',
      '负责系统架构、技术栈、PR 拆分、测试门禁和运行计划。',
      [
        submodule(projectId, 'development', 'architecture', '技术架构', 'idle', [
          'architecture.md'
        ]),
        submodule(projectId, 'development', 'coding_agent', '编码 Agent', 'ready', [
          '3 Engine',
          '文件树',
          '直接写入'
        ])
      ]
    ),
    sales: moduleCard(
      projectId,
      'sales',
      '销售模块',
      'idle',
      '负责定位、落地页文案、发布计划、Demo 和反馈循环。',
      [submodule(projectId, 'sales', 'launch_plan', '发布计划', 'idle', ['launch_checklist.md'])]
    )
  }
}

function defaultModuleConfigs(): Record<string, JsonRecord> {
  return Object.fromEntries(
    moduleIds.map((moduleId) => [
      moduleId,
      {
        enabled: true,
        defaultAgentIds: [],
        enabledSkillIds: [],
        enabledToolIds: ['tool_model_generate_stub'],
        enabledMcpServerIds: [],
        outputDir: `artifacts/${moduleId}`,
        inputSchema: {},
        parameters: {}
      }
    ])
  )
}

function moduleCard(
  projectId: string,
  moduleId: string,
  displayName: string,
  status: string,
  summary: string,
  submodules: JsonRecord[]
): JsonRecord {
  return {
    projectId,
    moduleId,
    displayName,
    status,
    summary,
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_blueprint'],
    enabledTools: ['tool_model_generate_stub', 'tool_artifact_write'],
    enabledMcpServers: [],
    outputArtifacts: [],
    nextBestAction: '选择子模块继续推进。',
    submodules,
    config: {}
  }
}

function submodule(
  projectId: string,
  moduleId: string,
  submoduleId: string,
  displayName: string,
  status: string,
  outputArtifacts: string[]
): JsonRecord {
  return {
    projectId,
    moduleId,
    submoduleId,
    displayName,
    status,
    summary: `${displayName} 工作区已就绪。`,
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_blueprint'],
    enabledTools: ['tool_model_generate_stub', 'tool_artifact_write'],
    outputArtifacts,
    nextBestAction: '进入工作区开始处理。',
    config: {}
  }
}

function createProfile(
  profileId: string,
  displayName: string,
  providerId: string,
  model: string,
  timestamp: string,
  overrides: JsonRecord = {}
): JsonRecord {
  return {
    profileId,
    displayName,
    providerId,
    model,
    temperature: 0.4,
    maxTokens: 4096,
    contextWindow: 128000,
    responseFormat: 'text',
    toolMode: 'auto',
    fallbackProfileId: 'profile_stub',
    timeoutMs: 120000,
    purpose: '默认模型配置',
    enabled: true,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  }
}

function createChatSession(sessionId: string, timestamp: string, projectId: string): JsonRecord {
  return {
    sessionId,
    title: '通用 Agent 工作台',
    projectId,
    agentId: 'agent_general_assistant',
    modelProfileId: 'profile_fast',
    status: 'active',
    lastMessageAt: null,
    createdAt: timestamp,
    updatedAt: timestamp
  }
}

function defaultAgents(timestamp: string): Record<string, JsonRecord> {
  return {
    agent_general_assistant: {
      agentId: 'agent_general_assistant',
      displayName: '通用助手',
      role: '通用 Agent 聊天入口',
      description: '处理日常问答、上下文整理和轻量任务拆解。',
      systemPrompt: '你是 DreamWorker 的通用助手，优先用中文清晰回答。',
      modelProfileId: 'profile_fast',
      providerId: 'provider_deepseek',
      model: process.env.DEEPSEEK_FAST_MODEL || process.env.DEEPSEEK_MODEL || 'deepseek-v4-flash',
      enabledSkills: ['skill_opportunity_scan'],
      enabledTools: ['tool_model_generate_stub', 'tool_human_input'],
      enabledMcpServers: [],
      runtimeConfig: { contextWindow: 128000, temperature: 0.4, maxTokens: 4096 },
      planner: { enabled: true, strategy: 'react' },
      executor: { timeoutMs: 120000, retryPolicy: 'none' },
      memoryScope: 'project',
      enabled: true,
      builtIn: true,
      createdAt: timestamp,
      updatedAt: timestamp
    }
  }
}

function defaultSkills(): Record<string, JsonRecord> {
  return {
    skill_opportunity_scan: skill(
      'skill_opportunity_scan',
      'opportunity-scan',
      '机会扫描',
      '探索机会与风险'
    ),
    skill_blueprint: skill('skill_blueprint', 'blueprint', '技术蓝图', '产出架构与工程计划'),
    skill_prd_draft: skill('skill_prd_draft', 'prd-draft', 'PRD 草稿', '整理产品需求'),
    skill_launch_plan: skill('skill_launch_plan', 'launch-plan', '发布计划', '设计发布节奏')
  }
}

function skill(
  skillId: string,
  commandName: string,
  displayName: string,
  description: string
): JsonRecord {
  return {
    skillId,
    commandName,
    displayName,
    description,
    whenToUse: description,
    instructions: description,
    category: 'general',
    version: '0.1.0',
    enabled: true,
    builtIn: true,
    sourcePath: '',
    requiredCapabilities: [],
    outputArtifacts: []
  }
}

function defaultTools(): Record<string, JsonRecord> {
  const tools = [
    [
      'tool_artifact_read',
      '读取产物',
      '读取项目空间内的 Artifact 元数据和内容。',
      'artifact',
      'low'
    ],
    [
      'tool_artifact_write',
      '写入产物',
      '只允许写入当前项目目录内的 Artifact。',
      'artifact',
      'medium'
    ],
    [
      'tool_model_generate_stub',
      '模型生成 Stub',
      '用于离线演示与 CI 的确定性模型能力。',
      'model',
      'low'
    ],
    ['tool_human_input', '人工输入', '把审批和 steering 交还给用户。', 'human', 'low']
  ] as const
  return Object.fromEntries(
    tools.map(([toolId, displayName, description, category, riskLevel]) => [
      toolId,
      { toolId, displayName, description, category, riskLevel, enabled: true, builtIn: true }
    ])
  )
}

function unique(values: string[]): string[] {
  return [...new Set(values.filter(Boolean))]
}

function maskInline(value: string): string {
  return value.length <= 8 ? '***' : `${value.slice(0, 4)}...${value.slice(-4)}`
}

export function ensureSnapshotShape(snapshot: WorkspaceSnapshot): WorkspaceSnapshot {
  return {
    schemaVersion: snapshot.schemaVersion || 'dreamworker.workspace.snapshot.v1',
    sequence: Number.isFinite(snapshot.sequence) ? snapshot.sequence : 0,
    providers: snapshot.providers ?? {},
    providerSecrets: snapshot.providerSecrets ?? {},
    profiles: snapshot.profiles ?? {},
    agents: snapshot.agents ?? {},
    skills: snapshot.skills ?? {},
    tools: snapshot.tools ?? {},
    mcpServers: snapshot.mcpServers ?? {},
    mcpServerSecrets: snapshot.mcpServerSecrets ?? {},
    mcpTools: snapshot.mcpTools ?? {},
    projects: snapshot.projects ?? {},
    modules: snapshot.modules ?? {},
    sessions: snapshot.sessions ?? {},
    messages: snapshot.messages ?? {},
    contextSummaries: snapshot.contextSummaries ?? {},
    settings: snapshot.settings ?? defaultSettings()
  }
}

export function touch(record: JsonRecord): JsonRecord {
  return { ...record, updatedAt: nowISO() }
}
