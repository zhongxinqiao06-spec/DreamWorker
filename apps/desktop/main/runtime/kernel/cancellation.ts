import { notFound, badRequest, type DeleteResult } from '../types'
import { asString } from '../shared/util'

export class CancellationRegistry {
  private readonly controllers = new Map<string, AbortController>()

  start(streamId: string): AbortController {
    const controller = new AbortController()
    this.controllers.set(streamId, controller)
    return controller
  }

  cancel(input: { readonly streamId?: unknown }): DeleteResult {
    const streamId = asString(input.streamId)
    if (!streamId) {
      throw badRequest('BAD_REQUEST', 'missing streamId', 'select an active stream')
    }
    const controller = this.controllers.get(streamId)
    if (!controller) {
      throw notFound('STREAM_NOT_FOUND', 'stream not found', 'the stream may already be finished')
    }
    controller.abort()
    this.controllers.delete(streamId)
    return { ok: true, deletedId: streamId }
  }

  complete(streamId: string): void {
    this.controllers.delete(streamId)
  }

  abortAll(): void {
    for (const controller of this.controllers.values()) {
      controller.abort()
    }
    this.controllers.clear()
  }
}
