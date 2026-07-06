import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { notFound } from '../../../kernel/errors'
import type { CodingEngineId, JsonRecord } from '../../../types'
import { resolveOpenCodeCommand, runtimeRoot } from '../opencode/opencode-cli-resolver'
import { ClaudeAgentEngine } from './claude-agent-engine'
import type { CodingEngine } from './coding-engine'
import { engineDescriptors } from './coding-engine'
import { CodexEngine } from './codex-engine'
import { OpenCodeEngine } from './opencode-engine'

export class CodingEngineRegistry {
  private readonly engines = new Map<CodingEngineId, CodingEngine>([
    ['claude_agent', new ClaudeAgentEngine()],
    ['codex', new CodexEngine()],
    ['opencode', new OpenCodeEngine()]
  ])

  async list(): Promise<JsonRecord> {
    const engineStatuses = await Promise.all([
      runtimeStatusFor('claude_agent', '@anthropic-ai/claude-agent-sdk', 'claude'),
      runtimeStatusFor('codex', '@openai/codex-sdk', 'codex'),
      runtimeStatusFor('opencode', '@opencode-ai/sdk', 'opencode')
    ])
    const available = engineStatuses.some((status) => status.installed === true)
    return {
      runtimeDir: runtimeRoot(),
      nodeBin: process.execPath,
      adapterPath: fileURLToPath(import.meta.url),
      available,
      message: available
        ? 'Node coding runtime is ready'
        : 'Node coding runtime SDK packages are missing',
      engines: engineDescriptors,
      engineStatuses
    }
  }

  get(engineId: CodingEngineId): CodingEngine {
    const engine = this.engines.get(engineId)
    if (!engine) {
      throw notFound('CODING_ENGINE_NOT_FOUND', 'coding engine not found', 'select another engine')
    }
    return engine
  }

  dispose(): void {
    for (const engine of this.engines.values()) {
      engine.dispose?.()
    }
  }
}

async function runtimeStatusFor(
  engineId: CodingEngineId,
  packageName: string,
  key: string
): Promise<JsonRecord> {
  try {
    await import(packageName)
    if (engineId === 'opencode') {
      const command = resolveOpenCodeCommand()
      if (!command) {
        return {
          engineId,
          packageName,
          installed: true,
          executable: false,
          status: 'error',
          message: 'OpenCode SDK is installed but CLI binary was not found',
          key
        }
      }
      const version = spawnSync(command.command, [...command.argsPrefix, '--version'], {
        encoding: 'utf8',
        timeout: 5000,
        windowsHide: true
      })
      if (version.status !== 0) {
        return {
          engineId,
          packageName,
          installed: true,
          executable: false,
          status: 'error',
          message:
            version.stderr.trim() || version.stdout.trim() || 'OpenCode CLI is not executable',
          key
        }
      }
    }
    return {
      engineId,
      packageName,
      installed: true,
      executable: true,
      status: 'ready',
      message: `${packageName} is installed in Main Runtime`,
      key
    }
  } catch (error) {
    return {
      engineId,
      packageName,
      installed: false,
      executable: false,
      status: 'missing',
      message: error instanceof Error ? error.message : `${packageName} is missing`,
      key
    }
  }
}
