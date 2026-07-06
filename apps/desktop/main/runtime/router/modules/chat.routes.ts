import { asRecord, asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, stream, type RuntimeRoute, type RuntimeStreamRoute } from '../route'

export function chatRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/chat/sessions', () => store.listChatSessions()),
    post('/chat/sessions/create', (body) => store.createChatSession(body)),
    post('/chat/sessions/update', (body) => store.updateChatSession(body)),
    post('/chat/messages', (body) => store.listChatMessages(asString(body.sessionId))),
    post('/chat/messages/send', (body) => store.sendChatMessage(body)),
    post('/chat/messages/cancel', (body) => ({ ok: true, deletedId: asString(body.streamId) })),
    post('/chat/images/generate', (body) =>
      store.sendChatMessage({ ...body, content: asString(body.prompt) || 'generate image' })
    ),
    post('/chat/sessions/delete', (body) => store.deleteChatSession(asString(body.sessionId)))
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
