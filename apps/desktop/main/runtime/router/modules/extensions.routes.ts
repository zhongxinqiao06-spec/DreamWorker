import { asString } from '../../shared/util'
import type { RuntimeContext } from '../../bootstrap/runtime-context'
import { get, post, type RuntimeRoute } from '../route'

export function extensionRoutes(context: RuntimeContext): RuntimeRoute[] {
  const { store } = context
  return [
    get('/extensions', () => store.listExtensions()),
    post('/extensions/status', (body) => store.extensionStatus(asString(body.extensionId))),
    post('/extensions/detect', (body) => store.extensionAction(body, '检测')),
    post('/extensions/install', (body) => store.extensionAction(body, '安装')),
    post('/extensions/start', (body) => store.extensionAction(body, '启动')),
    post('/extensions/stop', (body) => store.extensionAction(body, '停止')),
    post('/extensions/restart', (body) => store.extensionAction(body, '重启')),
    post('/extensions/test', (body) => store.extensionAction(body, '测试')),
    post('/extensions/refresh-models', (body) => ({
      ok: true,
      extensionId: asString(body.extensionId) || '9router',
      models: [],
      status: store.extensionStatus(asString(body.extensionId))
    })),
    post('/extensions/verify-streaming', (body) => ({
      ok: true,
      extensionId: asString(body.extensionId) || '9router',
      message: 'Main 内嵌 Runtime stream bridge 可用。',
      latencyMs: 0,
      status: store.extensionStatus(asString(body.extensionId))
    })),
    post('/extensions/logs/tail', () => []),
    post('/extensions/logs/clear', (body) => store.extensionAction(body, '清理日志'))
  ]
}
