<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import ChatWorkspace from './components/ChatWorkspace.vue'
import DiagnosticsWorkspace from './components/DiagnosticsWorkspace.vue'
import ModuleWorkspace from './components/modules/ModuleWorkspace.vue'
import ProjectsWorkspace from './components/ProjectsWorkspace.vue'
import ResourceCenter from './components/ResourceCenter.vue'
import SettingsWorkspace from './components/SettingsWorkspace.vue'
import SplashVortexCanvas from './components/SplashVortexCanvas.vue'
import WorkspaceTopBar from './components/WorkspaceTopBar.vue'
import { useAppShellStore } from './stores/app-shell'

const appShell = useAppShellStore()
const splashVisible = ref(true)
const minSplashMs = 1000
let splashStartedAt = performance.now()
let splashTimer: number | undefined

function handleCommandShortcut(event: KeyboardEvent): void {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    appShell.toggleCommand()
  }
}

function clearSplashTimer(): void {
  if (!splashTimer) {
    return
  }
  window.clearTimeout(splashTimer)
  splashTimer = undefined
}

function showSplash(): void {
  clearSplashTimer()
  if (!splashVisible.value) {
    splashStartedAt = performance.now()
  }
  splashVisible.value = true
}

function hideSplashAfterMinimum(): void {
  const elapsedMs = performance.now() - splashStartedAt
  const waitMs = Math.max(0, minSplashMs - elapsedMs)
  clearSplashTimer()
  splashTimer = window.setTimeout(() => {
    splashVisible.value = false
    splashTimer = undefined
  }, waitMs)
}

watch(
  () => appShell.bootStatus,
  (status) => {
    if (status === 'idle' || status === 'loading') {
      showSplash()
      return
    }
    hideSplashAfterMinimum()
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener('keydown', handleCommandShortcut)
  void appShell.loadWorkspace()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleCommandShortcut)
  clearSplashTimer()
})
</script>

<template>
  <main class="desktop-shell">
    <AppSidebar />
    <section class="main-stage">
      <WorkspaceTopBar />
      <p v-if="appShell.errorBanner" class="system-banner">{{ appShell.errorBanner }}</p>
      <section class="workspace-slot">
        <ChatWorkspace v-if="appShell.activePrimary === 'chat'" />
        <ProjectsWorkspace v-else-if="appShell.activePrimary === 'projects'" />
        <ResourceCenter v-else-if="appShell.activePrimary === 'resources'" />
        <ModuleWorkspace v-else-if="appShell.activeModuleWorkspace" />
        <SettingsWorkspace v-else-if="appShell.activePrimary === 'settings'" />
        <DiagnosticsWorkspace v-else />
      </section>
    </section>

    <section
      v-if="appShell.commandOpen"
      class="command-palette"
      role="dialog"
      aria-label="命令面板"
    >
      <div class="command-box">
        <strong>命令面板</strong>
        <button type="button" @click="appShell.runCommand('chat')">打开聊天工作台</button>
        <button type="button" @click="appShell.runCommand('projects')">打开项目配置</button>
        <button type="button" @click="appShell.runCommand('resources')">打开资源配置中心</button>
        <button type="button" @click="appShell.runCommand('explore')">进入探索模块</button>
        <button type="button" @click="appShell.runCommand('product')">进入产品模块</button>
        <button type="button" @click="appShell.runCommand('development')">进入开发模块</button>
        <button type="button" @click="appShell.runCommand('sales')">进入销售模块</button>
        <button type="button" @click="appShell.runCommand('providers')">管理模型服务商</button>
        <button type="button" @click="appShell.runCommand('mcp')">管理 MCP 服务</button>
        <button type="button" @click="appShell.runCommand('diagnostics')">查看诊断</button>
      </div>
    </section>

    <Transition name="splash-fade">
      <SplashVortexCanvas v-if="splashVisible" />
    </Transition>
  </main>
</template>

<style scoped>
.splash-fade-enter-active,
.splash-fade-leave-active {
  transition: opacity 420ms ease;
}

.splash-fade-enter-from,
.splash-fade-leave-to {
  opacity: 0;
}
</style>
