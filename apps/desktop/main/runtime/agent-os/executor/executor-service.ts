import type { JsonRecord } from '../../types'
import type { AgentPlan } from '../planner/plan-schema'
import { ApprovalGate } from './approval-gate'
import { ToolRouter } from './tool-router'

export class ExecutorService {
  constructor(
    private readonly approvals = new ApprovalGate(),
    private readonly tools = new ToolRouter()
  ) {}

  executePlan(plan: AgentPlan): JsonRecord {
    const steps = plan.steps.map((step) => {
      const decision = this.approvals.decide(step)
      if (!decision.approved) {
        return { stepId: step.id, status: 'waiting_approval', reason: decision.reason }
      }
      if (step.type === 'tool') {
        return {
          stepId: step.id,
          status: 'routed',
          result: this.tools.route({ toolName: step.title, input: step.input })
        }
      }
      return { stepId: step.id, status: 'reserved', type: step.type }
    })
    return { goal: plan.goal, steps }
  }
}
