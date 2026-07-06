import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function modelRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/models/providers', () => store.listProviders()),
    post('/models/providers/save', (body) => store.saveProvider(body)),
    post('/models/providers/delete', (body) => store.deleteProvider(asString(body.providerId))),
    post('/models/providers/test', (body) => store.testProvider(asString(body.providerId))),
    post('/models/providers/refresh-models', (body) =>
      store.refreshProviderModels(asString(body.providerId))
    ),
    get('/models/profiles', () => store.listProfiles()),
    post('/models/profiles/save', (body) => store.saveProfile(body)),
    post('/models/profiles/delete', (body) => store.deleteProfile(asString(body.profileId)))
  ]
}
