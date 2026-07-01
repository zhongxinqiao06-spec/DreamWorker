<script setup lang="ts">
import { Save, Trash2 } from 'lucide-vue-next'
import { useAppShellStore } from '../../stores/app-shell'

const appShell = useAppShellStore()

function confirmDeleteProject(): void {
  const title = appShell.activeProject?.title ?? '当前项目'
  if (window.confirm(`确认删除「${title}」吗？项目产物不会在本阶段自动删除。`)) {
    void appShell.deleteActiveProject()
  }
}
</script>

<template>
  <section class="project-center panel-surface" aria-label="项目基础配置">
    <div class="project-header">
      <div>
        <p class="eyebrow">基础信息</p>
        <h2>{{ appShell.activeProject?.title ?? '暂无项目' }}</h2>
        <p>项目页只维护对象、资源绑定和默认配置；运行闭环请从左侧探索、产品、开发、销售进入。</p>
      </div>
      <div class="context-pills">
        <span>{{ appShell.activeProject?.enabledAgents.length ?? 0 }} 个 Agent</span>
        <span>{{ appShell.activeProject?.enabledSkills.length ?? 0 }} 个 Skill</span>
        <span>{{ appShell.activeProject?.enabledTools.length ?? 0 }} 个工具</span>
      </div>
    </div>

    <section v-if="appShell.activeProject" class="project-config-scroll">
      <div class="editor-card">
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
            <select v-model="appShell.projectDraft.defaultModelProfileId" aria-label="默认模型配置">
              <option
                v-for="profile in appShell.profiles"
                :key="profile.profileId"
                :value="profile.profileId"
              >
                {{ profile.displayName }}
              </option>
            </select>
          </label>
        </div>
      </div>

      <div class="project-config-grid">
        <section class="editor-card">
          <h3>项目 Agent</h3>
          <label v-for="agent in appShell.agents" :key="agent.agentId" class="check-row">
            <input
              type="checkbox"
              :checked="appShell.projectDraft.enabledAgents.includes(agent.agentId)"
              @change="appShell.toggleProjectAgent(agent.agentId)"
            />
            <span>{{ agent.displayName }}</span>
          </label>
        </section>

        <section class="editor-card">
          <h3>项目 Skill</h3>
          <label v-for="skill in appShell.skills" :key="skill.skillId" class="check-row">
            <input
              type="checkbox"
              :checked="appShell.projectDraft.enabledSkills.includes(skill.skillId)"
              @change="appShell.toggleProjectSkill(skill.skillId)"
            />
            <span>{{ skill.displayName }}</span>
          </label>
        </section>

        <section class="editor-card">
          <h3>项目工具</h3>
          <label v-for="tool in appShell.tools" :key="tool.toolId" class="check-row">
            <input
              type="checkbox"
              :checked="appShell.projectDraft.enabledTools.includes(tool.toolId)"
              @change="appShell.toggleProjectTool(tool.toolId)"
            />
            <span>{{ tool.displayName }}</span>
          </label>
        </section>

        <section class="editor-card">
          <h3>MCP 服务</h3>
          <label v-for="server in appShell.mcpServers" :key="server.serverId" class="check-row">
            <input
              type="checkbox"
              :checked="appShell.projectDraft.enabledMcpServers.includes(server.serverId)"
              @change="appShell.toggleProjectMcpServer(server.serverId)"
            />
            <span>{{ server.displayName }}</span>
          </label>
        </section>
      </div>
    </section>

    <section v-else class="placeholder-panel">
      <h3>暂无项目</h3>
      <p>请先在左侧新增项目，然后再配置模型、Agent、Skill、工具和 MCP 绑定。</p>
    </section>

    <div class="project-action-bar">
      <button
        class="primary-button"
        type="button"
        :disabled="!appShell.activeProject"
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
        删除项目
      </button>
    </div>
  </section>
</template>
