import { asString, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { SettingsService } from '../settings/settings-service'

export class ExtensionService {
  constructor(
    private readonly settings: SettingsService,
    private readonly configDir: string
  ) {}

  listExtensions(): JsonRecord[] {
    const settings = this.settings.getSettings()
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
    const settings = this.settings.getSettings()
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

  refreshModels(input: JsonRecord): JsonRecord {
    const extensionId = asString(input.extensionId) || '9router'
    return {
      ok: true,
      extensionId,
      models: [],
      status: this.extensionStatus(extensionId)
    }
  }

  verifyStreaming(input: JsonRecord): JsonRecord {
    const extensionId = asString(input.extensionId) || '9router'
    return {
      ok: true,
      extensionId,
      message: 'Main 内嵌 Runtime stream bridge 可用。',
      latencyMs: 0,
      status: this.extensionStatus(extensionId)
    }
  }

  tailLogs(): JsonRecord[] {
    return []
  }
}
