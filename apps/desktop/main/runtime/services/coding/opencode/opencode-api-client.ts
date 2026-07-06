import { spawnSync } from 'node:child_process'
import { asRecord, asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'
import { resolveOpenCodeCommand } from './opencode-cli-resolver'

export class OpenCodeApiClient {
  apiJson(
    operation: string,
    args: string[],
    data: unknown,
    timeoutMs: number,
    cwd: string,
    env: NodeJS.ProcessEnv
  ): JsonRecord {
    const command = resolveOpenCodeCommand()
    if (!command) {
      throw new Error('OpenCode CLI was not found')
    }
    const commandArgs = [...command.argsPrefix, 'api', operation, ...args]
    if (data !== undefined) {
      commandArgs.push('--data', JSON.stringify(data))
    }
    const result = spawnSync(command.command, commandArgs, {
      cwd: cwd || process.cwd(),
      env,
      encoding: 'utf8',
      timeout: timeoutMs,
      windowsHide: true
    })
    if (result.status !== 0) {
      throw new Error(
        result.stderr.trim() || result.stdout.trim() || `OpenCode API ${operation} failed`
      )
    }
    const stdout = result.stdout.trim()
    if (!stdout) {
      return {}
    }
    let payload: JsonRecord
    try {
      const parsed = JSON.parse(stdout) as unknown
      payload = Array.isArray(parsed) ? { data: parsed } : asRecord(parsed)
    } catch (error) {
      throw new Error(`OpenCode API ${operation} returned invalid JSON: ${stdout.slice(0, 400)}`, {
        cause: error
      })
    }
    const tag = asString(payload._tag)
    if (tag.endsWith('Error') || tag === 'UnauthorizedError') {
      throw new Error(asString(payload.message) || `OpenCode API ${operation} failed with ${tag}`)
    }
    return payload
  }
}
