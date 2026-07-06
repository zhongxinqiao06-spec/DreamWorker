import type { ProjectModuleId, ProjectModuleStatus } from '../../../shared/dreamworker-api'

export type ModuleWorkspaceId = ProjectModuleId
export type PrimaryNavId =
  'home' | 'chat' | 'projects' | 'resources' | ModuleWorkspaceId | 'settings' | 'diagnostics'
export type ResourceTabId = 'providers' | 'extensions' | 'agents' | 'skills' | 'tools' | 'mcp'

export type PrimaryNavItem = {
  readonly id: PrimaryNavId
  readonly label: string
  readonly caption: string
}

export const moduleWorkspaceIds: readonly ModuleWorkspaceId[] = [
  'explore',
  'product',
  'development',
  'sales'
]

export const primaryNavItems: readonly PrimaryNavItem[] = [
  { id: 'home', label: '首页', caption: '项目指引、模型与 Token 总览' },
  { id: 'chat', label: '聊天', caption: '普通 Agent 工作台' },
  { id: 'projects', label: '项目', caption: '新增、修改、删除与基础配置' },
  { id: 'resources', label: '资源', caption: '模型 / Agent / MCP' },
  { id: 'explore', label: '探索', caption: '机会、用户、竞品、证据' },
  { id: 'product', label: '产品', caption: '需求、PRD、原型、蓝图' },
  { id: 'development', label: '开发', caption: '架构、成本、PR、测试' },
  { id: 'sales', label: '销售', caption: '定位、落地页、发布、反馈' },
  { id: 'settings', label: '设置', caption: '本地偏好与边界' },
  { id: 'diagnostics', label: '诊断', caption: '引擎与接口状态' }
]

export const resourceTabs: readonly {
  readonly id: ResourceTabId
  readonly label: string
}[] = [
  { id: 'providers', label: '模型服务商' },
  { id: 'extensions', label: '拓展能力' },
  { id: 'agents', label: 'Agent' },
  { id: 'skills', label: 'Skill' },
  { id: 'tools', label: '工具' },
  { id: 'mcp', label: 'MCP' }
]

export function isModuleWorkspace(id: PrimaryNavId): id is ModuleWorkspaceId {
  return moduleWorkspaceIds.includes(id as ModuleWorkspaceId)
}

export function moduleTitle(moduleId: ModuleWorkspaceId): string {
  const labels: Record<ModuleWorkspaceId, string> = {
    explore: '探索模块',
    product: '产品模块',
    development: '开发模块',
    sales: '销售模块'
  }
  return labels[moduleId]
}

export function moduleShortTitle(moduleId: ModuleWorkspaceId): string {
  const labels: Record<ModuleWorkspaceId, string> = {
    explore: '探索',
    product: '产品',
    development: '开发',
    sales: '销售'
  }
  return labels[moduleId]
}

export function statusLabel(status: ProjectModuleStatus | undefined): string {
  const labels: Record<ProjectModuleStatus, string> = {
    idle: '待启动',
    ready: '可运行',
    running: '运行中',
    blocked: '受阻',
    completed: '已完成'
  }
  return status ? labels[status] : '待加载'
}
