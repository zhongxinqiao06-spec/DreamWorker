import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type {
  DreamWorkerApi,
  Project,
  ProjectDirectoryCheck,
  ProjectModule,
  ProjectSubmodule
} from '../../../shared/dreamworker-api'
import {
  ALL_MODEL_ROUTE_SOURCE,
  isRoutedModelProvider,
  modelsForRouteSource,
  routeSourceOptionsForModels,
  useAppShellStore
} from './app-shell'
import { createDefaultProjectModuleConfigs } from './project-draft'

function projectWorkspaceDefaults() {
  return {
    localRootPath: null,
    localDirectoryStatus: 'not_set' as const,
    localDirectoryLastCheckedAt: null,
    defaultRouteProfileId: null,
    moduleConfigs: createDefaultProjectModuleConfigs(),
    memoryConfig: {
      projectMemoryEnabled: true,
      artifactIndexEnabled: true,
      localFileIndexEnabled: false,
      maxContextTokens: 64000
    },
    runPolicy: {
      plannerMode: 'plan_execute' as const,
      executorMode: 'safe' as const,
      maxRunCostUsd: 5,
      maxRunMinutes: 30,
      requireApprovalForHighRiskTools: true
    },
    securityPolicy: {
      fileAccessScope: 'project_directory_only' as const,
      allowWriteArtifacts: true,
      allowWriteSource: false,
      allowShellExecution: false,
      allowNetworkTools: true
    }
  }
}

function createExploreSubmodule(
  input: Pick<ProjectSubmodule, 'submoduleId' | 'displayName' | 'summary' | 'nextBestAction'> & {
    readonly status?: ProjectSubmodule['status']
    readonly artifacts?: readonly string[]
  }
): ProjectSubmodule {
  return {
    projectId: 'project_001',
    moduleId: 'explore',
    submoduleId: input.submoduleId,
    displayName: input.displayName,
    status: input.status ?? 'idle',
    summary: input.summary,
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_opportunity_scan'],
    enabledTools: ['tool_model_generate_stub'],
    outputArtifacts: input.artifacts ?? ['dream_brief.md'],
    nextBestAction: input.nextBestAction,
    config: { stage: 'Discover' }
  }
}

function createExploreModule(
  config: ProjectModule['config'] = { stage: 'Discover' }
): ProjectModule {
  return {
    projectId: 'project_001',
    moduleId: 'explore',
    displayName: '探索模块',
    status: 'ready',
    summary: '机会扫描和证据收集。',
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_opportunity_scan'],
    enabledTools: ['tool_model_generate_stub'],
    enabledMcpServers: [],
    outputArtifacts: ['dream_brief.md'],
    nextBestAction: '先跑机会扫描。',
    submodules: [
      createExploreSubmodule({
        submoduleId: 'opportunity_radar',
        displayName: '机会雷达',
        status: 'ready',
        summary: '扫描机会。',
        nextBestAction: '先生成机会清单。',
        artifacts: ['dream_brief.md', 'hypotheses.yaml']
      }),
      createExploreSubmodule({
        submoduleId: 'user_persona',
        displayName: '用户画像',
        summary: '结构化目标用户、场景和付费动机。',
        nextBestAction: '补齐 ICP 和痛点证据。',
        artifacts: ['personas.json']
      }),
      createExploreSubmodule({
        submoduleId: 'competitor_map',
        displayName: '竞品地图',
        summary: '整理替代方案和差异化空间。',
        nextBestAction: '确认竞品范围。',
        artifacts: ['competitor_matrix.xlsx']
      }),
      createExploreSubmodule({
        submoduleId: 'evidence_graph',
        displayName: '证据图谱',
        summary: '把假设、证据、风险和动作连成图谱。',
        nextBestAction: '证据不足时返回 ask_user。',
        artifacts: ['evidence_graph.json']
      })
    ],
    config
  }
}

function createDreamWorkerApiStub(): DreamWorkerApi {
  return {
    runtime: {
      ping: vi.fn().mockResolvedValue({
        schema_version: '0.1',
        ok: true,
        engineVersion: '0.1.0',
        trace_id: 'tr_store'
      })
    },
    system: {
      openExternal: vi.fn().mockResolvedValue({
        ok: true,
        url: 'http://localhost:20128/dashboard',
        message: null
      })
    },
    models: {
      listProviders: vi.fn().mockResolvedValue([
        {
          providerId: 'provider_deepseek',
          providerType: 'deepseek',
          displayName: 'DeepSeek 兼容服务',
          baseURL: 'https://api.deepseek.com',
          organization: null,
          project: null,
          defaultModel: 'deepseek-v4-flash',
          availableModels: ['deepseek-v4-flash'],
          enabled: true,
          status: 'connected',
          capabilities: ['chat', 'tools', 'json_schema'],
          supportsStreaming: true,
          healthStatus: 'connected',
          modelCount: 1,
          latencyMs: 18,
          lastDiscoveryAt: null,
          lastStreamAt: null,
          lastErrorCode: null,
          streamingVerified: true,
          hasApiKey: true,
          maskedKey: 'sk-b...4f3c',
          lastTestedAt: '2026-07-01T00:00:00Z',
          lastError: null,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      saveProvider: vi.fn().mockImplementation(async (input) => {
        const clonedInput = structuredClone(input)
        return {
          providerId: clonedInput.providerId,
          providerType: clonedInput.providerType,
          displayName: clonedInput.displayName,
          baseURL: clonedInput.baseURL,
          organization: clonedInput.organization,
          project: clonedInput.project,
          defaultModel: clonedInput.defaultModel,
          availableModels: clonedInput.availableModels,
          enabled: clonedInput.enabled,
          status: 'connected',
          capabilities: clonedInput.capabilities,
          supportsStreaming: true,
          healthStatus: 'connected',
          modelCount: clonedInput.availableModels.length,
          latencyMs: 18,
          lastDiscoveryAt: null,
          lastStreamAt: null,
          lastErrorCode: null,
          streamingVerified: true,
          hasApiKey: Boolean(clonedInput.apiKey),
          maskedKey: clonedInput.apiKey ? 'sk-t...cret' : null,
          lastTestedAt: '2026-07-01T00:00:00Z',
          lastError: null,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      }),
      deleteProvider: vi.fn(),
      testProvider: vi.fn().mockResolvedValue({
        ok: true,
        targetId: 'provider_deepseek',
        message: '连接检查已通过本地 Engine stub。',
        latencyMs: 18,
        trace_id: 'tr_provider'
      }),
      refreshProviderModels: vi.fn().mockResolvedValue({
        providerId: 'provider_deepseek',
        providerType: 'deepseek',
        displayName: 'DeepSeek 兼容服务',
        baseURL: 'https://api.deepseek.com',
        organization: null,
        project: null,
        defaultModel: 'deepseek-chat',
        availableModels: ['deepseek-chat', 'deepseek-reasoner'],
        enabled: true,
        status: 'connected',
        capabilities: ['chat', 'tools', 'json_schema'],
        supportsStreaming: true,
        healthStatus: 'connected',
        modelCount: 2,
        latencyMs: 22,
        lastDiscoveryAt: '2026-07-01T00:00:00Z',
        lastStreamAt: null,
        lastErrorCode: null,
        streamingVerified: true,
        hasApiKey: true,
        maskedKey: 'sk-b...4f3c',
        lastTestedAt: '2026-07-01T00:00:00Z',
        lastError: null,
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      }),
      listModelProfiles: vi.fn().mockResolvedValue([
        {
          profileId: 'profile_fast',
          displayName: '快速草稿模型',
          providerId: 'provider_deepseek',
          model: 'deepseek-v4-flash',
          temperature: 0.4,
          maxTokens: 4096,
          purpose: '聊天、探索、短产物生成',
          enabled: true,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      saveModelProfile: vi.fn().mockImplementation(async (input) => ({
        ...input,
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      })),
      deleteModelProfile: vi.fn().mockResolvedValue({ ok: true, deletedId: 'profile_fast' })
    },
    settings: {
      getSettings: vi.fn().mockResolvedValue({
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
      }),
      updateSettings: vi.fn().mockImplementation(async (input) => ({
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
        allowAgentsUseNineRouter: true,
        ...input
      })),
      resetExtensionSettings: vi.fn().mockResolvedValue({
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
      })
    },
    extensions: {
      listExtensions: vi.fn().mockResolvedValue([
        {
          extensionId: 'extension_9router',
          name: '9Router 本地模型路由器',
          kind: 'node_managed_provider',
          runtimeKind: 'node',
          description: 'OpenAI 兼容本地模型路由',
          install: {
            packageName: '9router',
            packageVersion: 'latest',
            runtimeDir: '',
            logDir: '',
            configDir: ''
          },
          process: { defaultCommand: '9router', defaultArgs: [], port: 20128, env: [] },
          health: {
            dashboardURL: 'http://localhost:20128',
            baseURL: 'http://localhost:20128/v1',
            modelsPath: '/models',
            chatPath: '/chat/completions'
          },
          providerBridge: {
            providerId: 'provider_9router_local',
            providerType: 'openai_compatible',
            displayName: '9Router 免费模型路由',
            baseURL: 'http://localhost:20128/v1',
            defaultModel: 'kr/claude-sonnet-4.5',
            sortOrder: 999,
            systemPreset: true,
            allowDeletion: false
          },
          capabilities: ['model_gateway'],
          security: {
            riskLevel: 'medium',
            allowedHosts: ['localhost'],
            secretKeys: ['NINEROUTER_API_KEY'],
            envAllowList: ['PATH'],
            managedRequiresExplicitEnable: true
          },
          systemPreset: true,
          enabled: true
        }
      ]),
      getExtensionStatus: vi.fn().mockResolvedValue({
        extensionId: 'extension_9router',
        installed: false,
        installSource: 'none',
        nodeAvailable: false,
        npmAvailable: false,
        runMode: 'external',
        processState: 'stopped',
        startedByDreamWorker: false,
        baseURL: 'http://localhost:20128/v1',
        dashboardURL: 'http://localhost:20128',
        healthStatus: 'unknown',
        modelCount: 1,
        models: ['kr/claude-sonnet-4.5'],
        streamingVerified: false,
        hasApiKey: false,
        logDir: '',
        workDir: '',
        runtime: {
          nodeAvailable: false,
          npmAvailable: false,
          commandAvailable: false,
          installSource: 'none'
        }
      }),
      detectExtension: vi.fn(),
      installExtension: vi.fn(),
      startExtension: vi.fn(),
      stopExtension: vi.fn(),
      restartExtension: vi.fn(),
      testExtension: vi.fn(),
      refreshExtensionModels: vi.fn(),
      verifyExtensionStreaming: vi.fn(),
      tailExtensionLogs: vi.fn().mockResolvedValue([]),
      clearExtensionLogs: vi.fn()
    },
    agents: {
      listAgents: vi.fn().mockResolvedValue([
        {
          agentId: 'agent_general_assistant',
          displayName: '通用助手',
          role: '普通 Agent 聊天入口',
          description: '处理日常问答、上下文整理和轻量任务拆解。',
          systemPrompt: '中文回答。',
          modelProfileId: 'profile_fast',
          enabledSkills: ['skill_opportunity_scan'],
          enabledTools: ['tool_model_generate_stub'],
          enabledMcpServers: [],
          runtimeConfig: { contextWindow: 128000, temperature: 0.4, maxTokens: 4096 },
          planner: { enabled: true, strategy: 'plan-execute' },
          executor: { timeoutMs: 120000, retryPolicy: 'retry_twice_then_ask' },
          memoryScope: 'project',
          enabled: true,
          builtIn: true,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      getAgent: vi.fn(),
      saveAgent: vi.fn(),
      duplicateAgent: vi.fn(),
      deleteAgent: vi.fn()
    },
    skills: {
      listSkills: vi.fn().mockResolvedValue([
        {
          skillId: 'skill_opportunity_scan',
          commandName: 'opportunity-scan',
          displayName: '机会扫描',
          description: '拆出目标人群和风险假设。',
          whenToUse: 'Use for opportunity scan.',
          instructions: '## Instructions\n\nScan the opportunity.',
          category: 'explore',
          version: '0.1.0',
          enabled: true,
          builtIn: true,
          sourcePath: 'C:/project/DreamWorker/.agent/skills/opportunity-scan/SKILL.md',
          requiredCapabilities: ['cap_model_generate_stub'],
          outputArtifacts: ['dream_brief.md']
        }
      ]),
      getSkill: vi.fn(),
      saveSkill: vi.fn(),
      deleteSkill: vi.fn()
    },
    tools: {
      listTools: vi.fn().mockResolvedValue([
        {
          toolId: 'tool_model_generate_stub',
          displayName: '模型生成 Stub',
          description: '确定性模型生成能力。',
          category: 'model',
          riskLevel: 'low',
          enabled: true,
          builtIn: true
        }
      ]),
      getTool: vi.fn(),
      saveTool: vi.fn().mockImplementation(async (input) => input),
      setToolEnabled: vi.fn().mockResolvedValue({
        toolId: 'tool_model_generate_stub',
        displayName: '模型生成 Stub',
        description: '确定性模型生成能力。',
        category: 'model',
        riskLevel: 'low',
        enabled: false,
        builtIn: true
      }),
      deleteTool: vi.fn().mockResolvedValue({ ok: true, deletedId: 'tool_model_generate_stub' })
    },
    mcp: {
      listServers: vi.fn().mockResolvedValue([
        {
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
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      saveServer: vi.fn(),
      deleteServer: vi.fn(),
      testServer: vi.fn(),
      refreshTools: vi.fn().mockResolvedValue([
        {
          toolId: 'mcp_mcp_local_files_read_file',
          displayName: 'read_file',
          description: 'Read file through MCP',
          category: 'project',
          riskLevel: 'low',
          enabled: true,
          builtIn: false
        }
      ])
    },
    projects: {
      listProjects: vi.fn().mockResolvedValue([
        {
          projectId: 'project_001',
          title: 'AI 项目孵化器',
          description: '项目空间种子数据。',
          status: 'active',
          ...projectWorkspaceDefaults(),
          defaultModelProfileId: 'profile_fast',
          enabledAgents: ['agent_general_assistant'],
          enabledSkills: ['skill_opportunity_scan'],
          enabledTools: ['tool_model_generate_stub'],
          enabledMcpServers: [],
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      createProject: vi.fn().mockResolvedValue({
        projectId: 'project_002',
        title: '新的 AI 项目',
        description: '新项目。',
        status: 'active',
        ...projectWorkspaceDefaults(),
        defaultModelProfileId: 'profile_fast',
        enabledAgents: [],
        enabledSkills: [],
        enabledTools: [],
        enabledMcpServers: [],
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      }),
      getProject: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        title: 'AI 项目孵化器',
        description: '项目空间种子数据。',
        status: 'active',
        ...projectWorkspaceDefaults(),
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        localDirectoryStatus: 'valid',
        localDirectoryLastCheckedAt: '2026-07-01T00:00:00Z',
        defaultModelProfileId: 'profile_fast',
        enabledAgents: ['agent_general_assistant'],
        enabledSkills: ['skill_opportunity_scan'],
        enabledTools: ['tool_model_generate_stub'],
        enabledMcpServers: [],
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      }),
      updateProject: vi.fn().mockImplementation(async (input) => {
        expect(() => structuredClone(input)).not.toThrow()
        return {
          projectId: input.projectId,
          title: input.title ?? 'AI 项目孵化器',
          description: input.description ?? '项目空间种子数据。',
          status: input.status ?? 'active',
          ...projectWorkspaceDefaults(),
          localRootPath: input.localRootPath ?? null,
          localDirectoryStatus: input.localRootPath ? 'invalid' : 'not_set',
          localDirectoryLastCheckedAt: null,
          defaultModelProfileId: input.defaultModelProfileId ?? 'profile_fast',
          defaultRouteProfileId: input.defaultRouteProfileId ?? null,
          enabledAgents: input.enabledAgents ?? ['agent_general_assistant'],
          enabledSkills: input.enabledSkills ?? ['skill_opportunity_scan'],
          enabledTools: input.enabledTools ?? ['tool_model_generate_stub'],
          enabledMcpServers: input.enabledMcpServers ?? [],
          moduleConfigs: input.moduleConfigs ?? createDefaultProjectModuleConfigs(),
          memoryConfig: input.memoryConfig ?? projectWorkspaceDefaults().memoryConfig,
          runPolicy: input.runPolicy ?? projectWorkspaceDefaults().runPolicy,
          securityPolicy: input.securityPolicy ?? projectWorkspaceDefaults().securityPolicy,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      }),
      deleteProject: vi.fn().mockResolvedValue({ ok: true, deletedId: 'project_001' }),
      pickLocalDirectory: vi.fn().mockResolvedValue('C:\\DreamWorkerProjects\\project_001'),
      validateLocalDirectory: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        status: 'valid',
        lastCheckedAt: '2026-07-01T00:00:00Z',
        exists: true,
        readable: true,
        writable: true,
        dreamworkerInitialized: true,
        requiredDirectories: [],
        message: '本地目录可用，项目结构完整。'
      }),
      initializeLocalDirectory: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        status: 'valid',
        lastCheckedAt: '2026-07-01T00:00:00Z',
        exists: true,
        readable: true,
        writable: true,
        dreamworkerInitialized: true,
        requiredDirectories: [],
        message: '本地目录可用，项目结构完整。'
      }),
      openLocalDirectory: vi.fn().mockResolvedValue({
        ok: true,
        projectId: 'project_001',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        message: '已打开项目本地目录。'
      }),
      exportProjectManifest: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        manifestPath: 'C:\\DreamWorkerProjects\\project_001\\.dreamworker\\manifest.json',
        manifest: {}
      }),
      listProjectModules: vi.fn().mockResolvedValue([createExploreModule()]),
      getProjectModule: vi.fn(),
      updateProjectModuleConfig: vi.fn().mockImplementation(async (input) =>
        createExploreModule({
          ...input.config,
          updatedByTest: true
        })
      ),
      importRequirementFiles: vi.fn().mockResolvedValue(null),
      listRequirementSources: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        sources: []
      }),
      previewRequirementSource: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        source: {
          sourceId: 'project_description',
          kind: 'project_description',
          fileName: '',
          relativePath: '',
          absolutePath: '',
          mimeType: 'text/plain',
          charCount: 0,
          importedAt: '2026-07-01T00:00:00Z',
          summary: ''
        },
        parser: 'direct_text',
        content: '',
        charCount: 0,
        truncated: false,
        traceId: 'tr_preview',
        createdAt: '2026-07-01T00:00:00Z'
      }),
      runRequirementAnalysis: vi.fn().mockResolvedValue({
        runId: 'req_001',
        projectId: 'project_001',
        status: 'completed',
        sources: [],
        analysis: {
          projectTitle: 'AI 项目孵化器',
          summary: '',
          sources: [],
          roles: [],
          features: [],
          nonFunctionalRequirements: [],
          risks: [],
          openQuestions: []
        },
        featureCount: 0,
        outputFiles: [],
        warnings: [],
        traceId: 'tr_req',
        createdAt: '2026-07-01T00:00:00Z'
      }),
      openRequirementOutputFile: vi.fn().mockResolvedValue({
        ok: true,
        path: 'C:\\DreamWorkerProjects\\project_001\\artifacts\\product\\requirements_spec.docx',
        message: '已打开文件。'
      })
    },
    chat: {
      listSessions: vi.fn().mockResolvedValue([
        {
          sessionId: 'chat_001',
          projectId: 'project_001',
          title: '普通 Agent 工作台',
          agentId: 'agent_general_assistant',
          modelProfileId: 'profile_fast',
          messageCount: 0,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        }
      ]),
      createSession: vi.fn(),
      updateSession: vi.fn().mockImplementation(async (input) => ({
        sessionId: input.sessionId,
        projectId: input.projectId,
        title: input.title,
        agentId: input.agentId,
        modelProfileId: input.modelProfileId,
        providerId: input.providerId,
        model: input.model,
        messageCount: 0,
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      })),
      getMessages: vi.fn().mockResolvedValue([]),
      sendMessage: vi.fn().mockResolvedValue({
        session: {
          sessionId: 'chat_001',
          projectId: 'project_001',
          title: '普通 Agent 工作台',
          agentId: 'agent_general_assistant',
          modelProfileId: 'profile_fast',
          messageCount: 2,
          createdAt: '2026-07-01T00:00:00Z',
          updatedAt: '2026-07-01T00:00:00Z'
        },
        messages: [
          {
            messageId: 'msg_001',
            sessionId: 'chat_001',
            role: 'user',
            content: '你好',
            trace_id: 'tr_chat',
            createdAt: '2026-07-01T00:00:00Z'
          },
          {
            messageId: 'msg_002',
            sessionId: 'chat_001',
            role: 'assistant',
            content: '已收到。',
            trace_id: 'tr_chat',
            createdAt: '2026-07-01T00:00:00Z'
          }
        ],
        executionSteps: [
          {
            stepId: 'step_plan',
            phase: 'PLAN',
            title: '生成计划',
            summary: '解析用户意图。',
            status: 'completed',
            startedAt: '2026-07-01T00:00:00Z',
            completedAt: '2026-07-01T00:00:00Z'
          }
        ],
        toolCalls: [
          {
            callId: 'call_tool_model_generate_stub',
            toolId: 'tool_model_generate_stub',
            displayName: '模型生成 Stub',
            riskLevel: 'low',
            approvalRequired: false,
            status: 'preview',
            summary: '工具调用预览。'
          }
        ],
        runtimeSummary: 'Agent=agent_general_assistant / Planner=plan-execute'
      }),
      streamMessage: vi.fn().mockImplementation(async (input) => ({
        streamId: input.streamId ?? 'stream_test',
        cancel: vi.fn()
      })),
      cancelStream: vi.fn(),
      deleteSession: vi.fn()
    },
    coding: {
      listEngines: vi.fn().mockResolvedValue({
        runtimeDir: '',
        nodeBin: 'node',
        adapterPath: '',
        available: true,
        message: 'ready',
        engines: [],
        engineStatuses: []
      }),
      createSession: vi.fn().mockResolvedValue({
        sessionId: 'coding_001',
        projectId: 'project_001',
        engineId: 'codex',
        providerId: 'provider_deepseek',
        model: 'deepseek-v4-flash',
        title: 'Coding session',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        engineThreadId: '',
        status: 'ready',
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      }),
      getSession: vi.fn().mockResolvedValue({
        sessionId: 'coding_001',
        projectId: 'project_001',
        engineId: 'codex',
        providerId: 'provider_deepseek',
        model: 'deepseek-v4-flash',
        title: 'Coding session',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001',
        engineThreadId: '',
        status: 'ready',
        createdAt: '2026-07-01T00:00:00Z',
        updatedAt: '2026-07-01T00:00:00Z'
      }),
      listFiles: vi.fn().mockResolvedValue([]),
      readFile: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        path: 'README.md',
        content: '',
        size: 0,
        truncated: false,
        mimeType: 'text/markdown'
      }),
      fileStatus: vi.fn().mockResolvedValue({
        projectId: 'project_001',
        branch: 'main',
        changes: [],
        clean: true,
        message: 'clean'
      }),
      streamTurn: vi.fn().mockImplementation(async (input) => ({
        streamId: input.streamId ?? 'coding_stream_test',
        cancel: vi.fn()
      })),
      cancelTurn: vi.fn().mockResolvedValue({ ok: true, deletedId: 'coding_stream_test' })
    }
  }
}

function stubDreamWorkerApi(api: DreamWorkerApi): void {
  vi.stubGlobal('window', { dreamworker: api })
}

describe('app shell workspace state', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.unstubAllGlobals()
  })

  it('loads resources, projects and chat sessions from the typed API', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()

    await store.loadWorkspace()

    expect(store.bootStatus).toBe('ready')
    expect(store.providers[0]?.maskedKey).toBe('sk-b...4f3c')
    expect(store.projects[0]?.projectId).toBe('project_001')
    expect(store.projectModules.every((module) => module.projectId === 'project_001')).toBe(true)
    expect(store.projectModules[0]?.submodules[0]?.displayName).toBe('机会雷达')
    expect(store.activeSubmodule?.submoduleId).toBe('opportunity_radar')
    expect(api.models.listProviders).toHaveBeenCalledTimes(1)
    expect(api.projects.listProjectModules).toHaveBeenCalledWith('project_001')
  })

  it('opens and closes every explore submodule detail page', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    for (const submoduleId of [
      'opportunity_radar',
      'user_persona',
      'competitor_map',
      'evidence_graph'
    ]) {
      store.enterSubmodule('explore', submoduleId)
      expect(store.activeSubmoduleDetail).toEqual({ moduleId: 'explore', submoduleId })
      expect(store.activeSubmodule?.submoduleId).toBe(submoduleId)
    }

    store.leaveSubmoduleDetail()

    expect(store.activeSubmoduleDetail).toBeNull()
  })

  it('persists explore module config through the typed API', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.updateProjectModuleConfig('explore', {
      stage: 'Discover',
      opportunity_radar_brief: '验证本地项目工作台机会'
    })

    expect(api.projects.updateProjectModuleConfig).toHaveBeenCalledWith({
      projectId: 'project_001',
      moduleId: 'explore',
      config: {
        stage: 'Discover',
        opportunity_radar_brief: '验证本地项目工作台机会'
      }
    })
    expect(store.projectModules[0]?.config).toMatchObject({
      opportunity_radar_brief: '验证本地项目工作台机会',
      updatedByTest: true
    })
  })

  it('imports requirement documents, previews MinerU output and stores generated artifacts', async () => {
    const api = createDreamWorkerApiStub()
    const projectSource = {
      sourceId: 'project_description',
      kind: 'project_description' as const,
      fileName: '',
      relativePath: '',
      absolutePath: '',
      mimeType: 'text/plain',
      charCount: 28,
      importedAt: '2026-07-01T00:00:00Z',
      summary: '项目描述'
    }
    const importedSource = {
      sourceId: 'src_contract_docx',
      kind: 'imported_file' as const,
      fileName: 'coding-agent-contract.docx',
      relativePath: 'imports/requirements/req_import/coding-agent-contract.docx',
      absolutePath:
        'C:\\DreamWorkerProjects\\project_001\\imports\\requirements\\req_import\\coding-agent-contract.docx',
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      charCount: 2400,
      importedAt: '2026-07-01T00:01:00Z',
      summary: '编码智能体合同式需求'
    }
    vi.mocked(api.projects.importRequirementFiles).mockResolvedValue({
      projectId: 'project_001',
      runId: 'req_import',
      sources: [importedSource],
      message: '已导入 1 个需求文件。'
    })
    vi.mocked(api.projects.listRequirementSources).mockResolvedValue({
      projectId: 'project_001',
      sources: [projectSource, importedSource]
    })
    vi.mocked(api.projects.previewRequirementSource).mockResolvedValue({
      projectId: 'project_001',
      source: importedSource,
      parser: 'mineru',
      content: 'FR-001 任务理解与执行计划\nFR-002 代码修改与验证',
      charCount: 38,
      truncated: false,
      traceId: 'tr_preview',
      createdAt: '2026-07-01T00:02:00Z'
    })
    vi.mocked(api.projects.runRequirementAnalysis).mockResolvedValue({
      runId: 'req_analysis',
      projectId: 'project_001',
      status: 'completed',
      sources: [projectSource, importedSource],
      featureCount: 1,
      outputFiles: [
        {
          kind: 'feature_excel',
          fileName: 'feature_list.xlsx',
          relativePath: 'artifacts/product/feature_list.xlsx',
          absolutePath:
            'C:\\DreamWorkerProjects\\project_001\\artifacts\\product\\feature_list.xlsx'
        },
        {
          kind: 'requirements_word',
          fileName: 'requirements_spec.docx',
          relativePath: 'artifacts/product/requirements_spec.docx',
          absolutePath:
            'C:\\DreamWorkerProjects\\project_001\\artifacts\\product\\requirements_spec.docx'
        }
      ],
      warnings: [],
      traceId: 'tr_req',
      createdAt: '2026-07-01T00:03:00Z',
      analysis: {
        projectTitle: '编码智能体',
        summary: '从合同式需求中抽取功能清单。',
        sources: ['coding-agent-contract.docx'],
        roles: ['项目负责人'],
        features: [
          {
            featureId: 'FR-001',
            module: '任务执行',
            name: '任务理解与执行计划',
            role: '项目负责人',
            scenario: '提交编码任务后获得计划与执行结果',
            description: '识别任务目标、约束和验收口径。',
            priority: 'P0',
            type: 'functional',
            inputs: ['合同式需求文档'],
            outputs: ['执行计划'],
            acceptanceCriteria: ['生成可追踪的计划项'],
            dependencies: [],
            sourceRefs: ['coding-agent-contract.docx'],
            notes: ''
          }
        ],
        nonFunctionalRequirements: ['生成文档可由 WPS 或 Word 打开'],
        risks: [],
        openQuestions: []
      }
    })
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.importRequirementFiles()

    expect(api.projects.previewRequirementSource).toHaveBeenCalledWith({
      projectId: 'project_001',
      sourceId: 'src_contract_docx'
    })
    expect(store.requirementSourcePreview?.parser).toBe('mineru')
    expect(store.selectedRequirementSourceIds).toEqual(['project_description', 'src_contract_docx'])

    await store.runRequirementAnalysis()

    expect(api.projects.runRequirementAnalysis).toHaveBeenCalledWith({
      projectId: 'project_001',
      sourceIds: ['project_description', 'src_contract_docx'],
      prompt: ''
    })
    expect(store.requirementAnalysisRun?.featureCount).toBe(1)
    expect(store.requirementAnalysisRun?.outputFiles.map((file) => file.fileName)).toEqual([
      'feature_list.xlsx',
      'requirements_spec.docx'
    ])
  })

  it('opens requirement output files through the typed API', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    const artifactPath =
      'C:\\DreamWorkerProjects\\project_001\\artifacts\\product\\requirements_spec.docx'
    await store.openRequirementOutputFile(artifactPath)

    expect(api.projects.openRequirementOutputFile).toHaveBeenCalledWith(artifactPath)
    expect(store.resourceNotice?.tone).toBe('success')
  })

  it('keeps runtime.ping as a status concern', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()

    await store.checkRuntimePing()

    expect(store.runtimePing).toEqual({
      status: 'ready',
      headline: '引擎已连接',
      detail: 'Main Runtime 0.1.0 已响应。',
      traceId: 'tr_store',
      errorCode: '暂无'
    })
  })

  it('groups routed models by upstream source prefix', () => {
    const models = ['cx/gpt-5.5', 'kr/claude-sonnet-4.5', 'kr/claude-haiku-4.5']

    expect(routeSourceOptionsForModels(models)).toEqual([
      { id: ALL_MODEL_ROUTE_SOURCE, label: '全部上游', modelCount: 3 },
      { id: 'cx', label: 'CX', modelCount: 1 },
      { id: 'kr', label: 'Kiro AI', modelCount: 2 }
    ])
    expect(modelsForRouteSource(models, 'kr')).toEqual([
      'kr/claude-sonnet-4.5',
      'kr/claude-haiku-4.5'
    ])
  })

  it('only enables routed model grouping for 9Router-like providers', () => {
    expect(
      isRoutedModelProvider({
        providerId: 'provider_9router_local',
        providerType: 'openai_compatible',
        displayName: '9Router 免费模型路由',
        baseURL: 'http://localhost:20128/v1'
      })
    ).toBe(true)
    expect(
      isRoutedModelProvider({
        providerId: 'provider_siliconflow',
        providerType: 'siliconflow',
        displayName: 'SiliconFlow',
        baseURL: 'https://api.siliconflow.cn/v1'
      })
    ).toBe(false)
  })

  it('saves provider drafts without exposing raw keys in provider state', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    store.providerDraft.apiKey = 'sk-test-secret'
    await store.saveProviderDraft()

    expect(api.models.saveProvider).toHaveBeenCalledWith(
      expect.objectContaining({ apiKey: 'sk-test-secret' })
    )
    expect(JSON.stringify(store.providers)).not.toContain('sk-test-secret')
    expect(store.providers[0]?.maskedKey).toBe('sk-t...cret')
  })

  it('switches blocked chat to a newly saved keyed provider', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    const currentProvider = store.providers[0]
    if (!currentProvider) {
      throw new Error('expected provider fixture')
    }
    store.providers = [{ ...currentProvider, hasApiKey: false, maskedKey: null }]
    expect(store.composerDisabledReason).toBe('缺少密钥')

    store.newProviderDraft('openai_compatible')
    store.providerDraft = {
      ...store.providerDraft,
      providerId: 'provider_custom_keyed',
      displayName: '自定义已配置服务商',
      baseURL: 'https://api.example.com/v1',
      defaultModel: 'custom-chat-model',
      availableModelsText: 'custom-chat-model',
      apiKey: 'sk-custom-secret'
    }

    await store.saveProviderDraft()

    expect(api.chat.updateSession).toHaveBeenCalledWith(
      expect.objectContaining({
        providerId: 'provider_custom_keyed',
        model: 'custom-chat-model'
      })
    )
    expect(store.activeChatProviderId).toBe('provider_custom_keyed')
    expect(store.composerDisabledReason).toBe('')
  })

  it('sends chat messages through the Engine chat API', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    store.chatDraft = '你好'
    await store.sendChatMessage()

    expect(api.chat.streamMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        sessionId: 'chat_001',
        content: '你好'
      }),
      expect.any(Function)
    )
    expect(store.chatMessages).toHaveLength(2)
    expect(store.chatDraft).toBe('')
  })

  it('persists active chat agent, model and project bindings through the Engine', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.setActiveChatAgent('agent_general_assistant')
    await store.setActiveChatModelProfile('profile_fast')
    await store.setActiveChatProject('')

    expect(api.chat.updateSession).toHaveBeenCalledWith(
      expect.objectContaining({
        sessionId: 'chat_001',
        agentId: 'agent_general_assistant',
        modelProfileId: 'profile_fast'
      })
    )
    expect(api.chat.updateSession).toHaveBeenLastCalledWith(
      expect.objectContaining({ projectId: null })
    )
  })

  it('saves and validates the active project local directory through typed APIs', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.chooseProjectLocalDirectory()

    expect(api.projects.pickLocalDirectory).toHaveBeenCalled()
    expect(api.projects.updateProject).toHaveBeenCalledWith(
      expect.objectContaining({
        projectId: 'project_001',
        localRootPath: 'C:\\DreamWorkerProjects\\project_001'
      })
    )
    expect(api.projects.validateLocalDirectory).toHaveBeenCalledWith('project_001')
    expect(store.activeProjectDirectoryCheck?.status).toBe('valid')
    expect(store.projectDraft.localRootPath).toBe('C:\\DreamWorkerProjects\\project_001')
  })

  it('saves a manually typed local directory before initializing it', async () => {
    const api = createDreamWorkerApiStub()
    let project: Project = {
      projectId: 'project_001',
      title: 'AI 项目孵化器',
      description: '项目空间种子数据。',
      status: 'active',
      ...projectWorkspaceDefaults(),
      defaultModelProfileId: 'profile_fast',
      enabledAgents: ['agent_general_assistant'],
      enabledSkills: ['skill_opportunity_scan'],
      enabledTools: ['tool_model_generate_stub'],
      enabledMcpServers: [],
      createdAt: '2026-07-01T00:00:00Z',
      updatedAt: '2026-07-01T00:00:00Z'
    }
    const validCheck = (path: string): ProjectDirectoryCheck => ({
      projectId: 'project_001',
      localRootPath: path,
      status: 'valid',
      lastCheckedAt: '2026-07-01T00:10:00Z',
      exists: true,
      readable: true,
      writable: true,
      dreamworkerInitialized: true,
      requiredDirectories: [],
      message: '本地目录可用，项目结构完整。'
    })
    Object.assign(api.projects, {
      listProjects: vi.fn().mockResolvedValue([project]),
      getProject: vi.fn().mockImplementation(async () => project),
      updateProject: vi.fn().mockImplementation(async (input) => {
        expect(() => structuredClone(input)).not.toThrow()
        project = {
          ...project,
          ...input,
          localRootPath: input.localRootPath ?? null,
          localDirectoryStatus: input.localRootPath ? 'invalid' : 'not_set',
          localDirectoryLastCheckedAt: null,
          updatedAt: '2026-07-01T00:05:00Z'
        }
        return project
      }),
      initializeLocalDirectory: vi.fn().mockImplementation(async () => {
        const check = validCheck(project.localRootPath ?? '')
        project = {
          ...project,
          localDirectoryStatus: check.status,
          localDirectoryLastCheckedAt: check.lastCheckedAt
        }
        return check
      })
    })
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    store.projectDraft.localRootPath = 'C:\\ManualDreamWorkerProject'
    await store.initializeActiveProjectDirectory()

    expect(api.projects.updateProject).toHaveBeenCalledWith(
      expect.objectContaining({
        projectId: 'project_001',
        localRootPath: 'C:\\ManualDreamWorkerProject'
      })
    )
    expect(api.projects.initializeLocalDirectory).toHaveBeenCalledWith('project_001')
    expect(vi.mocked(api.projects.updateProject).mock.invocationCallOrder[0]).toBeLessThan(
      vi.mocked(api.projects.initializeLocalDirectory).mock.invocationCallOrder[0] ??
        Number.MAX_SAFE_INTEGER
    )
    expect(store.activeProject?.localDirectoryStatus).toBe('valid')
    expect(store.activeProjectDirectoryCheck?.localRootPath).toBe('C:\\ManualDreamWorkerProject')
    expect(store.projectDraft.localRootPath).toBe('C:\\ManualDreamWorkerProject')
  })

  it('clears stale directory checks when a saved local path changes', async () => {
    const api = createDreamWorkerApiStub()
    let project: Project = {
      projectId: 'project_001',
      title: 'AI 项目孵化器',
      description: '项目空间种子数据。',
      status: 'active',
      ...projectWorkspaceDefaults(),
      localRootPath: 'C:\\OldDreamWorkerProject',
      localDirectoryStatus: 'valid',
      localDirectoryLastCheckedAt: '2026-07-01T00:00:00Z',
      defaultModelProfileId: 'profile_fast',
      enabledAgents: ['agent_general_assistant'],
      enabledSkills: ['skill_opportunity_scan'],
      enabledTools: ['tool_model_generate_stub'],
      enabledMcpServers: [],
      createdAt: '2026-07-01T00:00:00Z',
      updatedAt: '2026-07-01T00:00:00Z'
    }
    Object.assign(api.projects, {
      listProjects: vi.fn().mockResolvedValue([project]),
      getProject: vi.fn().mockImplementation(async () => project),
      updateProject: vi.fn().mockImplementation(async (input) => {
        expect(() => structuredClone(input)).not.toThrow()
        project = {
          ...project,
          ...input,
          localRootPath: input.localRootPath ?? null,
          localDirectoryStatus: input.localRootPath ? 'invalid' : 'not_set',
          localDirectoryLastCheckedAt: null,
          updatedAt: '2026-07-01T00:05:00Z'
        }
        return project
      })
    })
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()
    store.activeProjectDirectoryCheck = {
      projectId: 'project_001',
      localRootPath: 'C:\\OldDreamWorkerProject',
      status: 'valid',
      lastCheckedAt: '2026-07-01T00:00:00Z',
      exists: true,
      readable: true,
      writable: true,
      dreamworkerInitialized: true,
      requiredDirectories: [],
      message: '旧目录可用。'
    }

    store.projectDraft.localRootPath = 'C:\\NewDreamWorkerProject'
    await store.saveActiveProject()

    expect(api.projects.updateProject).toHaveBeenCalledWith(
      expect.objectContaining({
        projectId: 'project_001',
        localRootPath: 'C:\\NewDreamWorkerProject'
      })
    )
    expect(store.activeProjectDirectoryCheck).toBeNull()
    expect(store.activeProject?.localDirectoryStatus).toBe('invalid')
    expect(store.activeProject?.localDirectoryLastCheckedAt).toBeNull()
  })

  it('opens and exports the active project directory after saving dirty draft changes', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    store.projectDraft.localRootPath = 'C:\\ExportableDreamWorkerProject'
    await store.openActiveProjectDirectory()
    await store.exportActiveProjectManifest()

    expect(api.projects.updateProject).toHaveBeenCalledWith(
      expect.objectContaining({
        projectId: 'project_001',
        localRootPath: 'C:\\ExportableDreamWorkerProject'
      })
    )
    expect(api.projects.openLocalDirectory).toHaveBeenCalledWith('project_001')
    expect(api.projects.exportProjectManifest).toHaveBeenCalledWith('project_001')
  })

  it('deletes the active project through the typed API', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.deleteActiveProject()

    expect(api.projects.deleteProject).toHaveBeenCalledWith({ projectId: 'project_001' })
    expect(store.projects).toHaveLength(0)
    expect(store.projectModules).toHaveLength(0)
    expect(store.activeProjectId).toBe('')
  })

  it('navigates by primary, resource and module commands', () => {
    stubDreamWorkerApi(createDreamWorkerApiStub())
    const store = useAppShellStore()

    store.runCommand('resources')
    expect(store.activePrimary).toBe('resources')

    store.runCommand('mcp')
    expect(store.activePrimary).toBe('resources')
    expect(store.activeResourceTab).toBe('mcp')

    store.runCommand('explore')
    expect(store.activePrimary).toBe('explore')
  })

  it('applies context compaction and tool runtime stream events', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    store.chatStreamId = 'stream_test'
    store.chatStreamSessionId = 'chat_001'
    store.chatStreaming = true
    store.setChatMessagesForSession('chat_001', [
      {
        messageId: 'assistant_001',
        attemptId: 'attempt_001',
        sessionId: 'chat_001',
        role: 'assistant',
        content: '',
        status: 'streaming',
        providerId: 'provider_deepseek',
        model: 'deepseek-chat',
        usage: null,
        finishReason: '',
        runtimeSummary: '',
        trace_id: 'tr_stream',
        createdAt: '2026-07-01T00:00:00Z'
      }
    ])

    store.applyChatStreamEvent({
      type: 'context_compacted',
      streamId: 'stream_test',
      sessionId: 'chat_001',
      messageId: 'assistant_001',
      trace_id: 'tr_stream',
      sequence: 2,
      timestamp: '2026-07-01T00:00:00Z',
      contextBudget: {
        contextWindow: 768,
        maxOutputTokens: 128,
        inputBudgetTokens: 640,
        estimatedTokens: 512,
        systemTokens: 120,
        recentMessageTokens: 256,
        summaryTokens: 136,
        recentMessageCount: 4,
        compactedCount: 20,
        compacted: true,
        warnings: []
      },
      contextSummary: {
        summaryId: 'ctx_001',
        sessionId: 'chat_001',
        sourceMessageIds: ['msg_001'],
        content: 'summary',
        contentHash: 'hash',
        tokenEstimate: 12,
        createdBy: 'deterministic_extractive',
        contextVersion: 1,
        createdAt: '2026-07-01T00:00:00Z'
      }
    })

    store.applyChatStreamEvent({
      type: 'tool_result',
      streamId: 'stream_test',
      sessionId: 'chat_001',
      messageId: 'assistant_001',
      trace_id: 'tr_stream',
      sequence: 3,
      timestamp: '2026-07-01T00:00:00Z',
      toolCall: {
        callId: 'call_001',
        toolId: 'tool_model_generate_stub',
        displayName: 'Model Stub',
        riskLevel: 'low',
        approvalRequired: false,
        status: 'completed',
        summary: 'done',
        resultSummary: 'done'
      },
      toolResult: {
        callId: 'call_001',
        toolId: 'tool_model_generate_stub',
        status: 'completed',
        outputSummary: 'done',
        errorCode: '',
        errorMessage: '',
        latencyMs: 1
      }
    })

    expect(store.chatContextBudget.compacted).toBe(true)
    expect(store.chatContextSummary?.summaryId).toBe('ctx_001')
    expect(store.chatRuntimeToolState).toBe('completed')
    expect(store.chatToolCalls[0]?.status).toBe('completed')
  })

  it('refreshes active MCP tools into the resource tool list', async () => {
    const api = createDreamWorkerApiStub()
    stubDreamWorkerApi(api)
    const store = useAppShellStore()
    await store.loadWorkspace()

    await store.refreshActiveMcpTools()

    expect(api.mcp.refreshTools).toHaveBeenCalledWith('mcp_local_files')
    expect(store.tools[0]?.toolId).toBe('mcp_mcp_local_files_read_file')
  })
})
