import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function settingsRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/settings', () => context.settings.getSettings()),
    post('/settings/update', (body) => context.settings.updateSettings(body)),
    post('/settings/reset-extension', () => context.settings.resetExtensionSettings())
  ]
}
