<script setup lang="ts">
import ProjectInspector from './projects/ProjectInspector.vue'
import ProjectListPanel from './projects/ProjectListPanel.vue'
import ProjectSettingsPanel from './projects/ProjectSettingsPanel.vue'
</script>

<template>
  <section class="workspace-layout projects-layout">
    <ProjectListPanel />
    <ProjectSettingsPanel />
    <ProjectInspector />
  </section>
</template>
