import { notFound } from '../../kernel/errors'
import { asString, newTraceId, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { ChatRepository } from '../../store/repositories/chat-repository'

export class ChatService {
  constructor(private readonly chats: ChatRepository) {}

  listChatSessions(): JsonRecord[] {
    return this.chats.listSessions()
  }

  createChatSession(input: JsonRecord): JsonRecord {
    const sessionId = this.chats.nextSessionId()
    const now = nowISO()
    const session = {
      sessionId,
      title: asString(input.title) || '新会话',
      projectId: input.projectId ?? null,
      agentId: asString(input.agentId) || 'agent_general_assistant',
      modelProfileId: asString(input.modelProfileId) || 'profile_fast',
      status: 'active',
      lastMessageAt: null,
      createdAt: now,
      updatedAt: now
    }
    this.chats.saveSession(sessionId, session)
    this.chats.saveMessages(sessionId, [])
    return session
  }

  updateChatSession(input: JsonRecord): JsonRecord {
    const sessionId = asString(input.sessionId)
    const session = { ...this.getChatSession(sessionId), ...input, updatedAt: nowISO() }
    this.chats.saveSession(sessionId, session)
    return session
  }

  getChatSession(sessionId: string): JsonRecord {
    const session = this.chats.getSession(sessionId)
    if (!session) {
      throw notFound('CHAT_SESSION_NOT_FOUND', 'chat session not found', 'refresh list')
    }
    return session
  }

  listChatMessages(sessionId: string): JsonRecord[] {
    this.getChatSession(sessionId)
    return this.chats.listMessages(sessionId)
  }

  sendChatMessage(input: JsonRecord): JsonRecord {
    const sessionId = asString(input.sessionId)
    const session = this.getChatSession(sessionId)
    const now = nowISO()
    const messages = this.chats.listMessages(sessionId)
    const userMessage = {
      messageId: this.chats.nextMessageId(),
      sessionId,
      role: 'user',
      content: asString(input.content) || asString(input.prompt),
      createdAt: now,
      updatedAt: now
    }
    const assistantMessage = {
      messageId: this.chats.nextMessageId(),
      sessionId,
      role: 'assistant',
      content: 'Main Runtime 已接管会话链路；真实模型流式调用将在模型网关迁移步骤接入。',
      createdAt: nowISO(),
      updatedAt: nowISO()
    }
    this.chats.saveMessages(sessionId, [...messages, userMessage, assistantMessage])
    this.chats.saveSession(sessionId, {
      ...session,
      updatedAt: nowISO(),
      lastMessageAt: nowISO()
    })
    return {
      sessionId,
      userMessage,
      assistantMessage,
      traceId: newTraceId(),
      usage: { promptTokens: 0, completionTokens: 0, totalTokens: 0 }
    }
  }

  deleteChatSession(sessionId: string): DeleteResult {
    if (!this.chats.getSession(sessionId)) {
      throw notFound('CHAT_SESSION_NOT_FOUND', 'chat session not found', 'refresh sessions')
    }
    this.chats.deleteSession(sessionId)
    return { ok: true, deletedId: sessionId }
  }
}
