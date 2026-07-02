<script setup lang="ts">
import {
  Bell,
  Command,
  HelpCircle,
  RefreshCw,
  Search,
  ShieldCheck,
  UserCircle
} from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <header class="top-bar">
    <div class="top-title">
      <p class="eyebrow">可配置、可扩展的 AI 工作台</p>
      <h1>{{ appShell.activeWorkspaceTitle }}</h1>
    </div>
    <div class="top-center">
      <button
        class="global-search"
        type="button"
        title="打开命令面板"
        @click="appShell.toggleCommand()"
      >
        <Search :size="16" aria-hidden="true" />
        <span>搜索资源、任务、知识库...</span>
        <kbd>Ctrl K</kbd>
      </button>
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
        title="刷新通知状态"
        @click="appShell.checkRuntimePing()"
      >
        <Bell :size="18" aria-hidden="true" />
      </button>
      <button
        class="icon-button"
        type="button"
        title="查看设置"
        @click="appShell.setPrimary('settings')"
      >
        <HelpCircle :size="18" aria-hidden="true" />
      </button>
      <button
        class="icon-button"
        type="button"
        title="打开命令面板"
        @click="appShell.toggleCommand()"
      >
        <Command :size="18" aria-hidden="true" />
      </button>
      <button
        class="icon-button avatar-button"
        type="button"
        title="打开设置"
        @click="appShell.setPrimary('settings')"
      >
        <UserCircle :size="22" aria-hidden="true" />
      </button>
    </div>
  </header>
</template>
