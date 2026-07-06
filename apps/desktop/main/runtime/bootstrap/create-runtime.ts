import { CodingService } from '../coding/coding-service'
import { RuntimeLifecycle } from '../kernel/runtime-lifecycle'
import { ChatStreamService } from '../services/chat/chat-stream-service'
import { WorkspaceStore } from '../store/workspace-store'
import type { RuntimeContext } from './runtime-context'

export function createRuntimeContext(configDir?: string): RuntimeContext {
  const store = new WorkspaceStore(configDir)
  const coding = new CodingService(store)
  const chat = new ChatStreamService(store)
  const lifecycle = new RuntimeLifecycle()

  lifecycle.add({ stop: () => store.close() })
  lifecycle.add({ stop: () => coding.dispose() })

  return {
    store,
    coding,
    chat,
    lifecycle
  }
}
