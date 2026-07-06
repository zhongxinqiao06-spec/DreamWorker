import { badRequest, notFound } from '../../kernel/errors'
import { asString, nowISO } from '../../shared/util'
import type { DeleteResult, JsonRecord } from '../../types'
import type { ProfileRepository } from '../../store/repositories/profile-repository'

export class ProfileService {
  constructor(private readonly profiles: ProfileRepository) {}

  listProfiles(): JsonRecord[] {
    return this.profiles.list()
  }

  saveProfile(input: JsonRecord): JsonRecord {
    const profileId = asString(input.profileId) || this.profiles.nextId()
    const previous = this.profiles.get(profileId) ?? {}
    const now = nowISO()
    const profile = {
      ...previous,
      ...input,
      profileId,
      createdAt: asString(previous.createdAt) || now,
      updatedAt: now
    }
    this.profiles.save(profileId, profile)
    return profile
  }

  deleteProfile(profileId: string): DeleteResult {
    if (!profileId) {
      throw badRequest('BAD_REQUEST', 'missing profileId', 'select an item')
    }
    if (!this.profiles.get(profileId)) {
      throw notFound('RESOURCE_NOT_FOUND', 'resource not found', 'refresh list')
    }
    this.profiles.delete(profileId)
    return { ok: true, deletedId: profileId }
  }
}
