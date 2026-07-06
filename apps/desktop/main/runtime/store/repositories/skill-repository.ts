import { sortedValues } from '../../shared/util'
import type { JsonRecord } from '../../types'
import type { WorkspaceStore } from '../workspace-store'

export class SkillRepository {
  constructor(private readonly store: WorkspaceStore) {}

  nextId(): string {
    return this.store.nextId('skill')
  }

  list(): JsonRecord[] {
    return sortedValues(this.store.snapshot.skills, 'skillId')
  }

  get(skillId: string): JsonRecord | undefined {
    return this.store.snapshot.skills[skillId]
  }

  save(skillId: string, skill: JsonRecord): void {
    this.store.snapshot.skills[skillId] = skill
    this.store.save()
  }

  delete(skillId: string): void {
    delete this.store.snapshot.skills[skillId]
    this.store.save()
  }
}
