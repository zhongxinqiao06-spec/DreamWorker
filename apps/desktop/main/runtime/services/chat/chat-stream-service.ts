import type { ChatStreamEvent } from '../../../../shared/dreamworker-api'
import { asString, newTraceId, nowISO } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { ChatService } from './chat-service'

export class ChatStreamService {
  constructor(private readonly chat: ChatService) {}

  async *stream(body: JsonRecord): AsyncGenerator<ChatStreamEvent> {
    const streamId = asString(body.streamId) || `stream_${Date.now()}`
    const sessionId = asString(body.sessionId)
    const traceId = newTraceId()
    const started: ChatStreamEvent = {
      type: 'started',
      streamId,
      sessionId,
      messageId: '',
      trace_id: traceId,
      sequence: 1,
      timestamp: nowISO()
    }
    yield started
    yield {
      ...started,
      type: 'token_delta',
      sequence: 2,
      timestamp: nowISO(),
      delta: 'Main Runtime 已在 Main 进程内接管 Chat stream；模型网关接入后会输出真实模型增量。'
    }

    let result: unknown
    try {
      result = this.chat.sendChatMessage({
        ...body,
        content: asString(body.content) || asString(body.prompt)
      })
    } catch {
      result = undefined
    }
    const completed: ChatStreamEvent = {
      ...started,
      type: 'completed',
      sequence: 3,
      timestamp: nowISO()
    }
    if (result) {
      yield {
        ...completed,
        result: result as NonNullable<ChatStreamEvent['result']>
      }
      return
    }
    yield completed
  }
}
