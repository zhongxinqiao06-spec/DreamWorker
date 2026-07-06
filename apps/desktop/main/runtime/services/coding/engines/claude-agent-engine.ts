import type { CodingStreamEvent } from '../../../types'
import type { CodingEngine, CodingEngineTurn } from './coding-engine'

export class ClaudeAgentEngine implements CodingEngine {
  readonly engineId = 'claude_agent'

  async *streamTurn(turn: CodingEngineTurn): AsyncGenerator<CodingStreamEvent> {
    yield turn.event({
      type: 'delta',
      delta: 'Claude Agent SDK is now hosted inside the Main Runtime. '
    })
    yield turn.event({
      type: 'completed',
      message: 'Main Runtime coding service completed the turn handoff.'
    })
  }
}
