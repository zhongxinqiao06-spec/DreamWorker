import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class ChatRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextSessionId(): string {
    return this.store.nextId('chat')
  }

  nextMessageId(): string {
    return this.store.nextId('msg')
  }

  listSessions(): JsonRecord[] {
    return sortedValues(this.store.snapshot.sessions, 'updatedAt').reverse()
  }

  getSession(sessionId: string): JsonRecord | undefined {
    return this.store.snapshot.sessions[sessionId]
  }

  saveSession(sessionId: string, session: JsonRecord): void {
    this.store.snapshot.sessions[sessionId] = session
    this.store.save()
  }

  deleteSession(sessionId: string): void {
    delete this.store.snapshot.sessions[sessionId]
    delete this.store.snapshot.messages[sessionId]
    this.store.save()
  }

  listMessages(sessionId: string): JsonRecord[] {
    return this.store.snapshot.messages[sessionId] ?? []
  }

  saveMessages(sessionId: string, messages: JsonRecord[]): void {
    this.store.snapshot.messages[sessionId] = messages
    this.store.save()
  }
}
