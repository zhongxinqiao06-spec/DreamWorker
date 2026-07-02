import type {
  DreamWorkerError as GeneratedDreamWorkerError,
  RuntimePingResponse
} from './generated/contracts'

export const CONTRACT_SCHEMA_VERSION = '0.1'

export const CHANNELS = {
  runtimePing: 'runtime:ping',
  modelsListProviders: 'models:listProviders',
  modelsSaveProvider: 'models:saveProvider',
  modelsDeleteProvider: 'models:deleteProvider',
  modelsTestProvider: 'models:testProvider',
  modelsRefreshProviderModels: 'models:refreshProviderModels',
  modelsListProfiles: 'models:listModelProfiles',
  modelsSaveProfile: 'models:saveModelProfile',
  modelsDeleteProfile: 'models:deleteModelProfile',
  agentsList: 'agents:listAgents',
  agentsGet: 'agents:getAgent',
  agentsSave: 'agents:saveAgent',
  agentsDuplicate: 'agents:duplicateAgent',
  agentsDelete: 'agents:deleteAgent',
  skillsList: 'skills:listSkills',
  skillsGet: 'skills:getSkill',
  skillsSave: 'skills:saveSkill',
  skillsDelete: 'skills:deleteSkill',
  toolsList: 'tools:listTools',
  toolsGet: 'tools:getTool',
  toolsSave: 'tools:saveTool',
  toolsSetEnabled: 'tools:setToolEnabled',
  toolsDelete: 'tools:deleteTool',
  mcpListServers: 'mcp:listServers',
  mcpSaveServer: 'mcp:saveServer',
  mcpDeleteServer: 'mcp:deleteServer',
  mcpTestServer: 'mcp:testServer',
  mcpRefreshTools: 'mcp:refreshTools',
  projectsList: 'projects:listProjects',
  projectsCreate: 'projects:createProject',
  projectsGet: 'projects:getProject',
  projectsUpdate: 'projects:updateProject',
  projectsDelete: 'projects:deleteProject',
  projectsListModules: 'projects:listProjectModules',
  projectsGetModule: 'projects:getProjectModule',
  projectsUpdateModuleConfig: 'projects:updateProjectModuleConfig',
  chatListSessions: 'chat:listSessions',
  chatCreateSession: 'chat:createSession',
  chatUpdateSession: 'chat:updateSession',
  chatGetMessages: 'chat:getMessages',
  chatSendMessage: 'chat:sendMessage',
  chatStartStream: 'chat:startStream',
  chatCancelStream: 'chat:cancelStream',
  chatStreamEvent: 'chat:streamEvent',
  chatDeleteSession: 'chat:deleteSession'
} as const

export const RUNTIME_PING_CHANNEL = CHANNELS.runtimePing

export type DreamWorkerApiErrorCode = 'ENGINE_NOT_CONNECTED' | 'IPC_UNAVAILABLE'

export type DreamWorkerApiError = GeneratedDreamWorkerError & {
  readonly code: DreamWorkerApiErrorCode
}

export type { RuntimePingResponse }

export type ProviderType =
  | 'openai_compatible'
  | 'deepseek'
  | 'openai'
  | 'anthropic'
  | 'glm'
  | 'volcano'
  | 'siliconflow'
  | 'gemini'
  | 'ollama'
  | 'custom'

export type ProviderCapability = 'chat' | 'tools' | 'vision' | 'json_schema'

export type ProviderStatus = 'connected' | 'error' | 'unknown'

export type SafeModelProvider = {
  readonly providerId: string
  readonly providerType: ProviderType
  readonly displayName: string
  readonly baseURL: string
  readonly organization: string | null
  readonly project: string | null
  readonly defaultModel: string
  readonly availableModels: readonly string[]
  readonly enabled: boolean
  readonly status: ProviderStatus
  readonly capabilities: readonly ProviderCapability[]
  readonly supportsStreaming: boolean
  readonly healthStatus: ProviderStatus
  readonly modelCount: number
  readonly latencyMs: number
  readonly lastDiscoveryAt: string | null
  readonly lastStreamAt: string | null
  readonly lastErrorCode: string | null
  readonly streamingVerified: boolean
  readonly hasApiKey: boolean
  readonly maskedKey: string | null
  readonly lastTestedAt: string | null
  readonly lastError: string | null
  readonly createdAt: string
  readonly updatedAt: string
}

export type SaveModelProviderInput = {
  readonly providerId: string
  readonly providerType: ProviderType
  readonly displayName: string
  readonly baseURL: string
  readonly organization: string | null
  readonly project: string | null
  readonly defaultModel: string
  readonly availableModels: readonly string[]
  readonly enabled: boolean
  readonly capabilities: readonly ProviderCapability[]
  readonly apiKey?: string
}

export type ModelProfile = {
  readonly profileId: string
  readonly displayName: string
  readonly providerId: string
  readonly model: string
  readonly temperature: number
  readonly maxTokens: number
  readonly contextWindow: number
  readonly responseFormat: 'text' | 'json_object' | 'json_schema'
  readonly toolMode: 'none' | 'auto' | 'required'
  readonly fallbackProfileId: string | null
  readonly timeoutMs: number
  readonly purpose: string
  readonly enabled: boolean
  readonly createdAt: string
  readonly updatedAt: string
}

export type SaveModelProfileInput = Omit<ModelProfile, 'createdAt' | 'updatedAt'>

export type TestResult = {
  readonly ok: boolean
  readonly targetId: string
  readonly message: string
  readonly latencyMs: number
  readonly trace_id: string
}

export type AgentRuntimeConfig = {
  readonly contextWindow: number
  readonly temperature: number
  readonly maxTokens: number
}

export type AgentPlannerConfig = {
  readonly enabled: boolean
  readonly strategy: 'plan-execute' | 'react' | 'manual'
}

export type AgentExecutorConfig = {
  readonly timeoutMs: number
  readonly retryPolicy: string
}

export type AgentConfig = {
  readonly agentId: string
  readonly displayName: string
  readonly role: string
  readonly description: string
  readonly systemPrompt: string
  readonly modelProfileId: string
  readonly providerId: string
  readonly model: string
  readonly enabledSkills: readonly string[]
  readonly enabledTools: readonly string[]
  readonly enabledMcpServers: readonly string[]
  readonly runtimeConfig: AgentRuntimeConfig
  readonly planner: AgentPlannerConfig
  readonly executor: AgentExecutorConfig
  readonly memoryScope: 'short_term' | 'project' | 'semantic'
  readonly enabled: boolean
  readonly builtIn: boolean
  readonly createdAt: string
  readonly updatedAt: string
}

export type SaveAgentInput = Omit<AgentConfig, 'createdAt' | 'updatedAt' | 'builtIn'> & {
  readonly builtIn?: boolean
}

export type SkillConfig = {
  readonly skillId: string
  readonly commandName: string
  readonly displayName: string
  readonly description: string
  readonly whenToUse: string
  readonly instructions: string
  readonly category: 'explore' | 'product' | 'development' | 'sales' | 'general'
  readonly version: string
  readonly enabled: boolean
  readonly builtIn: boolean
  readonly sourcePath: string
  readonly requiredCapabilities: readonly string[]
  readonly outputArtifacts: readonly string[]
}

export type SaveSkillInput = Omit<SkillConfig, 'builtIn'> & {
  readonly builtIn?: boolean
}

export type ToolConfig = {
  readonly toolId: string
  readonly displayName: string
  readonly description: string
  readonly category: 'artifact' | 'browser' | 'search' | 'model' | 'human' | 'project'
  readonly riskLevel: 'low' | 'medium' | 'high' | 'critical'
  readonly enabled: boolean
  readonly builtIn: boolean
}

export type SaveToolInput = ToolConfig

export type McpServerConfig = {
  readonly serverId: string
  readonly displayName: string
  readonly command: string
  readonly args: readonly string[]
  readonly envKeys: readonly string[]
  readonly url: string | null
  readonly trustLevel:
    'trusted_builtin' | 'verified_partner' | 'community' | 'local_unverified' | 'remote_untrusted'
  readonly enabled: boolean
  readonly hasSecrets: boolean
  readonly maskedSecrets: readonly string[]
  readonly createdAt: string
  readonly updatedAt: string
}

export type SaveMcpServerInput = {
  readonly serverId: string
  readonly displayName: string
  readonly command: string
  readonly args: readonly string[]
  readonly url: string | null
  readonly trustLevel: McpServerConfig['trustLevel']
  readonly enabled: boolean
  readonly secrets?: Record<string, string>
}

export type Project = {
  readonly projectId: string
  readonly title: string
  readonly description: string
  readonly status: 'active' | 'paused' | 'archived'
  readonly defaultModelProfileId: string
  readonly enabledAgents: readonly string[]
  readonly enabledSkills: readonly string[]
  readonly enabledTools: readonly string[]
  readonly enabledMcpServers: readonly string[]
  readonly createdAt: string
  readonly updatedAt: string
}

export type CreateProjectInput = {
  readonly title: string
  readonly description: string
}

export type UpdateProjectInput = Partial<
  Pick<
    Project,
    | 'title'
    | 'description'
    | 'status'
    | 'defaultModelProfileId'
    | 'enabledAgents'
    | 'enabledSkills'
    | 'enabledTools'
    | 'enabledMcpServers'
  >
> & {
  readonly projectId: string
}

export type DeleteProjectInput = {
  readonly projectId: string
}

export type ProjectModuleId = 'explore' | 'product' | 'development' | 'sales'

export type ProjectModuleStatus = 'idle' | 'ready' | 'running' | 'blocked' | 'completed'

export type ProjectSubmodule = {
  readonly projectId: string
  readonly moduleId: ProjectModuleId
  readonly submoduleId: string
  readonly displayName: string
  readonly status: ProjectModuleStatus
  readonly summary: string
  readonly defaultAgents: readonly string[]
  readonly enabledSkills: readonly string[]
  readonly enabledTools: readonly string[]
  readonly outputArtifacts: readonly string[]
  readonly nextBestAction: string
  readonly config: Record<string, string | number | boolean>
}

export type ProjectModule = {
  readonly projectId: string
  readonly moduleId: ProjectModuleId
  readonly displayName: string
  readonly status: ProjectModuleStatus
  readonly summary: string
  readonly defaultAgents: readonly string[]
  readonly enabledSkills: readonly string[]
  readonly enabledTools: readonly string[]
  readonly enabledMcpServers: readonly string[]
  readonly outputArtifacts: readonly string[]
  readonly nextBestAction: string
  readonly submodules: readonly ProjectSubmodule[]
  readonly config: Record<string, string | number | boolean>
}

export type UpdateProjectModuleConfigInput = {
  readonly projectId: string
  readonly moduleId: ProjectModuleId
  readonly config: Record<string, string | number | boolean>
}

export type ChatSession = {
  readonly sessionId: string
  readonly projectId: string | null
  readonly title: string
  readonly agentId: string
  readonly modelProfileId: string
  readonly providerId: string
  readonly model: string
  readonly messageCount: number
  readonly createdAt: string
  readonly updatedAt: string
}

export type CreateChatSessionInput = {
  readonly projectId: string | null
  readonly title: string
  readonly agentId: string
  readonly modelProfileId: string
  readonly providerId?: string
  readonly model?: string
}

export type UpdateChatSessionInput = CreateChatSessionInput & {
  readonly sessionId: string
}

export type ChatMessage = {
  readonly messageId: string
  readonly attemptId: string
  readonly sessionId: string
  readonly role: 'user' | 'assistant' | 'system'
  readonly content: string
  readonly status: 'streaming' | 'completed' | 'failed' | 'cancelled'
  readonly providerId: string
  readonly model: string
  readonly usage: ChatModelUsage | null
  readonly finishReason: string
  readonly runtimeSummary: string
  readonly trace_id: string
  readonly createdAt: string
}

export type SendChatMessageInput = {
  readonly sessionId: string
  readonly content: string
  readonly streamId?: string
  readonly retryOfMessageId?: string
}

export type CancelChatStreamInput = {
  readonly streamId: string
}

export type ChatExecutionStep = {
  readonly stepId: string
  readonly phase: 'PLAN' | 'GRAPH' | 'EXECUTE' | 'OBSERVE' | 'REPLAN'
  readonly title: string
  readonly summary: string
  readonly status: 'ready' | 'running' | 'completed' | 'blocked' | 'error'
  readonly startedAt: string
  readonly completedAt: string
}

export type ChatToolCallPreview = {
  readonly callId: string
  readonly toolId: string
  readonly displayName: string
  readonly riskLevel: ToolConfig['riskLevel']
  readonly approvalRequired: boolean
  readonly status: 'preview' | 'pending_approval' | 'running' | 'completed' | 'blocked'
  readonly summary: string
  readonly arguments?: string
  readonly resultSummary?: string
  readonly errorCode?: string
}

export type ChatModelUsage = {
  readonly inputTokens: number
  readonly outputTokens: number
  readonly totalTokens: number
  readonly costUsd: number
}

export type ChatStreamError = {
  readonly code: string
  readonly message: string
  readonly recoverable: boolean
}

export type ChatStreamWarning = {
  readonly code: string
  readonly message: string
}

export type ChatTurnResult = {
  readonly session: ChatSession
  readonly messages: readonly ChatMessage[]
  readonly executionSteps: readonly ChatExecutionStep[]
  readonly toolCalls: readonly ChatToolCallPreview[]
  readonly contextSummary: ChatContextSummary | null
  readonly contextBudget: ContextBudgetReport
  readonly auditSummary: ChatAuditSummary
  readonly providerStatus: string
  readonly runtimeSnapshot: string
  readonly runtimeSummary: string
}

export type ChatAuditSummary = {
  readonly contentHash: string
  readonly providerId: string
  readonly model: string
  readonly latencyMs: number
  readonly errorCode: string
  readonly usage: ChatModelUsage | null
  readonly finishReason: string
}

export type ContextBudgetReport = {
  readonly contextWindow: number
  readonly maxOutputTokens: number
  readonly inputBudgetTokens: number
  readonly estimatedTokens: number
  readonly systemTokens: number
  readonly recentMessageTokens: number
  readonly summaryTokens: number
  readonly recentMessageCount: number
  readonly compactedCount: number
  readonly compacted: boolean
  readonly warnings: readonly string[]
}

export type ChatContextSummary = {
  readonly summaryId: string
  readonly sessionId: string
  readonly sourceMessageIds: readonly string[]
  readonly content: string
  readonly contentHash: string
  readonly tokenEstimate: number
  readonly createdBy: string
  readonly contextVersion: number
  readonly createdAt: string
}

export type SkillRuntimeDescriptor = {
  readonly skillId: string
  readonly displayName: string
  readonly instruction: string
  readonly requiredCapabilities: readonly string[]
  readonly outputArtifacts: readonly string[]
  readonly runtimePolicy: string
}

export type ToolRuntimeDescriptor = {
  readonly toolId: string
  readonly displayName: string
  readonly description: string
  readonly riskLevel: ToolConfig['riskLevel']
  readonly autoExecutable: boolean
  readonly approvalRequired: boolean
}

export type ChatRuntimeSelection = {
  readonly summary: string
  readonly skills: readonly SkillRuntimeDescriptor[]
  readonly tools: readonly ToolRuntimeDescriptor[]
  readonly mcpServers: readonly string[]
}

export type ToolExecutionResult = {
  readonly callId: string
  readonly toolId: string
  readonly status: 'completed' | 'blocked' | 'failed'
  readonly outputSummary: string
  readonly errorCode: string
  readonly errorMessage: string
  readonly latencyMs: number
}

export type ChatStreamEvent = {
  readonly type:
    | 'started'
    | 'step'
    | 'context_compacted'
    | 'reasoning_delta'
    | 'token_delta'
    | 'tool_call_delta'
    | 'tool_started'
    | 'tool_result'
    | 'tool_blocked'
    | 'skill_used'
    | 'usage'
    | 'completed'
    | 'failed'
    | 'cancelled'
  readonly streamId: string
  readonly sessionId: string
  readonly messageId: string
  readonly trace_id: string
  readonly sequence: number
  readonly timestamp: string
  readonly delta?: string
  readonly reasoningDelta?: string
  readonly step?: ChatExecutionStep
  readonly toolCall?: ChatToolCallPreview
  readonly toolResult?: ToolExecutionResult
  readonly runtimeSelection?: ChatRuntimeSelection
  readonly contextBudget?: ContextBudgetReport
  readonly contextSummary?: ChatContextSummary
  readonly usage?: ChatModelUsage
  readonly result?: ChatTurnResult
  readonly error?: ChatStreamError
  readonly warning?: ChatStreamWarning
  readonly providerId?: string
  readonly model?: string
  readonly finishReason?: string
  readonly attemptId?: string
  readonly latencyMs?: number
}

export type ChatStreamStartResult = {
  readonly streamId: string
}

export type ChatStreamController = {
  readonly streamId: string
  readonly cancel: () => Promise<void>
}

export type DeleteResult = {
  readonly ok: true
  readonly deletedId: string
}

export type DreamWorkerApi = {
  readonly runtime: {
    readonly ping: () => Promise<RuntimePingResponse>
  }
  readonly models: {
    readonly listProviders: () => Promise<readonly SafeModelProvider[]>
    readonly saveProvider: (input: SaveModelProviderInput) => Promise<SafeModelProvider>
    readonly deleteProvider: (providerId: string) => Promise<DeleteResult>
    readonly testProvider: (providerId: string) => Promise<TestResult>
    readonly refreshProviderModels: (providerId: string) => Promise<SafeModelProvider>
    readonly listModelProfiles: () => Promise<readonly ModelProfile[]>
    readonly saveModelProfile: (input: SaveModelProfileInput) => Promise<ModelProfile>
    readonly deleteModelProfile: (profileId: string) => Promise<DeleteResult>
  }
  readonly agents: {
    readonly listAgents: () => Promise<readonly AgentConfig[]>
    readonly getAgent: (agentId: string) => Promise<AgentConfig>
    readonly saveAgent: (input: SaveAgentInput) => Promise<AgentConfig>
    readonly duplicateAgent: (agentId: string) => Promise<AgentConfig>
    readonly deleteAgent: (agentId: string) => Promise<DeleteResult>
  }
  readonly skills: {
    readonly listSkills: () => Promise<readonly SkillConfig[]>
    readonly getSkill: (skillId: string) => Promise<SkillConfig>
    readonly saveSkill: (input: SaveSkillInput) => Promise<SkillConfig>
    readonly deleteSkill: (skillId: string) => Promise<DeleteResult>
  }
  readonly tools: {
    readonly listTools: () => Promise<readonly ToolConfig[]>
    readonly getTool: (toolId: string) => Promise<ToolConfig>
    readonly saveTool: (input: SaveToolInput) => Promise<ToolConfig>
    readonly setToolEnabled: (toolId: string, enabled: boolean) => Promise<ToolConfig>
    readonly deleteTool: (toolId: string) => Promise<DeleteResult>
  }
  readonly mcp: {
    readonly listServers: () => Promise<readonly McpServerConfig[]>
    readonly saveServer: (input: SaveMcpServerInput) => Promise<McpServerConfig>
    readonly deleteServer: (serverId: string) => Promise<DeleteResult>
    readonly testServer: (serverId: string) => Promise<TestResult>
    readonly refreshTools: (serverId: string) => Promise<readonly ToolConfig[]>
  }
  readonly projects: {
    readonly listProjects: () => Promise<readonly Project[]>
    readonly createProject: (input: CreateProjectInput) => Promise<Project>
    readonly getProject: (projectId: string) => Promise<Project>
    readonly updateProject: (input: UpdateProjectInput) => Promise<Project>
    readonly deleteProject: (input: DeleteProjectInput) => Promise<DeleteResult>
    readonly listProjectModules: (projectId: string) => Promise<readonly ProjectModule[]>
    readonly getProjectModule: (
      projectId: string,
      moduleId: ProjectModuleId
    ) => Promise<ProjectModule>
    readonly updateProjectModuleConfig: (
      input: UpdateProjectModuleConfigInput
    ) => Promise<ProjectModule>
  }
  readonly chat: {
    readonly listSessions: () => Promise<readonly ChatSession[]>
    readonly createSession: (input: CreateChatSessionInput) => Promise<ChatSession>
    readonly updateSession: (input: UpdateChatSessionInput) => Promise<ChatSession>
    readonly getMessages: (sessionId: string) => Promise<readonly ChatMessage[]>
    readonly sendMessage: (input: SendChatMessageInput) => Promise<ChatTurnResult>
    readonly streamMessage: (
      input: SendChatMessageInput,
      onEvent: (event: ChatStreamEvent) => void
    ) => Promise<ChatStreamController>
    readonly cancelStream: (input: CancelChatStreamInput) => Promise<DeleteResult>
    readonly deleteSession: (sessionId: string) => Promise<DeleteResult>
  }
}

export function createEngineNotConnectedResponse(traceId: string): RuntimePingResponse {
  return {
    schema_version: CONTRACT_SCHEMA_VERSION,
    ok: false,
    trace_id: traceId,
    error: {
      code: 'ENGINE_NOT_CONNECTED',
      message: 'Go Engine 尚未连接，后续阶段会接入本地引擎。',
      recoverable: true,
      user_action: '等待引擎接入后重试。',
      trace_id: traceId
    }
  }
}
