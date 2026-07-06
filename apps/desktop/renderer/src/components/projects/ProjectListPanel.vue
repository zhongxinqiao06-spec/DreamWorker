<script setup lang="ts">
import { FolderPlus } from 'lucide-vue-next'
import { useAppShellStore } from '../../stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <aside class="sub-rail project-list-rail" aria-label="项目列表">
    <div class="panel-heading compact">
      <div>
        <p class="eyebrow">项目</p>
        <h2>项目空间</h2>
      </div>
    </div>

    <div class="project-list-scroll">
      <button
        v-for="project in appShell.projects"
        :key="project.projectId"
        class="list-row project-list-row"
        :class="{ active: project.projectId === appShell.activeProjectId }"
        type="button"
        @click="appShell.selectProject(project.projectId)"
      >
        <strong>{{ project.title }}</strong>
        <span
          >{{ project.status }} / {{ project.localDirectoryStatus }} / {{ project.projectId }}</span
        >
      </button>
    </div>

    <section class="create-card">
      <strong>新建项目</strong>
      <input v-model="appShell.newProjectTitle" aria-label="项目名称" />
      <textarea v-model="appShell.newProjectDescription" aria-label="项目描述" />
      <button class="primary-button" type="button" @click="appShell.createProject()">
        <FolderPlus :size="15" aria-hidden="true" />
        创建
      </button>
    </section>
  </aside>
</template>
