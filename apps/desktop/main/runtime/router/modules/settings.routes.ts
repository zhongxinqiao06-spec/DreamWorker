import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function settingsRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/settings', () => store.getSettings()),
    post('/settings/update', (body) => store.updateSettings(body)),
    post('/settings/reset-extension', () => store.resetExtensionSettings())
  ]
}
