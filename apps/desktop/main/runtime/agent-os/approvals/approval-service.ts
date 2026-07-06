import type { JsonRecord } from '../../types'

export type ApprovalRequest = {
  approvalId: string
  runId: string
  reason: string
  payload: JsonRecord
}

export class ApprovalService {
  private readonly requests = new Map<string, ApprovalRequest>()

  request(input: ApprovalRequest): ApprovalRequest {
    this.requests.set(input.approvalId, input)
    return input
  }

  list(): ApprovalRequest[] {
    return [...this.requests.values()]
  }
}
