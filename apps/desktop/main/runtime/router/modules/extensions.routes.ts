import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function extensionRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { extensions } = context
  return [
    get('/extensions', () => extensions.listExtensions()),
    post('/extensions/status', (body) => extensions.extensionStatus(asString(body.extensionId))),
    post('/extensions/detect', (body) => extensions.extensionAction(body, '检测')),
    post('/extensions/install', (body) => extensions.extensionAction(body, '安装')),
    post('/extensions/start', (body) => extensions.extensionAction(body, '启动')),
    post('/extensions/stop', (body) => extensions.extensionAction(body, '停止')),
    post('/extensions/restart', (body) => extensions.extensionAction(body, '重启')),
    post('/extensions/test', (body) => extensions.extensionAction(body, '测试')),
    post('/extensions/refresh-models', (body) => extensions.refreshModels(body)),
    post('/extensions/verify-streaming', (body) => extensions.verifyStreaming(body)),
    post('/extensions/logs/tail', () => extensions.tailLogs()),
    post('/extensions/logs/clear', (body) => extensions.extensionAction(body, '清理日志'))
  ]
}
