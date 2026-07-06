export type RuntimeEventHandler<TEvent> = (event: TEvent) => void | Promise<void>

export class RuntimeEventBus<TEvent extends { readonly type: string }> {
  private readonly handlers = new Map<string, Set<RuntimeEventHandler<TEvent>>>()

  on(type: string, handler: RuntimeEventHandler<TEvent>): () => void {
    const handlers = this.handlers.get(type) ?? new Set<RuntimeEventHandler<TEvent>>()
    handlers.add(handler)
    this.handlers.set(type, handlers)
    return () => handlers.delete(handler)
  }

  async emit(event: TEvent): Promise<void> {
    const handlers = this.handlers.get(event.type) ?? new Set<RuntimeEventHandler<TEvent>>()
    for (const handler of handlers) {
      await handler(event)
    }
  }
}
