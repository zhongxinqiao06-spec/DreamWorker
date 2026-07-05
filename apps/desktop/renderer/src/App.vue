<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import ChatWorkspace from './components/ChatWorkspace.vue'
import DiagnosticsWorkspace from './components/DiagnosticsWorkspace.vue'
import HomeWorkspace from './components/HomeWorkspace.vue'
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
    <header class="app-title-bar" aria-label="窗口标题栏">
      <div class="app-title-lockup">
        <img src="/aios/brand-mark.png" alt="" />
        <strong>DreamWorker AI 工作台</strong>
        <span>AI OS 2.0</span>
      </div>
    </header>
    <AppSidebar />
    <section class="main-stage">
      <WorkspaceTopBar />
      <p v-if="appShell.errorBanner" class="system-banner">{{ appShell.errorBanner }}</p>
      <Transition name="toast-fade">
        <div
          v-if="appShell.resourceNotice"
          class="resource-toast global-resource-toast"
          :data-tone="appShell.resourceNotice.tone"
          role="status"
        >
          {{ appShell.resourceNotice.message }}
        </div>
      </Transition>
      <section class="workspace-slot">
        <HomeWorkspace v-if="appShell.activePrimary === 'home'" />
        <ChatWorkspace v-else-if="appShell.activePrimary === 'chat'" />
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
        <button type="button" @click="appShell.runCommand('home')">打开工作台首页</button>
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

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition:
    opacity 160ms ease,
    transform 160ms ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
