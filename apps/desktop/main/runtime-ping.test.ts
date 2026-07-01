import { describe, expect, it } from 'vitest'
import { createRuntimePingStubResponse } from './runtime-ping'

describe('runtime ping main stub', () => {
  it('returns a structured Chinese engine-not-connected response', () => {
    const response = createRuntimePingStubResponse('tr_contract')

    expect(response).toEqual({
      schema_version: '0.1',
      ok: false,
      trace_id: 'tr_contract',
      error: {
        code: 'ENGINE_NOT_CONNECTED',
        message: 'Go Engine 尚未连接，后续阶段会接入本地引擎。',
        recoverable: true,
        user_action: '等待引擎接入后重试。',
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
