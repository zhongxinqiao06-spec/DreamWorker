import { asRecord, asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, stream, type RuntimeRoute, type RuntimeStreamRoute } from '../route'

export function chatRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { chatService } = context
  return [
    get('/chat/sessions', () => chatService.listChatSessions()),
    post('/chat/sessions/create', (body) => chatService.createChatSession(body)),
    post('/chat/sessions/update', (body) => chatService.updateChatSession(body)),
    post('/chat/messages', (body) => chatService.listChatMessages(asString(body.sessionId))),
    post('/chat/messages/send', (body) => chatService.sendChatMessage(body)),
    post('/chat/messages/cancel', (body) => ({ ok: true, deletedId: asString(body.streamId) })),
    post('/chat/images/generate', (body) =>
      chatService.sendChatMessage({ ...body, content: asString(body.prompt) || 'generate image' })
    ),
    post('/chat/sessions/delete', (body) => chatService.deleteChatSession(asString(body.sessionId)))
  ]
}

export function chatStreamRoutes(context: RuntimeContext): RuntimeStreamRoute[] {
  return [
    stream('/chat/messages/stream', async (body, onEvent) => {
      for await (const event of context.chat.stream(asRecord(body))) {
        onEvent(event)
      }
    })
  ]
}
