import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { platform } from 'node:process'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  ENGINE_READY_TIMEOUT_MS,
  cancelEngineDaemonStream,
  parseEngineReadyLine,
  pingEngineDaemon,
  resolveEngineLaunchCommand,
  startEngineDaemonStream
} from './engine-daemon'

describe('engine daemon bridge', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('parses the engine ready line', () => {
    const ready = parseEngineReadyLine(
      JSON.stringify({
        ok: true,
        event: 'engine.ready',
        baseUrl: 'http://127.0.0.1:12345',
        engineVersion: '0.1.0',
        trace_id: 'tr_ready'
      })
    )

    expect(ready).toEqual({
      ok: true,
      event: 'engine.ready',
      baseUrl: 'http://127.0.0.1:12345',
      engineVersion: '0.1.0',
      trace_id: 'tr_ready'
    })
  })

  it('returns null for non-ready output', () => {
    expect(parseEngineReadyLine('warming up')).toBeNull()
  })

  it('calls the local engine ping endpoint with a bearer token', async () => {
    const fetchLike = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        schema_version: '0.1',
        ok: true,
        engineVersion: '0.1.0',
        trace_id: 'tr_engine'
      })
    })

    await expect(
      pingEngineDaemon('http://127.0.0.1:54321', 'secret-token', fetchLike)
    ).resolves.toEqual({
      schema_version: '0.1',
      ok: true,
      engineVersion: '0.1.0',
      trace_id: 'tr_engine'
    })
    expect(fetchLike).toHaveBeenCalledWith('http://127.0.0.1:54321/runtime/ping', {
      headers: {
        Authorization: 'Bearer secret-token'
      }
    })
  })

  it('uses go run as the development launch command', () => {
    const rootDir = mkdtempSync(join(tmpdir(), 'dreamworker-engine-test-'))
    const engineDir = join(rootDir, 'engine')
    mkdirSync(engineDir, { recursive: true })
    writeFileSync(join(engineDir, 'go.mod'), 'module test/engine\n')

    const launch = resolveEngineLaunchCommand('secret-token', rootDir)

    expect(launch.command).toBe('go')
    expect(launch.args).toEqual([
      'run',
      './cmd/dreamworker-engine',
      'serve',
      '--token',
      'secret-token'
    ])
    expect(launch.cwd).toContain('engine')
  })

  it('uses go run during electron-vite dev even when an engine binary exists', () => {
    const rootDir = mkdtempSync(join(tmpdir(), 'dreamworker-engine-dev-test-'))
    const engineDir = join(rootDir, 'engine')
    const engineBinDir = join(engineDir, 'bin')
    const executableName = platform === 'win32' ? 'dreamworker-engine.exe' : 'dreamworker-engine'
    mkdirSync(engineBinDir, { recursive: true })
    writeFileSync(join(engineDir, 'go.mod'), 'module test/engine\n')
    writeFileSync(join(engineBinDir, executableName), '')
    vi.stubEnv('ELECTRON_RENDERER_URL', 'http://localhost:5173')

    const launch = resolveEngineLaunchCommand('secret-token', rootDir)

    expect(launch.command).toBe('go')
    expect(launch.args).toEqual([
      'run',
      './cmd/dreamworker-engine',
      'serve',
      '--token',
      'secret-token'
    ])
    expect(launch.cwd).toBe(engineDir)
  })

  it('uses packaged engine binary and .agent resources when present', () => {
    const rootDir = mkdtempSync(join(tmpdir(), 'dreamworker-packaged-test-'))
    const engineBinDir = join(rootDir, 'engine', 'bin')
    const executableName = platform === 'win32' ? 'dreamworker-engine.exe' : 'dreamworker-engine'
    mkdirSync(engineBinDir, { recursive: true })
    mkdirSync(join(rootDir, '.agent'), { recursive: true })
    writeFileSync(join(engineBinDir, executableName), '')

    const launch = resolveEngineLaunchCommand('secret-token', rootDir)

    expect(launch.command).toContain(executableName)
    expect(launch.args).toEqual(['serve', '--token', 'secret-token'])
    expect(launch.env?.DREAMWORKER_AGENT_DIR).toBe(join(rootDir, '.agent'))
  })

  it('injects .env.local variables into the engine process', () => {
    const rootDir = mkdtempSync(join(tmpdir(), 'dreamworker-env-test-'))
    mkdirSync(join(rootDir, 'engine'), { recursive: true })
    writeFileSync(join(rootDir, 'engine', 'go.mod'), 'module test/engine\n')
    writeFileSync(
      join(rootDir, '.env.local'),
      'DEEPSEEK_API_KEY=sk-test\nSILICONFLOW_MODEL="deepseek-ai/DeepSeek-V3"\n'
    )

    const launch = resolveEngineLaunchCommand('secret-token', rootDir)

    expect(launch.env?.DEEPSEEK_API_KEY).toBe('sk-test')
    expect(launch.env?.SILICONFLOW_MODEL).toBe('deepseek-ai/DeepSeek-V3')
  })

  it('allows cold go run compilation before reporting engine not connected', () => {
    expect(ENGINE_READY_TIMEOUT_MS).toBe(60000)
  })

  it('can abort an active local SSE request by stream id', async () => {
    const closed = new Promise<void>((resolve) => {
      const server = createServer((_, response) => {
        response.writeHead(200, { 'Content-Type': 'text/event-stream' })
        response.write(
          'data: {"type":"started","streamId":"stream_abort","sessionId":"chat","messageId":"msg","trace_id":"tr","sequence":1,"timestamp":"2026-07-01T00:00:00Z"}\n\n'
        )
        response.on('close', () => {
          server.close(() => resolve())
        })
      })
      server.listen(0, '127.0.0.1', () => {
        const address = server.address()
        if (typeof address === 'object' && address) {
          startEngineDaemonStream(
            `http://127.0.0.1:${address.port}`,
            'token',
            '/chat/messages/stream',
            { streamId: 'stream_abort' },
            'stream_abort',
            () => cancelEngineDaemonStream('stream_abort')
          )
        }
      })
    })

    await expect(closed).resolves.toBeUndefined()
  }, 10000)
})
