import type { CodingEngineId, CodingSession, CodingStreamEvent, JsonRecord } from '../../../types'

export const engineDescriptors = [
  {
    engineId: 'claude_agent',
    displayName: 'Claude Agent',
    description: 'Anthropic Claude Agent SDK, cwd scoped to project workspace/code.',
    supportedProviderTypes: ['anthropic'],
    preferredProviderIds: ['provider_anthropic'],
    directWrite: true,
    streaming: true
  },
  {
    engineId: 'codex',
    displayName: 'Codex',
    description: 'OpenAI Codex SDK thread run with workspace-write sandbox.',
    supportedProviderTypes: [
      'openai',
      'openai_compatible',
      'deepseek',
      'siliconflow',
      'glm',
      'custom'
    ],
    preferredProviderIds: ['provider_9router_local', 'provider_openai'],
    directWrite: true,
    streaming: false
  },
  {
    engineId: 'opencode',
    displayName: 'OpenCode',
    description: 'OpenCode SDK/CLI managed by the Main Runtime process.',
    supportedProviderTypes: [
      'openai',
      'openai_compatible',
      'deepseek',
      'siliconflow',
      'glm',
      'ollama',
      'custom'
    ],
    preferredProviderIds: ['provider_9router_local'],
    directWrite: true,
    streaming: true
  }
] as const

export type CodingEventInput = Omit<
  CodingStreamEvent,
  | 'streamId'
  | 'sessionId'
  | 'engineId'
  | 'providerId'
  | 'model'
  | 'trace_id'
  | 'sequence'
  | 'timestamp'
>

export type CodingEventFactory = (eventInput: CodingEventInput) => CodingStreamEvent

export type CodingEngineTurn = {
  readonly prompt: string
  readonly session: CodingSession
  readonly provider: JsonRecord
  readonly signal: AbortSignal
  readonly event: CodingEventFactory
  readonly updateSession: (session: CodingSession) => void
}

export interface CodingEngine {
  readonly engineId: CodingEngineId
  streamTurn(turn: CodingEngineTurn): AsyncGenerator<CodingStreamEvent>
  dispose?(): void
}

export function normalizeEngine(value: string): CodingEngineId {
  if (value === 'codex' || value === 'opencode' || value === 'claude_agent') {
    return value
  }
  return 'claude_agent'
}
