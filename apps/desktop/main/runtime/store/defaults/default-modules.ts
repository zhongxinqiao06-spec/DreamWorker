import type { JsonRecord } from '../../types'

export const moduleIds = ['explore', 'product', 'development', 'sales'] as const

export function createProjectModules(projectId: string): Record<string, JsonRecord> {
  return {
    explore: moduleCard(
      projectId,
      'explore',
      '探索模块',
      'ready',
      '负责机会扫描、用户细分、竞品地图和证据收集。',
      [
        submodule(projectId, 'explore', 'opportunity_radar', '机会雷达', 'ready', [
          'dream_brief.md',
          'hypotheses.yaml'
        ])
      ]
    ),
    product: moduleCard(
      projectId,
      'product',
      '产品模块',
      'idle',
      '负责需求分析、PRD、原型说明和 Blueprint Canvas 输入。',
      [
        submodule(projectId, 'product', 'requirement_analysis', '需求分析', 'ready', [
          'feature_list.xlsx',
          'requirements_spec.docx',
          'requirements_analysis.json'
        ])
      ]
    ),
    development: moduleCard(
      projectId,
      'development',
      '开发模块',
      'idle',
      '负责系统架构、技术栈、PR 拆分、测试门禁和运行计划。',
      [
        submodule(projectId, 'development', 'architecture', '技术架构', 'idle', [
          'architecture.md'
        ]),
        submodule(projectId, 'development', 'coding_agent', '编码 Agent', 'ready', [
          '3 Engine',
          '文件树',
          '直接写入'
        ])
      ]
    ),
    sales: moduleCard(
      projectId,
      'sales',
      '销售模块',
      'idle',
      '负责定位、落地页文案、发布计划、Demo 和反馈循环。',
      [submodule(projectId, 'sales', 'launch_plan', '发布计划', 'idle', ['launch_checklist.md'])]
    )
  }
}

export function defaultModuleConfigs(): Record<string, JsonRecord> {
  return Object.fromEntries(
    moduleIds.map((moduleId) => [
      moduleId,
      {
        enabled: true,
        defaultAgentIds: [],
        enabledSkillIds: [],
        enabledToolIds: ['tool_model_generate_stub'],
        enabledMcpServerIds: [],
        outputDir: `artifacts/${moduleId}`,
        inputSchema: {},
        parameters: {}
      }
    ])
  )
}

function moduleCard(
  projectId: string,
  moduleId: string,
  displayName: string,
  status: string,
  summary: string,
  submodules: JsonRecord[]
): JsonRecord {
  return {
    projectId,
    moduleId,
    displayName,
    status,
    summary,
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_blueprint'],
    enabledTools: ['tool_model_generate_stub', 'tool_artifact_write'],
    enabledMcpServers: [],
    outputArtifacts: [],
    nextBestAction: '选择子模块继续推进。',
    submodules,
    config: {}
  }
}

function submodule(
  projectId: string,
  moduleId: string,
  submoduleId: string,
  displayName: string,
  status: string,
  outputArtifacts: string[]
): JsonRecord {
  return {
    projectId,
    moduleId,
    submoduleId,
    displayName,
    status,
    summary: `${displayName} 工作区已就绪。`,
    defaultAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_blueprint'],
    enabledTools: ['tool_model_generate_stub', 'tool_artifact_write'],
    outputArtifacts,
    nextBestAction: '进入工作区开始处理。',
    config: {}
  }
}
