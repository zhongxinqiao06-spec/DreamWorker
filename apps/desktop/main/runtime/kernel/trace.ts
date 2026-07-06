import { newTraceId } from '../shared/util'

export function createTraceId(): string {
  return newTraceId()
}
