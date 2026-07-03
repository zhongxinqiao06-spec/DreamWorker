import { defineStore } from 'pinia'
import type {
  AgentConfig,
  AppSettings,
  ChatContextSummary,
  ChatExecutionStep,
  ChatRuntimeSelection,
  ChatStreamController,
  ChatStreamEvent,
  ChatMessage,
  ChatSession,
  ChatToolCallPreview,
  ContextBudgetReport,
  ExtensionActionResult,
  ExtensionLogLine,
  ExtensionSpec,
  ExtensionStatus,
  McpServerConfig,
  ModelProfile,
  ProviderCapability,
  Project,
  ProjectDirectoryCheck,
  ProjectModule,
  RuntimePingResponse,
  SafeModelProvider,
  SaveAgentInput,
  SaveMcpServerInput,
  SaveModelProfileInput,
  SaveModelProviderInput,
  SaveSkillInput,
  SaveToolInput,
  SkillConfig,
  ToolConfig
} from '../../../shared/dreamworker-api'
import {
  createEmptyProjectDraft,
  createProjectDraft,
  toggleSelection,
  type ProjectConfigDraft
} from './project-draft'
import {
  isModuleWorkspace,
  moduleShortTitle,
  moduleTitle,
  primaryNavItems,
  resourceTabs,
  type ModuleWorkspaceId,
  type PrimaryNavId,
  type ResourceTabId
} from './workspace-navigation'

export type { ModuleWorkspaceId, PrimaryNavId, ResourceTabId }

type RuntimePingStatus = 'idle' | 'checking' | 'ready' | 'engine_not_connected' | 'error'

export type RuntimePingState = {
  status: RuntimePingStatus
  headline: string
  detail: string
  traceId: string
  errorCode: string
}

type ProviderDraft = {
  providerId: string
  providerType: SaveModelProviderInput['providerType']
  displayName: string
  baseURL: string
  defaultModel: string
  availableModelsText: string
  capabilities: ProviderCapability[]
  enabled: boolean
  apiKey: string
}

type ProfileDraft = SaveModelProfileInput

type AgentDraft = SaveAgentInput
type SkillDraft = SaveSkillInput
type ToolDraft = SaveToolInput
type McpDraft = Omit<SaveMcpServerInput, 'args' | 'secrets'> & {
  argsText: string
  secretsText: string
}

type ResourceNoticeTone = 'success' | 'info' | 'error'

type ResourceNotice = {
  id: number
  tone: ResourceNoticeTone
  message: string
}

export type ProjectSettingsTabId =
  | 'basic'
  | 'directory'
  | 'resources'
  | 'modules'
  | 'run-policy'
  | 'security'

export type ProjectResourceType = 'agents' | 'skills' | 'tools' | 'mcp'

export type ProviderTemplateId =
  'deepseek' | 'siliconflow' | 'glm' | 'openai_compatible' | 'anthropic' | 'ollama'

let activeChatStreamCancel: (() => Promise<void>) | null = null
let resourceNoticeTimer: ReturnType<typeof setTimeout> | null = null

function resourceFailureMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message.trim()) {
    return `${fallback}：${error.message.trim()}`
  }
  if (typeof error === 'object' && error !== null && 'message' in error) {
    const message = String((error as { message?: unknown }).message ?? '').trim()
    if (message) {
      return `${fallback}：${message}`
    }
  }
  return fallback
}

function splitDraftLines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

export const ALL_MODEL_ROUTE_SOURCE = '__all__'

export type ModelRouteSourceOption = {
  readonly id: string
  readonly label: string
  readonly modelCount: number
}

const routeSourceLabels: Record<string, string> = {
  cx: 'CX',
  gc: 'Gemini CLI',
  gemini: 'Gemini CLI',
  kiro: 'Kiro AI',
  kr: 'Kiro AI',
  mimo: 'MiMo Code',
  mm: 'MiMo Code',
  oc: 'OpenCode',
  opencode: 'OpenCode',
  openrouter: 'OpenRouter',
  or: 'OpenRouter',
  qd: 'Qoder',
  qoder: 'Qoder'
}

export function routeSourceForModel(model: string): string {
  const trimmed = model.trim()
  const slashIndex = trimmed.indexOf('/')
  if (slashIndex <= 0) {
    return 'direct'
  }
  return trimmed.slice(0, slashIndex).trim() || 'direct'
}

export function routeSourceLabel(source: string): string {
  if (source === ALL_MODEL_ROUTE_SOURCE) {
    return '全部上游'
  }
  if (source === 'direct') {
    return '直连模型'
  }
  return routeSourceLabels[source.toLowerCase()] ?? source.toUpperCase()
}

export function routeSourceOptionsForModels(
  models: readonly string[]
): readonly ModelRouteSourceOption[] {
  if (models.length === 0) {
    return []
  }
  const counts = new Map<string, number>()
  for (const model of models) {
    const source = routeSourceForModel(model)
    counts.set(source, (counts.get(source) ?? 0) + 1)
  }
  return [
    {
      id: ALL_MODEL_ROUTE_SOURCE,
      label: routeSourceLabel(ALL_MODEL_ROUTE_SOURCE),
      modelCount: models.length
    },
    ...[...counts.entries()].map(([id, modelCount]) => ({
      id,
      label: routeSourceLabel(id),
      modelCount
    }))
  ]
}

export function modelsForRouteSource(models: readonly string[], source: string): readonly string[] {
  if (!source || source === ALL_MODEL_ROUTE_SOURCE) {
    return models
  }
  return models.filter((model) => routeSourceForModel(model) === source)
}

export function isRoutedModelProvider(
  provider:
    | Pick<SafeModelProvider, 'providerId' | 'providerType' | 'displayName' | 'baseURL'>
    | null
    | undefined
): boolean {
  if (!provider) {
    return false
  }
  const providerId = provider.providerId.toLowerCase()
  const displayName = provider.displayName.toLowerCase()
  const baseURL = provider.baseURL.toLowerCase()
  return (
    providerId === 'provider_9router_local' ||
    providerId.includes('9router') ||
    displayName.includes('9router') ||
    baseURL.includes('9router') ||
    baseURL.includes('localhost:20128') ||
    baseURL.includes('127.0.0.1:20128')
  )
}

function providerDraftToSaveInput(draft: ProviderDraft): SaveModelProviderInput {
  const input: SaveModelProviderInput = {
    providerId: draft.providerId,
    providerType: draft.providerType,
    displayName: draft.displayName,
    baseURL: draft.baseURL,
    organization: null,
    project: null,
    defaultModel: draft.defaultModel,
    availableModels: splitDraftLines(draft.availableModelsText),
    enabled: draft.enabled,
    capabilities: [...draft.capabilities]
  }
  const apiKey = draft.apiKey.trim()
  return apiKey ? { ...input, apiKey } : input
}

function profileDraftToSaveInput(draft: ProfileDraft): SaveModelProfileInput {
  return {
    profileId: draft.profileId,
    displayName: draft.displayName,
    providerId: draft.providerId,
    model: draft.model,
    temperature: Number(draft.temperature),
    maxTokens: Number(draft.maxTokens),
    contextWindow: Number(draft.contextWindow),
    responseFormat: draft.responseFormat,
    toolMode: draft.toolMode,
    fallbackProfileId: draft.fallbackProfileId,
    timeoutMs: Number(draft.timeoutMs),
    purpose: draft.purpose,
    enabled: draft.enabled
  }
}

function agentDraftToSaveInput(
  draft: AgentDraft,
  modelProfileId: string | undefined
): SaveAgentInput {
  const input: SaveAgentInput = {
    agentId: draft.agentId,
    displayName: draft.displayName,
    role: draft.role,
    description: draft.description,
    systemPrompt: draft.systemPrompt,
    modelProfileId: modelProfileId ?? draft.modelProfileId,
    providerId: draft.providerId,
    model: draft.model,
    enabledSkills: [...draft.enabledSkills],
    enabledTools: [...draft.enabledTools],
    enabledMcpServers: [...draft.enabledMcpServers],
    runtimeConfig: {
      contextWindow: Number(draft.runtimeConfig.contextWindow),
      temperature: Number(draft.runtimeConfig.temperature),
      maxTokens: Number(draft.runtimeConfig.maxTokens)
    },
    planner: {
      enabled: draft.planner.enabled,
      strategy: draft.planner.strategy
    },
    executor: {
      timeoutMs: Number(draft.executor.timeoutMs),
      retryPolicy: draft.executor.retryPolicy
    },
    memoryScope: draft.memoryScope,
    enabled: draft.enabled
  }
  return draft.builtIn === undefined ? input : { ...input, builtIn: draft.builtIn }
}

function skillDraftToSaveInput(draft: SkillDraft): SaveSkillInput {
  const input: SaveSkillInput = {
    skillId: draft.skillId,
    commandName: draft.commandName,
    displayName: draft.displayName,
    description: draft.description,
    whenToUse: draft.whenToUse,
    instructions: draft.instructions,
    category: draft.category,
    version: draft.version,
    enabled: draft.enabled,
    sourcePath: draft.sourcePath,
    requiredCapabilities: [...draft.requiredCapabilities],
    outputArtifacts: [...draft.outputArtifacts]
  }
  return draft.builtIn === undefined ? input : { ...input, builtIn: draft.builtIn }
}

function toolDraftToSaveInput(draft: ToolDraft): SaveToolInput {
  return {
    toolId: draft.toolId,
    displayName: draft.displayName,
    description: draft.description,
    category: draft.category,
    riskLevel: draft.riskLevel,
    enabled: draft.enabled,
    builtIn: draft.builtIn
  }
}

export const providerTypeOptions: readonly {
  readonly value: SaveModelProviderInput['providerType']
  readonly label: string
}[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'glm', label: 'GLM 智谱' },
  { value: 'volcano', label: '火山引擎' },
  { value: 'siliconflow', label: 'SiliconFlow' },
  { value: 'openai_compatible', label: 'OpenAI 兼容' },
  { value: 'ollama', label: 'Ollama 本地模型' }
]

export const providerCapabilityOptions: readonly {
  readonly value: ProviderCapability
  readonly label: string
}[] = [
  { value: 'chat', label: '对话' },
  { value: 'tools', label: '工具调用' },
  { value: 'vision', label: '视觉' },
  { value: 'json_schema', label: 'JSON Schema' }
]

export const providerTemplateOptions: readonly {
  readonly id: ProviderTemplateId
  readonly label: string
}[] = [
  { id: 'deepseek', label: 'DeepSeek' },
  { id: 'siliconflow', label: '硅基流动' },
  { id: 'glm', label: 'GLM 智谱' },
  { id: 'openai_compatible', label: 'OpenAI 兼容' },
  { id: 'anthropic', label: 'Anthropic' },
  { id: 'ollama', label: 'Ollama' }
]

function createIdleRuntimePingState(): RuntimePingState {
  return {
    status: 'idle',
    headline: '等待检查',
    detail: 'runtime.ping 会在诊断区和状态栏显示，不再占据主工作区。',
    traceId: '暂无',
    errorCode: '暂无'
  }
}

function toRuntimePingState(response: RuntimePingResponse): RuntimePingState {
  if (response.ok) {
    return {
      status: 'ready',
      headline: '引擎已连接',
      detail: `Go Engine ${response.engineVersion} 已响应。`,
      traceId: response.trace_id,
      errorCode: '暂无'
    }
  }

  return {
    status: response.error.code === 'ENGINE_NOT_CONNECTED' ? 'engine_not_connected' : 'error',
    headline: '引擎未连接',
    detail: `${response.error.message}${response.error.user_action}`,
    traceId: response.trace_id,
    errorCode: response.error.code
  }
}

function createProviderDraft(provider?: SafeModelProvider): ProviderDraft {
  if (!provider) {
    return createProviderDraftFromTemplate('deepseek')
  }
  return {
    providerId: provider.providerId,
    providerType: provider.providerType,
    displayName: provider.displayName,
    baseURL: provider.baseURL,
    defaultModel: provider.defaultModel,
    availableModelsText: provider.availableModels.join('\n'),
    capabilities: [...provider.capabilities],
    enabled: provider.enabled,
    apiKey: ''
  }
}

function createProviderDraftFromTemplate(template: ProviderTemplateId): ProviderDraft {
  const templates: Record<ProviderTemplateId, ProviderDraft> = {
    deepseek: {
      providerId: 'provider_deepseek',
      providerType: 'deepseek',
      displayName: 'DeepSeek 兼容服务',
      baseURL: 'https://api.deepseek.com',
      defaultModel: 'deepseek-v4-flash',
      availableModelsText: 'deepseek-v4-flash\ndeepseek-v4-pro\ndeepseek-chat\ndeepseek-reasoner',
      capabilities: ['chat', 'tools', 'json_schema'],
      enabled: true,
      apiKey: ''
    },
    siliconflow: {
      providerId: 'provider_siliconflow',
      providerType: 'siliconflow',
      displayName: 'SiliconFlow 硅基流动',
      baseURL: 'https://api.siliconflow.cn/v1',
      defaultModel: 'deepseek-ai/DeepSeek-V4-Flash',
      availableModelsText:
        'deepseek-ai/DeepSeek-V4-Flash\ndeepseek-ai/DeepSeek-V4-Pro\nzai-org/GLM-5.2\nQwen/Qwen3.5-4B',
      capabilities: ['chat', 'tools', 'vision', 'json_schema'],
      enabled: true,
      apiKey: ''
    },
    glm: {
      providerId: 'provider_glm',
      providerType: 'glm',
      displayName: 'GLM 智谱',
      baseURL: 'https://open.bigmodel.cn/api/paas/v4',
      defaultModel: 'glm-5.2',
      availableModelsText: 'glm-5.2\nglm-5.1\nglm-5\nglm-5-turbo\nglm-4.7\nglm-4.7-flashx\nglm-4.6',
      capabilities: ['chat', 'tools', 'vision', 'json_schema'],
      enabled: true,
      apiKey: ''
    },
    openai_compatible: {
      providerId: `provider_custom_${Date.now()}`,
      providerType: 'openai_compatible',
      displayName: '自定义 OpenAI 兼容服务',
      baseURL: 'https://api.example.com/v1',
      defaultModel: 'model-name',
      availableModelsText: 'model-name',
      capabilities: ['chat', 'tools', 'json_schema'],
      enabled: true,
      apiKey: ''
    },
    anthropic: {
      providerId: 'provider_anthropic',
      providerType: 'anthropic',
      displayName: 'Anthropic',
      baseURL: 'https://api.anthropic.com',
      defaultModel: 'claude-sonnet-4-5',
      availableModelsText: 'claude-sonnet-4-5\nclaude-haiku-4-5',
      capabilities: ['chat', 'tools', 'vision', 'json_schema'],
      enabled: true,
      apiKey: ''
    },
    ollama: {
      providerId: 'provider_ollama',
      providerType: 'ollama',
      displayName: 'Ollama 本地模型',
      baseURL: 'http://127.0.0.1:11434',
      defaultModel: 'llama3.1',
      availableModelsText: 'llama3.1\nqwen3\ndeepseek-r1',
      capabilities: ['chat', 'tools'],
      enabled: true,
      apiKey: ''
    }
  }
  return { ...templates[template], capabilities: [...templates[template].capabilities] }
}

function createProfileDraft(profile?: ModelProfile, provider?: SafeModelProvider): ProfileDraft {
  return {
    profileId: profile?.profileId ?? `profile_custom_${Date.now()}`,
    displayName: profile?.displayName ?? 'DeepSeek 模型配置',
    providerId: profile?.providerId ?? provider?.providerId ?? 'provider_deepseek',
    model: profile?.model ?? provider?.defaultModel ?? 'deepseek-v4-flash',
    temperature: profile?.temperature ?? 0.4,
    maxTokens: profile?.maxTokens ?? 4096,
    contextWindow: profile?.contextWindow ?? 128000,
    responseFormat: profile?.responseFormat ?? 'text',
    toolMode: profile?.toolMode ?? 'auto',
    fallbackProfileId: profile?.fallbackProfileId ?? null,
    timeoutMs: profile?.timeoutMs ?? 120000,
    purpose: profile?.purpose ?? '聊天与项目孵化通用生成',
    enabled: profile?.enabled ?? true
  }
}

function createAgentDraft(agent?: AgentConfig, profile?: ModelProfile): AgentDraft {
  const providerId = agent?.providerId || profile?.providerId || 'provider_deepseek'
  const model = agent?.model || profile?.model || 'deepseek-v4-flash'
  return {
    agentId: agent?.agentId ?? `agent_custom_${Date.now()}`,
    displayName: agent?.displayName ?? '自定义 Agent',
    role: agent?.role ?? '工作台助手',
    description: agent?.description ?? '处理当前项目中的专门任务。',
    systemPrompt: agent?.systemPrompt ?? '你是 DreamWorker 的专业 Agent，优先用中文清晰回答。',
    modelProfileId: agent?.modelProfileId ?? profile?.profileId ?? 'profile_fast',
    providerId,
    model,
    enabledSkills: [...(agent?.enabledSkills ?? [])],
    enabledTools: [...(agent?.enabledTools ?? [])],
    enabledMcpServers: [...(agent?.enabledMcpServers ?? [])],
    runtimeConfig: {
      contextWindow: agent?.runtimeConfig.contextWindow ?? 128000,
      temperature: agent?.runtimeConfig.temperature ?? 0.4,
      maxTokens: agent?.runtimeConfig.maxTokens ?? 4096
    },
    planner: {
      enabled: agent?.planner.enabled ?? true,
      strategy: agent?.planner.strategy ?? 'plan-execute'
    },
    executor: {
      timeoutMs: agent?.executor.timeoutMs ?? 120000,
      retryPolicy: agent?.executor.retryPolicy ?? 'retry_twice_then_ask'
    },
    memoryScope: agent?.memoryScope ?? 'project',
    enabled: agent?.enabled ?? true
  }
}

function createSkillDraft(skill?: SkillConfig): SkillDraft {
  return {
    skillId: skill?.skillId ?? `skill_custom_${Date.now()}`,
    commandName: skill?.commandName ?? 'custom-skill',
    displayName: skill?.displayName ?? '自定义 Skill',
    description: skill?.description ?? '把一类高频任务沉淀成可复用能力。',
    whenToUse: skill?.whenToUse ?? '当用户需要执行该类任务时使用。',
    instructions: skill?.instructions ?? '# Instructions\n\n用中文完成任务，并输出可执行下一步。',
    category: skill?.category ?? 'general',
    version: skill?.version ?? '0.1.0',
    enabled: skill?.enabled ?? true,
    sourcePath: skill?.sourcePath ?? '',
    requiredCapabilities: [...(skill?.requiredCapabilities ?? ['cap_model_generate_stub'])],
    outputArtifacts: [...(skill?.outputArtifacts ?? [])]
  }
}

function createToolDraft(tool?: ToolConfig): ToolDraft {
  return {
    toolId: tool?.toolId ?? `tool_custom_${Date.now()}`,
    displayName: tool?.displayName ?? '自定义工具',
    description: tool?.description ?? '在当前工作台中执行一个明确动作。',
    category: tool?.category ?? 'project',
    riskLevel: tool?.riskLevel ?? 'medium',
    enabled: tool?.enabled ?? true,
    builtIn: tool?.builtIn ?? false
  }
}

function createMcpDraft(server?: McpServerConfig): McpDraft {
  return {
    serverId: server?.serverId ?? `mcp_custom_${Date.now()}`,
    displayName: server?.displayName ?? '自定义 MCP 服务',
    command: server?.command ?? '',
    argsText: server?.args.join('\n') ?? '',
    url: server?.url ?? null,
    trustLevel: server?.trustLevel ?? 'local_unverified',
    enabled: server?.enabled ?? false,
    secretsText: ''
  }
}

function createEmptyContextBudget(): ContextBudgetReport {
  return {
    contextWindow: 0,
    maxOutputTokens: 0,
    inputBudgetTokens: 0,
    estimatedTokens: 0,
    systemTokens: 0,
    recentMessageTokens: 0,
    summaryTokens: 0,
    recentMessageCount: 0,
    compactedCount: 0,
    compacted: false,
    warnings: []
  }
}

function createDefaultAppSettings(): AppSettings {
  return {
    enableNineRouterIntegration: true,
    nineRouterRunMode: 'external',
    nineRouterBaseURL: 'http://localhost:20128/v1',
    nineRouterDashboardURL: 'http://localhost:20128',
    nineRouterDefaultModel: 'kr/claude-sonnet-4.5',
    nineRouterAutoDetectOnStart: true,
    nineRouterManagedAutoStart: false,
    nineRouterManagedAutoRestart: false,
    nineRouterManagedInstallVersion: 'latest',
    nineRouterManagedPackageName: '9router',
    nineRouterManagedCommand: '9router',
    nineRouterManagedWorkDir: '',
    nineRouterManagedLogDir: '',
    nineRouterManagedTimeoutMs: 30000,
    allowNineRouterAsFreeRoute: true,
    allowAgentsUseNineRouter: true
  }
}

function statusFromExtensionResult(result: ExtensionActionResult): ExtensionStatus {
  return result.status
}

function preferDeepSeekProvider(providers: SafeModelProvider[]): SafeModelProvider | undefined {
  return (
    providers.find((provider) => provider.providerId === 'provider_deepseek') ??
    providers.find((provider) => provider.providerType === 'deepseek') ??
    providers[0]
  )
}

function preferDeepSeekProfile(
  profiles: ModelProfile[],
  providerId = 'provider_deepseek'
): ModelProfile | undefined {
  return (
    profiles.find((profile) => profile.profileId === 'profile_fast') ??
    profiles.find((profile) => profile.providerId === providerId) ??
    profiles.find((profile) => profile.profileId === 'profile_deepseek') ??
    profiles[0]
  )
}

function profileForProviderModel(
  profiles: ModelProfile[],
  providerId: string,
  model: string
): ModelProfile | undefined {
  return (
    profiles.find((profile) => profile.providerId === providerId && profile.model === model) ??
    profiles.find((profile) => profile.providerId === providerId) ??
    preferDeepSeekProfile(profiles, providerId)
  )
}

function providerSatisfiesChatKey(provider: SafeModelProvider): boolean {
  return (
    provider.enabled &&
    (provider.providerType === 'ollama' ||
      provider.providerId === 'provider_local_stub' ||
      provider.hasApiKey)
  )
}

export const useAppShellStore = defineStore('app-shell', {
  state: () => ({
    primaryNavItems,
    resourceTabs,
    activePrimary: 'chat' as PrimaryNavId,
    activeResourceTab: 'providers' as ResourceTabId,
    activeSubmoduleByModule: {} as Partial<Record<ModuleWorkspaceId, string>>,
    bootStatus: 'idle' as 'idle' | 'loading' | 'ready' | 'error',
    errorBanner: '',
    runtimePing: createIdleRuntimePingState(),
    settings: createDefaultAppSettings(),
    extensions: [] as ExtensionSpec[],
    extensionStatuses: {} as Record<string, ExtensionStatus>,
    extensionLogs: {} as Record<string, ExtensionLogLine[]>,
    activeExtensionId: 'extension_9router',
    extensionActionStatus: '',
    providers: [] as SafeModelProvider[],
    profiles: [] as ModelProfile[],
    agents: [] as AgentConfig[],
    skills: [] as SkillConfig[],
    tools: [] as ToolConfig[],
    mcpServers: [] as McpServerConfig[],
    projects: [] as Project[],
    projectModules: [] as ProjectModule[],
    chatSessions: [] as ChatSession[],
    chatMessages: [] as ChatMessage[],
    chatMessagesBySession: {} as Record<string, ChatMessage[]>,
    chatExecutionSteps: [] as ChatExecutionStep[],
    chatToolCalls: [] as ChatToolCallPreview[],
    chatRuntimeSummary: '',
    chatRuntimeProvider: '',
    chatRuntimeModel: '',
    chatRuntimeLatencyMs: 0,
    chatRuntimeFinishReason: '',
    chatContextBudget: createEmptyContextBudget(),
    chatContextSummary: null as ChatContextSummary | null,
    chatRuntimeToolState: 'ready',
    chatRuntimeSkillState: 'ready',
    chatRuntimeSelection: null as ChatRuntimeSelection | null,
    chatReasoningByMessage: {} as Record<string, string>,
    chatRuntimeAttemptStatus: 'ready' as
      'ready' | 'streaming' | 'blocked' | 'failed' | 'cancelled' | 'completed',
    providerSearch: '',
    chatSending: false,
    chatStreaming: false,
    chatStreamId: '',
    chatStreamSessionId: '',
    chatStreamError: '',
    chatStreamTraceId: '',
    lastRetryUserMessageId: '',
    resourceNotice: null as ResourceNotice | null,
    providerActionStatus: '',
    activeProjectId: '',
    activeProjectSettingsTab: 'basic' as ProjectSettingsTabId,
    projectResourceType: 'agents' as ProjectResourceType,
    projectResourceSearch: '',
    activeProjectDirectoryCheck: null as ProjectDirectoryCheck | null,
    activeChatSessionId: '',
    activeAgentId: 'agent_general_assistant',
    activeProviderId: 'provider_deepseek',
    activeSkillId: '',
    activeToolId: '',
    activeMcpServerId: '',
    chatDraft: '我想把一个 AI 产品想法拆成可执行计划，先帮我看机会和风险。',
    providerDraft: createProviderDraft(),
    profileDraft: createProfileDraft(),
    agentDraft: createAgentDraft(),
    skillDraft: createSkillDraft(),
    toolDraft: createToolDraft(),
    mcpDraft: createMcpDraft(),
    activeProfileId: '',
    projectDraft: createEmptyProjectDraft() as ProjectConfigDraft,
    newProjectTitle: '新的 AI 项目',
    newProjectDescription: '把一个想法推进到探索、产品、开发和销售闭环。',
    commandOpen: false
  }),
  getters: {
    activeExtension: (state): ExtensionSpec | undefined =>
      state.extensions.find((extension) => extension.extensionId === state.activeExtensionId) ??
      state.extensions[0],
    activeExtensionStatus: (state): ExtensionStatus | undefined =>
      state.extensionStatuses[state.activeExtensionId],
    nineRouterStatus: (state): ExtensionStatus | undefined =>
      state.extensionStatuses.extension_9router,
    chatSelectableProviders: (state): SafeModelProvider[] =>
      state.providers.filter(
        (provider) =>
          provider.enabled &&
          (provider.providerId !== 'provider_9router_local' ||
            state.settings.allowAgentsUseNineRouter)
      ),
    activeProvider: (state): SafeModelProvider | undefined =>
      state.providers.find((provider) => provider.providerId === state.activeProviderId) ??
      preferDeepSeekProvider(state.providers),
    activeProfile: (state): ModelProfile | undefined =>
      state.profiles.find((profile) => profile.profileId === state.activeProfileId) ??
      preferDeepSeekProfile(state.profiles, state.activeProviderId),
    modelsForProvider:
      (state) =>
      (providerId: string): readonly string[] =>
        state.providers.find((provider) => provider.providerId === providerId)?.availableModels ??
        [],
    providerModelProfile:
      (state) =>
      (providerId: string, model: string): ModelProfile | undefined =>
        profileForProviderModel(state.profiles, providerId, model),
    activeAgent: (state): AgentConfig | undefined =>
      state.agents.find((agent) => agent.agentId === state.activeAgentId) ?? state.agents[0],
    activeSkill: (state): SkillConfig | undefined =>
      state.skills.find((skill) => skill.skillId === state.activeSkillId) ?? state.skills[0],
    activeTool: (state): ToolConfig | undefined =>
      state.tools.find((tool) => tool.toolId === state.activeToolId) ?? state.tools[0],
    activeMcpServer: (state): McpServerConfig | undefined =>
      state.mcpServers.find((server) => server.serverId === state.activeMcpServerId) ??
      state.mcpServers[0],
    activeProject: (state): Project | undefined =>
      state.projects.find((project) => project.projectId === state.activeProjectId) ??
      state.projects[0],
    activeModuleWorkspace: (state): ModuleWorkspaceId | null =>
      isModuleWorkspace(state.activePrimary) ? state.activePrimary : null,
    activeModule: (state): ProjectModule | undefined => {
      const moduleId = isModuleWorkspace(state.activePrimary) ? state.activePrimary : 'explore'
      return state.projectModules.find((module) => module.moduleId === moduleId)
    },
    activeSubmodule: (state) => {
      const moduleId = isModuleWorkspace(state.activePrimary) ? state.activePrimary : 'explore'
      const module = state.projectModules.find((item) => item.moduleId === moduleId)
      const activeSubmoduleId = state.activeSubmoduleByModule[moduleId]
      return (
        module?.submodules.find((submodule) => submodule.submoduleId === activeSubmoduleId) ??
        module?.submodules[0]
      )
    },
    activeChatSession: (state): ChatSession | undefined =>
      state.chatSessions.find((session) => session.sessionId === state.activeChatSessionId) ??
      state.chatSessions[0],
    activeChatModelProfileId: (state): string =>
      (
        state.chatSessions.find((session) => session.sessionId === state.activeChatSessionId) ??
        state.chatSessions[0]
      )?.modelProfileId ??
      state.agents.find((agent) => agent.agentId === state.activeAgentId)?.modelProfileId ??
      preferDeepSeekProfile(state.profiles, state.activeProviderId)?.profileId ??
      'profile_fast',
    activeChatProviderId: (state): string => {
      const session =
        state.chatSessions.find((item) => item.sessionId === state.activeChatSessionId) ??
        state.chatSessions[0]
      if (session?.providerId) {
        return session.providerId
      }
      const agent = state.agents.find(
        (item) => item.agentId === (session?.agentId ?? state.activeAgentId)
      )
      if (agent?.providerId) {
        return agent.providerId
      }
      const profile = state.profiles.find(
        (item) => item.profileId === (session?.modelProfileId ?? agent?.modelProfileId)
      )
      return profile?.providerId ?? state.activeProviderId ?? 'provider_deepseek'
    },
    activeChatModel: (state): string => {
      const session =
        state.chatSessions.find((item) => item.sessionId === state.activeChatSessionId) ??
        state.chatSessions[0]
      if (session?.model) {
        return session.model
      }
      const agent = state.agents.find(
        (item) => item.agentId === (session?.agentId ?? state.activeAgentId)
      )
      if (agent?.model) {
        return agent.model
      }
      const profile = state.profiles.find(
        (item) => item.profileId === (session?.modelProfileId ?? agent?.modelProfileId)
      )
      const provider = state.providers.find(
        (item) =>
          item.providerId === (session?.providerId ?? agent?.providerId ?? profile?.providerId)
      )
      return profile?.model ?? provider?.defaultModel ?? 'deepseek-v4-flash'
    },
    activeChatProjectId: (state): string =>
      (
        state.chatSessions.find((session) => session.sessionId === state.activeChatSessionId) ??
        state.chatSessions[0]
      )?.projectId ?? '',
    activeSessionStreaming: (state): boolean =>
      state.chatStreaming && state.chatStreamSessionId === state.activeChatSessionId,
    composerDisabledReason: (state): string => {
      const session =
        state.chatSessions.find((item) => item.sessionId === state.activeChatSessionId) ??
        state.chatSessions[0]
      const agent = state.agents.find((item) => item.agentId === session?.agentId)
      const profile = state.profiles.find((item) => item.profileId === session?.modelProfileId)
      const providerId = session?.providerId || agent?.providerId || profile?.providerId
      const model = session?.model || agent?.model || profile?.model
      const provider = state.providers.find((item) => item.providerId === providerId)
      if (!agent || !agent.enabled) {
        return '无可用 Agent'
      }
      if (!provider || !provider.enabled) {
        return '服务商已停用'
      }
      if (
        provider.providerType !== 'ollama' &&
        provider.providerId !== 'provider_local_stub' &&
        !provider.hasApiKey
      ) {
        return '缺少密钥'
      }
      if (!model && provider.modelCount === 0 && !provider.defaultModel) {
        return '模型不可用'
      }
      return ''
    },
    resourceSummary: (state): string =>
      `${state.providers.length} 个模型服务商 / ${state.agents.length} 个 Agent / ${state.skills.length} 个 Skill / ${state.tools.length} 个工具`,
    activeProviderStatus: (state): string => {
      const provider =
        state.providers.find((item) => item.providerId === state.activeProviderId) ??
        preferDeepSeekProvider(state.providers)
      if (!provider) {
        return '暂无服务商'
      }
      const statusMap = {
        connected: '已连接',
        error: '异常',
        unknown: '未检查'
      } as const
      return statusMap[provider.status]
    },
    projectModuleSummary: (state): string =>
      state.projectModules.map((module) => `${module.displayName}:${module.status}`).join(' / '),
    projectDraftDirty: (state): boolean => {
      const activeProject =
        state.projects.find((project) => project.projectId === state.activeProjectId) ??
        state.projects[0]
      return JSON.stringify(createProjectDraft(activeProject)) !== JSON.stringify(state.projectDraft)
    },
    activeWorkspaceTitle: (state): string => {
      if (state.activePrimary === 'chat') {
        return '普通 Agent 聊天工作台'
      }
      if (state.activePrimary === 'projects') {
        return '项目配置'
      }
      if (state.activePrimary === 'resources') {
        return '资源配置中心'
      }
      if (isModuleWorkspace(state.activePrimary)) {
        return moduleTitle(state.activePrimary)
      }
      if (state.activePrimary === 'settings') {
        return '设置'
      }
      return '诊断'
    },
    activeModuleShortTitle: (state): string =>
      isModuleWorkspace(state.activePrimary) ? moduleShortTitle(state.activePrimary) : '探索',
    hasProjects: (state): boolean => state.projects.length > 0
  },
  actions: {
    async loadWorkspace(): Promise<void> {
      this.bootStatus = 'loading'
      this.errorBanner = ''
      try {
        await this.checkRuntimePing()
        const [
          settings,
          extensions,
          providers,
          profiles,
          agents,
          skills,
          tools,
          mcpServers,
          projects,
          chatSessions
        ] = await Promise.all([
          window.dreamworker.settings.getSettings(),
          window.dreamworker.extensions.listExtensions(),
          window.dreamworker.models.listProviders(),
          window.dreamworker.models.listModelProfiles(),
          window.dreamworker.agents.listAgents(),
          window.dreamworker.skills.listSkills(),
          window.dreamworker.tools.listTools(),
          window.dreamworker.mcp.listServers(),
          window.dreamworker.projects.listProjects(),
          window.dreamworker.chat.listSessions()
        ])

        this.settings = settings
        this.extensions = [...extensions]
        this.providers = [...providers]
        this.profiles = [...profiles]
        this.agents = [...agents]
        this.skills = [...skills]
        this.tools = [...tools]
        this.mcpServers = [...mcpServers]
        this.projects = [...projects]
        this.chatSessions = [...chatSessions]
        const defaultProvider = preferDeepSeekProvider(this.providers)
        const defaultProfile = preferDeepSeekProfile(this.profiles, defaultProvider?.providerId)
        this.activeProviderId = defaultProvider?.providerId ?? ''
        this.activeProfileId = defaultProfile?.profileId ?? ''
        this.activeAgentId = this.agents[0]?.agentId ?? 'agent_general_assistant'
        this.activeProjectId = this.projects[0]?.projectId ?? ''
        this.activeChatSessionId = this.chatSessions[0]?.sessionId ?? ''
        this.activeSkillId = this.skills[0]?.skillId ?? ''
        this.activeToolId = this.tools[0]?.toolId ?? ''
        this.activeMcpServerId = this.mcpServers[0]?.serverId ?? ''
        this.activeExtensionId = this.extensions[0]?.extensionId ?? 'extension_9router'
        await this.refreshExtensionStatus(this.activeExtensionId)
        if (this.activeProvider) {
          this.providerDraft = createProviderDraft(this.activeProvider)
        }
        if (this.activeProfile) {
          this.profileDraft = createProfileDraft(this.activeProfile, this.activeProvider)
        }
        this.agentDraft = createAgentDraft(this.activeAgent, this.activeProfile)
        this.skillDraft = createSkillDraft(this.activeSkill)
        this.toolDraft = createToolDraft(this.activeTool)
        this.mcpDraft = createMcpDraft(this.activeMcpServer)
        if (this.activeChatSessionId) {
          this.syncActiveChatContext()
        }
        if (this.activeProjectId) {
          await this.loadProjectModules(this.activeProjectId)
        }
        if (this.activeChatSessionId) {
          await this.loadChatMessages(this.activeChatSessionId)
        }
        this.syncProjectDraft()
        this.bootStatus = 'ready'
      } catch (error) {
        this.bootStatus = 'error'
        this.errorBanner = error instanceof Error ? error.message : '工作台数据加载失败。'
      }
    },
    setPrimary(id: PrimaryNavId): void {
      this.activePrimary = id
      if (isModuleWorkspace(id)) {
        this.ensureActiveSubmodule(id)
      }
      this.commandOpen = false
    },
    setResourceTab(id: ResourceTabId): void {
      this.activeResourceTab = id
    },
    syncActiveChatContext(): void {
      const session = this.activeChatSession
      if (!session) {
        return
      }
      this.activeAgentId = session.agentId
      if (session.providerId) {
        this.activeProviderId = session.providerId
      }
      if (session.projectId) {
        this.activeProjectId = session.projectId
      }
    },
    async checkRuntimePing(): Promise<void> {
      this.runtimePing = {
        status: 'checking',
        headline: '正在检查引擎',
        detail: '正在通过 typed preload API 调用 runtime.ping。',
        traceId: '等待返回',
        errorCode: '暂无'
      }
      try {
        this.runtimePing = toRuntimePingState(await window.dreamworker.runtime.ping())
      } catch {
        this.runtimePing = {
          status: 'error',
          headline: '预加载调用失败',
          detail: 'runtime.ping 调用失败，请查看开发者日志。',
          traceId: '暂无',
          errorCode: 'IPC_UNAVAILABLE'
        }
      }
    },
    showResourceNotice(message: string, tone: ResourceNoticeTone = 'success'): void {
      this.resourceNotice = {
        id: Date.now(),
        tone,
        message
      }
      if (resourceNoticeTimer) {
        clearTimeout(resourceNoticeTimer)
      }
      resourceNoticeTimer = setTimeout(() => {
        this.resourceNotice = null
        resourceNoticeTimer = null
      }, 2600)
    },
    showResourceFailure(error: unknown, fallback: string): void {
      this.showResourceNotice(resourceFailureMessage(error, fallback), 'error')
    },
    setActiveExtension(extensionId: string): void {
      this.activeExtensionId = extensionId
      void this.refreshExtensionStatus(extensionId)
    },
    async refreshExtensionStatus(extensionId?: string): Promise<void> {
      const targetExtensionId = extensionId ?? this.activeExtensionId
      if (!targetExtensionId) {
        return
      }
      const status = await window.dreamworker.extensions.getExtensionStatus(targetExtensionId)
      this.extensionStatuses = {
        ...this.extensionStatuses,
        [targetExtensionId]: status
      }
    },
    async refreshExtensionLogs(extensionId?: string, showNotice = false): Promise<void> {
      const targetExtensionId = extensionId ?? this.activeExtensionId
      if (!targetExtensionId) {
        return
      }
      const logs = await window.dreamworker.extensions.tailExtensionLogs(targetExtensionId, {
        limit: 160
      })
      this.extensionLogs = {
        ...this.extensionLogs,
        [targetExtensionId]: [...logs]
      }
      if (showNotice) {
        this.showResourceNotice('日志已刷新')
      }
    },
    async refreshProviders(): Promise<void> {
      this.providers = [...(await window.dreamworker.models.listProviders())]
      if (!this.providers.some((provider) => provider.providerId === this.activeProviderId)) {
        this.activeProviderId = preferDeepSeekProvider(this.providers)?.providerId ?? ''
      }
    },
    async applyExtensionResult(result: ExtensionActionResult): Promise<void> {
      this.extensionStatuses = {
        ...this.extensionStatuses,
        [result.extensionId]: statusFromExtensionResult(result)
      }
      this.extensionActionStatus = result.message
      this.showResourceNotice(
        result.message || (result.ok ? '拓展操作已完成' : '拓展操作未完成'),
        result.ok ? 'success' : 'error'
      )
      await this.refreshExtensionLogs(result.extensionId)
      await this.refreshProviders()
    },
    async updateNineRouterSettings(partial: Partial<AppSettings>): Promise<void> {
      this.settings = await window.dreamworker.settings.updateSettings(partial)
      await this.refreshExtensionStatus('extension_9router')
      await this.refreshProviders()
      this.extensionActionStatus = '9Router 设置已保存'
      this.showResourceNotice('9Router 设置已保存')
    },
    async resetNineRouterSettings(): Promise<void> {
      this.settings = await window.dreamworker.settings.resetExtensionSettings('extension_9router')
      await this.refreshExtensionStatus('extension_9router')
      await this.refreshProviders()
      this.extensionActionStatus = '9Router 设置已恢复默认'
      this.showResourceNotice('9Router 设置已恢复默认')
    },
    async detectActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.detectExtension(this.activeExtensionId)
      await this.applyExtensionResult(result)
    },
    async installActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.installExtension({
        extensionId: this.activeExtensionId,
        version: this.settings.nineRouterManagedInstallVersion
      })
      await this.applyExtensionResult(result)
      await this.refreshExtensionLogs(this.activeExtensionId)
    },
    async startActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.startExtension(this.activeExtensionId)
      await this.applyExtensionResult(result)
    },
    async stopActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.stopExtension(this.activeExtensionId)
      await this.applyExtensionResult(result)
    },
    async restartActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.restartExtension(this.activeExtensionId)
      await this.applyExtensionResult(result)
    },
    async testActiveExtension(): Promise<void> {
      const result = await window.dreamworker.extensions.testExtension(this.activeExtensionId)
      await this.applyExtensionResult(result)
    },
    async refreshActiveExtensionModels(): Promise<void> {
      const result = await window.dreamworker.extensions.refreshExtensionModels(
        this.activeExtensionId
      )
      this.extensionStatuses = {
        ...this.extensionStatuses,
        [result.extensionId]: result.status
      }
      this.extensionActionStatus = result.ok
        ? `已刷新 ${result.models.length} 个 9Router 模型`
        : '模型刷新未完成'
      this.showResourceNotice(
        result.ok ? `已刷新 ${result.models.length} 个 9Router 模型` : '模型刷新未完成',
        result.ok ? 'success' : 'error'
      )
      await this.refreshProviders()
    },
    async verifyActiveExtensionStreaming(): Promise<void> {
      const result = await window.dreamworker.extensions.verifyExtensionStreaming(
        this.activeExtensionId
      )
      this.extensionStatuses = {
        ...this.extensionStatuses,
        [result.extensionId]: result.status
      }
      this.extensionActionStatus = result.message
      this.showResourceNotice(
        result.message || (result.ok ? '流式输出已验证' : '流式输出验证失败'),
        result.ok ? 'success' : 'error'
      )
      await this.refreshProviders()
    },
    async clearActiveExtensionLogs(): Promise<void> {
      const result = await window.dreamworker.extensions.clearExtensionLogs(this.activeExtensionId)
      await this.applyExtensionResult(result)
      this.extensionLogs = {
        ...this.extensionLogs,
        [this.activeExtensionId]: []
      }
    },
    selectProvider(providerId: string): void {
      this.activeProviderId = providerId
      const provider = this.providers.find((item) => item.providerId === providerId)
      this.providerDraft = createProviderDraft(provider)
    },
    newProviderDraft(template: ProviderTemplateId = 'deepseek'): void {
      this.providerDraft = createProviderDraftFromTemplate(template)
      this.activeProviderId = ''
      this.providerActionStatus = '已创建服务商草稿，保存后生效'
      this.showResourceNotice('已创建服务商草稿')
    },
    selectProfile(profileId: string): void {
      this.activeProfileId = profileId
      const profile = this.profiles.find((item) => item.profileId === profileId)
      const provider = this.providers.find((item) => item.providerId === profile?.providerId)
      this.profileDraft = createProfileDraft(profile, provider)
    },
    newProfileDraft(): void {
      this.profileDraft = createProfileDraft(undefined, this.activeProvider)
      this.activeProfileId = ''
      this.providerActionStatus = '已创建模型配置草稿，保存后生效'
      this.showResourceNotice('已创建模型配置草稿')
    },
    selectAgent(agentId: string): void {
      this.activeAgentId = agentId
      const agent = this.agents.find((item) => item.agentId === agentId)
      this.agentDraft = createAgentDraft(agent, this.activeProfile)
    },
    newAgentDraft(): void {
      this.agentDraft = createAgentDraft(undefined, this.activeProfile)
      this.activeAgentId = ''
      this.providerActionStatus = '已创建 Agent 草稿，保存后生效'
      this.showResourceNotice('已创建 Agent 草稿')
    },
    setAgentDraftProvider(providerId: string): void {
      const provider = this.providers.find((item) => item.providerId === providerId)
      this.agentDraft = {
        ...this.agentDraft,
        providerId,
        model: provider?.defaultModel ?? provider?.availableModels[0] ?? '',
        modelProfileId:
          profileForProviderModel(
            this.profiles,
            providerId,
            provider?.defaultModel ?? provider?.availableModels[0] ?? ''
          )?.profileId ?? this.agentDraft.modelProfileId
      }
    },
    setAgentDraftModel(model: string): void {
      this.agentDraft = {
        ...this.agentDraft,
        model,
        modelProfileId:
          profileForProviderModel(this.profiles, this.agentDraft.providerId, model)?.profileId ??
          this.agentDraft.modelProfileId
      }
    },
    selectSkill(skillId: string): void {
      this.activeSkillId = skillId
      const skill = this.skills.find((item) => item.skillId === skillId)
      this.skillDraft = createSkillDraft(skill)
    },
    newSkillDraft(): void {
      this.skillDraft = createSkillDraft()
      this.activeSkillId = ''
      this.providerActionStatus = '已创建 Skill 草稿，保存后生效'
      this.showResourceNotice('已创建 Skill 草稿')
    },
    selectTool(toolId: string): void {
      this.activeToolId = toolId
      const tool = this.tools.find((item) => item.toolId === toolId)
      this.toolDraft = createToolDraft(tool)
    },
    newToolDraft(): void {
      this.toolDraft = createToolDraft()
      this.activeToolId = ''
      this.providerActionStatus = '已创建工具草稿，保存后生效'
      this.showResourceNotice('已创建工具草稿')
    },
    selectMcpServer(serverId: string): void {
      this.activeMcpServerId = serverId
      const server = this.mcpServers.find((item) => item.serverId === serverId)
      this.mcpDraft = createMcpDraft(server)
    },
    newMcpDraft(): void {
      this.mcpDraft = createMcpDraft()
      this.activeMcpServerId = ''
      this.providerActionStatus = '已创建 MCP 草稿，保存后生效'
      this.showResourceNotice('已创建 MCP 草稿')
    },
    async saveProfileDraft(): Promise<void> {
      const profile = await window.dreamworker.models.saveModelProfile(
        profileDraftToSaveInput(this.profileDraft)
      )
      this.profiles = this.profiles.filter((item) => item.profileId !== profile.profileId)
      this.profiles.unshift(profile)
      this.selectProfile(profile.profileId)
      this.providerActionStatus = '模型配置已保存'
      this.showResourceNotice('模型配置已保存')
    },
    async deleteActiveProfile(): Promise<void> {
      const profileId = this.activeProfileId || this.profileDraft.profileId
      if (!profileId) {
        return
      }
      await window.dreamworker.models.deleteModelProfile(profileId)
      this.profiles = this.profiles.filter((item) => item.profileId !== profileId)
      this.activeProfileId =
        preferDeepSeekProfile(this.profiles, this.activeProviderId)?.profileId ?? ''
      this.profileDraft = createProfileDraft(this.activeProfile, this.activeProvider)
      this.providerActionStatus = '模型配置已删除'
      this.showResourceNotice('模型配置已删除')
    },
    async saveProviderDraft(): Promise<void> {
      try {
        const missingKeyBlockedBeforeSave = this.composerDisabledReason === '缺少密钥'
        const previousChatProviderId = this.activeChatProviderId
        const previousChatModel = this.activeChatModel
        const input = providerDraftToSaveInput(this.providerDraft)
        const provider = await window.dreamworker.models.saveProvider(input)
        this.providers = this.providers.filter((item) => item.providerId !== provider.providerId)
        this.providers.unshift(provider)
        this.selectProvider(provider.providerId)
        if (
          providerSatisfiesChatKey(provider) &&
          (missingKeyBlockedBeforeSave || previousChatProviderId === provider.providerId)
        ) {
          await this.updateActiveChatSessionBinding({
            providerId: provider.providerId,
            model:
              previousChatProviderId === provider.providerId
                ? previousChatModel
                : provider.defaultModel || provider.availableModels[0] || ''
          })
          this.providerActionStatus = '模型服务商配置已保存，聊天已同步到该服务商'
          this.showResourceNotice('模型服务商配置已保存，聊天已同步')
          return
        }
        this.providerActionStatus = '模型服务商配置已保存'
        this.showResourceNotice('模型服务商配置已保存')
      } catch (error) {
        this.providerActionStatus = resourceFailureMessage(error, '模型服务商配置保存失败')
        this.showResourceFailure(error, '模型服务商配置保存失败')
      }
    },
    async deleteActiveProvider(): Promise<void> {
      const providerId = this.activeProviderId || this.providerDraft.providerId
      if (!providerId) {
        return
      }
      await window.dreamworker.models.deleteProvider(providerId)
      this.providers = this.providers.filter((item) => item.providerId !== providerId)
      this.activeProviderId = preferDeepSeekProvider(this.providers)?.providerId ?? ''
      this.providerDraft = createProviderDraft(this.activeProvider)
      this.providerActionStatus = '模型服务商已删除'
      this.showResourceNotice('模型服务商已删除')
    },
    async toggleActiveProvider(enabled: boolean): Promise<void> {
      this.providerDraft.enabled = enabled
      await this.saveProviderDraft()
    },
    async testActiveProvider(): Promise<void> {
      if (!this.activeProviderId) {
        return
      }
      const result = await window.dreamworker.models.testProvider(this.activeProviderId)
      this.providerActionStatus = `${result.message} trace_id ${result.trace_id}`
      this.showResourceNotice(
        result.message || (result.ok ? '连接检查已通过' : '连接检查失败'),
        result.ok ? 'success' : 'error'
      )
      const provider =
        (await window.dreamworker.models.listProviders()).find(
          (item) => item.providerId === this.activeProviderId
        ) ?? this.activeProvider
      if (!provider) {
        return
      }
      this.providers = this.providers.map((item) =>
        item.providerId === provider.providerId ? provider : item
      )
      this.selectProvider(provider.providerId)
    },
    async refreshActiveProviderModels(): Promise<void> {
      if (!this.activeProviderId) {
        return
      }
      const provider = await window.dreamworker.models.refreshProviderModels(this.activeProviderId)
      this.providers = this.providers.map((item) =>
        item.providerId === provider.providerId ? provider : item
      )
      this.selectProvider(provider.providerId)
      this.providerActionStatus = `已自动获取 ${provider.availableModels.length} 个模型`
      this.showResourceNotice(`已自动获取 ${provider.availableModels.length} 个模型`)
    },
    async saveAgentDraft(): Promise<void> {
      const modelProfile =
        profileForProviderModel(this.profiles, this.agentDraft.providerId, this.agentDraft.model) ??
        this.activeProfile
      const agent = await window.dreamworker.agents.saveAgent(
        agentDraftToSaveInput(this.agentDraft, modelProfile?.profileId)
      )
      this.agents = this.agents.filter((item) => item.agentId !== agent.agentId)
      this.agents.unshift(agent)
      this.selectAgent(agent.agentId)
      this.providerActionStatus = 'Agent 已保存'
      this.showResourceNotice('Agent 已保存')
    },
    async duplicateActiveAgent(): Promise<void> {
      if (!this.activeAgentId) {
        return
      }
      const agent = await window.dreamworker.agents.duplicateAgent(this.activeAgentId)
      this.agents.unshift(agent)
      this.selectAgent(agent.agentId)
      this.providerActionStatus = 'Agent 副本已创建'
      this.showResourceNotice('Agent 副本已创建')
    },
    async deleteActiveAgent(): Promise<void> {
      const agentId = this.activeAgentId || this.agentDraft.agentId
      if (!agentId) {
        return
      }
      await window.dreamworker.agents.deleteAgent(agentId)
      this.agents = this.agents.filter((item) => item.agentId !== agentId)
      this.activeAgentId = this.agents[0]?.agentId ?? ''
      this.agentDraft = createAgentDraft(this.activeAgent, this.activeProfile)
      this.providerActionStatus = 'Agent 已删除'
      this.showResourceNotice('Agent 已删除')
    },
    async saveSkillDraft(): Promise<void> {
      const skill = await window.dreamworker.skills.saveSkill(
        skillDraftToSaveInput(this.skillDraft)
      )
      this.skills = this.skills.filter((item) => item.skillId !== skill.skillId)
      this.skills.unshift(skill)
      this.selectSkill(skill.skillId)
      this.providerActionStatus = 'Skill 已保存'
      this.showResourceNotice('Skill 已保存')
    },
    async deleteActiveSkill(): Promise<void> {
      const skillId = this.activeSkillId || this.skillDraft.skillId
      if (!skillId) {
        return
      }
      await window.dreamworker.skills.deleteSkill(skillId)
      this.skills = this.skills.filter((item) => item.skillId !== skillId)
      this.activeSkillId = this.skills[0]?.skillId ?? ''
      this.skillDraft = createSkillDraft(this.activeSkill)
      this.providerActionStatus = 'Skill 已删除'
      this.showResourceNotice('Skill 已删除')
    },
    async saveToolDraft(): Promise<void> {
      const tool = await window.dreamworker.tools.saveTool(toolDraftToSaveInput(this.toolDraft))
      this.tools = this.tools.filter((item) => item.toolId !== tool.toolId)
      this.tools.unshift(tool)
      this.selectTool(tool.toolId)
      this.providerActionStatus = '工具已保存'
      this.showResourceNotice('工具已保存')
    },
    async deleteActiveTool(): Promise<void> {
      const toolId = this.activeToolId || this.toolDraft.toolId
      if (!toolId) {
        return
      }
      await window.dreamworker.tools.deleteTool(toolId)
      this.tools = this.tools.filter((item) => item.toolId !== toolId)
      this.activeToolId = this.tools[0]?.toolId ?? ''
      this.toolDraft = createToolDraft(this.activeTool)
      this.providerActionStatus = '工具已删除'
      this.showResourceNotice('工具已删除')
    },
    async saveMcpDraft(): Promise<void> {
      const secrets: Record<string, string> = {}
      for (const rawLine of this.mcpDraft.secretsText.split(/\r?\n/)) {
        const line = rawLine.trim()
        const separatorIndex = line.indexOf('=')
        if (separatorIndex <= 0) {
          continue
        }
        const envKey = line.slice(0, separatorIndex).trim()
        const value = line.slice(separatorIndex + 1).trim()
        if (envKey.length > 0 && value.length > 0) {
          secrets[envKey] = value
        }
      }
      const server = await window.dreamworker.mcp.saveServer({
        serverId: this.mcpDraft.serverId,
        displayName: this.mcpDraft.displayName,
        command: this.mcpDraft.command,
        args: this.mcpDraft.argsText
          .split(/\r?\n/)
          .map((line) => line.trim())
          .filter(Boolean),
        url: this.mcpDraft.url || null,
        trustLevel: this.mcpDraft.trustLevel,
        enabled: this.mcpDraft.enabled,
        ...(Object.keys(secrets).length > 0 ? { secrets } : {})
      })
      this.mcpServers = this.mcpServers.filter((item) => item.serverId !== server.serverId)
      this.mcpServers.unshift(server)
      this.selectMcpServer(server.serverId)
      this.providerActionStatus = 'MCP 服务已保存'
      this.showResourceNotice('MCP 服务已保存')
    },
    async deleteActiveMcpServer(): Promise<void> {
      const serverId = this.activeMcpServerId || this.mcpDraft.serverId
      if (!serverId) {
        return
      }
      await window.dreamworker.mcp.deleteServer(serverId)
      this.mcpServers = this.mcpServers.filter((item) => item.serverId !== serverId)
      this.activeMcpServerId = this.mcpServers[0]?.serverId ?? ''
      this.mcpDraft = createMcpDraft(this.activeMcpServer)
      this.providerActionStatus = 'MCP 服务已删除'
      this.showResourceNotice('MCP 服务已删除')
    },
    async testActiveMcpServer(): Promise<void> {
      if (!this.activeMcpServerId) {
        return
      }
      const result = await window.dreamworker.mcp.testServer(this.activeMcpServerId)
      this.providerActionStatus = `${result.message} trace_id ${result.trace_id}`
      this.showResourceNotice(
        result.message || (result.ok ? 'MCP 检查已通过' : 'MCP 检查失败'),
        result.ok ? 'success' : 'error'
      )
    },
    async selectProject(projectId: string): Promise<void> {
      this.activeProjectId = projectId
      this.activeProjectDirectoryCheck = null
      await this.loadProjectModules(projectId)
      this.syncProjectDraft()
    },
    setProjectSettingsTab(tabId: ProjectSettingsTabId): void {
      this.activeProjectSettingsTab = tabId
    },
    setProjectResourceType(type: ProjectResourceType): void {
      this.projectResourceType = type
    },
    syncProjectDraft(): void {
      this.projectDraft = createProjectDraft(this.activeProject)
    },
    async refreshActiveProject(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      const project = await window.dreamworker.projects.getProject(this.activeProjectId)
      if (
        this.activeProjectDirectoryCheck &&
        this.activeProjectDirectoryCheck.localRootPath !== project.localRootPath
      ) {
        this.activeProjectDirectoryCheck = null
      }
      this.projects = this.projects.map((item) =>
        item.projectId === project.projectId ? project : item
      )
      this.syncProjectDraft()
    },
    async loadProjectModules(projectId: string): Promise<void> {
      if (!projectId) {
        this.projectModules = []
        return
      }
      this.projectModules = [...(await window.dreamworker.projects.listProjectModules(projectId))]
      this.ensureActiveSubmodule('explore')
      this.ensureActiveSubmodule('product')
      this.ensureActiveSubmodule('development')
      this.ensureActiveSubmodule('sales')
    },
    ensureActiveSubmodule(moduleId: ModuleWorkspaceId): void {
      const module = this.projectModules.find((item) => item.moduleId === moduleId)
      if (!module?.submodules.length) {
        return
      }
      const current = this.activeSubmoduleByModule[moduleId]
      if (!current || !module.submodules.some((submodule) => submodule.submoduleId === current)) {
        const firstSubmodule = module.submodules[0]
        if (firstSubmodule) {
          this.activeSubmoduleByModule[moduleId] = firstSubmodule.submoduleId
        }
      }
    },
    selectSubmodule(moduleId: ModuleWorkspaceId, submoduleId: string): void {
      this.activeSubmoduleByModule[moduleId] = submoduleId
    },
    async createProject(): Promise<void> {
      const project = await window.dreamworker.projects.createProject({
        title: this.newProjectTitle.trim(),
        description: this.newProjectDescription.trim()
      })
      this.projects.unshift(project)
      this.newProjectTitle = '新的 AI 项目'
      this.newProjectDescription = '把一个想法推进到探索、产品、开发和销售闭环。'
      await this.selectProject(project.projectId)
      this.activePrimary = 'projects'
    },
    async saveActiveProject(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      const previousLocalRootPath = this.activeProject?.localRootPath ?? null
      const project = await window.dreamworker.projects.updateProject({
        projectId: this.activeProjectId,
        ...this.projectDraft
      })
      if (previousLocalRootPath !== project.localRootPath) {
        this.activeProjectDirectoryCheck = null
      }
      this.projects = this.projects.map((item) =>
        item.projectId === project.projectId ? project : item
      )
      this.syncProjectDraft()
      this.showResourceNotice('项目配置已保存')
    },
    async chooseProjectLocalDirectory(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      const pickedPath = await window.dreamworker.projects.pickLocalDirectory()
      if (!pickedPath) {
        return
      }
      this.projectDraft.localRootPath = pickedPath
      await this.saveActiveProject()
      await this.validateActiveProjectDirectory()
    },
    async validateActiveProjectDirectory(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      if (this.projectDraftDirty) {
        await this.saveActiveProject()
      }
      const check = await window.dreamworker.projects.validateLocalDirectory(this.activeProjectId)
      this.activeProjectDirectoryCheck = check
      await this.refreshActiveProject()
      this.showResourceNotice(check.message, check.status === 'valid' ? 'success' : 'info')
    },
    async initializeActiveProjectDirectory(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      if (this.projectDraftDirty) {
        await this.saveActiveProject()
      }
      const check = await window.dreamworker.projects.initializeLocalDirectory(this.activeProjectId)
      this.activeProjectDirectoryCheck = check
      await this.refreshActiveProject()
      this.showResourceNotice(check.message, check.status === 'valid' ? 'success' : 'info')
    },
    async openActiveProjectDirectory(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      if (this.projectDraftDirty) {
        await this.saveActiveProject()
      }
      const result = await window.dreamworker.projects.openLocalDirectory(this.activeProjectId)
      if (result.check) {
        this.activeProjectDirectoryCheck = result.check
      }
      await this.refreshActiveProject()
      this.showResourceNotice(result.message, result.ok ? 'success' : 'error')
    },
    async exportActiveProjectManifest(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      if (this.projectDraftDirty) {
        await this.saveActiveProject()
      }
      const result = await window.dreamworker.projects.exportProjectManifest(this.activeProjectId)
      this.showResourceNotice(
        result.manifestPath ? `项目 manifest 已导出：${result.manifestPath}` : '项目 manifest 已生成'
      )
    },
    async deleteActiveProject(): Promise<void> {
      if (!this.activeProjectId) {
        return
      }
      const deletedProjectId = this.activeProjectId
      await window.dreamworker.projects.deleteProject({ projectId: deletedProjectId })
      this.projects = this.projects.filter((project) => project.projectId !== deletedProjectId)
      this.chatSessions = this.chatSessions.map((session) =>
        session.projectId === deletedProjectId ? { ...session, projectId: null } : session
      )
      this.activeProjectId = this.projects[0]?.projectId ?? ''
      if (this.activeProjectId) {
        await this.loadProjectModules(this.activeProjectId)
      } else {
        this.projectModules = []
      }
      this.syncProjectDraft()
    },
    toggleProjectAgent(agentId: string): void {
      this.projectDraft.enabledAgents = toggleSelection(this.projectDraft.enabledAgents, agentId)
    },
    toggleProjectSkill(skillId: string): void {
      this.projectDraft.enabledSkills = toggleSelection(this.projectDraft.enabledSkills, skillId)
    },
    toggleProjectTool(toolId: string): void {
      this.projectDraft.enabledTools = toggleSelection(this.projectDraft.enabledTools, toolId)
    },
    toggleProjectMcpServer(serverId: string): void {
      this.projectDraft.enabledMcpServers = toggleSelection(
        this.projectDraft.enabledMcpServers,
        serverId
      )
    },
    async createChatSession(): Promise<void> {
      const providerId = this.activeChatProviderId
      const model = this.activeChatModel
      const profile = profileForProviderModel(this.profiles, providerId, model)
      const session = await window.dreamworker.chat.createSession({
        projectId: this.activeProjectId || null,
        title: '新的 Agent 对话',
        agentId: this.activeAgentId,
        modelProfileId: profile?.profileId ?? this.activeChatModelProfileId,
        providerId,
        model
      })
      this.chatSessions.unshift(session)
      this.activeChatSessionId = session.sessionId
      this.syncActiveChatContext()
      this.setChatMessagesForSession(session.sessionId, [])
      this.chatExecutionSteps = []
      this.chatToolCalls = []
      this.chatRuntimeSummary = ''
      this.chatContextBudget = createEmptyContextBudget()
      this.chatContextSummary = null
      this.chatRuntimeSelection = null
    },
    async selectChatSession(sessionId: string): Promise<void> {
      this.activeChatSessionId = sessionId
      this.syncActiveChatContext()
      if (this.activeProjectId) {
        await this.loadProjectModules(this.activeProjectId)
      }
      if (
        this.chatStreaming &&
        this.chatStreamSessionId === sessionId &&
        this.chatMessagesBySession[sessionId]
      ) {
        this.chatMessages = [...this.chatMessagesBySession[sessionId]]
      } else {
        await this.loadChatMessages(sessionId)
      }
      this.chatExecutionSteps = []
      this.chatToolCalls = []
      this.chatRuntimeSummary = ''
      this.chatRuntimeSelection = null
    },
    async updateActiveChatSessionBinding(input: {
      agentId?: string
      modelProfileId?: string
      providerId?: string
      model?: string
      projectId?: string | null
    }): Promise<void> {
      const session = this.activeChatSession
      if (!session) {
        if (input.agentId) {
          this.activeAgentId = input.agentId
        }
        if (input.projectId) {
          this.activeProjectId = input.projectId
        }
        if (input.providerId) {
          this.activeProviderId = input.providerId
        }
        return
      }
      const providerId = input.providerId ?? session.providerId ?? this.activeChatProviderId
      const model = input.model ?? session.model ?? this.activeChatModel
      const profile = profileForProviderModel(this.profiles, providerId, model)
      const updated = await window.dreamworker.chat.updateSession({
        sessionId: session.sessionId,
        title: session.title,
        projectId: input.projectId !== undefined ? input.projectId : session.projectId,
        agentId: input.agentId ?? session.agentId,
        modelProfileId: input.modelProfileId ?? profile?.profileId ?? session.modelProfileId,
        providerId,
        model
      })
      this.chatSessions = this.chatSessions.map((item) =>
        item.sessionId === updated.sessionId ? updated : item
      )
      this.activeChatSessionId = updated.sessionId
      this.syncActiveChatContext()
    },
    async setActiveChatAgent(agentId: string): Promise<void> {
      this.activeAgentId = agentId
      await this.updateActiveChatSessionBinding({ agentId })
    },
    async setActiveChatModelProfile(modelProfileId: string): Promise<void> {
      await this.updateActiveChatSessionBinding({ modelProfileId })
    },
    async setActiveChatProvider(providerId: string): Promise<void> {
      const provider = this.providers.find((item) => item.providerId === providerId)
      const model = provider?.defaultModel ?? provider?.availableModels[0] ?? ''
      await this.updateActiveChatSessionBinding({ providerId, model })
    },
    async setActiveChatModel(model: string): Promise<void> {
      await this.updateActiveChatSessionBinding({ providerId: this.activeChatProviderId, model })
    },
    async setActiveChatProject(projectId: string): Promise<void> {
      const normalizedProjectId = projectId || null
      await this.updateActiveChatSessionBinding({ projectId: normalizedProjectId })
      if (normalizedProjectId) {
        await this.selectProject(normalizedProjectId)
      }
    },
    setChatMessagesForSession(sessionId: string, messages: ChatMessage[]): void {
      this.chatMessagesBySession = {
        ...this.chatMessagesBySession,
        [sessionId]: [...messages]
      }
      if (this.activeChatSessionId === sessionId) {
        this.chatMessages = [...messages]
      }
    },
    async loadChatMessages(sessionId: string): Promise<void> {
      if (!sessionId) {
        this.chatMessages = []
        return
      }
      this.setChatMessagesForSession(sessionId, [
        ...(await window.dreamworker.chat.getMessages(sessionId))
      ])
    },
    async sendChatMessage(): Promise<void> {
      if (this.composerDisabledReason) {
        return
      }
      await this.startChatStream({ content: this.chatDraft.trim() })
    },
    async retryLastChatMessage(): Promise<void> {
      const retryOfMessageId = this.lastRetryUserMessageId || this.findLastUserMessageId()
      if (!retryOfMessageId) {
        return
      }
      await this.startChatStream({ content: '', retryOfMessageId })
    },
    async startChatStream(input: { content: string; retryOfMessageId?: string }): Promise<void> {
      if (!input.content && !input.retryOfMessageId) {
        return
      }
      if (!this.activeChatSessionId) {
        await this.createChatSession()
      }
      const sessionId = this.activeChatSessionId
      const streamId = `stream_${Date.now()}_${Math.random().toString(36).slice(2)}`
      const createdAt = new Date().toISOString()
      this.chatSending = true
      this.chatStreaming = true
      this.chatStreamId = streamId
      this.chatStreamSessionId = sessionId
      this.chatStreamError = ''
      this.chatStreamTraceId = streamId
      this.chatRuntimeProvider = ''
      this.chatRuntimeModel = this.activeChatModel
      this.chatRuntimeLatencyMs = 0
      this.chatRuntimeFinishReason = ''
      this.chatContextBudget = createEmptyContextBudget()
      this.chatContextSummary = null
      this.chatRuntimeToolState = 'ready'
      this.chatRuntimeSkillState = 'ready'
      this.chatRuntimeSelection = null
      this.chatRuntimeAttemptStatus = 'streaming'
      this.lastRetryUserMessageId = input.retryOfMessageId ?? ''
      if (!input.retryOfMessageId) {
        this.chatDraft = ''
      }
      const existing = this.chatMessagesBySession[sessionId] ?? this.chatMessages
      const optimistic: ChatMessage[] = [
        ...(input.retryOfMessageId
          ? []
          : [
              {
                messageId: `local_user_${streamId}`,
                attemptId: '',
                sessionId,
                role: 'user' as const,
                content: input.content,
                status: 'completed' as const,
                providerId: '',
                model: '',
                usage: null,
                finishReason: '',
                runtimeSummary: '',
                trace_id: streamId,
                createdAt
              }
            ]),
        {
          messageId: `local_assistant_${streamId}`,
          attemptId: '',
          sessionId,
          role: 'assistant',
          content: '',
          status: 'streaming',
          providerId: '',
          model: this.activeChatModel,
          usage: null,
          finishReason: '',
          runtimeSummary: '',
          trace_id: streamId,
          createdAt
        }
      ]
      this.setChatMessagesForSession(sessionId, [...existing, ...optimistic])
      try {
        const request =
          input.retryOfMessageId === undefined
            ? { sessionId, content: input.content, streamId }
            : {
                sessionId,
                content: input.content,
                streamId,
                retryOfMessageId: input.retryOfMessageId
              }
        const controller: ChatStreamController = await window.dreamworker.chat.streamMessage(
          request,
          (event) => this.applyChatStreamEvent(event)
        )
        activeChatStreamCancel = controller.cancel
      } catch (error) {
        this.chatStreamError = error instanceof Error ? error.message : '对话流式调用失败。'
        this.chatStreaming = false
        this.chatSending = false
        this.chatRuntimeAttemptStatus = 'failed'
        activeChatStreamCancel = null
      }
    },
    applyChatStreamEvent(event: ChatStreamEvent): void {
      if (event.streamId !== this.chatStreamId) {
        return
      }
      const sessionId = event.sessionId || this.chatStreamSessionId
      let messages = [...(this.chatMessagesBySession[sessionId] ?? this.chatMessages)]
      if (event.type === 'started') {
        this.chatRuntimeProvider = event.providerId ?? this.chatRuntimeProvider
        this.chatRuntimeModel = event.model ?? this.chatRuntimeModel
        this.chatStreamTraceId = event.trace_id
        this.chatRuntimeSelection = event.runtimeSelection ?? this.chatRuntimeSelection
        if (event.contextBudget) {
          this.chatContextBudget = event.contextBudget
        }
        messages = messages.map((message) =>
          message.messageId === `local_assistant_${event.streamId}`
            ? {
                ...message,
                messageId: event.messageId,
                attemptId: event.attemptId ?? message.attemptId,
                trace_id: event.trace_id,
                providerId: event.providerId ?? message.providerId,
                model: event.model ?? message.model
              }
            : message
        )
        this.setChatMessagesForSession(sessionId, messages)
        return
      }
      if (event.type === 'reasoning_delta' && event.reasoningDelta) {
        this.chatReasoningByMessage = {
          ...this.chatReasoningByMessage,
          [event.messageId]: `${this.chatReasoningByMessage[event.messageId] ?? ''}${event.reasoningDelta}`
        }
        return
      }
      if (event.type === 'context_compacted') {
        if (event.contextBudget) {
          this.chatContextBudget = event.contextBudget
        }
        if (event.contextSummary) {
          this.chatContextSummary = event.contextSummary
        }
        if (event.warning) {
          this.chatStreamError = event.warning.message
        }
        return
      }
      if (event.type === 'step' && event.step) {
        if (this.activeChatSessionId === sessionId) {
          this.chatExecutionSteps = [
            ...this.chatExecutionSteps.filter((step) => step.stepId !== event.step?.stepId),
            event.step
          ]
        }
        if (event.warning) {
          this.chatStreamError = event.warning.message
        }
        return
      }
      if (event.type === 'tool_call_delta' && event.toolCall) {
        if (this.activeChatSessionId === sessionId) {
          this.chatToolCalls = [
            ...this.chatToolCalls.filter((call) => call.callId !== event.toolCall?.callId),
            event.toolCall
          ]
        }
        return
      }
      if (
        (event.type === 'tool_started' ||
          event.type === 'tool_result' ||
          event.type === 'tool_blocked') &&
        event.toolCall
      ) {
        this.chatRuntimeToolState =
          event.type === 'tool_started'
            ? 'running'
            : event.type === 'tool_blocked'
              ? 'blocked'
              : 'completed'
        if (this.activeChatSessionId === sessionId) {
          this.chatToolCalls = [
            ...this.chatToolCalls.filter((call) => call.callId !== event.toolCall?.callId),
            event.toolCall
          ]
        }
        if (event.toolResult?.errorMessage) {
          this.chatStreamError = event.toolResult.errorMessage
        }
        return
      }
      if (event.type === 'skill_used') {
        this.chatRuntimeSkillState = 'used'
        return
      }
      if (event.type === 'token_delta' && event.delta) {
        messages = messages.map((message) =>
          message.messageId === event.messageId
            ? { ...message, content: `${message.content}${event.delta}` }
            : message
        )
        this.setChatMessagesForSession(sessionId, messages)
        return
      }
      if (event.type === 'usage' && event.usage) {
        messages = messages.map((message) =>
          message.messageId === event.messageId
            ? { ...message, usage: event.usage ?? null }
            : message
        )
        this.setChatMessagesForSession(sessionId, messages)
        return
      }
      if (event.type === 'completed' && event.result) {
        this.setChatMessagesForSession(event.result.session.sessionId, [...event.result.messages])
        if (this.activeChatSessionId === event.result.session.sessionId) {
          this.chatExecutionSteps = [...event.result.executionSteps]
          this.chatToolCalls = [...event.result.toolCalls]
          this.chatRuntimeSummary = event.result.runtimeSummary
          this.chatContextBudget = event.result.contextBudget
          this.chatContextSummary = event.result.contextSummary
        }
        this.chatRuntimeLatencyMs = event.result.auditSummary.latencyMs
        this.chatRuntimeFinishReason = event.result.auditSummary.finishReason
        this.chatRuntimeAttemptStatus = 'completed'
        this.chatSessions = this.chatSessions.map((session) =>
          session.sessionId === event.result?.session.sessionId ? event.result.session : session
        )
        this.chatSending = false
        this.chatStreaming = false
        this.chatStreamId = ''
        this.chatStreamSessionId = ''
        this.lastRetryUserMessageId = ''
        activeChatStreamCancel = null
        return
      }
      if (event.type === 'failed') {
        this.chatStreamError = event.error?.message ?? '对话流式调用失败。'
        this.chatStreamTraceId = event.trace_id
        this.chatRuntimeLatencyMs = event.latencyMs ?? this.chatRuntimeLatencyMs
        this.chatRuntimeFinishReason = event.finishReason ?? 'error'
        this.chatRuntimeAttemptStatus = 'failed'
        if (event.result) {
          this.setChatMessagesForSession(event.result.session.sessionId, [...event.result.messages])
          this.chatContextBudget = event.result.contextBudget
          this.chatContextSummary = event.result.contextSummary
          this.chatSessions = this.chatSessions.map((session) =>
            session.sessionId === event.result?.session.sessionId ? event.result.session : session
          )
        } else {
          messages = messages.map((message) =>
            message.messageId === event.messageId ? { ...message, status: 'failed' } : message
          )
          this.setChatMessagesForSession(sessionId, messages)
        }
        this.lastRetryUserMessageId = this.findLastUserMessageId(sessionId)
        this.chatSending = false
        this.chatStreaming = false
        this.chatStreamId = ''
        this.chatStreamSessionId = ''
        activeChatStreamCancel = null
        return
      }
      if (event.type === 'cancelled') {
        this.chatRuntimeLatencyMs = event.latencyMs ?? this.chatRuntimeLatencyMs
        this.chatRuntimeFinishReason = event.finishReason ?? 'cancelled'
        this.chatRuntimeAttemptStatus = 'cancelled'
        if (event.result) {
          this.setChatMessagesForSession(event.result.session.sessionId, [...event.result.messages])
          this.chatContextBudget = event.result.contextBudget
          this.chatContextSummary = event.result.contextSummary
          this.chatSessions = this.chatSessions.map((session) =>
            session.sessionId === event.result?.session.sessionId ? event.result.session : session
          )
        } else {
          messages = messages.map((message) =>
            message.messageId === event.messageId ? { ...message, status: 'cancelled' } : message
          )
          this.setChatMessagesForSession(sessionId, messages)
        }
        this.lastRetryUserMessageId = this.findLastUserMessageId(sessionId)
        this.chatSending = false
        this.chatStreaming = false
        this.chatStreamId = ''
        this.chatStreamSessionId = ''
        activeChatStreamCancel = null
      }
    },
    findLastUserMessageId(sessionId?: string): string {
      const targetSessionId = sessionId ?? this.activeChatSessionId
      const messages = this.chatMessagesBySession[targetSessionId] ?? this.chatMessages
      for (let index = messages.length - 1; index >= 0; index -= 1) {
        const message = messages[index]
        if (message?.role === 'user') {
          return message.messageId
        }
      }
      return ''
    },
    async cancelChatStream(): Promise<void> {
      if (!this.chatStreaming || !activeChatStreamCancel) {
        return
      }
      await activeChatStreamCancel()
      activeChatStreamCancel = null
      this.chatSending = false
      this.chatStreaming = false
    },
    async sendChatMessageLegacy(): Promise<void> {
      if (!this.chatDraft.trim()) {
        return
      }
      if (!this.activeChatSessionId) {
        await this.createChatSession()
      }
      this.chatSending = true
      try {
        const result = await window.dreamworker.chat.sendMessage({
          sessionId: this.activeChatSessionId,
          content: this.chatDraft.trim()
        })
        this.chatDraft = ''
        this.chatMessages = [...result.messages]
        this.chatExecutionSteps = [...result.executionSteps]
        this.chatToolCalls = [...result.toolCalls]
        this.chatRuntimeSummary = result.runtimeSummary
        this.chatSessions = this.chatSessions.map((session) =>
          session.sessionId === result.session.sessionId ? result.session : session
        )
      } finally {
        this.chatSending = false
      }
    },
    async setToolEnabled(toolId: string, enabled: boolean): Promise<void> {
      const tool = await window.dreamworker.tools.setToolEnabled(toolId, enabled)
      this.tools = this.tools.map((item) => (item.toolId === tool.toolId ? tool : item))
      if (this.activeToolId === tool.toolId) {
        this.toolDraft = createToolDraft(tool)
      }
      this.showResourceNotice(enabled ? '工具已启用' : '工具已停用')
    },
    async refreshActiveMcpTools(): Promise<void> {
      if (!this.activeMcpServerId) {
        return
      }
      const tools = await window.dreamworker.mcp.refreshTools(this.activeMcpServerId)
      const discoveredIds = new Set(tools.map((tool) => tool.toolId))
      this.tools = [
        ...tools,
        ...this.tools.filter(
          (tool) => !discoveredIds.has(tool.toolId) && !tool.toolId.startsWith(`mcp_`)
        )
      ]
      this.providerActionStatus = `已刷新 MCP 工具：${tools.length} 个`
      this.showResourceNotice(`已刷新 MCP 工具：${tools.length} 个`)
    },
    toggleCommand(): void {
      this.commandOpen = !this.commandOpen
    },
    runCommand(target: PrimaryNavId | ResourceTabId): void {
      this.commandOpen = false
      if (target === 'providers' || target === 'extensions' || target === 'agents') {
        this.setPrimary('resources')
        this.setResourceTab(target)
        return
      }
      if (target === 'skills' || target === 'tools' || target === 'mcp') {
        this.setPrimary('resources')
        this.setResourceTab(target)
        return
      }
      this.setPrimary(target)
    }
  }
})
