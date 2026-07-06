import type { JsonRecord } from '../../types'
import { defaultModuleConfigs } from './default-modules'

export const projectDirectoryLayout = [
  '.dreamworker',
  '.dreamworker/runs',
  '.dreamworker/logs',
  '.dreamworker/cache',
  '.dreamworker/indexes',
  'docs',
  'artifacts',
  'artifacts/explore',
  'artifacts/product',
  'artifacts/development',
  'artifacts/sales',
  'workspace',
  'workspace/code',
  'workspace/imports',
  'workspace/exports',
  'workspace/temp',
  'source',
  'source/repo'
] as const

export const projectDocumentStubs: Record<string, string> = {
  'docs/dream_brief.md': '# Dream Brief\n\n',
  'docs/research_pack.md': '# Research Pack\n\n',
  'docs/prd.md': '# PRD\n\n',
  'docs/architecture_blueprint.md': '# Architecture Blueprint\n\n',
  'docs/launch_plan.md': '# Launch Plan\n\n'
}

export function createProject(
  projectId: string,
  timestamp: string,
  input: { title: string; description: string; localRootPath?: string | null }
): JsonRecord {
  return {
    projectId,
    title: input.title,
    description: input.description,
    status: 'active',
    localRootPath: input.localRootPath ?? null,
    localDirectoryStatus: input.localRootPath ? 'invalid' : 'not_set',
    localDirectoryLastCheckedAt: null,
    defaultModelProfileId: 'profile_fast',
    defaultRouteProfileId: null,
    enabledAgents: ['agent_general_assistant'],
    enabledSkills: ['skill_opportunity_scan'],
    enabledTools: ['tool_model_generate_stub', 'tool_human_input'],
    enabledMcpServers: [],
    moduleConfigs: defaultModuleConfigs(),
    memoryConfig: {
      projectMemoryEnabled: true,
      artifactIndexEnabled: true,
      localFileIndexEnabled: false,
      maxContextTokens: 64000
    },
    runPolicy: {
      plannerMode: 'plan_execute',
      executorMode: 'safe',
      maxRunCostUsd: 5,
      maxRunMinutes: 30,
      requireApprovalForHighRiskTools: true
    },
    securityPolicy: {
      fileAccessScope: 'project_directory_only',
      allowWriteArtifacts: true,
      allowWriteSource: false,
      allowShellExecution: false,
      allowNetworkTools: true
    },
    createdAt: timestamp,
    updatedAt: timestamp
  }
}
