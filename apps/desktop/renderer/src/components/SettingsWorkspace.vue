<script setup lang="ts">
import {
  CheckCircle2,
  Code2,
  LockKeyhole,
  MonitorCog,
  PlugZap,
  RefreshCw,
  Shield,
  XCircle
} from 'lucide-vue-next'
import { computed, onMounted, ref } from 'vue'
import type {
  CodingRuntimeEngineStatus,
  CodingRuntimeStatus
} from '../../../shared/dreamworker-api'
import { useAppShellStore } from '../stores/app-shell'

const appShell = useAppShellStore()
const codingRuntime = ref<CodingRuntimeStatus | null>(null)
const codingRuntimeLoading = ref(false)
const codingRuntimeError = ref('')

const codingEngineStatuses = computed<CodingRuntimeEngineStatus[]>(() => {
  const statuses = codingRuntime.value?.engineStatuses ?? []
  if (statuses.length > 0) {
    return [...statuses]
  }
  return [
    {
      engineId: 'claude_agent',
      packageName: '@anthropic-ai/claude-agent-sdk',
      installed: false,
      executable: false,
      status: 'missing',
      message: '等待检测'
    },
    {
      engineId: 'codex',
      packageName: '@openai/codex-sdk',
      installed: false,
      executable: false,
      status: 'missing',
      message: '等待检测'
    },
    {
      engineId: 'opencode',
      packageName: '@opencode-ai/sdk',
      installed: false,
      executable: false,
      status: 'missing',
      message: '等待检测'
    }
  ]
})

const codingRuntimeMessage = computed(() => {
  if (codingRuntime.value?.message) {
    return codingRuntime.value.message
  }
  return codingRuntimeError.value || '等待检测'
})

onMounted(() => {
  void refreshCodingRuntime()
})

async function refreshCodingRuntime(): Promise<void> {
  codingRuntimeLoading.value = true
  codingRuntimeError.value = ''
  try {
    codingRuntime.value = await window.dreamworker.coding.listEngines()
  } catch (error) {
    codingRuntimeError.value =
      error instanceof Error ? error.message : '编码 Agent runtime 检测失败。'
  } finally {
    codingRuntimeLoading.value = false
  }
}

function updateRunMode(event: Event): void {
  const value = (event.target as HTMLSelectElement).value
  void appShell.updateNineRouterSettings({
    nineRouterRunMode: value === 'managed' ? 'managed' : 'external'
  })
}

function updateText(
  field: 'nineRouterBaseURL' | 'nineRouterDashboardURL' | 'nineRouterDefaultModel',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).value
  })
}

function updateBoolean(
  field: 'enableNineRouterIntegration' | 'allowAgentsUseNineRouter',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).checked
  })
}

function codingEngineLabel(engineId: CodingRuntimeEngineStatus['engineId']): string {
  if (engineId === 'claude_agent') {
    return 'Claude Agent'
  }
  if (engineId === 'codex') {
    return 'Codex'
  }
  return 'OpenCode'
}

function statusText(status: string): string {
  const labels: Record<string, string> = {
    missing: '未安装',
    checking: '检测中',
    ready: '待命',
    running: '运行中',
    error: '异常'
  }
  return labels[status] ?? status
}
</script>

<template>
  <section class="settings-grid">
    <article class="panel-surface settings-card">
      <MonitorCog :size="22" aria-hidden="true" />
      <h2>桌面工作台偏好</h2>
      <p>
        设置区保留本地优先、中文界面和资源隔离原则。持久化后，所有配置仍通过 Main 转发到 Node
        Engine。
      </p>
      <label class="check-row">
        <input type="checkbox" checked disabled />
        界面文字使用中文
      </label>
      <label class="check-row">
        <input type="checkbox" checked disabled />
        Renderer 不保存密钥
      </label>
    </article>

    <article class="panel-surface settings-card">
      <Shield :size="22" aria-hidden="true" />
      <h2>安全边界</h2>
      <p>
        Renderer 只通过 typed preload API 访问能力；Main 管理桌面生命周期、本地代理和内嵌 Runtime；
        负责资源、策略和运行时。
      </p>
      <ul>
        <li>不直接访问 Node、文件系统、SQLite 或 secret。</li>
        <li>外部链接由 Main 安全打开。</li>
        <li>高风险能力必须进入 PolicyEngine 和 Approval。</li>
      </ul>
    </article>

    <article class="panel-surface settings-card">
      <LockKeyhole :size="22" aria-hidden="true" />
      <h2>密钥策略</h2>
      <p>
        API Key、MCP Secret 和远程连接令牌只允许保存在 Engine 侧。界面最多展示 hasApiKey、maskedKey
        和 maskedSecrets。
      </p>
    </article>

    <article class="panel-surface settings-card coding-runtime-settings-card">
      <div class="settings-card-heading">
        <Code2 :size="22" aria-hidden="true" />
        <div>
          <h2>本地编码 Agent SDK</h2>
          <p>安装包内置 Claude Agent、Codex 和 OpenCode SDK，运行时不执行 npm install。</p>
        </div>
        <button type="button" title="刷新编码 Agent 状态" @click="refreshCodingRuntime">
          <RefreshCw :size="15" aria-hidden="true" />
        </button>
      </div>

      <div class="coding-runtime-summary">
        <span :class="{ ready: codingRuntime?.available }">
          {{ codingRuntime?.available ? 'runtime ready' : 'runtime missing' }}
        </span>
        <small>{{ codingRuntimeMessage }}</small>
      </div>

      <div class="coding-runtime-paths">
        <span>Node</span>
        <strong>{{ codingRuntime?.nodeBin || 'node' }}</strong>
        <span>Adapter</span>
        <strong>{{ codingRuntime?.adapterPath || '未找到' }}</strong>
      </div>

      <div class="coding-engine-status-list">
        <article v-for="engine in codingEngineStatuses" :key="engine.engineId">
          <div>
            <strong>{{ codingEngineLabel(engine.engineId) }}</strong>
            <small>{{ engine.packageName }}</small>
          </div>
          <span :class="engine.status">
            <CheckCircle2
              v-if="engine.installed && engine.executable"
              :size="14"
              aria-hidden="true"
            />
            <XCircle v-else :size="14" aria-hidden="true" />
            {{ statusText(engine.status) }}
          </span>
          <p>{{ engine.message }}</p>
        </article>
      </div>

      <small v-if="codingRuntimeLoading">正在检测本地 SDK runtime...</small>
    </article>

    <article class="panel-surface settings-card settings-extension-card">
      <PlugZap :size="22" aria-hidden="true" />
      <h2>9Router 拓展能力</h2>
      <p>
        9Router 可以作为 DreamWorker 的 OpenAI
        兼容上游模型路由。外部服务模式只连接本机服务，受管模式由 Main Runtime
        管理安装、启动、停止、健康检查和日志。
      </p>
      <label>
        运行模式
        <select :value="appShell.settings.nineRouterRunMode" @change="updateRunMode">
          <option value="external">外部服务</option>
          <option value="managed">DreamWorker 受管</option>
        </select>
      </label>
      <label>
        Base URL
        <input
          :value="appShell.settings.nineRouterBaseURL"
          @change="updateText('nineRouterBaseURL', $event)"
        />
      </label>
      <label>
        Dashboard URL
        <input
          :value="appShell.settings.nineRouterDashboardURL"
          @change="updateText('nineRouterDashboardURL', $event)"
        />
      </label>
      <label>
        默认模型
        <input
          :value="appShell.settings.nineRouterDefaultModel"
          @change="updateText('nineRouterDefaultModel', $event)"
        />
      </label>
      <label class="check-row">
        <input
          :checked="appShell.settings.enableNineRouterIntegration"
          type="checkbox"
          @change="updateBoolean('enableNineRouterIntegration', $event)"
        />
        启用 9Router 集成
      </label>
      <label class="check-row">
        <input
          :checked="appShell.settings.allowAgentsUseNineRouter"
          type="checkbox"
          @change="updateBoolean('allowAgentsUseNineRouter', $event)"
        />
        允许 Agent 和聊天使用
      </label>
      <button type="button" @click="appShell.resetNineRouterSettings()">恢复默认配置</button>
    </article>
  </section>
</template>

<style scoped>
.settings-card-heading,
.coding-runtime-summary,
.coding-engine-status-list article,
.coding-engine-status-list span {
  display: flex;
  align-items: center;
}

.settings-card-heading {
  gap: 12px;
}

.settings-card-heading > svg {
  flex: 0 0 auto;
  color: var(--purple);
}

.settings-card-heading h2,
.settings-card-heading p {
  margin: 0;
}

.settings-card-heading button {
  display: grid;
  width: 32px;
  height: 32px;
  margin-left: auto;
  place-items: center;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(248, 250, 252, 0.76);
  color: var(--muted);
  cursor: pointer;
}

.coding-runtime-settings-card {
  grid-column: span 2;
  gap: 12px;
}

.coding-runtime-summary {
  justify-content: space-between;
  gap: 12px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  background: rgba(248, 250, 252, 0.72);
  padding: 10px 12px;
}

.coding-runtime-summary span,
.coding-engine-status-list span {
  gap: 6px;
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 999px;
  background: rgba(254, 242, 242, 0.76);
  padding: 4px 9px;
  color: var(--danger);
  font-size: 12px;
  font-weight: 800;
}

.coding-runtime-summary span.ready,
.coding-engine-status-list span.ready,
.coding-engine-status-list span.running {
  border-color: rgba(16, 185, 129, 0.2);
  background: rgba(236, 253, 245, 0.82);
  color: var(--green);
}

.coding-engine-status-list span.checking {
  border-color: rgba(245, 158, 11, 0.22);
  background: rgba(255, 251, 235, 0.84);
  color: var(--amber);
}

.coding-runtime-summary small {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
}

.coding-runtime-paths {
  display: grid;
  grid-template-columns: 70px minmax(0, 1fr);
  gap: 7px 10px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.18);
  padding-bottom: 12px;
  font-size: 12px;
}

.coding-runtime-paths span {
  color: var(--subtle);
  font-weight: 800;
}

.coding-runtime-paths strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
  font-weight: 700;
}

.coding-engine-status-list {
  display: grid;
  gap: 8px;
}

.coding-engine-status-list article {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 4px 10px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.62);
  padding: 10px;
}

.coding-engine-status-list div {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.coding-engine-status-list strong,
.coding-engine-status-list small,
.coding-engine-status-list p {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coding-engine-status-list small {
  color: var(--subtle);
  font-size: 12px;
}

.coding-engine-status-list p {
  grid-column: 1 / -1;
  margin: 0;
  color: var(--muted);
  font-size: 12px;
}
</style>
