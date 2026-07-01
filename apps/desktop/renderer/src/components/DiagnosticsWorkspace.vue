<script setup lang="ts">
import { Activity, Bug, RefreshCw } from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()
</script>

<template>
  <section class="diagnostics-page">
    <article class="panel-surface diagnostics-card">
      <div class="section-title">
        <Activity :size="20" aria-hidden="true" />
        <span>运行时检查</span>
      </div>
      <h2>{{ appShell.runtimePing.headline }}</h2>
      <p>{{ appShell.runtimePing.detail }}</p>
      <dl>
        <div>
          <dt>接口</dt>
          <dd>runtime.ping</dd>
        </div>
        <div>
          <dt>trace_id</dt>
          <dd>{{ appShell.runtimePing.traceId }}</dd>
        </div>
        <div>
          <dt>错误码</dt>
          <dd>{{ appShell.runtimePing.errorCode }}</dd>
        </div>
      </dl>
      <button class="primary-button" type="button" @click="appShell.checkRuntimePing()">
        <RefreshCw :size="15" aria-hidden="true" />
        重新检查
      </button>
    </article>

    <article class="panel-surface diagnostics-card">
      <div class="section-title">
        <Bug :size="20" aria-hidden="true" />
        <span>接口覆盖</span>
      </div>
      <p>
        当前桌面端已经通过 typed preload API 暴露
        runtime、models、agents、skills、tools、mcp、projects 和 chat 命名空间。
      </p>
      <dl>
        <div>
          <dt>资源数量</dt>
          <dd>{{ appShell.resourceSummary }}</dd>
        </div>
        <div>
          <dt>项目数量</dt>
          <dd>{{ appShell.projects.length }}</dd>
        </div>
        <div>
          <dt>对话数量</dt>
          <dd>{{ appShell.chatSessions.length }}</dd>
        </div>
      </dl>
    </article>
  </section>
</template>
