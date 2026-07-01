<script setup lang="ts">
import { Settings2 } from 'lucide-vue-next'
import { useAppShellStore } from '../../stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <aside class="right-panel" aria-label="项目配置摘要">
    <section class="inspector-card">
      <p class="eyebrow">项目隔离</p>
      <h3>{{ appShell.activeProject?.projectId ?? '暂无项目' }}</h3>
      <p>
        项目只负责基本信息和资源绑定。Agent、Skill、工具、MCP、产物、对话和运行记录都必须绑定
        projectId。
      </p>
      <dl>
        <div>
          <dt>模型配置</dt>
          <dd>{{ appShell.activeProject?.defaultModelProfileId ?? '暂无' }}</dd>
        </div>
        <div>
          <dt>项目状态</dt>
          <dd>{{ appShell.activeProject?.status ?? '暂无' }}</dd>
        </div>
        <div>
          <dt>更新时间</dt>
          <dd>{{ appShell.activeProject?.updatedAt ?? '暂无' }}</dd>
        </div>
      </dl>
    </section>

    <section class="inspector-card">
      <p class="eyebrow">闭环入口</p>
      <h3>探索 / 产品 / 开发 / 销售</h3>
      <p>
        四个模块已经提升为左侧主模块。它们不是固定 workflow，而是当前项目空间里的可配置能力组合。
      </p>
      <button type="button" @click="appShell.setPrimary('explore')">
        <Settings2 :size="15" aria-hidden="true" />
        进入探索模块
      </button>
    </section>

    <section class="inspector-card">
      <p class="eyebrow">绑定摘要</p>
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
          <dt>工具</dt>
          <dd>{{ appShell.projectDraft.enabledTools.length }}</dd>
        </div>
        <div>
          <dt>MCP</dt>
          <dd>{{ appShell.projectDraft.enabledMcpServers.length }}</dd>
        </div>
      </dl>
    </section>
  </aside>
</template>
