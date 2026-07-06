import type { AgentPlanStep } from '../planner/plan-schema'

export type ApprovalDecision = {
  approved: boolean
  reason: string
}

export class ApprovalGate {
  decide(step: AgentPlanStep): ApprovalDecision {
    if (step.risk === 'high') {
      return { approved: false, reason: 'high risk steps require explicit approval' }
    }
    return { approved: true, reason: 'step is within automatic execution policy' }
  }
}
