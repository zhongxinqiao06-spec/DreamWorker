import type { RuntimeContext } from '../bootstrap/runtime-context'
import { agentRoutes } from './modules/agents.routes'
import { chatRoutes, chatStreamRoutes } from './modules/chat.routes'
import { codingRoutes, codingStreamRoutes } from './modules/coding.routes'
import { extensionRoutes } from './modules/extensions.routes'
import { mcpRoutes } from './modules/mcp.routes'
import { modelRoutes } from './modules/models.routes'
import { projectRoutes } from './modules/projects.routes'
import { requirementRoutes } from './modules/requirements.routes'
import { settingsRoutes } from './modules/settings.routes'
import { skillRoutes } from './modules/skills.routes'
import { toolRoutes } from './modules/tools.routes'
import { flattenRoutes, flattenStreamRoutes, RuntimeRouter } from './runtime-router'

export function createRuntimeRouter(context: RuntimeContext): RuntimeRouter {
  return new RuntimeRouter(
    flattenRoutes([
      modelRoutes(context),
      settingsRoutes(context),
      extensionRoutes(context),
      agentRoutes(context),
      skillRoutes(context),
      toolRoutes(context),
      mcpRoutes(context),
      projectRoutes(context),
      requirementRoutes(context),
      chatRoutes(context),
      codingRoutes(context)
    ]),
    flattenStreamRoutes([chatStreamRoutes(context), codingStreamRoutes(context)])
  )
}
