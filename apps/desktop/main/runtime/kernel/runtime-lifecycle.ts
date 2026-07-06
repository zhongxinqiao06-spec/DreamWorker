export type RuntimeDisposable = {
  readonly stop: () => void
}

export class RuntimeLifecycle {
  private readonly disposables: RuntimeDisposable[] = []

  add(disposable: RuntimeDisposable): void {
    this.disposables.push(disposable)
  }

  stop(): void {
    for (const disposable of [...this.disposables].reverse()) {
      disposable.stop()
    }
    this.disposables.length = 0
  }
}
