import { existsSync, readFileSync } from 'node:fs'

const env = readLocalEnv('.env.local')
const apiKey = process.env.DEEPSEEK_API_KEY ?? env.DEEPSEEK_API_KEY
const baseUrl = trimTrailingSlash(
  process.env.DEEPSEEK_BASE_URL ?? env.DEEPSEEK_BASE_URL ?? 'https://api.deepseek.com'
)
const model = process.env.DEEPSEEK_MODEL ?? env.DEEPSEEK_MODEL ?? 'deepseek-v4-flash'
const isLongTask = process.argv.includes('--long-task')

if (!apiKey) {
  console.error('DeepSeek smoke skipped: DEEPSEEK_API_KEY is not configured.')
  process.exit(2)
}

const startedAt = Date.now()
const response = await fetch(`${baseUrl}/chat/completions`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${apiKey}`
  },
  body: JSON.stringify({
    model,
    messages: [
      {
        role: 'system',
        content: '你是 DreamWorker 的工程验证助手。只输出 JSON，不要包含密钥、环境变量或私人信息。'
      },
      {
        role: 'user',
        content: isLongTask
          ? '围绕 DreamWorker PR-07 的 Electron/Vue 桌面 AI 工作台，生成一个可执行的长任务 QA 清单。必须覆盖：普通 Agent 聊天工作台、资源配置中心、模型服务商脱敏、Agent/Skill/Tool/MCP 管理、项目空间、探索/产品/开发/销售四大模块、Command-K、中文 UI、runtime.ping 诊断、Main 到 Go Engine typed API 联通。JSON 字段：summary、checks、risks、next_best_action。checks 至少 8 条。'
          : '返回一个最小 JSON：{"ok":true,"message":"DeepSeek flash 连通成功"}。'
      }
    ],
    temperature: 0.2,
    max_tokens: isLongTask ? 1200 : 120,
    stream: false
  })
})

const elapsedMs = Date.now() - startedAt
const payload = await response.json().catch(() => ({}))

if (!response.ok) {
  console.error(
    JSON.stringify(
      {
        ok: false,
        model,
        status: response.status,
        elapsed_ms: elapsedMs,
        error: sanitizeError(payload)
      },
      null,
      2
    )
  )
  process.exit(1)
}

const content = payload?.choices?.[0]?.message?.content ?? ''
if (!content.trim()) {
  console.error(
    JSON.stringify(
      {
        ok: false,
        model,
        elapsed_ms: elapsedMs,
        error: 'empty model response'
      },
      null,
      2
    )
  )
  process.exit(1)
}
if (isLongTask && !/资源配置中心|项目空间|Agent|runtime\.ping|Go Engine/.test(content)) {
  console.error(
    JSON.stringify(
      {
        ok: false,
        model,
        elapsed_ms: elapsedMs,
        error: 'OFF_DOMAIN_RESPONSE'
      },
      null,
      2
    )
  )
  process.exit(1)
}

console.log(
  JSON.stringify(
    {
      ok: true,
      model,
      elapsed_ms: elapsedMs,
      usage: payload.usage ?? null,
      content_preview: content.slice(0, 600)
    },
    null,
    2
  )
)

function readLocalEnv(path) {
  if (!existsSync(path)) {
    return {}
  }

  const values = {}
  for (const line of readFileSync(path, 'utf8').split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      continue
    }
    const separator = trimmed.indexOf('=')
    if (separator === -1) {
      continue
    }
    const key = trimmed.slice(0, separator).trim()
    const value = trimmed.slice(separator + 1).trim()
    values[key] = value.replace(/^["']|["']$/g, '')
  }
  return values
}

function trimTrailingSlash(value) {
  return value.replace(/\/+$/, '')
}

function sanitizeError(payload) {
  const error = payload?.error ?? payload
  if (typeof error === 'string') {
    return error
  }
  if (!error || typeof error !== 'object') {
    return 'unknown error'
  }
  return {
    message: error.message ?? 'unknown error',
    type: error.type ?? undefined,
    code: error.code ?? undefined
  }
}
