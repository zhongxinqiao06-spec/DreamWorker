import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function skillRoutes(context: RuntimeContext): RuntimeRoute[] {
  return [
    get('/skills', () => context.skills.listSkills()),
    post('/skills/get', (body) => context.skills.getSkill(asString(body.skillId))),
    post('/skills/save', (body) => context.skills.saveSkill(body)),
    post('/skills/delete', (body) => context.skills.deleteSkill(asString(body.skillId)))
  ]
}
