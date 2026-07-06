import type { JsonRecord } from '../../types'

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
