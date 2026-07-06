import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { SkillRepository } from '../../store/repositories/skill-repository'

export class SkillService {
  constructor(private readonly skills: SkillRepository) {}

  listSkills(): JsonRecord[] {
    return this.skills.list()
  }

  getSkill(skillId: string): JsonRecord {
    const skill = this.skills.get(skillId)
    if (!skill) {
      throw notFound('SKILL_NOT_FOUND', 'skill not found', 'refresh list')
    }
    return skill
  }

  saveSkill(input: JsonRecord): JsonRecord {
    const skillId = asString(input.skillId) || this.skills.nextId()
    const previous = this.skills.get(skillId) ?? {}
    const now = nowISO()
    const skill = {
      ...previous,
      ...input,
      skillId,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    this.skills.save(skillId, skill)
    return skill
  }

  deleteSkill(skillId: string): DeleteResult {
    if (!skillId) {
      throw badRequest('BAD_REQUEST', 'missing skillId', 'select an item')
    }
    if (!this.skills.get(skillId)) {
      throw notFound('RESOURCE_NOT_FOUND', 'resource not found', 'refresh list')
    }
    this.skills.delete(skillId)
    return { ok: true, deletedId: skillId }
  }
}
