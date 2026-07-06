import type { JsonRecord } from '../../types'

export function defaultSkills(): Record<string, JsonRecord> {
  return {
    skill_opportunity_scan: skill(
      'skill_opportunity_scan',
      'opportunity-scan',
      '机会扫描',
      '探索机会与风险'
    ),
    skill_blueprint: skill('skill_blueprint', 'blueprint', '技术蓝图', '产出架构与工程计划'),
    skill_prd_draft: skill('skill_prd_draft', 'prd-draft', 'PRD 草稿', '整理产品需求'),
    skill_launch_plan: skill('skill_launch_plan', 'launch-plan', '发布计划', '设计发布节奏')
  }
}

function skill(
  skillId: string,
  commandName: string,
  displayName: string,
  description: string
): JsonRecord {
  return {
    skillId,
    commandName,
    displayName,
    description,
    whenToUse: description,
    instructions: description,
    category: 'general',
    version: '0.1.0',
    enabled: true,
    builtIn: true,
    sourcePath: '',
    requiredCapabilities: [],
    outputArtifacts: []
  }
}
