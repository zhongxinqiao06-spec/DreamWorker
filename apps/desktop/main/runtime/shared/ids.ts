import { randomBytes } from 'node:crypto'

export function newTraceId(): string {
  return `tr_${new Date().toISOString().replace(/[-:.]/g, '').slice(0, 17)}_${randomBytes(4).toString('hex')}`
}
