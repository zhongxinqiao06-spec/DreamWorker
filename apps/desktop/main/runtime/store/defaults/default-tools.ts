import type { JsonRecord } from '../../types'

export function defaultTools(): Record<string, JsonRecord> {
  const tools = [
    [
      'tool_artifact_read',
      '读取产物',
      '读取项目空间内的 Artifact 元数据和内容。',
      'artifact',
      'low'
    ],
    [
      'tool_artifact_write',
      '写入产物',
      '只允许写入当前项目目录内的 Artifact。',
      'artifact',
      'medium'
    ],
    [
      'tool_model_generate_stub',
      '模型生成 Stub',
      '用于离线演示与 CI 的确定性模型能力。',
      'model',
      'low'
    ],
    ['tool_human_input', '人工输入', '把审批和 steering 交还给用户。', 'human', 'low']
  ] as const
  return Object.fromEntries(
    tools.map(([toolId, displayName, description, category, riskLevel]) => [
      toolId,
      { toolId, displayName, description, category, riskLevel, enabled: true, builtIn: true }
    ])
  )
}
