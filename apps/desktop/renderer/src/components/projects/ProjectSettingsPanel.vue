<script setup lang="ts">
import { computed } from 'vue'
import {
  CheckCircle2,
  Download,
  Folder,
  FolderOpen,
  PlayCircle,
  RefreshCw,
  Save,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  Trash2,
  TriangleAlert
} from 'lucide-vue-next'
import type { ProjectModuleId } from '../../../../shared/dreamworker-api'
import { projectModuleIds, toggleSelection } from '../../stores/project-draft'
import { useAppShellStore, type ProjectResourceType } from '../../stores/app-shell'

const appShell = useAppShellStore()

const tabs = [
  { id: 'basic', label: '基础信息' },
  { id: 'directory', label: '本地目录' },
  { id: 'resources', label: '资源绑定' },
  { id: 'modules', label: '模块配置' },
  { id: 'run-policy', label: '运行策略' },
  { id: 'security', label: '安全与导出' }
] as const

const resourceTypes: readonly { id: ProjectResourceType; label: string }[] = [
  { id: 'agents', label: 'Agent' },
  { id: 'skills', label: 'Skill' },
  { id: 'tools', label: '工具' },
  { id: 'mcp', label: 'MCP' }
]

const moduleMeta: Record<
  ProjectModuleId,
  { label: string; intent: string; accent: string; next: string }
> = {
  explore: {
    label: '探索',
    intent: '商机挖掘 / 竞品对比 / 客群分析',
    accent: 'teal',
    next: '跑机会扫描'
  },
  product: {
    label: '产品',
    intent: '功能设计 / 原型设计 / PRD',
    accent: 'amber',
    next: '生成 PRD 草案'
  },
  development: {
    label: '开发',
    intent: '架构蓝图 / 技术选型 / 开发落地',
    accent: 'blue',
    next: '输出架构蓝图'
  },
  sales: {
    label: '销售',
    intent: '销售方案 / 演示设计 / 发布计划',
    accent: 'pink',
    next: '准备发布计划'
  }
}

const resourcePacks = [
  {
    id: 'lightweight',
    label: '轻量项目包',
    description: '通用助理、PRD 草案、人工输入和产物写入。',
    agents: ['agent_general_assistant'],
    skills: ['skill_prd_draft'],
    tools: ['tool_model_generate_stub', 'tool_human_input', 'tool_artifact_write'],
    mcp: []
  },
  {
    id: 'explore',
    label: '探索验证包',
    description: '机会侦察、竞品地图、搜索和产物沉淀。',
    agents: ['agent_opportunity_scout', 'agent_competitor_analyst'],
    skills: ['skill_opportunity_scan', 'skill_competitor_map'],
    tools: ['tool_web_search_stub', 'tool_model_generate_stub', 'tool_artifact_write'],
    mcp: []
  },
  {
    id: 'product',
    label: '产品设计包',
    description: '产品设计、原型说明、PRD 和评估。',
    agents: ['agent_product_designer', 'agent_prototype_designer', 'agent_evaluator'],
    skills: ['skill_prd_draft'],
    tools: ['tool_model_generate_stub', 'tool_artifact_write'],
    mcp: []
  },
  {
    id: 'development',
    label: '开发落地包',
    description: '系统架构、开发编排、蓝图和产物写入。',
    agents: ['agent_system_architect', 'agent_dev_orchestrator'],
    skills: ['skill_blueprint'],
    tools: ['tool_model_generate_stub', 'tool_artifact_write'],
    mcp: ['mcp_local_files']
  },
  {
    id: 'sales',
    label: '销售发布包',
    description: '销售策略、演示方案、发布计划和反馈闭环。',
    agents: ['agent_sales_strategist', 'agent_demo_designer'],
    skills: ['skill_launch_plan'],
    tools: ['tool_model_generate_stub', 'tool_artifact_write', 'tool_human_input'],
    mcp: []
  }
] as const

type ResourceCard = {
  id: string
  name: string
  description: string
  type: ProjectResourceType
  category: string
  risk: string
  builtIn: boolean
  enabled: boolean
  moduleHint: string
}

const boundResourceCount = computed(
  () =>
    appShell.projectDraft.enabledAgents.length +
    appShell.projectDraft.enabledSkills.length +
    appShell.projectDraft.enabledTools.length +
    appShell.projectDraft.enabledMcpServers.length
)

const directoryCheck = computed(() => appShell.activeProjectDirectoryCheck)

const localDirectoryStatus = computed(
  () => directoryCheck.value?.status ?? appShell.activeProject?.localDirectoryStatus ?? 'not_set'
)

const resourceCards = computed<ResourceCard[]>(() => {
  const query = appShell.projectResourceSearch.trim().toLowerCase()
  const cards = resourcesForType(appShell.projectResourceType)
  if (!query) {
    return cards
  }
  return cards.filter((item) =>
    [item.name, item.description, item.category, item.moduleHint, item.risk]
      .join(' ')
      .toLowerCase()
      .includes(query)
  )
})

const selectedResourceCards = computed<ResourceCard[]>(() =>
  [
    ...resourcesForType('agents'),
    ...resourcesForType('skills'),
    ...resourcesForType('tools'),
    ...resourcesForType('mcp')
  ].filter((item) => isResourceSelected(item))
)

function resourcesForType(type: ProjectResourceType): ResourceCard[] {
  if (type === 'agents') {
    return appShell.agents.map((agent) => ({
      id: agent.agentId,
      name: agent.displayName,
      description: agent.description,
      type,
      category: agent.role,
      risk: agent.memoryScope,
      builtIn: agent.builtIn,
      enabled: agent.enabled,
      moduleHint: moduleHintForIds(agent.enabledSkills, agent.enabledTools)
    }))
  }
  if (type === 'skills') {
    return appShell.skills.map((skill) => ({
      id: skill.skillId,
      name: skill.displayName,
      description: skill.description || skill.whenToUse,
      type,
      category: skill.category,
      risk: skill.outputArtifacts.length > 0 ? 'artifact' : 'low',
      builtIn: skill.builtIn,
      enabled: skill.enabled,
      moduleHint: skill.category
    }))
  }
  if (type === 'tools') {
    return appShell.tools.map((tool) => ({
      id: tool.toolId,
      name: tool.displayName,
      description: tool.description,
      type,
      category: tool.category,
      risk: tool.riskLevel,
      builtIn: tool.builtIn,
      enabled: tool.enabled,
      moduleHint: tool.category
    }))
  }
  return appShell.mcpServers.map((server) => ({
    id: server.serverId,
    name: server.displayName,
    description: server.url ?? server.command,
    type,
    category: server.trustLevel,
    risk: server.hasSecrets ? 'secret' : 'low',
    builtIn: server.trustLevel === 'trusted_builtin',
    enabled: server.enabled,
    moduleHint: server.enabled ? 'connected' : 'needs setup'
  }))
}

function moduleHintForIds(skillIds: readonly string[], toolIds: readonly string[]): string {
  if (skillIds.some((id) => id.includes('opportunity') || id.includes('competitor'))) {
    return 'explore'
  }
  if (skillIds.some((id) => id.includes('prd'))) {
    return 'product'
  }
  if (skillIds.some((id) => id.includes('blueprint'))) {
    return 'development'
  }
  if (skillIds.some((id) => id.includes('launch'))) {
    return 'sales'
  }
  if (toolIds.some((id) => id.includes('artifact'))) {
    return 'artifact'
  }
  return 'general'
}

function isResourceSelected(resource: ResourceCard): boolean {
  if (resource.type === 'agents') {
    return appShell.projectDraft.enabledAgents.includes(resource.id)
  }
  if (resource.type === 'skills') {
    return appShell.projectDraft.enabledSkills.includes(resource.id)
  }
  if (resource.type === 'tools') {
    return appShell.projectDraft.enabledTools.includes(resource.id)
  }
  return appShell.projectDraft.enabledMcpServers.includes(resource.id)
}

function toggleResource(resource: ResourceCard): void {
  if (resource.type === 'agents') {
    appShell.toggleProjectAgent(resource.id)
    return
  }
  if (resource.type === 'skills') {
    appShell.toggleProjectSkill(resource.id)
    return
  }
  if (resource.type === 'tools') {
    appShell.toggleProjectTool(resource.id)
    return
  }
  appShell.toggleProjectMcpServer(resource.id)
}

function applyResourcePack(pack: (typeof resourcePacks)[number]): void {
  appShell.projectDraft.enabledAgents = mergeIds(appShell.projectDraft.enabledAgents, pack.agents)
  appShell.projectDraft.enabledSkills = mergeIds(appShell.projectDraft.enabledSkills, pack.skills)
  appShell.projectDraft.enabledTools = mergeIds(appShell.projectDraft.enabledTools, pack.tools)
  appShell.projectDraft.enabledMcpServers = mergeIds(
    appShell.projectDraft.enabledMcpServers,
    pack.mcp
  )
}

function mergeIds(current: readonly string[], next: readonly string[]): string[] {
  return [...new Set([...current, ...next])]
}

function setFirstModuleResource(
  moduleId: ProjectModuleId,
  field: 'agent' | 'skill' | 'tool' | 'mcp',
  value: string
): void {
  const config = appShell.projectDraft.moduleConfigs[moduleId]
  if (field === 'agent') {
    config.defaultAgentIds = value ? mergeIds([value], config.defaultAgentIds) : []
  }
  if (field === 'skill') {
    config.enabledSkillIds = value ? mergeIds([value], config.enabledSkillIds) : []
  }
  if (field === 'tool') {
    config.enabledToolIds = value ? mergeIds([value], config.enabledToolIds) : []
  }
  if (field === 'mcp') {
    config.enabledMcpServerIds = value ? mergeIds([value], config.enabledMcpServerIds) : []
  }
}

function toggleModuleMcp(moduleId: ProjectModuleId, serverId: string): void {
  const config = appShell.projectDraft.moduleConfigs[moduleId]
  config.enabledMcpServerIds = toggleSelection(config.enabledMcpServerIds, serverId)
}

function eventValue(event: Event): string {
  return event.target instanceof HTMLSelectElement ? event.target.value : ''
}

function confirmDeleteProject(): void {
  const title = appShell.activeProject?.title ?? '当前项目'
  if (
    window.confirm(
      `确认仅删除 DreamWorker 项目记录「${title}」吗？本地目录和产物不会被删除。`
    )
  ) {
    void appShell.deleteActiveProject()
  }
}
</script>

<template>
  <section class="project-center panel-surface" aria-label="项目空间配置">
    <div class="project-header">
      <div>
        <p class="eyebrow">项目空间配置</p>
        <h2>{{ appShell.activeProject?.title ?? '暂无项目' }}</h2>
        <p>本地目录、资源能力、模块默认值、运行策略和安全边界统一绑定到 projectId。</p>
      </div>
      <div class="context-pills">
        <span>{{ localDirectoryStatus }}</span>
        <span>{{ boundResourceCount }} 个资源</span>
        <span>{{ appShell.activeProject?.updatedAt ?? '尚未同步' }}</span>
      </div>
    </div>

    <section v-if="appShell.activeProject" class="project-config-shell">
      <nav class="tab-strip project-tabs" aria-label="项目配置分区">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          :class="{ active: appShell.activeProjectSettingsTab === tab.id }"
          @click="appShell.setProjectSettingsTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </nav>

      <div class="project-config-scroll">
        <section v-if="appShell.activeProjectSettingsTab === 'basic'" class="project-tab-panel">
          <div class="project-metrics-strip">
            <article>
              <span>Project ID</span>
              <strong>{{ appShell.activeProject.projectId }}</strong>
            </article>
            <article>
              <span>状态</span>
              <strong>{{ appShell.projectDraft.status }}</strong>
            </article>
            <article>
              <span>目录</span>
              <strong>{{ localDirectoryStatus }}</strong>
            </article>
            <article>
              <span>资源</span>
              <strong>{{ boundResourceCount }}</strong>
            </article>
          </div>

          <div class="editor-card project-form-card">
            <label>
              项目名称
              <input v-model="appShell.projectDraft.title" aria-label="编辑项目名称" />
            </label>
            <label>
              项目描述
              <textarea v-model="appShell.projectDraft.description" aria-label="编辑项目描述" />
            </label>
            <div class="form-grid two">
              <label>
                项目状态
                <select v-model="appShell.projectDraft.status" aria-label="项目状态">
                  <option value="active">进行中</option>
                  <option value="paused">已暂停</option>
                  <option value="archived">已归档</option>
                </select>
              </label>
              <label>
                默认模型配置
                <select
                  v-model="appShell.projectDraft.defaultModelProfileId"
                  aria-label="默认模型配置"
                >
                  <option
                    v-for="profile in appShell.profiles"
                    :key="profile.profileId"
                    :value="profile.profileId"
                  >
                    {{ profile.displayName }}
                  </option>
                </select>
              </label>
              <label>
                默认路由配置
                <select
                  v-model="appShell.projectDraft.defaultRouteProfileId"
                  aria-label="默认路由配置"
                >
                  <option :value="null">跟随默认模型</option>
                  <option
                    v-for="profile in appShell.profiles"
                    :key="`route-${profile.profileId}`"
                    :value="profile.profileId"
                  >
                    {{ profile.displayName }}
                  </option>
                </select>
              </label>
              <label>
                项目标签
                <input value="incubation, ai-product" aria-label="项目标签" readonly />
              </label>
            </div>
          </div>
        </section>

        <section
          v-else-if="appShell.activeProjectSettingsTab === 'directory'"
          class="project-tab-panel"
        >
          <div class="editor-card project-directory-card">
            <div class="section-title">
              <Folder :size="17" aria-hidden="true" />
              <strong>本地项目目录</strong>
            </div>
            <div class="directory-path-row">
              <input
                v-model="appShell.projectDraft.localRootPath"
                placeholder="选择或粘贴项目根目录"
                aria-label="本地项目目录"
              />
              <button type="button" title="选择目录" @click="appShell.chooseProjectLocalDirectory">
                <Folder :size="16" aria-hidden="true" />
              </button>
              <button type="button" title="打开目录" @click="appShell.openActiveProjectDirectory">
                <FolderOpen :size="16" aria-hidden="true" />
              </button>
              <button type="button" title="重新检测" @click="appShell.validateActiveProjectDirectory">
                <RefreshCw :size="16" aria-hidden="true" />
              </button>
            </div>
            <div class="horizontal-actions">
              <button class="primary-button" type="button" @click="appShell.initializeActiveProjectDirectory">
                <CheckCircle2 :size="15" aria-hidden="true" />
                初始化项目目录
              </button>
            </div>
          </div>

          <div class="directory-grid">
            <section class="editor-card">
              <h3>目录状态</h3>
              <dl class="compact-dl">
                <div>
                  <dt>存在</dt>
                  <dd>{{ directoryCheck?.exists ? '是' : '否' }}</dd>
                </div>
                <div>
                  <dt>可读</dt>
                  <dd>{{ directoryCheck?.readable ? '是' : '否' }}</dd>
                </div>
                <div>
                  <dt>可写</dt>
                  <dd>{{ directoryCheck?.writable ? '是' : '否' }}</dd>
                </div>
                <div>
                  <dt>.dreamworker</dt>
                  <dd>{{ directoryCheck?.dreamworkerInitialized ? '已初始化' : '未初始化' }}</dd>
                </div>
                <div>
                  <dt>最近检测</dt>
                  <dd>{{ directoryCheck?.lastCheckedAt ?? appShell.activeProject.localDirectoryLastCheckedAt ?? '暂无' }}</dd>
                </div>
              </dl>
            </section>

            <section class="editor-card">
              <h3>目录结构</h3>
              <div class="directory-tree">
                <span>.dreamworker/</span>
                <span>docs/</span>
                <span>artifacts/explore/ product/ development/ sales/</span>
                <span>workspace/imports/ exports/ temp/</span>
                <span>source/repo/</span>
              </div>
            </section>

            <section class="editor-card directory-risk-card">
              <div class="section-title">
                <TriangleAlert :size="17" aria-hidden="true" />
                <strong>当前风险</strong>
              </div>
              <p>{{ directoryCheck?.message ?? '尚未检测本地目录。' }}</p>
              <p v-if="localDirectoryStatus === 'not_set'">项目只能保存在数据库，文件产物没有落地根目录。</p>
              <p v-else-if="localDirectoryStatus === 'permission_denied'">目录权限不足，产物导出和开发落地会失败。</p>
              <p v-else-if="localDirectoryStatus === 'invalid'">目录结构未初始化，工具写入会被项目策略阻断。</p>
              <p v-else-if="localDirectoryStatus === 'valid'">目录结构完整，可以作为项目产物和后续开发落地边界。</p>
            </section>
          </div>
        </section>

        <section
          v-else-if="appShell.activeProjectSettingsTab === 'resources'"
          class="project-tab-panel resource-binding-panel"
        >
          <div class="project-metrics-strip">
            <article>
              <span>Agent</span>
              <strong>{{ appShell.projectDraft.enabledAgents.length }}</strong>
            </article>
            <article>
              <span>Skill</span>
              <strong>{{ appShell.projectDraft.enabledSkills.length }}</strong>
            </article>
            <article>
              <span>Tool</span>
              <strong>{{ appShell.projectDraft.enabledTools.length }}</strong>
            </article>
            <article>
              <span>MCP</span>
              <strong>{{ appShell.projectDraft.enabledMcpServers.length }}</strong>
            </article>
          </div>

          <section class="resource-pack-row" aria-label="推荐资源包">
            <button v-for="pack in resourcePacks" :key="pack.id" type="button" @click="applyResourcePack(pack)">
              <strong>{{ pack.label }}</strong>
              <span>{{ pack.description }}</span>
            </button>
          </section>

          <div class="resource-selector-layout">
            <aside class="resource-type-rail" aria-label="资源类型">
              <button
                v-for="type in resourceTypes"
                :key="type.id"
                type="button"
                :class="{ active: appShell.projectResourceType === type.id }"
                @click="appShell.setProjectResourceType(type.id)"
              >
                {{ type.label }}
              </button>
            </aside>

            <section class="resource-picker-column">
              <label class="rail-search">
                <Search :size="15" aria-hidden="true" />
                <input
                  v-model="appShell.projectResourceSearch"
                  placeholder="搜索名称、用途、风险或模块"
                  aria-label="搜索项目资源"
                />
              </label>
              <div class="resource-card-list">
                <button
                  v-for="resource in resourceCards"
                  :key="resource.id"
                  class="project-resource-card"
                  :class="{ selected: isResourceSelected(resource) }"
                  type="button"
                  @click="toggleResource(resource)"
                >
                  <span class="resource-kind">{{ resource.type }}</span>
                  <strong>{{ resource.name }}</strong>
                  <p>{{ resource.description }}</p>
                  <span>{{ resource.category }}</span>
                  <span>{{ resource.risk }}</span>
                  <span>{{ resource.moduleHint }}</span>
                  <span>{{ resource.builtIn ? '内置' : '自定义' }}</span>
                </button>
              </div>
            </section>

            <aside class="selected-resource-panel" aria-label="已选资源">
              <h3>已选资源</h3>
              <div class="selected-resource-list">
                <button
                  v-for="resource in selectedResourceCards"
                  :key="`${resource.type}-${resource.id}`"
                  type="button"
                  @click="toggleResource(resource)"
                >
                  <span>{{ resource.type }}</span>
                  {{ resource.name }}
                </button>
              </div>
            </aside>
          </div>
        </section>

        <section v-else-if="appShell.activeProjectSettingsTab === 'modules'" class="project-tab-panel">
          <div class="module-config-grid">
            <section
              v-for="moduleId in projectModuleIds"
              :key="moduleId"
              class="editor-card module-config-card"
              :data-accent="moduleMeta[moduleId].accent"
            >
              <header>
                <div>
                  <p class="eyebrow">{{ moduleMeta[moduleId].label }}</p>
                  <h3>{{ moduleMeta[moduleId].intent }}</h3>
                </div>
                <label class="switch-row">
                  <input
                    v-model="appShell.projectDraft.moduleConfigs[moduleId].enabled"
                    type="checkbox"
                    :aria-label="`${moduleMeta[moduleId].label}模块启用状态`"
                  />
                  启用
                </label>
              </header>
              <label>
                输出目录
                <input
                  v-model="appShell.projectDraft.moduleConfigs[moduleId].outputDir"
                  :aria-label="`${moduleMeta[moduleId].label}输出目录`"
                />
              </label>
              <div class="form-grid two">
                <label>
                  默认 Agent
                  <select
                    :value="appShell.projectDraft.moduleConfigs[moduleId].defaultAgentIds[0] ?? ''"
                    @change="
                      setFirstModuleResource(
                        moduleId,
                        'agent',
                        eventValue($event)
                      )
                    "
                  >
                    <option value="">未指定</option>
                    <option v-for="agent in appShell.agents" :key="agent.agentId" :value="agent.agentId">
                      {{ agent.displayName }}
                    </option>
                  </select>
                </label>
                <label>
                  默认 Skill
                  <select
                    :value="appShell.projectDraft.moduleConfigs[moduleId].enabledSkillIds[0] ?? ''"
                    @change="
                      setFirstModuleResource(
                        moduleId,
                        'skill',
                        eventValue($event)
                      )
                    "
                  >
                    <option value="">未指定</option>
                    <option v-for="skill in appShell.skills" :key="skill.skillId" :value="skill.skillId">
                      {{ skill.displayName }}
                    </option>
                  </select>
                </label>
                <label>
                  默认 Tool
                  <select
                    :value="appShell.projectDraft.moduleConfigs[moduleId].enabledToolIds[0] ?? ''"
                    @change="
                      setFirstModuleResource(
                        moduleId,
                        'tool',
                        eventValue($event)
                      )
                    "
                  >
                    <option value="">未指定</option>
                    <option v-for="tool in appShell.tools" :key="tool.toolId" :value="tool.toolId">
                      {{ tool.displayName }}
                    </option>
                  </select>
                </label>
                <label>
                  MCP
                  <select
                    :value="appShell.projectDraft.moduleConfigs[moduleId].enabledMcpServerIds[0] ?? ''"
                    @change="
                      setFirstModuleResource(
                        moduleId,
                        'mcp',
                        eventValue($event)
                      )
                    "
                  >
                    <option value="">未指定</option>
                    <option
                      v-for="server in appShell.mcpServers"
                      :key="server.serverId"
                      :value="server.serverId"
                    >
                      {{ server.displayName }}
                    </option>
                  </select>
                </label>
              </div>
              <div class="tag-list">
                <span>{{ moduleMeta[moduleId].next }}</span>
                <span>{{ appShell.projectDraft.moduleConfigs[moduleId].outputDir }}</span>
              </div>
              <div v-if="appShell.mcpServers.length" class="module-mcp-row">
                <button
                  v-for="server in appShell.mcpServers"
                  :key="server.serverId"
                  type="button"
                  :class="{
                    active: appShell.projectDraft.moduleConfigs[moduleId].enabledMcpServerIds.includes(
                      server.serverId
                    )
                  }"
                  @click="toggleModuleMcp(moduleId, server.serverId)"
                >
                  {{ server.displayName }}
                </button>
              </div>
            </section>
          </div>
        </section>

        <section
          v-else-if="appShell.activeProjectSettingsTab === 'run-policy'"
          class="project-tab-panel"
        >
          <div class="policy-grid">
            <section class="editor-card">
              <div class="section-title">
                <PlayCircle :size="17" aria-hidden="true" />
                <strong>运行策略</strong>
              </div>
              <div class="form-grid two">
                <label>
                  Planner 模式
                  <select v-model="appShell.projectDraft.runPolicy.plannerMode">
                    <option value="plan_execute">plan-execute</option>
                    <option value="manual">manual</option>
                    <option value="react">react</option>
                  </select>
                </label>
                <label>
                  Executor 模式
                  <select v-model="appShell.projectDraft.runPolicy.executorMode">
                    <option value="safe">safe</option>
                    <option value="balanced">balanced</option>
                    <option value="aggressive">aggressive</option>
                  </select>
                </label>
                <label>
                  最大运行分钟
                  <input v-model.number="appShell.projectDraft.runPolicy.maxRunMinutes" type="number" min="1" />
                </label>
                <label>
                  最大成本 USD
                  <input
                    v-model.number="appShell.projectDraft.runPolicy.maxRunCostUsd"
                    type="number"
                    min="0"
                    step="0.1"
                  />
                </label>
              </div>
              <label class="check-row">
                <input
                  v-model="appShell.projectDraft.runPolicy.requireApprovalForHighRiskTools"
                  type="checkbox"
                />
                高风险工具需要审批
              </label>
            </section>

            <section class="editor-card">
              <div class="section-title">
                <SlidersHorizontal :size="17" aria-hidden="true" />
                <strong>上下文与记忆</strong>
              </div>
              <label class="check-row">
                <input v-model="appShell.projectDraft.memoryConfig.projectMemoryEnabled" type="checkbox" />
                启用项目记忆
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.memoryConfig.artifactIndexEnabled" type="checkbox" />
                启用产物索引
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.memoryConfig.localFileIndexEnabled" type="checkbox" />
                启用本地文件索引
              </label>
              <label>
                最大上下文 Token
                <input
                  v-model.number="appShell.projectDraft.memoryConfig.maxContextTokens"
                  type="number"
                  min="1000"
                  step="1000"
                />
              </label>
            </section>
          </div>
        </section>

        <section v-else class="project-tab-panel">
          <div class="policy-grid">
            <section class="editor-card">
              <div class="section-title">
                <ShieldCheck :size="17" aria-hidden="true" />
                <strong>安全边界</strong>
              </div>
              <label>
                文件访问范围
                <select v-model="appShell.projectDraft.securityPolicy.fileAccessScope">
                  <option value="project_directory_only">仅项目目录</option>
                  <option value="selected_directories">选定目录</option>
                  <option value="read_only">只读</option>
                </select>
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.securityPolicy.allowWriteArtifacts" type="checkbox" />
                允许写入 artifacts
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.securityPolicy.allowWriteSource" type="checkbox" />
                允许写入 source
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.securityPolicy.allowShellExecution" type="checkbox" />
                允许 shell/code execution
              </label>
              <label class="check-row">
                <input v-model="appShell.projectDraft.securityPolicy.allowNetworkTools" type="checkbox" />
                允许网络工具
              </label>
            </section>

            <section class="editor-card">
              <div class="section-title">
                <Download :size="17" aria-hidden="true" />
                <strong>导出与危险操作</strong>
              </div>
              <button class="primary-button" type="button" @click="appShell.exportActiveProjectManifest">
                <Download :size="15" aria-hidden="true" />
                导出项目 manifest
              </button>
              <button type="button" @click="appShell.projectDraft.status = 'archived'">
                归档项目
              </button>
              <button class="danger-button" type="button" @click="confirmDeleteProject">
                <Trash2 :size="15" aria-hidden="true" />
                仅删除项目记录
              </button>
              <p>本地目录删除未接入自动执行，避免误删用户文件。</p>
            </section>
          </div>
        </section>
      </div>
    </section>

    <section v-else class="placeholder-panel">
      <h3>暂无项目</h3>
      <p>先在左侧创建项目，再配置目录、资源、模块策略和安全边界。</p>
    </section>

    <div class="project-action-bar">
      <button
        class="primary-button"
        :class="{ 'is-dirty': appShell.projectDraftDirty }"
        type="button"
        :disabled="!appShell.activeProject || !appShell.projectDraftDirty"
        @click="appShell.saveActiveProject()"
      >
        <Save :size="15" aria-hidden="true" />
        保存配置
      </button>
      <button
        class="danger-button"
        type="button"
        :disabled="!appShell.activeProject"
        @click="confirmDeleteProject"
      >
        <Trash2 :size="15" aria-hidden="true" />
        删除项目记录
      </button>
    </div>
  </section>
</template>
