import { asRecord, asString, redactSecrets } from '../../../shared/util'
import type { JsonRecord } from '../../../types'

export function normalizeOpenCodeMessage(value: unknown): JsonRecord {
  const record = asRecord(value)
  const info = asRecord(record.info)
  if (Object.keys(info).length > 0) {
    return { ...info, parts: Array.isArray(record.parts) ? record.parts : [] }
  }
  return record
}

export function extractOpenCodeMessageText(message: JsonRecord): string {
  if (typeof message.text === 'string') {
    return message.text
  }
  const parts = Array.isArray(message.parts) ? message.parts : []
  for (const part of parts) {
    const record = asRecord(part)
    if (typeof record.text === 'string') {
      return record.text
    }
  }
  const content = Array.isArray(message.content) ? message.content : []
  const chunks: string[] = []
  for (const part of content) {
    const record = asRecord(part)
    if (typeof record.text === 'string') {
      chunks.push(record.text)
    }
    if (typeof record.content === 'string') {
      chunks.push(record.content)
    }
    if (typeof record.output === 'string') {
      chunks.push(record.output)
    }
  }
  return chunks.join('')
}

export function openCodeMessageError(message: JsonRecord): string {
  const error = asRecord(message.error)
  const data = asRecord(error.data)
  const text = asString(error.message) || asString(data.message) || asString(error.name)
  return text ? redactSecrets(text) : ''
}
