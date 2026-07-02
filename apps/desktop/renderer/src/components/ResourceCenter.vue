<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Bot,
  Copy,
  DatabaseZap,
  Eye,
  EyeOff,
  KeyRound,
  Plus,
  PlugZap,
  RefreshCw,
  Save,
  Search,
  ShieldCheck,
  Trash2,
  Wrench
} from 'lucide-vue-next'
import {
  ALL_MODEL_ROUTE_SOURCE,
  isRoutedModelProvider,
  modelsForRouteSource,
  providerCapabilityOptions,
  providerTemplateOptions,
  providerTypeOptions,
  routeSourceForModel,
  routeSourceOptionsForModels,
  useAppShellStore,
  type ModelRouteSourceOption,
  type ProviderTemplateId
} from '../stores/app-shell'
import type {
  ProviderStatus,
  ProviderType,
  SafeModelProvider
} from '../../../shared/dreamworker-api'

const appShell = useAppShellStore()
const extensionPanelMode = ref<'config' | 'console'>('config')
const extensionConsoleRevision = ref(0)
const showProviderApiKey = ref(false)
const activeProviderRouteSource = ref(ALL_MODEL_ROUTE_SOURCE)

const providerLogoSrc: Record<ProviderType, string> = {
  deepseek: '/provider-icons/deepseek.svg',
  siliconflow: '/provider-icons/siliconflow.svg',
  glm: '/provider-icons/glm.png',
  openai: '/provider-icons/openai.svg',
  anthropic: '/provider-icons/anthropic.ico',
  openai_compatible: '/provider-icons/openai.svg',
  volcano: '/provider-icons/volcano.png',
  gemini: '/provider-icons/gemini.png',
  ollama: '/provider-icons/ollama.png',
  custom: '/provider-icons/openai.svg'
}

const filteredProviders = computed(() => {
  const keyword = appShell.providerSearch.trim().toLowerCase()
  if (!keyword) {
    return appShell.providers
  }
  return appShell.providers.filter((provider) =>
    [provider.displayName, provider.providerId, provider.defaultModel, provider.providerType]
      .join(' ')
      .toLowerCase()
      .includes(keyword)
  )
})

const activeMcpTools = computed(() =>
  appShell.tools.filter((tool) =>
    tool.toolId.startsWith(`mcp_${appShell.activeMcpServer?.serverId ?? ''}`)
  )
)
const providerModelLines = computed(() => splitLines(appShell.providerDraft.availableModelsText))
const providerDraftRouteTarget = computed(() => ({
  providerId: appShell.providerDraft.providerId,
  providerType: appShell.providerDraft.providerType,
  displayName: appShell.providerDraft.displayName,
  baseURL: appShell.providerDraft.baseURL
}))
const providerUsesRouteSources = computed(() =>
  isRoutedModelProvider(providerDraftRouteTarget.value)
)
const providerRouteSourceOptions = computed(() =>
  providerUsesRouteSources.value ? routeSourceOptionsForModels(providerModelLines.value) : []
)
const showProviderRouteSource = computed(
  () => providerUsesRouteSources.value && providerRouteSourceOptions.value.length > 1
)
const filteredProviderModels = computed(() =>
  showProviderRouteSource.value
    ? modelsForRouteSource(providerModelLines.value, activeProviderRouteSource.value)
    : providerModelLines.value
)

const extensionConsoleUrl = computed(() =>
  normalizeDashboardUrl(
    appShell.nineRouterStatus?.dashboardURL || appShell.settings.nineRouterDashboardURL
  )
)
const providerApiKeyPlaceholder = computed(
  () => appShell.activeProvider?.maskedKey ?? '留空保留当前密钥，保存后只显示脱敏值'
)

function providerInitial(providerName: string): string {
  return providerName.trim().slice(0, 1).toUpperCase() || 'AI'
}

function providerLogo(provider: SafeModelProvider): string {
  if (provider.providerId === 'provider_9router_local') {
    return '/provider-icons/9router.svg'
  }
  return providerLogoSrc[provider.providerType]
}

function statusText(status: ProviderStatus | string | undefined): string {
  if (status === 'connected') {
    return '已连接'
  }
  if (status === 'error') {
    return '异常'
  }
  return '未检查'
}

function extensionStateText(status: string | undefined): string {
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
  return status ? (labels[status] ?? status) : '未检查'
}

function runModeText(mode: string | undefined): string {
  return mode === 'managed' ? 'DreamWorker 受管' : '外部服务'
}

function isLoopbackHost(hostname: string): boolean {
  const host = hostname.replace(/^\[|\]$/g, '').toLowerCase()
  return host === 'localhost' || host === '0.0.0.0' || host === '::1' || host.startsWith('127.')
}

function normalizeDashboardUrl(value: string): string {
  const trimmed = value.trim() || 'http://localhost:20128'
  try {
    const url = new URL(trimmed)
    if (url.protocol === 'https:' && isLoopbackHost(url.hostname)) {
      url.protocol = 'http:'
    }
    if (url.pathname === '' || url.pathname === '/') {
      url.pathname = '/dashboard'
    }
    return url.toString()
  } catch {
    return trimmed
  }
}

function updateNineRouterRunMode(event: Event): void {
  const value = (event.target as HTMLSelectElement).value
  void appShell.updateNineRouterSettings({
    nineRouterRunMode: value === 'managed' ? 'managed' : 'external'
  })
}

function updateNineRouterText(
  field:
    | 'nineRouterBaseURL'
    | 'nineRouterDashboardURL'
    | 'nineRouterDefaultModel'
    | 'nineRouterManagedInstallVersion'
    | 'nineRouterManagedCommand',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).value
  })
}

function updateNineRouterBoolean(
  field: 'enableNineRouterIntegration' | 'allowAgentsUseNineRouter',
  event: Event
): void {
  void appShell.updateNineRouterSettings({
    [field]: (event.target as HTMLInputElement).checked
  })
}

function splitLines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

function routeOptionLabel(option: ModelRouteSourceOption): string {
  return `${option.label} (${option.modelCount})`
}

function syncProviderRouteSource(): void {
  if (!showProviderRouteSource.value) {
    activeProviderRouteSource.value = ALL_MODEL_ROUTE_SOURCE
    return
  }
  const optionIds = providerRouteSourceOptions.value.map((option) => option.id)
  const defaultSource = routeSourceForModel(appShell.providerDraft.defaultModel)
  if (optionIds.includes(defaultSource)) {
    activeProviderRouteSource.value = defaultSource
    return
  }
  if (!optionIds.includes(activeProviderRouteSource.value)) {
    activeProviderRouteSource.value = ALL_MODEL_ROUTE_SOURCE
  }
}

function setProviderRouteSource(event: Event): void {
  const source = (event.target as HTMLSelectElement).value
  activeProviderRouteSource.value = source
  if (source === ALL_MODEL_ROUTE_SOURCE) {
    return
  }
  const sourceModels = modelsForRouteSource(providerModelLines.value, source)
  if (sourceModels.length > 0 && !sourceModels.includes(appShell.providerDraft.defaultModel)) {
    appShell.providerDraft.defaultModel = sourceModels[0] ?? ''
  }
}

watch(
  () => [
    appShell.providerDraft.providerId,
    appShell.providerDraft.providerType,
    appShell.providerDraft.displayName,
    appShell.providerDraft.baseURL,
    appShell.providerDraft.availableModelsText,
    appShell.providerDraft.defaultModel
  ],
  syncProviderRouteSource,
  { immediate: true }
)

function setSkillRequiredCapabilities(event: Event): void {
  appShell.skillDraft = {
    ...appShell.skillDraft,
    requiredCapabilities: splitLines((event.target as HTMLTextAreaElement).value)
  }
}

function setSkillOutputArtifacts(event: Event): void {
  appShell.skillDraft = {
    ...appShell.skillDraft,
    outputArtifacts: splitLines((event.target as HTMLTextAreaElement).value)
  }
}

function addProvider(template: ProviderTemplateId): void {
  appShell.newProviderDraft(template)
}

async function copyProviderApiKey(): Promise<void> {
  const value = appShell.providerDraft.apiKey.trim()
  if (!value) {
    appShell.showResourceNotice('当前没有可复制的明文 API Key', 'info')
    return
  }
  try {
    await navigator.clipboard.writeText(value)
    appShell.showResourceNotice('API Key 已复制')
  } catch {
    appShell.showResourceNotice('复制失败，请手动选择复制', 'error')
  }
}
</script>

<template>
  <section class="resource-page panel-surface">
    <div
      v-if="appShell.resourceNotice"
      class="resource-toast"
      :data-tone="appShell.resourceNotice.tone"
      role="status"
    >
      {{ appShell.resourceNotice.message }}
    </div>

    <header class="resource-header resource-header-compact">
      <div>
        <p class="eyebrow">资源配置中心</p>
        <h2>模型、Agent、Skill、工具和 MCP</h2>
        <p>统一管理工作台资源。密钥只进入 Go Engine，前端仅展示脱敏值。</p>
      </div>
      <strong>{{ appShell.resourceSummary }}</strong>
    </header>

    <nav class="tab-strip" aria-label="资源类型">
      <button
        v-for="tab in appShell.resourceTabs"
        :key="tab.id"
        type="button"
        :class="{ active: tab.id === appShell.activeResourceTab }"
        @click="appShell.setResourceTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <section v-if="appShell.activeResourceTab === 'providers'" class="provider-console">
      <aside class="provider-rail" aria-label="模型服务商">
        <div class="rail-search">
          <Search :size="16" aria-hidden="true" />
          <input v-model="appShell.providerSearch" placeholder="搜索服务商、模型或类型" />
        </div>
        <div class="provider-template-row">
          <button
            v-for="template in providerTemplateOptions"
            :key="template.id"
            type="button"
            @click="addProvider(template.id)"
          >
            <Plus :size="14" aria-hidden="true" />
            {{ template.label }}
          </button>
        </div>
        <div class="provider-list">
          <button
            v-for="provider in filteredProviders"
            :key="provider.providerId"
            class="provider-row"
            :class="{ active: provider.providerId === appShell.activeProviderId }"
            type="button"
            @click="appShell.selectProvider(provider.providerId)"
          >
            <span class="provider-logo" aria-hidden="true">
              <img :src="providerLogo(provider)" alt="" />
              <b>{{ providerInitial(provider.displayName) }}</b>
            </span>
            <span class="provider-row-main">
              <strong>{{ provider.displayName }}</strong>
              <small>{{ provider.defaultModel }} / {{ provider.maskedKey ?? '未配置密钥' }}</small>
              <small
                >{{ provider.modelCount }} 个模型 /
                {{ provider.supportsStreaming ? '支持流式' : '不支持流式' }}</small
              >
            </span>
            <span class="status-dot" :data-status="provider.status">{{
              statusText(provider.status)
            }}</span>
          </button>
        </div>
      </aside>

      <form class="editor-card provider-editor" @submit.prevent="appShell.saveProviderDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <KeyRound :size="18" aria-hidden="true" />
            <span>模型服务商配置</span>
          </div>
          <div class="horizontal-actions">
            <button type="button" @click="appShell.newProviderDraft('deepseek')">
              <Plus :size="15" aria-hidden="true" />
              新增
            </button>
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button type="button" @click="appShell.testActiveProvider()">检查连接</button>
            <button type="button" @click="appShell.refreshActiveProviderModels()">
              <RefreshCw :size="15" aria-hidden="true" />
              刷新模型
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveProvider()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>

        <div class="provider-editor-body">
          <section class="provider-form-column" aria-label="基础配置">
            <div class="provider-status-strip">
              <span>{{ appShell.activeProviderStatus }}</span>
              <span>{{ appShell.activeProvider?.latencyMs ?? 0 }} ms</span>
              <span>{{ appShell.activeProvider?.lastErrorCode ?? '无错误' }}</span>
              <span>{{ appShell.providerActionStatus || '等待操作' }}</span>
            </div>

            <div class="form-grid two provider-form-grid">
              <label>
                服务商 ID
                <input v-model="appShell.providerDraft.providerId" />
              </label>
              <label>
                服务商类型
                <select v-model="appShell.providerDraft.providerType">
                  <option
                    v-for="option in providerTypeOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </label>
              <label>
                名称
                <input v-model="appShell.providerDraft.displayName" />
              </label>
              <label>
                API Host
                <input v-model="appShell.providerDraft.baseURL" />
              </label>
              <label>
                默认模型
                <div
                  class="route-model-selectors"
                  :class="{ 'single-select': !showProviderRouteSource }"
                >
                  <select
                    v-if="showProviderRouteSource"
                    :value="activeProviderRouteSource"
                    aria-label="上游服务商"
                    @change="setProviderRouteSource"
                  >
                    <option
                      v-for="source in providerRouteSourceOptions"
                      :key="source.id"
                      :value="source.id"
                    >
                      {{ routeOptionLabel(source) }}
                    </option>
                  </select>
                  <select v-model="appShell.providerDraft.defaultModel">
                    <option v-for="model in filteredProviderModels" :key="model" :value="model">
                      {{ model }}
                    </option>
                  </select>
                </div>
              </label>
              <label>
                API Key
                <div class="secret-input-shell">
                  <input
                    v-model="appShell.providerDraft.apiKey"
                    :type="showProviderApiKey ? 'text' : 'password'"
                    :placeholder="providerApiKeyPlaceholder"
                    autocomplete="off"
                    spellcheck="false"
                  />
                  <button
                    type="button"
                    :aria-label="showProviderApiKey ? '隐藏 API Key' : '显示 API Key'"
                    :title="showProviderApiKey ? '隐藏 API Key' : '显示 API Key'"
                    @click="showProviderApiKey = !showProviderApiKey"
                  >
                    <EyeOff v-if="showProviderApiKey" :size="17" aria-hidden="true" />
                    <Eye v-else :size="17" aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    aria-label="复制当前输入的 API Key"
                    title="复制当前输入的 API Key"
                    @click="copyProviderApiKey"
                  >
                    <Copy :size="17" aria-hidden="true" />
                  </button>
                </div>
                <small v-if="appShell.activeProvider?.hasApiKey" class="secret-field-note">
                  已保存到本地；留空保存会继续使用当前密钥
                  <span v-if="appShell.activeProvider.maskedKey">
                    {{ appShell.activeProvider.maskedKey }}
                  </span>
                </small>
              </label>
            </div>

            <label class="provider-model-textarea">
              <span class="field-label-row">
                <span>可用模型清单</span>
                <small>{{ providerModelLines.length }} 个模型</small>
              </span>
              <textarea
                v-model="appShell.providerDraft.availableModelsText"
                placeholder="每行一个模型 ID，例如 gpt-4.1 或 claude-sonnet-4-20250514"
                spellcheck="false"
              />
            </label>
          </section>

          <aside class="provider-model-column" aria-label="能力与模型">
            <div class="provider-column-heading">
              <div class="section-title">
                <DatabaseZap :size="16" aria-hidden="true" />
                <span>能力与模型</span>
              </div>
              <strong
                >{{ filteredProviderModels.length }} /
                {{ providerModelLines.length }} 个模型</strong
              >
            </div>

            <div class="capability-picker provider-capability-grid" aria-label="模型能力">
              <label
                v-for="capability in providerCapabilityOptions"
                :key="capability.value"
                class="check-row"
              >
                <input
                  v-model="appShell.providerDraft.capabilities"
                  type="checkbox"
                  :value="capability.value"
                />
                {{ capability.label }}
              </label>
              <label class="check-row">
                <input v-model="appShell.providerDraft.enabled" type="checkbox" />
                启用
              </label>
            </div>

            <dl class="provider-health-grid">
              <div>
                <dt>当前密钥</dt>
                <dd>{{ appShell.activeProvider?.maskedKey ?? '未配置' }}</dd>
              </div>
              <div>
                <dt>模型发现</dt>
                <dd>{{ appShell.activeProvider?.lastDiscoveryAt ?? '未刷新' }}</dd>
              </div>
              <div>
                <dt>流式验证</dt>
                <dd>{{ appShell.activeProvider?.streamingVerified ? '已验证' : '未验证' }}</dd>
              </div>
              <div>
                <dt>最近调用</dt>
                <dd>{{ appShell.activeProvider?.lastStreamAt ?? '暂无' }}</dd>
              </div>
            </dl>

            <section class="model-preview-list provider-model-list" aria-label="当前模型">
              <article v-for="model in filteredProviderModels" :key="model">
                <DatabaseZap :size="16" aria-hidden="true" />
                <span>{{ model }}</span>
              </article>
            </section>
          </aside>
        </div>
      </form>
    </section>

    <section v-else-if="appShell.activeResourceTab === 'extensions'" class="resource-grid">
      <aside class="resource-list" aria-label="拓展能力列表">
        <button
          v-for="extension in appShell.extensions"
          :key="extension.extensionId"
          class="list-row"
          :class="{ active: extension.extensionId === appShell.activeExtensionId }"
          type="button"
          @click="appShell.setActiveExtension(extension.extensionId)"
        >
          <strong>{{ extension.name }}</strong>
          <span>{{ extension.kind }} / {{ extension.runtimeKind }}</span>
          <span>{{ runModeText(appShell.extensionStatuses[extension.extensionId]?.runMode) }}</span>
        </button>
      </aside>

      <section
        class="editor-card extension-editor"
        :class="{ 'extension-editor-console': extensionPanelMode === 'console' }"
      >
        <div class="editor-toolbar">
          <div class="section-title">
            <PlugZap :size="18" aria-hidden="true" />
            <span>受管拓展能力</span>
          </div>
          <div class="extension-view-switch" aria-label="拓展视图切换">
            <button
              type="button"
              :class="{ active: extensionPanelMode === 'config' }"
              @click="extensionPanelMode = 'config'"
            >
              配置
            </button>
            <button
              type="button"
              :class="{ active: extensionPanelMode === 'console' }"
              @click="extensionPanelMode = 'console'"
            >
              控制台预览
            </button>
          </div>
          <div class="horizontal-actions">
            <button
              v-if="extensionPanelMode === 'console'"
              type="button"
              @click="extensionConsoleRevision += 1"
            >
              刷新预览
            </button>
            <button type="button" @click="appShell.detectActiveExtension()">检测环境</button>
            <button type="button" @click="appShell.installActiveExtension()">安装</button>
            <button type="button" @click="appShell.startActiveExtension()">启动</button>
            <button type="button" @click="appShell.stopActiveExtension()">停止</button>
            <button type="button" @click="appShell.restartActiveExtension()">重启</button>
          </div>
        </div>

        <template v-if="extensionPanelMode === 'config'">
          <section class="extension-log-panel" aria-label="9Router 安装与运行日志">
            <div class="extension-log-panel-header">
              <strong>安装与运行日志</strong>
              <span>{{ appShell.extensionActionStatus || '等待操作' }}</span>
            </div>
            <div class="extension-log-list">
              <article
                v-for="line in appShell.extensionLogs[appShell.activeExtensionId] ?? []"
                :key="`${line.timestamp}-${line.stream}-${line.line}`"
              >
                <small>{{ line.timestamp }} / {{ line.stream }}</small>
                <span>{{ line.line }}</span>
              </article>
              <p v-if="(appShell.extensionLogs[appShell.activeExtensionId] ?? []).length === 0">
                暂无日志。点击“检测环境”“安装”或“启动”后会在这里显示脱敏日志。
              </p>
            </div>
          </section>

          <div class="provider-editor-body">
            <section class="provider-form-column">
              <div class="provider-status-strip">
                <span>{{ runModeText(appShell.nineRouterStatus?.runMode) }}</span>
                <span>{{ extensionStateText(appShell.nineRouterStatus?.processState) }}</span>
                <span>{{ extensionStateText(appShell.nineRouterStatus?.healthStatus) }}</span>
                <span>{{ appShell.nineRouterStatus?.modelCount ?? 0 }} 个模型</span>
              </div>

              <div class="form-grid two provider-form-grid">
                <label>
                  运行模式
                  <select
                    :value="appShell.settings.nineRouterRunMode"
                    @change="updateNineRouterRunMode"
                  >
                    <option value="external">外部服务</option>
                    <option value="managed">DreamWorker 受管</option>
                  </select>
                </label>
                <label>
                  启用 9Router
                  <input
                    :checked="appShell.settings.enableNineRouterIntegration"
                    type="checkbox"
                    @change="updateNineRouterBoolean('enableNineRouterIntegration', $event)"
                  />
                </label>
                <label>
                  API Base URL
                  <input
                    :value="appShell.settings.nineRouterBaseURL"
                    @change="updateNineRouterText('nineRouterBaseURL', $event)"
                  />
                </label>
                <label>
                  Dashboard URL
                  <input
                    :value="appShell.settings.nineRouterDashboardURL"
                    @change="updateNineRouterText('nineRouterDashboardURL', $event)"
                  />
                </label>
                <label>
                  默认模型
                  <input
                    :value="appShell.settings.nineRouterDefaultModel"
                    @change="updateNineRouterText('nineRouterDefaultModel', $event)"
                  />
                </label>
                <label>
                  受管版本
                  <input
                    :value="appShell.settings.nineRouterManagedInstallVersion"
                    @change="updateNineRouterText('nineRouterManagedInstallVersion', $event)"
                  />
                </label>
                <label>
                  启动命令
                  <input
                    :value="appShell.settings.nineRouterManagedCommand"
                    @change="updateNineRouterText('nineRouterManagedCommand', $event)"
                  />
                </label>
                <label class="check-row">
                  <input
                    :checked="appShell.settings.allowAgentsUseNineRouter"
                    type="checkbox"
                    @change="updateNineRouterBoolean('allowAgentsUseNineRouter', $event)"
                  />
                  允许 Agent 和聊天使用
                </label>
              </div>

              <div class="horizontal-actions">
                <button type="button" @click="appShell.testActiveExtension()">测试连接</button>
                <button type="button" @click="appShell.refreshActiveExtensionModels()">
                  刷新模型
                </button>
                <button type="button" @click="appShell.verifyActiveExtensionStreaming()">
                  验证流式输出
                </button>
                <button type="button" @click="appShell.refreshExtensionLogs(undefined, true)">
                  查看日志
                </button>
                <button type="button" @click="appShell.clearActiveExtensionLogs()">清空日志</button>
              </div>

              <p class="form-hint">
                受管模式会使用本机 Node/npm 在 DreamWorker 拓展目录安装并启动 9Router。DreamWorker
                只管理自己启动的进程，并且只通过 OpenAI 兼容接口访问 9Router。
              </p>
              <p class="form-hint">{{ appShell.extensionActionStatus }}</p>

              <dl class="provider-health-grid">
                <div>
                  <dt>Node</dt>
                  <dd>{{ appShell.nineRouterStatus?.nodeVersion || '未检测' }}</dd>
                </div>
                <div>
                  <dt>npm</dt>
                  <dd>{{ appShell.nineRouterStatus?.npmVersion || '未检测' }}</dd>
                </div>
                <div>
                  <dt>命令</dt>
                  <dd>{{ appShell.nineRouterStatus?.command || '未检测' }}</dd>
                </div>
                <div>
                  <dt>最近错误</dt>
                  <dd>{{ appShell.nineRouterStatus?.lastErrorMessage || '无' }}</dd>
                </div>
              </dl>

              <section class="model-preview-list provider-model-list" aria-label="9Router 模型">
                <article v-for="model in appShell.nineRouterStatus?.models ?? []" :key="model">
                  <DatabaseZap :size="16" aria-hidden="true" />
                  <span>{{ model }}</span>
                </article>
              </section>
            </section>
          </div>
        </template>

        <section v-else class="extension-console-panel" aria-label="9Router Web 控制台预览">
          <div class="extension-console-frame">
            <iframe
              :key="`${extensionConsoleUrl}-${extensionConsoleRevision}`"
              :src="extensionConsoleUrl"
              title="9Router Web 控制台"
              sandbox="allow-forms allow-modals allow-popups allow-same-origin allow-scripts"
            />
          </div>
        </section>
      </section>
    </section>

    <section v-else-if="appShell.activeResourceTab === 'agents'" class="resource-grid">
      <aside class="resource-list" aria-label="Agent 列表">
        <button type="button" class="list-row create-row" @click="appShell.newAgentDraft()">
          <strong>新增 Agent</strong>
          <span>配置角色、模型、Skill、工具和 MCP</span>
        </button>
        <button
          v-for="agent in appShell.agents"
          :key="agent.agentId"
          class="list-row"
          :class="{ active: agent.agentId === appShell.activeAgentId }"
          type="button"
          @click="appShell.selectAgent(agent.agentId)"
        >
          <strong>{{ agent.displayName }}</strong>
          <span>{{ agent.role }}</span>
          <span>{{ agent.providerId }} / {{ agent.model }}</span>
        </button>
      </aside>
      <form class="editor-card" @submit.prevent="appShell.saveAgentDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <Bot :size="18" aria-hidden="true" />
            <span>Agent 配置</span>
          </div>
          <div class="horizontal-actions">
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button type="button" @click="appShell.duplicateActiveAgent()">
              <Copy :size="15" aria-hidden="true" />
              复制
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveAgent()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>
        <div class="form-grid two">
          <label>
            Agent ID
            <input v-model="appShell.agentDraft.agentId" />
          </label>
          <label>
            名称
            <input v-model="appShell.agentDraft.displayName" />
          </label>
          <label>
            角色
            <input v-model="appShell.agentDraft.role" />
          </label>
          <label>
            模型服务商
            <select
              :value="appShell.agentDraft.providerId"
              @change="appShell.setAgentDraftProvider(($event.target as HTMLSelectElement).value)"
            >
              <option
                v-for="provider in appShell.providers"
                :key="provider.providerId"
                :value="provider.providerId"
              >
                {{ provider.displayName }}
              </option>
            </select>
          </label>
          <label>
            模型
            <select
              :value="appShell.agentDraft.model"
              @change="appShell.setAgentDraftModel(($event.target as HTMLSelectElement).value)"
            >
              <option
                v-for="model in appShell.modelsForProvider(appShell.agentDraft.providerId)"
                :key="model"
                :value="model"
              >
                {{ model }}
              </option>
            </select>
          </label>
          <label>
            规划策略
            <select v-model="appShell.agentDraft.planner.strategy">
              <option value="plan-execute">计划后执行</option>
              <option value="react">边想边做</option>
              <option value="manual">手动推进</option>
            </select>
          </label>
          <label>
            记忆范围
            <select v-model="appShell.agentDraft.memoryScope">
              <option value="short_term">短期</option>
              <option value="project">项目</option>
              <option value="semantic">语义</option>
            </select>
          </label>
          <label>
            上下文窗口
            <input
              v-model.number="appShell.agentDraft.runtimeConfig.contextWindow"
              type="number"
              min="1024"
            />
          </label>
          <label>
            最大输出
            <input
              v-model.number="appShell.agentDraft.runtimeConfig.maxTokens"
              type="number"
              min="1"
            />
          </label>
        </div>
        <label>
          描述
          <textarea v-model="appShell.agentDraft.description" />
        </label>
        <label>
          系统提示词
          <textarea v-model="appShell.agentDraft.systemPrompt" />
        </label>
        <div class="resource-selector-grid">
          <section>
            <strong>Skill</strong>
            <label v-for="skill in appShell.skills" :key="skill.skillId" class="check-row">
              <input
                v-model="appShell.agentDraft.enabledSkills"
                type="checkbox"
                :value="skill.skillId"
              />
              {{ skill.displayName }}
            </label>
          </section>
          <section>
            <strong>工具</strong>
            <label v-for="tool in appShell.tools" :key="tool.toolId" class="check-row">
              <input
                v-model="appShell.agentDraft.enabledTools"
                type="checkbox"
                :value="tool.toolId"
              />
              {{ tool.displayName }}
            </label>
          </section>
          <section>
            <strong>MCP</strong>
            <label v-for="server in appShell.mcpServers" :key="server.serverId" class="check-row">
              <input
                v-model="appShell.agentDraft.enabledMcpServers"
                type="checkbox"
                :value="server.serverId"
              />
              {{ server.displayName }}
            </label>
          </section>
        </div>
        <label class="check-row">
          <input v-model="appShell.agentDraft.enabled" type="checkbox" />
          启用
        </label>
      </form>
    </section>

    <section v-else-if="appShell.activeResourceTab === 'skills'" class="resource-grid">
      <aside class="resource-list" aria-label="Skill 列表">
        <button type="button" class="list-row create-row" @click="appShell.newSkillDraft()">
          <strong>新增 Skill</strong>
          <span>沉淀可复用的任务能力</span>
        </button>
        <button
          v-for="skill in appShell.skills"
          :key="skill.skillId"
          class="list-row"
          :class="{ active: skill.skillId === appShell.activeSkillId }"
          type="button"
          @click="appShell.selectSkill(skill.skillId)"
        >
          <strong>{{ skill.displayName }}</strong>
          <span>{{ skill.category }} / {{ skill.version }}</span>
          <span>{{ skill.enabled ? '已启用' : '未启用' }}</span>
        </button>
      </aside>
      <form class="editor-card" @submit.prevent="appShell.saveSkillDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <ShieldCheck :size="18" aria-hidden="true" />
            <span>Skill 配置</span>
          </div>
          <div class="horizontal-actions">
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveSkill()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>
        <div class="form-grid two">
          <label>
            Skill ID
            <input v-model="appShell.skillDraft.skillId" />
          </label>
          <label>
            命令名
            <input v-model="appShell.skillDraft.commandName" />
          </label>
          <label>
            名称
            <input v-model="appShell.skillDraft.displayName" />
          </label>
          <label>
            分类
            <select v-model="appShell.skillDraft.category">
              <option value="general">通用</option>
              <option value="explore">探索</option>
              <option value="product">产品</option>
              <option value="development">开发</option>
              <option value="sales">销售</option>
            </select>
          </label>
          <label>
            版本
            <input v-model="appShell.skillDraft.version" />
          </label>
          <label>
            来源路径
            <input v-model="appShell.skillDraft.sourcePath" />
          </label>
        </div>
        <label>
          描述
          <textarea v-model="appShell.skillDraft.description" />
        </label>
        <label>
          使用时机
          <textarea v-model="appShell.skillDraft.whenToUse" />
        </label>
        <label>
          指令
          <textarea v-model="appShell.skillDraft.instructions" />
        </label>
        <div class="form-grid two">
          <label>
            所需能力，每行一个
            <textarea
              :value="appShell.skillDraft.requiredCapabilities.join('\n')"
              @input="setSkillRequiredCapabilities"
            />
          </label>
          <label>
            输出产物，每行一个
            <textarea
              :value="appShell.skillDraft.outputArtifacts.join('\n')"
              @input="setSkillOutputArtifacts"
            />
          </label>
        </div>
        <label class="check-row">
          <input v-model="appShell.skillDraft.enabled" type="checkbox" />
          启用
        </label>
      </form>
    </section>

    <section v-else-if="appShell.activeResourceTab === 'tools'" class="resource-grid">
      <aside class="resource-list" aria-label="工具列表">
        <button type="button" class="list-row create-row" @click="appShell.newToolDraft()">
          <strong>新增工具</strong>
          <span>定义可执行能力和风险等级</span>
        </button>
        <button
          v-for="tool in appShell.tools"
          :key="tool.toolId"
          class="list-row"
          :class="{ active: tool.toolId === appShell.activeToolId }"
          type="button"
          @click="appShell.selectTool(tool.toolId)"
        >
          <strong>{{ tool.displayName }}</strong>
          <span>{{ tool.category }} / 风险 {{ tool.riskLevel }}</span>
          <span>{{ tool.enabled ? '已启用' : '未启用' }}</span>
        </button>
      </aside>
      <form class="editor-card" @submit.prevent="appShell.saveToolDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <Wrench :size="18" aria-hidden="true" />
            <span>工具配置</span>
          </div>
          <div class="horizontal-actions">
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveTool()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>
        <div class="form-grid two">
          <label>
            工具 ID
            <input v-model="appShell.toolDraft.toolId" />
          </label>
          <label>
            名称
            <input v-model="appShell.toolDraft.displayName" />
          </label>
          <label>
            分类
            <select v-model="appShell.toolDraft.category">
              <option value="artifact">产物</option>
              <option value="browser">浏览器</option>
              <option value="search">搜索</option>
              <option value="model">模型</option>
              <option value="human">人工</option>
              <option value="project">项目</option>
            </select>
          </label>
          <label>
            风险等级
            <select v-model="appShell.toolDraft.riskLevel">
              <option value="low">低</option>
              <option value="medium">中</option>
              <option value="high">高</option>
              <option value="critical">关键</option>
            </select>
          </label>
        </div>
        <label>
          描述
          <textarea v-model="appShell.toolDraft.description" />
        </label>
        <label class="check-row">
          <input v-model="appShell.toolDraft.enabled" type="checkbox" />
          启用
        </label>
      </form>
    </section>

    <section v-else class="resource-grid">
      <aside class="resource-list" aria-label="MCP 服务列表">
        <button type="button" class="list-row create-row" @click="appShell.newMcpDraft()">
          <strong>新增 MCP</strong>
          <span>接入本地或远程工具服务</span>
        </button>
        <button
          v-for="server in appShell.mcpServers"
          :key="server.serverId"
          class="list-row"
          :class="{ active: server.serverId === appShell.activeMcpServerId }"
          type="button"
          @click="appShell.selectMcpServer(server.serverId)"
        >
          <strong>{{ server.displayName }}</strong>
          <span>{{ server.trustLevel }} / {{ server.enabled ? '已启用' : '未启用' }}</span>
          <span>{{ server.hasSecrets ? '已配置密钥' : '无密钥' }}</span>
        </button>
      </aside>
      <form class="editor-card" @submit.prevent="appShell.saveMcpDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <PlugZap :size="18" aria-hidden="true" />
            <span>MCP 服务</span>
          </div>
          <div class="horizontal-actions">
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button type="button" @click="appShell.testActiveMcpServer()">检查</button>
            <button type="button" @click="appShell.refreshActiveMcpTools()">
              <RefreshCw :size="15" aria-hidden="true" />
              刷新工具
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveMcpServer()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>
        <div class="form-grid two">
          <label>
            服务 ID
            <input v-model="appShell.mcpDraft.serverId" />
          </label>
          <label>
            名称
            <input v-model="appShell.mcpDraft.displayName" />
          </label>
          <label>
            命令
            <input v-model="appShell.mcpDraft.command" />
          </label>
          <label>
            URL
            <input
              v-model="appShell.mcpDraft.url"
              placeholder="远程服务可填写 URL，本地命令可留空"
            />
          </label>
          <label>
            信任等级
            <select v-model="appShell.mcpDraft.trustLevel">
              <option value="trusted_builtin">内置可信</option>
              <option value="verified_partner">已验证伙伴</option>
              <option value="community">社区</option>
              <option value="local_unverified">本地未验证</option>
              <option value="remote_untrusted">远程未信任</option>
            </select>
          </label>
        </div>
        <div class="form-grid two">
          <label>
            参数，每行一个
            <textarea v-model="appShell.mcpDraft.argsText" />
          </label>
          <label>
            密钥，每行 KEY=VALUE
            <textarea v-model="appShell.mcpDraft.secretsText" placeholder="保存后只显示脱敏值" />
          </label>
        </div>
        <label class="check-row">
          <input v-model="appShell.mcpDraft.enabled" type="checkbox" />
          启用
        </label>
        <section v-if="activeMcpTools.length > 0" class="model-preview-list" aria-label="MCP 工具">
          <article v-for="tool in activeMcpTools" :key="tool.toolId">
            <Wrench :size="16" aria-hidden="true" />
            <span>{{ tool.displayName }} / {{ tool.riskLevel }}</span>
          </article>
        </section>
      </form>
    </section>
  </section>
</template>
