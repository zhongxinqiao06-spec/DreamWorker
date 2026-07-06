import type {
  Project,
  ProjectMemoryConfig,
  ProjectModuleConfig,
  ProjectModuleConfigs,
  ProjectModuleId,
  ProjectRunPolicy,
  ProjectSecurityPolicy
} from '../../../shared/dreamworker-api'

export const projectModuleIds: readonly ProjectModuleId[] = [
  'explore',
  'product',
  'development',
  'sales'
]

export type ProjectModuleConfigDraft = {
  enabled: boolean
  defaultAgentIds: string[]
  enabledSkillIds: string[]
  enabledToolIds: string[]
  enabledMcpServerIds: string[]
  outputDir: string
  inputSchema: Record<string, unknown>
  parameters: Record<string, unknown>
}

export type ProjectModuleConfigsDraft = Record<ProjectModuleId, ProjectModuleConfigDraft>

export type ProjectConfigDraft = {
  title: string
  description: string
  status: Project['status']
  localRootPath: string | null
  defaultModelProfileId: string
  defaultRouteProfileId: string | null
  enabledAgents: string[]
  enabledSkills: string[]
  enabledTools: string[]
  enabledMcpServers: string[]
  moduleConfigs: ProjectModuleConfigsDraft
  memoryConfig: ProjectMemoryConfig
  runPolicy: ProjectRunPolicy
  securityPolicy: ProjectSecurityPolicy
}

export function createEmptyProjectDraft(): ProjectConfigDraft {
  return {
    title: '',
    description: '',
    status: 'active',
    localRootPath: null,
    defaultModelProfileId: 'profile_fast',
    defaultRouteProfileId: null,
    enabledAgents: [],
    enabledSkills: [],
    enabledTools: [],
    enabledMcpServers: [],
    moduleConfigs: createDefaultProjectModuleConfigsDraft(),
    memoryConfig: createDefaultProjectMemoryConfig(),
    runPolicy: createDefaultProjectRunPolicy(),
    securityPolicy: createDefaultProjectSecurityPolicy()
  }
}

export function createProjectDraft(project?: Project): ProjectConfigDraft {
  if (!project) {
    return createEmptyProjectDraft()
  }
  const legacyProject = project as Partial<Project>
  return {
    title: project.title ?? '',
    description: project.description ?? '',
    status: project.status ?? 'active',
    localRootPath: legacyProject.localRootPath ?? null,
    defaultModelProfileId: project.defaultModelProfileId ?? 'profile_fast',
    defaultRouteProfileId: legacyProject.defaultRouteProfileId ?? null,
    enabledAgents: [...(project.enabledAgents ?? [])],
    enabledSkills: [...(project.enabledSkills ?? [])],
    enabledTools: [...(project.enabledTools ?? [])],
    enabledMcpServers: [...(project.enabledMcpServers ?? [])],
    moduleConfigs: cloneProjectModuleConfigs(
      legacyProject.moduleConfigs ?? createDefaultProjectModuleConfigs()
    ),
    memoryConfig: {
      ...createDefaultProjectMemoryConfig(),
      ...(legacyProject.memoryConfig ?? {})
    },
    runPolicy: {
      ...createDefaultProjectRunPolicy(),
      ...(legacyProject.runPolicy ?? {})
    },
    securityPolicy: {
      ...createDefaultProjectSecurityPolicy(),
      ...(legacyProject.securityPolicy ?? {})
    }
  }
}

export function toggleSelection(values: readonly string[], value: string): string[] {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value]
}

export function createDefaultProjectModuleConfigs(): ProjectModuleConfigs {
  return {
    explore: createDefaultProjectModuleConfig({
      outputDir: 'artifacts/explore',
      defaultAgentIds: [
        'agent_opportunity_scout',
        'agent_competitor_analyst',
        'agent_customer_segment'
      ],
      enabledSkillIds: ['skill_opportunity_scan', 'skill_competitor_map'],
      enabledToolIds: ['tool_web_search_stub', 'tool_model_generate_stub', 'tool_artifact_write'],
      parameters: { stage: 'Discover', evidenceRequired: true }
    }),
    product: createDefaultProjectModuleConfig({
      outputDir: 'artifacts/product',
      defaultAgentIds: ['agent_product_designer', 'agent_prototype_designer', 'agent_evaluator'],
      enabledSkillIds: ['skill_prd_draft'],
      enabledToolIds: ['tool_model_generate_stub', 'tool_artifact_write'],
      parameters: { stage: 'Shape', requiresDecisionGate: true }
    }),
    development: createDefaultProjectModuleConfig({
      outputDir: 'artifacts/development',
      defaultAgentIds: [
        'agent_system_architect',
        'agent_tech_stack_advisor',
        'agent_dev_orchestrator'
      ],
      enabledSkillIds: ['skill_blueprint'],
      enabledToolIds: ['tool_model_generate_stub', 'tool_artifact_write'],
      parameters: { stage: 'Build', writeCodeAutomatically: false }
    }),
    sales: createDefaultProjectModuleConfig({
      outputDir: 'artifacts/sales',
      defaultAgentIds: ['agent_sales_strategist', 'agent_demo_designer', 'agent_evaluator'],
      enabledSkillIds: ['skill_launch_plan'],
      enabledToolIds: ['tool_model_generate_stub', 'tool_artifact_write', 'tool_human_input'],
      parameters: { stage: 'Launch', publishRequiresApproval: true }
    })
  }
}

function createDefaultProjectModuleConfigsDraft(): ProjectModuleConfigsDraft {
  return cloneProjectModuleConfigs(createDefaultProjectModuleConfigs())
}

function createDefaultProjectModuleConfig(input: {
  outputDir: string
  defaultAgentIds: readonly string[]
  enabledSkillIds: readonly string[]
  enabledToolIds: readonly string[]
  parameters: Record<string, unknown>
}): ProjectModuleConfig {
  return {
    enabled: true,
    defaultAgentIds: input.defaultAgentIds,
    enabledSkillIds: input.enabledSkillIds,
    enabledToolIds: input.enabledToolIds,
    enabledMcpServerIds: [],
    outputDir: input.outputDir,
    inputSchema: {},
    parameters: input.parameters
  }
}

function cloneProjectModuleConfigs(configs: ProjectModuleConfigs): ProjectModuleConfigsDraft {
  const defaults = createDefaultProjectModuleConfigs()
  return projectModuleIds.reduce((result, moduleId) => {
    const config = configs[moduleId] ?? defaults[moduleId]
    result[moduleId] = cloneProjectModuleConfig(config)
    return result
  }, {} as ProjectModuleConfigsDraft)
}

function cloneProjectModuleConfig(config: ProjectModuleConfig): ProjectModuleConfigDraft {
  return {
    enabled: config.enabled,
    defaultAgentIds: [...config.defaultAgentIds],
    enabledSkillIds: [...config.enabledSkillIds],
    enabledToolIds: [...config.enabledToolIds],
    enabledMcpServerIds: [...config.enabledMcpServerIds],
    outputDir: config.outputDir,
    inputSchema: { ...config.inputSchema },
    parameters: { ...config.parameters }
  }
}

function createDefaultProjectMemoryConfig(): ProjectMemoryConfig {
  return {
    projectMemoryEnabled: true,
    artifactIndexEnabled: true,
    localFileIndexEnabled: false,
    maxContextTokens: 64000
  }
}

function createDefaultProjectRunPolicy(): ProjectRunPolicy {
  return {
    plannerMode: 'plan_execute',
    executorMode: 'safe',
    maxRunCostUsd: 5,
    maxRunMinutes: 30,
    requireApprovalForHighRiskTools: true
  }
}

function createDefaultProjectSecurityPolicy(): ProjectSecurityPolicy {
  return {
    fileAccessScope: 'project_directory_only',
    allowWriteArtifacts: true,
    allowWriteSource: false,
    allowShellExecution: false,
    allowNetworkTools: true
  }
}
