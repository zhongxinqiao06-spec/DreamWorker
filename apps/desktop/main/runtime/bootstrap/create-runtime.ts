import { CodingService } from '../coding/coding-service'
import { RuntimeLifecycle } from '../kernel/runtime-lifecycle'
import { WorkspaceStore } from '../store/workspace-store'
import type { RuntimeContext } from './runtime-context'

export function createRuntimeContext(configDir?: string): RuntimeContext {
  const store = new WorkspaceStore(configDir)
  const coding = new CodingService(store)
  const lifecycle = new RuntimeLifecycle()

  lifecycle.add({ stop: () => store.close() })
  lifecycle.add({ stop: () => coding.dispose() })

  return {
    store,
    coding,
    lifecycle
  }
}
