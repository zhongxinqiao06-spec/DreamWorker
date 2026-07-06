<script setup lang="ts">
import { computed } from 'vue'
import { Boxes, FolderCheck, Settings2, ShieldCheck, Zap } from 'lucide-vue-next'
import { projectModuleIds } from '../../stores/project-draft'
import { useAppShellStore } from '../../stores/app-shell'

const appShell = useAppShellStore()

const directoryStatus = computed(
  () =>
    appShell.activeProjectDirectoryCheck?.status ??
    appShell.activeProject?.localDirectoryStatus ??
    'not_set'
)

const enabledModuleCount = computed(
  () =>
    projectModuleIds.filter((moduleId) => appShell.projectDraft.moduleConfigs[moduleId].enabled)
      .length
)

const highRiskToolCount = computed(
  () =>
    appShell.tools.filter(
      (tool) =>
        appShell.projectDraft.enabledTools.includes(tool.toolId) &&
        (tool.riskLevel === 'high' || tool.riskLevel === 'critical')
    ).length
)

const disconnectedMcpCount = computed(
  () =>
    appShell.mcpServers.filter(
      (server) =>
        appShell.projectDraft.enabledMcpServers.includes(server.serverId) && !server.enabled
    ).length
)
</script>

<template>
  <aside class="right-panel" aria-label="项目配置检查器">
    <section class="inspector-card">
      <p class="eyebrow">当前分区</p>
      <h3>
        {{
          {
            basic: '基础信息',
            directory: '本地目录',
            resources: '资源绑定',
            modules: '模块配置',
            'run-policy': '运行策略',
            security: '安全与导出'
          }[appShell.activeProjectSettingsTab]
        }}
      </h3>
      <p>{{ appShell.activeProject?.projectId ?? '暂无项目' }}</p>
    </section>

    <section v-if="appShell.activeProjectSettingsTab === 'basic'" class="inspector-card">
      <Settings2 :size="18" aria-hidden="true" />
      <h3>项目元信息</h3>
      <dl>
        <div>
          <dt>状态</dt>
          <dd>{{ appShell.projectDraft.status }}</dd>
        </div>
        <div>
          <dt>默认模型</dt>
          <dd>{{ appShell.projectDraft.defaultModelProfileId }}</dd>
        </div>
        <div>
          <dt>更新</dt>
          <dd>{{ appShell.activeProject?.updatedAt ?? '暂无' }}</dd>
        </div>
      </dl>
    </section>

    <section v-else-if="appShell.activeProjectSettingsTab === 'directory'" class="inspector-card">
      <FolderCheck :size="18" aria-hidden="true" />
      <h3>目录校验</h3>
      <dl>
        <div>
          <dt>状态</dt>
          <dd>{{ directoryStatus }}</dd>
        </div>
        <div>
          <dt>可读</dt>
          <dd>{{ appShell.activeProjectDirectoryCheck?.readable ? '是' : '否' }}</dd>
        </div>
        <div>
          <dt>可写</dt>
          <dd>{{ appShell.activeProjectDirectoryCheck?.writable ? '是' : '否' }}</dd>
        </div>
        <div>
          <dt>结构</dt>
          <dd>
            {{ appShell.activeProjectDirectoryCheck?.dreamworkerInitialized ? '完整' : '待初始化' }}
          </dd>
        </div>
      </dl>
      <p>{{ appShell.activeProjectDirectoryCheck?.message ?? '目录尚未检测。' }}</p>
    </section>

    <section v-else-if="appShell.activeProjectSettingsTab === 'resources'" class="inspector-card">
      <Boxes :size="18" aria-hidden="true" />
      <h3>绑定摘要</h3>
      <dl>
        <div>
          <dt>Agent</dt>
          <dd>{{ appShell.projectDraft.enabledAgents.length }}</dd>
        </div>
        <div>
          <dt>Skill</dt>
          <dd>{{ appShell.projectDraft.enabledSkills.length }}</dd>
        </div>
        <div>
          <dt>Tool</dt>
          <dd>{{ appShell.projectDraft.enabledTools.length }}</dd>
        </div>
        <div>
          <dt>MCP</dt>
          <dd>{{ appShell.projectDraft.enabledMcpServers.length }}</dd>
        </div>
        <div>
          <dt>高风险</dt>
          <dd>{{ highRiskToolCount }}</dd>
        </div>
        <div>
          <dt>未连接 MCP</dt>
          <dd>{{ disconnectedMcpCount }}</dd>
        </div>
      </dl>
    </section>

    <section v-else-if="appShell.activeProjectSettingsTab === 'modules'" class="inspector-card">
      <Settings2 :size="18" aria-hidden="true" />
      <h3>模块默认值</h3>
      <dl>
        <div>
          <dt>启用模块</dt>
          <dd>{{ enabledModuleCount }}/4</dd>
        </div>
        <div v-for="moduleId in projectModuleIds" :key="moduleId">
          <dt>{{ moduleId }}</dt>
          <dd>{{ appShell.projectDraft.moduleConfigs[moduleId].outputDir }}</dd>
        </div>
      </dl>
    </section>

    <section v-else-if="appShell.activeProjectSettingsTab === 'run-policy'" class="inspector-card">
      <Zap :size="18" aria-hidden="true" />
      <h3>运行约束</h3>
      <dl>
        <div>
          <dt>Planner</dt>
          <dd>{{ appShell.projectDraft.runPolicy.plannerMode }}</dd>
        </div>
        <div>
          <dt>Executor</dt>
          <dd>{{ appShell.projectDraft.runPolicy.executorMode }}</dd>
        </div>
        <div>
          <dt>成本</dt>
          <dd>${{ appShell.projectDraft.runPolicy.maxRunCostUsd }}</dd>
        </div>
        <div>
          <dt>时长</dt>
          <dd>{{ appShell.projectDraft.runPolicy.maxRunMinutes }} 分钟</dd>
        </div>
      </dl>
    </section>

    <section v-else class="inspector-card">
      <ShieldCheck :size="18" aria-hidden="true" />
      <h3>安全边界</h3>
      <dl>
        <div>
          <dt>文件范围</dt>
          <dd>{{ appShell.projectDraft.securityPolicy.fileAccessScope }}</dd>
        </div>
        <div>
          <dt>Artifacts</dt>
          <dd>{{ appShell.projectDraft.securityPolicy.allowWriteArtifacts ? '可写' : '禁止' }}</dd>
        </div>
        <div>
          <dt>Source</dt>
          <dd>{{ appShell.projectDraft.securityPolicy.allowWriteSource ? '可写' : '禁止' }}</dd>
        </div>
        <div>
          <dt>Shell</dt>
          <dd>{{ appShell.projectDraft.securityPolicy.allowShellExecution ? '允许' : '禁止' }}</dd>
        </div>
        <div>
          <dt>Network</dt>
          <dd>{{ appShell.projectDraft.securityPolicy.allowNetworkTools ? '允许' : '禁止' }}</dd>
        </div>
      </dl>
    </section>
  </aside>
</template>
