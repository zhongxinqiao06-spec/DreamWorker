import type { Project } from '../../../shared/dreamworker-api'

export type ProjectConfigDraft = {
  title: string
  description: string
  status: Project['status']
  defaultModelProfileId: string
  enabledAgents: string[]
  enabledSkills: string[]
  enabledTools: string[]
  enabledMcpServers: string[]
}

export function createEmptyProjectDraft(): ProjectConfigDraft {
  return {
    title: '',
    description: '',
    status: 'active',
    defaultModelProfileId: 'profile_fast',
    enabledAgents: [],
    enabledSkills: [],
    enabledTools: [],
    enabledMcpServers: []
  }
}

export function createProjectDraft(project?: Project): ProjectConfigDraft {
  if (!project) {
    return createEmptyProjectDraft()
  }
  return {
    title: project.title,
    description: project.description,
    status: project.status,
    defaultModelProfileId: project.defaultModelProfileId,
    enabledAgents: [...project.enabledAgents],
    enabledSkills: [...project.enabledSkills],
    enabledTools: [...project.enabledTools],
    enabledMcpServers: [...project.enabledMcpServers]
  }
}

export function toggleSelection(values: readonly string[], value: string): string[] {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value]
}
