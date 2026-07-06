import { existsSync, mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { createDreamWorkerRuntime, type DreamWorkerRuntime } from './app'
import type {
  ChatStreamEvent,
  ExtensionSpec,
  ExtensionStatus,
  Project
} from '../../shared/dreamworker-api'

const runtimes: DreamWorkerRuntime[] = []

function createTestRuntime(): DreamWorkerRuntime {
  const configDir = mkdtempSync(join(tmpdir(), 'dreamworker-main-runtime-test-'))
  const runtime = createDreamWorkerRuntime(configDir)
  runtimes.push(runtime)
  return runtime
}

describe('main embedded runtime', () => {
  afterEach(() => {
    for (const runtime of runtimes.splice(0)) {
      runtime.stop()
    }
  })

  it('responds to runtime ping without spawning a daemon or using HTTP', () => {
    const runtime = createTestRuntime()

    expect(runtime.ping()).toEqual(
      expect.objectContaining({
        schema_version: '0.1',
        ok: true,
        engineVersion: '0.1.0-main-runtime',
        runtime: 'desktop-main-runtime'
      })
    )
  })

  it('dispatches workspace routes in memory and initializes workspace/code', async () => {
    const runtime = createTestRuntime()
    const localRootPath = mkdtempSync(join(tmpdir(), 'dreamworker-project-'))
    const project = await runtime.request<Project>('/projects/create', {
      method: 'POST',
      body: {
        title: '编码项目',
        description: '验证 Main Runtime 项目目录初始化。',
        localRootPath
      }
    })

    expect(project.localRootPath).toBe(localRootPath)
    expect(existsSync(join(localRootPath, 'workspace', 'code'))).toBe(true)
    await expect(
      runtime.request('/projects/local-directory/validate', {
        method: 'POST',
        body: { projectId: project.projectId }
      })
    ).resolves.toEqual(expect.objectContaining({ status: 'valid' }))
  })

  it('reports 9Router extension status with the renderer-facing id and dashboard url', async () => {
    const runtime = createTestRuntime()
    const extensions = await runtime.request<ExtensionSpec[]>('/extensions')
    const status = await runtime.request<ExtensionStatus>('/extensions/status', {
      method: 'POST',
      body: { extensionId: 'extension_9router' }
    })

    expect(extensions[0]).toEqual(
      expect.objectContaining({
        extensionId: 'extension_9router',
        health: expect.objectContaining({ dashboardURL: 'http://127.0.0.1:20128' })
      })
    )
    expect(status).toEqual(
      expect.objectContaining({
        extensionId: 'extension_9router',
        dashboardURL: 'http://127.0.0.1:20128'
      })
    )
  })

  it('streams chat events directly through the runtime callback path', async () => {
    const runtime = createTestRuntime()
    const events: ChatStreamEvent[] = []

    await runtime.stream(
      '/chat/messages/stream',
      {
        streamId: 'stream_test',
        body: { streamId: 'stream_test', sessionId: 'chat_001', content: '你好' }
      },
      (event) => {
        if (event.streamId === 'stream_test') {
          events.push(event as ChatStreamEvent)
        }
      }
    )

    expect(events.map((event) => event.type)).toEqual(['started', 'token_delta', 'completed'])
  })
})
