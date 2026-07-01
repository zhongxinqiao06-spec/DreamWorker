<script setup lang="ts">
import {
  Bot,
  Boxes,
  Code2,
  Compass,
  FileText,
  FolderKanban,
  Megaphone,
  MessageSquareText,
  Settings,
  Stethoscope
} from 'lucide-vue-next'
import { useAppShellStore, type PrimaryNavId } from '../stores/app-shell'

const appShell = useAppShellStore()

const icons: Record<PrimaryNavId, typeof MessageSquareText> = {
  chat: MessageSquareText,
  projects: FolderKanban,
  resources: Boxes,
  explore: Compass,
  product: FileText,
  development: Code2,
  sales: Megaphone,
  settings: Settings,
  diagnostics: Stethoscope
}
</script>

<template>
  <aside class="app-sidebar" aria-label="主导航">
    <div class="brand-block">
      <div class="brand-mark">DW</div>
      <div>
        <strong>DreamWorker</strong>
        <span>AI 工作台</span>
      </div>
    </div>

    <nav class="primary-nav">
      <button
        v-for="item in appShell.primaryNavItems"
        :key="item.id"
        class="nav-item"
        :class="{ active: appShell.activePrimary === item.id }"
        type="button"
        @click="appShell.setPrimary(item.id)"
      >
        <component :is="icons[item.id]" :size="18" aria-hidden="true" />
        <span>{{ item.label }}</span>
        <small>{{ item.caption }}</small>
      </button>
    </nav>

    <section class="sidebar-status" aria-label="引擎状态">
      <Bot :size="16" aria-hidden="true" />
      <div>
        <strong>{{ appShell.runtimePing.headline }}</strong>
        <span>{{ appShell.runtimePing.status === 'ready' ? '本地引擎在线' : '查看诊断' }}</span>
      </div>
    </section>
  </aside>
</template>
