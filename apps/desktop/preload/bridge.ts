import {
  CHANNELS,
  type CancelChatStreamInput,
  type ChatMessage,
  type AgentConfig,
  type ChatSession,
  type ChatStreamController,
  type ChatStreamEvent,
  type ChatStreamStartResult,
  type ChatTurnResult,
  type CreateChatSessionInput,
  type CreateProjectInput,
  type DeleteProjectInput,
  type DeleteResult,
  type DreamWorkerApi,
  type AppSettings,
  type ExtensionActionResult,
  type ExtensionLogLine,
  type ExtensionModelRefreshResult,
  type ExtensionSpec,
  type ExtensionStatus,
  type ExtensionStreamingResult,
  type InstallExtensionInput,
  type McpServerConfig,
  type ModelProfile,
  type Project,
  type ProjectModule,
  type ProjectModuleId,
  type RuntimePingResponse,
  type SafeModelProvider,
  type SaveAgentInput,
  type SaveMcpServerInput,
  type SaveModelProfileInput,
  type SaveModelProviderInput,
  type SaveSkillInput,
  type SaveToolInput,
  type SendChatMessageInput,
  type SkillConfig,
  type TestResult,
  type ToolConfig,
  type UpdateSettingsInput,
  type UpdateChatSessionInput,
  type UpdateProjectInput,
  type UpdateProjectModuleConfigInput
} from '../shared/dreamworker-api'

type IpcInvoke = (channel: string, ...args: readonly unknown[]) => Promise<unknown>
type IpcListen = (channel: string, listener: (payload: unknown) => void) => () => void

async function invokeTyped<T>(
  invoke: IpcInvoke,
  channel: string,
  ...args: readonly unknown[]
): Promise<T> {
  return (await invoke(channel, ...args)) as T
}

function createClientStreamId(): string {
  return `stream_${Date.now()}_${Math.random().toString(36).slice(2)}`
}

export function createDreamWorkerApi(invoke: IpcInvoke, listen?: IpcListen): DreamWorkerApi {
  return {
    runtime: {
      ping: () => invokeTyped<RuntimePingResponse>(invoke, CHANNELS.runtimePing)
    },
    models: {
      listProviders: () =>
        invokeTyped<readonly SafeModelProvider[]>(invoke, CHANNELS.modelsListProviders),
      saveProvider: (input: SaveModelProviderInput) =>
        invokeTyped<SafeModelProvider>(invoke, CHANNELS.modelsSaveProvider, input),
      deleteProvider: (providerId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.modelsDeleteProvider, { providerId }),
      testProvider: (providerId: string) =>
        invokeTyped<TestResult>(invoke, CHANNELS.modelsTestProvider, { providerId }),
      refreshProviderModels: (providerId: string) =>
        invokeTyped<SafeModelProvider>(invoke, CHANNELS.modelsRefreshProviderModels, {
          providerId
        }),
      listModelProfiles: () =>
        invokeTyped<readonly ModelProfile[]>(invoke, CHANNELS.modelsListProfiles),
      saveModelProfile: (input: SaveModelProfileInput) =>
        invokeTyped<ModelProfile>(invoke, CHANNELS.modelsSaveProfile, input),
      deleteModelProfile: (profileId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.modelsDeleteProfile, { profileId })
    },
    settings: {
      getSettings: () => invokeTyped<AppSettings>(invoke, CHANNELS.settingsGet),
      updateSettings: (input: UpdateSettingsInput) =>
        invokeTyped<AppSettings>(invoke, CHANNELS.settingsUpdate, input),
      resetExtensionSettings: (extensionId: string) =>
        invokeTyped<AppSettings>(invoke, CHANNELS.settingsResetExtension, { extensionId })
    },
    extensions: {
      listExtensions: () => invokeTyped<readonly ExtensionSpec[]>(invoke, CHANNELS.extensionsList),
      getExtensionStatus: (extensionId: string) =>
        invokeTyped<ExtensionStatus>(invoke, CHANNELS.extensionsGetStatus, { extensionId }),
      detectExtension: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsDetect, { extensionId }),
      installExtension: (input: InstallExtensionInput) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsInstall, input),
      startExtension: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsStart, { extensionId }),
      stopExtension: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsStop, { extensionId }),
      restartExtension: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsRestart, { extensionId }),
      testExtension: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsTest, { extensionId }),
      refreshExtensionModels: (extensionId: string) =>
        invokeTyped<ExtensionModelRefreshResult>(invoke, CHANNELS.extensionsRefreshModels, {
          extensionId
        }),
      verifyExtensionStreaming: (extensionId: string) =>
        invokeTyped<ExtensionStreamingResult>(invoke, CHANNELS.extensionsVerifyStreaming, {
          extensionId
        }),
      tailExtensionLogs: (extensionId: string, options?: { readonly limit?: number }) =>
        invokeTyped<readonly ExtensionLogLine[]>(invoke, CHANNELS.extensionsTailLogs, {
          extensionId,
          limit: options?.limit
        }),
      clearExtensionLogs: (extensionId: string) =>
        invokeTyped<ExtensionActionResult>(invoke, CHANNELS.extensionsClearLogs, { extensionId })
    },
    agents: {
      listAgents: () => invokeTyped<readonly AgentConfig[]>(invoke, CHANNELS.agentsList),
      getAgent: (agentId: string) =>
        invokeTyped<AgentConfig>(invoke, CHANNELS.agentsGet, { agentId }),
      saveAgent: (input: SaveAgentInput) =>
        invokeTyped<AgentConfig>(invoke, CHANNELS.agentsSave, input),
      duplicateAgent: (agentId: string) =>
        invokeTyped<AgentConfig>(invoke, CHANNELS.agentsDuplicate, { agentId }),
      deleteAgent: (agentId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.agentsDelete, { agentId })
    },
    skills: {
      listSkills: () => invokeTyped<readonly SkillConfig[]>(invoke, CHANNELS.skillsList),
      getSkill: (skillId: string) =>
        invokeTyped<SkillConfig>(invoke, CHANNELS.skillsGet, { skillId }),
      saveSkill: (input: SaveSkillInput) =>
        invokeTyped<SkillConfig>(invoke, CHANNELS.skillsSave, input),
      deleteSkill: (skillId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.skillsDelete, { skillId })
    },
    tools: {
      listTools: () => invokeTyped<readonly ToolConfig[]>(invoke, CHANNELS.toolsList),
      getTool: (toolId: string) => invokeTyped<ToolConfig>(invoke, CHANNELS.toolsGet, { toolId }),
      saveTool: (input: SaveToolInput) =>
        invokeTyped<ToolConfig>(invoke, CHANNELS.toolsSave, input),
      setToolEnabled: (toolId: string, enabled: boolean) =>
        invokeTyped<ToolConfig>(invoke, CHANNELS.toolsSetEnabled, { toolId, enabled }),
      deleteTool: (toolId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.toolsDelete, { toolId })
    },
    mcp: {
      listServers: () => invokeTyped<readonly McpServerConfig[]>(invoke, CHANNELS.mcpListServers),
      saveServer: (input: SaveMcpServerInput) =>
        invokeTyped<McpServerConfig>(invoke, CHANNELS.mcpSaveServer, input),
      deleteServer: (serverId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.mcpDeleteServer, { serverId }),
      testServer: (serverId: string) =>
        invokeTyped<TestResult>(invoke, CHANNELS.mcpTestServer, { serverId }),
      refreshTools: (serverId: string) =>
        invokeTyped<readonly ToolConfig[]>(invoke, CHANNELS.mcpRefreshTools, { serverId })
    },
    projects: {
      listProjects: () => invokeTyped<readonly Project[]>(invoke, CHANNELS.projectsList),
      createProject: (input: CreateProjectInput) =>
        invokeTyped<Project>(invoke, CHANNELS.projectsCreate, input),
      getProject: (projectId: string) =>
        invokeTyped<Project>(invoke, CHANNELS.projectsGet, { projectId }),
      updateProject: (input: UpdateProjectInput) =>
        invokeTyped<Project>(invoke, CHANNELS.projectsUpdate, input),
      deleteProject: (input: DeleteProjectInput) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.projectsDelete, input),
      listProjectModules: (projectId: string) =>
        invokeTyped<readonly ProjectModule[]>(invoke, CHANNELS.projectsListModules, { projectId }),
      getProjectModule: (projectId: string, moduleId: ProjectModuleId) =>
        invokeTyped<ProjectModule>(invoke, CHANNELS.projectsGetModule, { projectId, moduleId }),
      updateProjectModuleConfig: (input: UpdateProjectModuleConfigInput) =>
        invokeTyped<ProjectModule>(invoke, CHANNELS.projectsUpdateModuleConfig, input)
    },
    chat: {
      listSessions: () => invokeTyped<readonly ChatSession[]>(invoke, CHANNELS.chatListSessions),
      createSession: (input: CreateChatSessionInput) =>
        invokeTyped<ChatSession>(invoke, CHANNELS.chatCreateSession, input),
      updateSession: (input: UpdateChatSessionInput) =>
        invokeTyped<ChatSession>(invoke, CHANNELS.chatUpdateSession, input),
      getMessages: (sessionId: string) =>
        invokeTyped<readonly ChatMessage[]>(invoke, CHANNELS.chatGetMessages, { sessionId }),
      sendMessage: (input: SendChatMessageInput) =>
        invokeTyped<ChatTurnResult>(invoke, CHANNELS.chatSendMessage, input),
      streamMessage: async (
        input: SendChatMessageInput,
        onEvent: (event: ChatStreamEvent) => void
      ): Promise<ChatStreamController> => {
        if (!listen) {
          const result = await invokeTyped<ChatTurnResult>(invoke, CHANNELS.chatSendMessage, input)
          onEvent({
            type: 'completed',
            streamId: input.streamId ?? createClientStreamId(),
            sessionId: result.session.sessionId,
            messageId: result.messages[result.messages.length - 1]?.messageId ?? '',
            trace_id: result.messages[result.messages.length - 1]?.trace_id ?? '',
            sequence: 1,
            timestamp: result.session.updatedAt,
            result
          })
          return {
            streamId: input.streamId ?? '',
            cancel: async () => undefined
          }
        }
        const streamId = input.streamId ?? createClientStreamId()
        let unsubscribe = (): void => undefined
        unsubscribe = listen(CHANNELS.chatStreamEvent, (payload) => {
          const event = payload as ChatStreamEvent
          if (event.streamId === streamId) {
            onEvent(event)
            if (
              event.type === 'completed' ||
              event.type === 'failed' ||
              event.type === 'cancelled'
            ) {
              unsubscribe()
            }
          }
        })
        try {
          const result = await invokeTyped<ChatStreamStartResult>(
            invoke,
            CHANNELS.chatStartStream,
            { ...input, streamId }
          )
          return {
            streamId: result.streamId,
            cancel: async () => {
              unsubscribe()
              try {
                await invokeTyped<unknown>(invoke, CHANNELS.chatCancelStream, {
                  streamId: result.streamId
                })
              } catch {
                // Local stream cancellation is best-effort; terminal events may already have closed it.
              }
            }
          }
        } catch (error) {
          unsubscribe()
          throw error
        }
      },
      cancelStream: (input: CancelChatStreamInput) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.chatCancelStream, input),
      deleteSession: (sessionId: string) =>
        invokeTyped<DeleteResult>(invoke, CHANNELS.chatDeleteSession, { sessionId })
    }
  }
}
