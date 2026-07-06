<script setup lang="ts">
import { Activity, Bug, PlugZap, RefreshCw } from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()

function stateText(value: string | undefined): string {
  const labels: Record<string, string> = {
    unknown: '未检查',
    connected: '已连接',
    disconnected: '未连接',
    error: '异常',
    running: '运行中',
    stopped: '已停止',
    starting: '启动中',
    failed: '失败'
  }
  return value ? (labels[value] ?? value) : '未检查'
}
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
        <PlugZap :size="20" aria-hidden="true" />
        <span>9Router 拓展诊断</span>
      </div>
      <p>
        这里展示 Main Runtime 管理的拓展状态。诊断摘要只展示脱敏信息，不包含 API Key、Runtime token
        或完整环境变量。
      </p>
      <dl>
        <div>
          <dt>运行模式</dt>
          <dd>{{ appShell.nineRouterStatus?.runMode ?? 'external' }}</dd>
        </div>
        <div>
          <dt>进程状态</dt>
          <dd>{{ stateText(appShell.nineRouterStatus?.processState) }}</dd>
        </div>
        <div>
          <dt>健康状态</dt>
          <dd>{{ stateText(appShell.nineRouterStatus?.healthStatus) }}</dd>
        </div>
        <div>
          <dt>PID</dt>
          <dd>{{ appShell.nineRouterStatus?.pid ?? '无' }}</dd>
        </div>
        <div>
          <dt>Node</dt>
          <dd>{{ appShell.nineRouterStatus?.nodeVersion || '未检测' }}</dd>
        </div>
        <div>
          <dt>npm</dt>
          <dd>{{ appShell.nineRouterStatus?.npmVersion || '未检测' }}</dd>
        </div>
        <div>
          <dt>模型数量</dt>
          <dd>{{ appShell.nineRouterStatus?.modelCount ?? 0 }}</dd>
        </div>
        <div>
          <dt>最近错误</dt>
          <dd>{{ appShell.nineRouterStatus?.lastErrorMessage || '无' }}</dd>
        </div>
      </dl>
      <div class="horizontal-actions">
        <button type="button" @click="appShell.detectActiveExtension()">检测环境</button>
        <button type="button" @click="appShell.testActiveExtension()">测试连接</button>
        <button type="button" @click="appShell.refreshExtensionLogs()">刷新日志</button>
      </div>
    </article>

    <article class="panel-surface diagnostics-card">
      <div class="section-title">
        <Bug :size="20" aria-hidden="true" />
        <span>接口覆盖</span>
      </div>
      <p>
        当前桌面端通过 typed preload API 暴露
        runtime、settings、extensions、models、agents、skills、tools、mcp、projects 和 chat
        命名空间。
      </p>
      <dl>
        <div>
          <dt>资源数量</dt>
          <dd>{{ appShell.resourceSummary }}</dd>
        </div>
        <div>
          <dt>拓展数量</dt>
          <dd>{{ appShell.extensions.length }}</dd>
        </div>
        <div>
          <dt>对话数量</dt>
          <dd>{{ appShell.chatSessions.length }}</dd>
        </div>
      </dl>
    </article>
  </section>
</template>
