import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function modelRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/models/providers', () => context.providers.listProviders()),
    post('/models/providers/save', (body) => context.providers.saveProvider(body)),
    post('/models/providers/delete', (body) =>
      context.providers.deleteProvider(asString(body.providerId))
    ),
    post('/models/providers/test', (body) =>
      context.providers.testProvider(asString(body.providerId))
    ),
    post('/models/providers/refresh-models', (body) =>
      context.providers.refreshProviderModels(asString(body.providerId))
    ),
    get('/models/profiles', () => context.profiles.listProfiles()),
    post('/models/profiles/save', (body) => context.profiles.saveProfile(body)),
    post('/models/profiles/delete', (body) =>
      context.profiles.deleteProfile(asString(body.profileId))
    )
  ]
}
