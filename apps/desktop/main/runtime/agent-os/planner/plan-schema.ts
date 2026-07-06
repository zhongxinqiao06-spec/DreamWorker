import type { JsonRecord } from '../../types'

export type AgentPlanStep = {
  id: string
  type: 'model' | 'tool' | 'coding' | 'artifact' | 'human'
  title: string
  input: JsonRecord
  risk: 'low' | 'medium' | 'high'
}

export type AgentPlan = {
  goal: string
  steps: AgentPlanStep[]
}
