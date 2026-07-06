import { createServer, type IncomingMessage, type ServerResponse } from 'node:http'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createDreamWorkerRuntime, type DreamWorkerRuntime } from './app'
import type { CodingStreamEvent, Project } from '../../shared/dreamworker-api'

const runOpenCodeSmoke = process.env.DREAMWORKER_OPENCODE_SMOKE === '1'
const runtimes: DreamWorkerRuntime[] = []

describe.skipIf(!runOpenCodeSmoke)('opencode main runtime smoke', () => {
  let server: Awaited<ReturnType<typeof startFakeOpenAiCompatibleServer>> | null = null

  beforeEach(async () => {
    vi.stubEnv('DREAMWORKER_OPENCODE_IDLE_TIMEOUT_MS', '30000')
    vi.stubEnv('DREAMWORKER_OPENCODE_TURN_TIMEOUT_MS', '90000')
    server = await startFakeOpenAiCompatibleServer()
  })

  afterEach(async () => {
    for (const runtime of runtimes.splice(0)) {
      runtime.stop()
    }
    await server?.close()
    server = null
    vi.unstubAllEnvs()
  })

  it('connects OpenCode server, session, prompt and event polling through Main Runtime', async () => {
    const runtime = createTestRuntime()
    const localRootPath = mkdtempSync(join(tmpdir(), 'dreamworker-opencode-project-'))
    const project = await runtime.request<Project>('/projects/create', {
      method: 'POST',
      body: {
        title: 'OpenCode Smoke',
        description: 'Main Runtime OpenCode 链路验证。',
        localRootPath
      }
    })
    mkdirSync(join(localRootPath, 'workspace', 'code'), { recursive: true })
    writeFileSync(join(localRootPath, 'workspace', 'code', 'README.md'), '# Smoke\n')

    await runtime.request('/models/providers/save', {
      method: 'POST',
      body: {
        providerId: 'provider_smoke_opencode',
        providerType: 'openai_compatible',
        displayName: 'Smoke OpenAI Compatible',
        baseURL: server?.baseURL,
        defaultModel: 'smoke-model',
        availableModels: ['smoke-model'],
        enabled: true,
        capabilities: ['chat', 'tools'],
        apiKey: 'sk-smoke'
      }
    })

    const events: CodingStreamEvent[] = []
    await runtime.stream(
      '/coding/turns/stream',
      {
        streamId: 'coding_smoke',
        body: {
          streamId: 'coding_smoke',
          projectId: project.projectId,
          engineId: 'opencode',
          providerId: 'provider_smoke_opencode',
          model: 'smoke-model',
          prompt: 'Reply with exactly SMOKE_OK.'
        }
      },
      (event) => events.push(event as CodingStreamEvent)
    )

    expect(server?.llmHits()).toBeGreaterThan(0)
    expect(events.map((event) => event.type)).toContain('completed')
    expect(events.map((event) => event.type)).not.toContain('error')
    expect(
      events
        .filter((event) => event.type === 'delta')
        .map((event) => event.delta)
        .join('')
    ).toContain('SMOKE_OK')
    expect(
      events.filter((event) => event.type === 'tool_call').map((event) => event.toolCall?.toolName)
    ).toEqual(
      expect.arrayContaining([
        'workspace.code_root',
        'opencode.server',
        'opencode.session.create',
        'opencode.session.prompt',
        'opencode.event.poll'
      ])
    )
  }, 120000)
})

function createTestRuntime(): DreamWorkerRuntime {
  const configDir = mkdtempSync(join(tmpdir(), 'dreamworker-opencode-runtime-'))
  const runtime = createDreamWorkerRuntime(configDir)
  runtimes.push(runtime)
  return runtime
}

async function startFakeOpenAiCompatibleServer(): Promise<{
  readonly baseURL: string
  readonly llmHits: () => number
  readonly close: () => Promise<void>
}> {
  let hits = 0
  const server = createServer(async (request, response) => {
    const url = new URL(request.url || '/', 'http://127.0.0.1')
    if (request.method === 'GET' && url.pathname === '/v1/models') {
      writeJson(response, 200, {
        object: 'list',
        data: [{ id: 'smoke-model', object: 'model', created: 0, owned_by: 'dreamworker' }]
      })
      return
    }
    if (request.method === 'POST' && url.pathname === '/v1/chat/completions') {
      hits += 1
      const body = await readJson(request)
      if (body.stream === true) {
        writeChatStream(response)
        return
      }
      writeJson(response, 200, chatCompletionPayload())
      return
    }
    if (request.method === 'POST' && url.pathname === '/v1/responses') {
      hits += 1
      writeJson(response, 200, {
        id: 'resp_smoke',
        object: 'response',
        status: 'completed',
        output_text: 'SMOKE_OK',
        output: [
          {
            type: 'message',
            role: 'assistant',
            content: [{ type: 'output_text', text: 'SMOKE_OK' }]
          }
        ]
      })
      return
    }
    writeJson(response, 404, { error: { message: `not found: ${url.pathname}` } })
  })

  await new Promise<void>((resolve, reject) => {
    server.once('error', reject)
    server.listen(0, '127.0.0.1', () => {
      server.off('error', reject)
      resolve()
    })
  })

  const address = server.address()
  if (!address || typeof address === 'string') {
    throw new Error('fake OpenAI-compatible server did not bind to a TCP port')
  }
  return {
    baseURL: `http://127.0.0.1:${address.port}/v1`,
    llmHits: () => hits,
    close: () =>
      new Promise<void>((resolveClose) => {
        server.close(() => resolveClose())
      })
  }
}

function chatCompletionPayload() {
  return {
    id: 'chatcmpl_smoke',
    object: 'chat.completion',
    created: Math.floor(Date.now() / 1000),
    model: 'smoke-model',
    choices: [
      {
        index: 0,
        message: { role: 'assistant', content: 'SMOKE_OK' },
        finish_reason: 'stop'
      }
    ],
    usage: { prompt_tokens: 1, completion_tokens: 1, total_tokens: 2 }
  }
}

function writeChatStream(response: ServerResponse): void {
  response.writeHead(200, {
    'Content-Type': 'text/event-stream; charset=utf-8',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive'
  })
  response.write(`data: ${JSON.stringify(chatDeltaPayload('SMOKE_OK', null))}\n\n`)
  response.write(`data: ${JSON.stringify(chatDeltaPayload('', 'stop'))}\n\n`)
  response.write('data: [DONE]\n\n')
  response.end()
}

function chatDeltaPayload(content: string, finishReason: string | null) {
  return {
    id: 'chatcmpl_smoke',
    object: 'chat.completion.chunk',
    created: Math.floor(Date.now() / 1000),
    model: 'smoke-model',
    choices: [
      {
        index: 0,
        delta: content ? { role: 'assistant', content } : {},
        finish_reason: finishReason
      }
    ]
  }
}

function writeJson(response: ServerResponse, status: number, payload: unknown): void {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' })
  response.end(JSON.stringify(payload))
}

async function readJson(request: IncomingMessage): Promise<Record<string, unknown>> {
  let raw = ''
  for await (const chunk of request) {
    raw += String(chunk)
  }
  return raw.trim() ? (JSON.parse(raw) as Record<string, unknown>) : {}
}
