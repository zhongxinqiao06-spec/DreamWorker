import {
  type RuntimePingResponse,
  createEngineNotConnectedResponse
} from '../shared/dreamworker-api'

function createTraceId(): string {
  const timestamp = Date.now().toString(36)
  const suffix = Math.random().toString(36).slice(2, 10)
  return `tr_${timestamp}_${suffix}`
}

export function createRuntimePingStubResponse(traceId = createTraceId()): RuntimePingResponse {
  return createEngineNotConnectedResponse(traceId)
}
