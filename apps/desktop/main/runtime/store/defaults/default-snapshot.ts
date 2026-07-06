import { nowISO } from '../../shared/util'
import type { JsonRecord, WorkspaceSnapshot } from '../../types'
import { defaultAgents } from './default-agents'
import { createProjectModules } from './default-modules'
import { createProfile } from './default-profiles'
import { createDefaultProviders } from './default-providers'
import { createProject } from './default-projects'
import { defaultSettings } from './default-settings'
import { defaultSkills } from './default-skills'
import { defaultTools } from './default-tools'

export function createDefaultSnapshot(): WorkspaceSnapshot {
  const timestamp = '2026-07-01T00:00:00Z'
  const providerSeed = createDefaultProviders(timestamp)
  const project = createProject('project_001', timestamp, {
    title: '独立开发者 AI 项目孵化器',
    description: '从机会探索、产品定义、工程开发到销售发布的默认项目空间。'
  })
  const projectId = String(project.projectId)
  const modules = createProjectModules(projectId)
  const session = createChatSession('chat_001', timestamp, projectId)
  const sessionId = String(session.sessionId)

  return {
    schemaVersion: 'dreamworker.workspace.snapshot.v1',
    sequence: 1,
    providers: providerSeed.providers,
    providerSecrets: providerSeed.providerSecrets,
    profiles: {
      profile_fast: createProfile(
        'profile_fast',
        'DeepSeek V4 Flash 快速',
        'provider_deepseek',
        providerSeed.deepseekModel,
        timestamp
      ),
      profile_pro: createProfile(
        'profile_pro',
        'DeepSeek V4 Pro 高级',
        'provider_deepseek',
        providerSeed.deepseekProModel,
        timestamp
      ),
      profile_stub: createProfile(
        'profile_stub',
        '离线确定性模型',
        'provider_local_stub',
        'model_generate_stub',
        timestamp,
        {
          temperature: 0,
          toolMode: 'none',
          fallbackProfileId: null,
          timeoutMs: 30000
        }
      )
    },
    agents: defaultAgents(timestamp),
    skills: defaultSkills(),
    tools: defaultTools(),
    mcpServers: {
      mcp_local_files: {
        serverId: 'mcp_local_files',
        displayName: '本地文件 MCP',
        command: 'dreamworker-mcp-files',
        args: ['--project-root', '.'],
        envKeys: [],
        url: null,
        trustLevel: 'trusted_builtin',
        enabled: false,
        hasSecrets: false,
        maskedSecrets: [],
        createdAt: timestamp,
        updatedAt: timestamp
      }
    },
    mcpServerSecrets: {},
    mcpTools: {},
    projects: { [projectId]: project },
    modules: { [projectId]: modules },
    sessions: { [sessionId]: session },
    messages: { [sessionId]: [] },
    contextSummaries: {},
    settings: defaultSettings()
  }
}

export function ensureSnapshotShape(snapshot: WorkspaceSnapshot): WorkspaceSnapshot {
  return {
    schemaVersion: snapshot.schemaVersion || 'dreamworker.workspace.snapshot.v1',
    sequence: Number.isFinite(snapshot.sequence) ? snapshot.sequence : 0,
    providers: snapshot.providers ?? {},
    providerSecrets: snapshot.providerSecrets ?? {},
    profiles: snapshot.profiles ?? {},
    agents: snapshot.agents ?? {},
    skills: snapshot.skills ?? {},
    tools: snapshot.tools ?? {},
    mcpServers: snapshot.mcpServers ?? {},
    mcpServerSecrets: snapshot.mcpServerSecrets ?? {},
    mcpTools: snapshot.mcpTools ?? {},
    projects: snapshot.projects ?? {},
    modules: snapshot.modules ?? {},
    sessions: snapshot.sessions ?? {},
    messages: snapshot.messages ?? {},
    contextSummaries: snapshot.contextSummaries ?? {},
    settings: snapshot.settings ?? defaultSettings()
  }
}

export function touch(record: JsonRecord): JsonRecord {
  return { ...record, updatedAt: nowISO() }
}

function createChatSession(sessionId: string, timestamp: string, projectId: string): JsonRecord {
  return {
    sessionId,
    title: '通用 Agent 工作台',
    projectId,
    agentId: 'agent_general_assistant',
    modelProfileId: 'profile_fast',
    status: 'active',
    lastMessageAt: null,
    createdAt: timestamp,
    updatedAt: timestamp
  }
}
