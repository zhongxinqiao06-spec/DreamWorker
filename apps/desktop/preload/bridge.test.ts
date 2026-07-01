import { describe, expect, it, vi } from 'vitest'
import { CHANNELS, type RuntimePingResponse } from '../shared/dreamworker-api'
import { createDreamWorkerApi } from './bridge'

describe('preload typed API contract', () => {
  it('exposes only the DreamWorker workspace namespaces', () => {
    const api = createDreamWorkerApi(vi.fn())

    expect(Object.keys(api)).toEqual([
      'runtime',
      'models',
      'agents',
      'skills',
      'tools',
      'mcp',
      'projects',
      'chat'
    ])
  })

  it('routes runtime.ping through the expected IPC channel', async () => {
    const response: RuntimePingResponse = {
      schema_version: '0.1',
      ok: false,
      trace_id: 'tr_test',
      error: {
        code: 'ENGINE_NOT_CONNECTED',
        message: 'Go Engine 尚未连接，后续阶段会接入本地引擎。',
        recoverable: true,
        user_action: '等待引擎接入后重试。',
        trace_id: 'tr_test'
      }
    }
    const invoke = vi.fn().mockResolvedValue(response)
    const api = createDreamWorkerApi(invoke)

    await expect(api.runtime.ping()).resolves.toEqual(response)
    expect(invoke).toHaveBeenCalledWith(CHANNELS.runtimePing)
  })

  it('routes resource calls through explicit IPC channels without raw IPC exposure', async () => {
    const invoke = vi.fn().mockResolvedValue({ ok: true, deletedId: 'provider_custom' })
    const api = createDreamWorkerApi(invoke)

    await api.models.deleteProvider('provider_custom')

    expect(invoke).toHaveBeenCalledWith(CHANNELS.modelsDeleteProvider, {
      providerId: 'provider_custom'
    })
    expect('ipcRenderer' in api).toBe(false)
  })

  it('keeps chat and project calls behind the typed preload API', async () => {
    const invoke = vi.fn().mockResolvedValue([])
    const api = createDreamWorkerApi(invoke)

    await api.projects.listProjectModules('project_001')
    await api.projects.deleteProject({ projectId: 'project_001' })
    await api.chat.updateSession({
      sessionId: 'chat_001',
      projectId: 'project_001',
      title: '普通 Agent 工作台',
      agentId: 'agent_general_assistant',
      modelProfileId: 'profile_fast'
    })
    await api.chat.sendMessage({ sessionId: 'chat_001', content: '你好' })

    expect(invoke).toHaveBeenCalledWith(CHANNELS.projectsListModules, {
      projectId: 'project_001'
    })
    expect(invoke).toHaveBeenCalledWith(CHANNELS.projectsDelete, {
      projectId: 'project_001'
    })
    expect(invoke).toHaveBeenCalledWith(CHANNELS.chatUpdateSession, {
      sessionId: 'chat_001',
      projectId: 'project_001',
      title: '普通 Agent 工作台',
      agentId: 'agent_general_assistant',
      modelProfileId: 'profile_fast'
    })
    expect(invoke).toHaveBeenCalledWith(CHANNELS.chatSendMessage, {
      sessionId: 'chat_001',
      content: '你好'
    })
  })

  it('cleans up stream listeners and forwards cancel through typed IPC', async () => {
    const unsubscribe = vi.fn()
    const invoke = vi.fn().mockImplementation(async (channel: string, payload: unknown) => {
      if (channel === CHANNELS.chatStartStream) {
        return { streamId: (payload as { streamId: string }).streamId }
      }
      if (channel === CHANNELS.chatCancelStream) {
        return { ok: true, deletedId: (payload as { streamId: string }).streamId }
      }
      return {}
    })
    const listen = vi.fn().mockReturnValue(unsubscribe)
    const api = createDreamWorkerApi(invoke, listen)

    const controller = await api.chat.streamMessage(
      { sessionId: 'chat_001', content: 'hello', streamId: 'stream_test' },
      vi.fn()
    )
    await controller.cancel()

    expect(unsubscribe).toHaveBeenCalledTimes(1)
    expect(invoke).toHaveBeenCalledWith(CHANNELS.chatCancelStream, { streamId: 'stream_test' })
  })
})
