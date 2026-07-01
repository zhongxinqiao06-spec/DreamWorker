<script setup lang="ts">
import { Command, RefreshCw, ShieldCheck } from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <header class="top-bar">
    <div>
      <p class="eyebrow">可配置、可扩展的 AI 工作台</p>
      <h1>{{ appShell.activeWorkspaceTitle }}</h1>
    </div>
    <div class="top-actions">
      <span class="health-pill" :data-status="appShell.runtimePing.status">
        <ShieldCheck :size="15" aria-hidden="true" />
        {{ appShell.runtimePing.headline }}
      </span>
      <button class="icon-text-button" type="button" @click="appShell.checkRuntimePing()">
        <RefreshCw :size="15" aria-hidden="true" />
        检查引擎
      </button>
      <button
        class="icon-button"
        type="button"
        title="打开命令面板"
        @click="appShell.toggleCommand()"
      >
        <Command :size="18" aria-hidden="true" />
      </button>
    </div>
  </header>
</template>
