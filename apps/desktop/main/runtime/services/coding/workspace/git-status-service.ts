import { spawnSync } from 'node:child_process'
import { asString } from '../../../shared/util'
import type { JsonRecord } from '../../../types'

export class GitStatusService {
  changes(root: string): JsonRecord[] {
    const result = spawnSync('git', ['status', '--short', '--porcelain=v1'], {
      cwd: root,
      encoding: 'utf8',
      windowsHide: true
    })
    if (result.status !== 0 || !result.stdout.trim()) {
      return []
    }
    return result.stdout
      .trim()
      .split(/\r?\n/)
      .filter(Boolean)
      .map((line) => {
        const status = line.slice(0, 2).trim()
        const path = line.slice(3).trim().split(' -> ').pop() ?? ''
        return { path: path.replaceAll('\\', '/'), status }
      })
  }

  statusMap(root: string): Map<string, string> {
    return new Map(
      this.changes(root).map((change) => [asString(change.path), asString(change.status)])
    )
  }

  branch(root: string): string {
    const result = spawnSync('git', ['branch', '--show-current'], {
      cwd: root,
      encoding: 'utf8',
      windowsHide: true
    })
    return result.status === 0 ? result.stdout.trim() : ''
  }
}
