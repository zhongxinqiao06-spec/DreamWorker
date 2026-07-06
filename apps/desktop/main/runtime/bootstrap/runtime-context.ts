import type { CodingService } from '../coding/coding-service'
import type { RuntimeLifecycle } from '../kernel/runtime-lifecycle'
import type { WorkspaceStore } from '../store/workspace-store'

export type RuntimeContext = {
  readonly store: WorkspaceStore
  readonly coding: CodingService
  readonly lifecycle: RuntimeLifecycle
}
