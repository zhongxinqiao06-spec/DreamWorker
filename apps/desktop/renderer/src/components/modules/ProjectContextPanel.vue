<script setup lang="ts">
import { FolderKanban } from 'lucide-vue-next'
import { useAppShellStore } from '../../stores/app-shell'
import { moduleShortTitle, statusLabel } from '../../stores/workspace-navigation'

const appShell = useAppShellStore()
</script>

<template>
  <aside class="sub-rail" aria-label="项目上下文">
    <div class="panel-heading compact">
      <div>
        <p class="eyebrow">当前项目</p>
        <h2>{{ appShell.activeProject?.title ?? '暂无项目' }}</h2>
      </div>
    </div>

    <button
      v-for="project in appShell.projects"
      :key="project.projectId"
      class="list-row"
      :class="{ active: project.projectId === appShell.activeProjectId }"
      type="button"
      @click="appShell.selectProject(project.projectId)"
    >
      <strong>{{ project.title }}</strong>
      <span>{{ project.status }} / {{ project.projectId }}</span>
    </button>

    <section class="inspector-card">
      <FolderKanban :size="18" aria-hidden="true" />
      <p class="eyebrow">{{ moduleShortTitle(appShell.activeModuleWorkspace ?? 'explore') }}闭环</p>
      <h3>{{ appShell.activeModule?.displayName ?? '模块未加载' }}</h3>
      <p>{{ appShell.activeModule?.summary ?? '请选择项目后查看模块配置。' }}</p>
      <dl>
        <div>
          <dt>模块状态</dt>
          <dd>{{ statusLabel(appShell.activeModule?.status) }}</dd>
        </div>
        <div>
          <dt>子模块</dt>
          <dd>{{ appShell.activeModule?.submodules.length ?? 0 }}</dd>
        </div>
      </dl>
    </section>
  </aside>
</template>
