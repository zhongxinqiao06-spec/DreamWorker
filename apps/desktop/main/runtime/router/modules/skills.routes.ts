import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function skillRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/skills', () => store.listSkills()),
    post('/skills/get', (body) => store.getSkill(asString(body.skillId))),
    post('/skills/save', (body) => store.saveSkill(body)),
    post('/skills/delete', (body) => store.deleteSkill(asString(body.skillId)))
  ]
}
