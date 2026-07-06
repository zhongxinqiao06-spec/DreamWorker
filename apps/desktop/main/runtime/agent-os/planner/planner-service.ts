import type { AgentPlan } from './plan-schema'

export class PlannerService {
  createSingleStepPlan(goal: string): AgentPlan {
    return {
      goal,
      steps: [
        {
          id: 'step_001',
          type: 'model',
          title: goal,
          input: { goal },
          risk: 'low'
        }
      ]
    }
  }
}
