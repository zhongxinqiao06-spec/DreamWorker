import { describe, expect, it } from 'vitest'
import { createRuntimePingStubResponse } from './runtime-ping'

describe('runtime ping main stub', () => {
  it('returns a structured engine-not-connected response', () => {
    const response = createRuntimePingStubResponse('tr_contract')

    expect(response).toEqual({
      schema_version: '0.1',
      ok: false,
      trace_id: 'tr_contract',
      error: {
        code: 'ENGINE_NOT_CONNECTED',
        message: 'Main Runtime 尚未连接。',
        recoverable: true,
        user_action: '等待本地 Runtime 启动后重试。',
        trace_id: 'tr_contract'
      }
    })
  })

  it('generates a trace_id when none is provided', () => {
    const response = createRuntimePingStubResponse()

    expect(response.trace_id).toMatch(/^tr_/)
    expect(response.ok).toBe(false)
    if (!response.ok) {
      expect(response.error.trace_id).toBe(response.trace_id)
    }
  })
})
