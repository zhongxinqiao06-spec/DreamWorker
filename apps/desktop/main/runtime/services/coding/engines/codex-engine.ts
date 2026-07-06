import type { CodingStreamEvent } from '../../../types'
import type { CodingEngine, CodingEngineTurn } from './coding-engine'

export class CodexEngine implements CodingEngine {
  readonly engineId = 'codex'

  async *streamTurn(turn: CodingEngineTurn): AsyncGenerator<CodingStreamEvent> {
    yield turn.event({
      type: 'delta',
      delta: 'Codex SDK is now hosted inside the Main Runtime. '
    })
    yield turn.event({
      type: 'completed',
      message: 'Main Runtime coding service completed the turn handoff.'
    })
  }
}
