<script setup lang="ts">
import {
  AlertTriangle,
  ArrowLeft,
  Bot,
  CheckCircle2,
  FileText,
  FolderTree,
  GitBranch,
  Loader2,
  RefreshCw,
  Send,
  Settings,
  Square,
  Terminal
} from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import type {
  CodingEngineDescriptor,
  CodingEngineId,
  CodingFileChange,
  CodingFileEntry,
  CodingReadFileResult,
  CodingRuntimeStatus,
  CodingSession,
  CodingStreamController,
  CodingStreamEvent,
  ProviderType,
  SafeModelProvider
} from '../../../../shared/dreamworker-api'
import { useAppShellStore } from '../../stores/app-shell'
import { isRoutedModelProvider } from '../../stores/app-shell'

type CodingMessage = {
  readonly id: string
  readonly role: 'user' | 'assistant' | 'system'
  content: string
  readonly createdAt: string
  status: 'streaming' | 'completed' | 'error'
}

type CommandLog = {
  readonly id: string
  readonly command: string
  readonly output: string
  readonly createdAt: string
}

const appShell = useAppShellStore()
const runtime = ref<CodingRuntimeStatus | null>(null)
const engineId = ref<CodingEngineId>('claude_agent')
const providerId = ref('')
const model = ref('')
const session = ref<CodingSession | null>(null)
const files = ref<CodingFileEntry[]>([])
const fileQuery = ref('')
const selectedFile = ref<CodingReadFileResult | null>(null)
const selectedPath = ref('')
const fileStatus = ref<{
  branch: string
  changes: readonly CodingFileChange[]
  clean: boolean
  message: string
} | null>(null)
const messages = ref<CodingMessage[]>([])
const draft = ref('')
const streaming = ref(false)
const streamError = ref('')
const commandLogs = ref<CommandLog[]>([])
const changedFiles = ref<CodingFileChange[]>([])
const controller = ref<CodingStreamController | null>(null)
const assistantMessageId = ref('')

const activeProject = computed(() => appShell.activeProject)
const hasProjectRoot = computed(() => Boolean(activeProject.value?.localRootPath))
const engines = computed(() => runtime.value?.engines ?? [])
const selectedEngine = computed<CodingEngineDescriptor | undefined>(
  () => engines.value.find((engine) => engine.engineId === engineId.value) ?? engines.value[0]
)
const providerOptions = computed(() =>
  appShell.providers.map((provider) => ({
    provider,
    ...providerSupport(provider, selectedEngine.value)
  }))
)
const compatibleProviders = computed(() =>
  providerOptions.value.filter((option) => option.supported).map((option) => option.provider)
)
const selectedProvider = computed(
  () =>
    appShell.providers.find((provider) => provider.providerId === providerId.value) ??
    compatibleProviders.value[0]
)
const availableModels = computed(() => {
  const provider = selectedProvider.value
  if (!provider) {
    return []
  }
  return provider.availableModels.length > 0 ? provider.availableModels : [provider.defaultModel]
})
const runtimeReady = computed(() => runtime.value?.available === true)
const canSend = computed(
  () =>
    runtimeReady.value &&
    hasProjectRoot.value &&
    Boolean(selectedProvider.value) &&
    Boolean(model.value) &&
    draft.value.trim().length > 0 &&
    !streaming.value
)

onMounted(() => {
  void refreshRuntime()
})

watch(
  () => [activeProject.value?.projectId, activeProject.value?.localRootPath] as const,
  () => {
    session.value = null
    selectedFile.value = null
    selectedPath.value = ''
    changedFiles.value = []
    if (hasProjectRoot.value) {
      void refreshFiles()
      void refreshStatus()
    }
  },
  { immediate: true }
)

watch(
  () => [engineId.value, appShell.providers.length, runtime.value?.available] as const,
  () => ensureProviderSelection(),
  { immediate: true }
)

watch(
  () => providerId.value,
  () => ensureModelSelection(),
  { immediate: true }
)

async function refreshRuntime(): Promise<void> {
  runtime.value = await window.dreamworker.coding.listEngines()
  const firstEngine = runtime.value.engines[0]
  if (firstEngine && !runtime.value.engines.some((engine) => engine.engineId === engineId.value)) {
    engineId.value = firstEngine.engineId
  }
  ensureProviderSelection()
}

async function refreshFiles(): Promise<void> {
  const projectId = activeProject.value?.projectId
  if (!projectId || !hasProjectRoot.value) {
    files.value = []
    return
  }
  files.value = [
    ...(await window.dreamworker.coding.listFiles({
      projectId,
      query: fileQuery.value,
      limit: 600
    }))
  ]
}

async function refreshStatus(): Promise<void> {
  const projectId = activeProject.value?.projectId
  if (!projectId || !hasProjectRoot.value) {
    fileStatus.value = null
    return
  }
  const status = await window.dreamworker.coding.fileStatus(projectId)
  fileStatus.value = status
  changedFiles.value = [...status.changes]
}

async function selectFile(path: string): Promise<void> {
  const projectId = activeProject.value?.projectId
  if (!projectId || !path) {
    return
  }
  selectedPath.value = path
  selectedFile.value = await window.dreamworker.coding.readFile({ projectId, path })
}

async function sendPrompt(): Promise<void> {
  const project = activeProject.value
  const provider = selectedProvider.value
  const prompt = draft.value.trim()
  if (!project || !provider || !prompt || !model.value) {
    return
  }
  streamError.value = ''
  const activeSession = await ensureSession(project.projectId, provider.providerId)
  const streamId = `coding_${Date.now()}_${Math.random().toString(36).slice(2)}`
  const now = new Date().toISOString()
  assistantMessageId.value = `assistant_${streamId}`
  messages.value = [
    ...messages.value,
    {
      id: `user_${streamId}`,
      role: 'user',
      content: prompt,
      createdAt: now,
      status: 'completed'
    },
    {
      id: assistantMessageId.value,
      role: 'assistant',
      content: '',
      createdAt: now,
      status: 'streaming'
    }
  ]
  draft.value = ''
  streaming.value = true
  controller.value = await window.dreamworker.coding.streamTurn(
    {
      sessionId: activeSession.sessionId,
      projectId: project.projectId,
      engineId: engineId.value,
      providerId: provider.providerId,
      model: model.value,
      prompt,
      streamId
    },
    applyStreamEvent
  )
}

async function stopTurn(): Promise<void> {
  await controller.value?.cancel()
  controller.value = null
  streaming.value = false
  markAssistant('completed')
}

async function ensureSession(projectId: string, nextProviderId: string): Promise<CodingSession> {
  if (
    session.value &&
    session.value.projectId === projectId &&
    session.value.engineId === engineId.value &&
    session.value.providerId === nextProviderId &&
    session.value.model === model.value
  ) {
    return session.value
  }
  session.value = await window.dreamworker.coding.createSession({
    projectId,
    engineId: engineId.value,
    providerId: nextProviderId,
    model: model.value,
    title: 'Coding Agent'
  })
  return session.value
}

function applyStreamEvent(event: CodingStreamEvent): void {
  if (event.type === 'started') {
    addSystemMessage(event.message || `${engineLabel(event.engineId)} 已启动`)
    return
  }
  if (event.type === 'delta' && event.delta) {
    appendAssistant(event.delta)
    return
  }
  if (event.type === 'tool_call' && event.toolCall) {
    commandLogs.value = [
      {
        id: `${event.streamId}_${event.sequence}`,
        command: event.toolCall.toolName,
        output: JSON.stringify(event.toolCall.arguments ?? {}, null, 2),
        createdAt: event.timestamp
      },
      ...commandLogs.value
    ].slice(0, 20)
    return
  }
  if (event.type === 'shell_output') {
    commandLogs.value = [
      {
        id: `${event.streamId}_${event.sequence}`,
        command: event.command || 'shell',
        output: event.output || event.message || '',
        createdAt: event.timestamp
      },
      ...commandLogs.value
    ].slice(0, 20)
    return
  }
  if (event.type === 'file_changed' && event.file) {
    changedFiles.value = upsertChange(changedFiles.value, event.file)
    return
  }
  if (event.type === 'completed') {
    markAssistant('completed')
    streaming.value = false
    controller.value = null
    if (event.engineThreadId && session.value) {
      session.value = { ...session.value, engineThreadId: event.engineThreadId, status: 'ready' }
    }
    void refreshFiles()
    void refreshStatus()
    return
  }
  if (event.type === 'cancelled') {
    markAssistant('completed')
    streaming.value = false
    controller.value = null
    addSystemMessage('本轮已停止。')
    return
  }
  if (event.type === 'error') {
    streamError.value = event.error?.message ?? event.message ?? '编码 Agent 运行失败。'
    markAssistant('error')
    streaming.value = false
    controller.value = null
  }
}

function ensureProviderSelection(): void {
  const engine = selectedEngine.value
  if (!engine || compatibleProviders.value.length === 0) {
    providerId.value = ''
    model.value = ''
    return
  }
  if (compatibleProviders.value.some((provider) => provider.providerId === providerId.value)) {
    ensureModelSelection()
    return
  }
  const preferred =
    engine.preferredProviderIds
      .map((id) => compatibleProviders.value.find((provider) => provider.providerId === id))
      .find(Boolean) ??
    compatibleProviders.value.find((provider) => isRoutedModelProvider(provider)) ??
    compatibleProviders.value[0]
  providerId.value = preferred?.providerId ?? ''
  ensureModelSelection()
}

function ensureModelSelection(): void {
  const models = availableModels.value
  if (models.length === 0) {
    model.value = ''
    return
  }
  if (!models.includes(model.value)) {
    model.value = selectedProvider.value?.defaultModel || models[0] || ''
  }
}

function providerSupport(
  provider: SafeModelProvider,
  engine: CodingEngineDescriptor | undefined
): { supported: boolean; reason: string } {
  if (!engine) {
    return { supported: false, reason: 'engine unavailable' }
  }
  if (!provider.enabled) {
    return { supported: false, reason: 'provider disabled' }
  }
  if (engine.engineId === 'claude_agent') {
    return provider.providerType === 'anthropic'
      ? { supported: true, reason: '' }
      : { supported: false, reason: 'Claude Agent uses Anthropic provider' }
  }
  const allowed: readonly ProviderType[] =
    engine.engineId === 'opencode'
      ? ['openai', 'openai_compatible', 'deepseek', 'siliconflow', 'glm', 'ollama', 'custom']
      : ['openai', 'openai_compatible', 'deepseek', 'siliconflow', 'glm', 'custom']
  if (allowed.includes(provider.providerType) || isRoutedModelProvider(provider)) {
    return { supported: true, reason: '' }
  }
  return {
    supported: false,
    reason: `${engine.displayName} prefers 9Router or OpenAI-compatible providers`
  }
}

function appendAssistant(delta: string): void {
  messages.value = messages.value.map((message) =>
    message.id === assistantMessageId.value
      ? { ...message, content: `${message.content}${delta}` }
      : message
  )
}

function markAssistant(status: CodingMessage['status']): void {
  messages.value = messages.value.map((message) =>
    message.id === assistantMessageId.value ? { ...message, status } : message
  )
}

function addSystemMessage(content: string): void {
  messages.value = [
    ...messages.value,
    {
      id: `system_${Date.now()}_${Math.random().toString(36).slice(2)}`,
      role: 'system',
      content,
      createdAt: new Date().toISOString(),
      status: 'completed'
    }
  ]
}

function upsertChange(
  current: readonly CodingFileChange[],
  change: CodingFileChange
): CodingFileChange[] {
  return [change, ...current.filter((item) => item.path !== change.path)].slice(0, 30)
}

function engineLabel(id: CodingEngineId): string {
  return engines.value.find((engine) => engine.engineId === id)?.displayName ?? id
}

function openProjectDirectorySettings(): void {
  appShell.setProjectSettingsTab('directory')
  appShell.setPrimary('projects')
}

function formatFileSize(size: number): string {
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${Math.round(size / 1024)} KB`
  }
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}
</script>

<template>
  <section class="coding-agent-workspace panel-surface" aria-label="编码 Agent 工作台">
    <header class="coding-topbar">
      <button
        class="icon-button small"
        type="button"
        title="返回开发模块"
        @click="appShell.leaveSubmoduleDetail()"
      >
        <ArrowLeft :size="16" aria-hidden="true" />
      </button>
      <div class="coding-title">
        <p class="eyebrow">开发模块 / 编码 Agent</p>
        <h2>编码 Agent</h2>
      </div>
      <div class="coding-controls">
        <div class="segmented-control" aria-label="引擎">
          <button
            v-for="engine in engines"
            :key="engine.engineId"
            type="button"
            :class="{ active: engine.engineId === engineId }"
            @click="engineId = engine.engineId"
          >
            {{ engine.displayName }}
          </button>
        </div>
        <select v-model="providerId" title="供应商">
          <option
            v-for="option in providerOptions"
            :key="option.provider.providerId"
            :value="option.provider.providerId"
            :disabled="!option.supported"
          >
            {{ option.provider.displayName }}{{ option.supported ? '' : ` - ${option.reason}` }}
          </option>
        </select>
        <select v-model="model" title="模型">
          <option v-for="item in availableModels" :key="item" :value="item">{{ item }}</option>
        </select>
        <span class="runtime-pill" :class="{ ready: runtimeReady, error: !runtimeReady }">
          <CheckCircle2 v-if="runtimeReady" :size="15" aria-hidden="true" />
          <AlertTriangle v-else :size="15" aria-hidden="true" />
          {{ runtimeReady ? 'SDK ready' : 'runtime missing' }}
        </span>
      </div>
    </header>

    <section v-if="!hasProjectRoot" class="coding-empty-state">
      <FolderTree :size="38" aria-hidden="true" />
      <h3>需要绑定项目目录</h3>
      <p>编码 Agent 只会在当前项目 localRootPath/workspace/code 内读写文件。</p>
      <button class="primary-button" type="button" @click="openProjectDirectorySettings">
        <Settings :size="16" aria-hidden="true" />
        绑定 localRootPath
      </button>
    </section>

    <section v-else class="coding-grid">
      <aside class="coding-pane file-pane">
        <div class="pane-heading">
          <FolderTree :size="17" aria-hidden="true" />
          <strong>文件树</strong>
          <button type="button" title="刷新文件" @click="refreshFiles">
            <RefreshCw :size="15" aria-hidden="true" />
          </button>
        </div>
        <input
          v-model="fileQuery"
          class="coding-input"
          type="search"
          placeholder="搜索文件"
          @keydown.enter="refreshFiles"
        />
        <div class="file-list">
          <button
            v-for="file in files"
            :key="file.path"
            type="button"
            :class="{ active: selectedPath === file.path }"
            @click="file.isDir ? undefined : selectFile(file.path)"
          >
            <span>{{ file.isDir ? '▸' : '·' }}</span>
            <strong>{{ file.path }}</strong>
            <small v-if="file.gitStatus">{{ file.gitStatus }}</small>
          </button>
        </div>
        <div class="file-preview">
          <div class="pane-heading compact">
            <FileText :size="15" aria-hidden="true" />
            <strong>{{ selectedFile?.path ?? '预览' }}</strong>
            <span v-if="selectedFile">{{ formatFileSize(selectedFile.size) }}</span>
          </div>
          <pre v-if="selectedFile?.content">{{ selectedFile.content }}</pre>
          <p v-else>
            {{ selectedFile ? '二进制或超大文件不展示内容。' : '选择一个文件查看内容。' }}
          </p>
        </div>
      </aside>

      <main class="coding-pane chat-pane">
        <div class="message-stream">
          <article
            v-for="message in messages"
            :key="message.id"
            class="coding-message"
            :class="[message.role, message.status]"
          >
            <span>{{
              message.role === 'user'
                ? 'You'
                : message.role === 'assistant'
                  ? engineLabel(engineId)
                  : 'System'
            }}</span>
            <p>{{ message.content || (message.status === 'streaming' ? '运行中...' : '') }}</p>
          </article>
          <article v-if="messages.length === 0" class="coding-message system">
            <span>System</span>
            <p>选择引擎、供应商和模型后，可以直接让 Agent 在 workspace/code 内修改代码。</p>
          </article>
        </div>
        <div v-if="streamError" class="coding-error">
          <AlertTriangle :size="15" aria-hidden="true" />
          {{ streamError }}
        </div>
        <form class="coding-composer" @submit.prevent="sendPrompt">
          <textarea
            v-model="draft"
            rows="3"
            placeholder="描述要实现、修复或分析的代码任务"
            :disabled="streaming"
          />
          <button v-if="streaming" type="button" title="停止" @click="stopTurn">
            <Square :size="16" aria-hidden="true" />
          </button>
          <button v-else class="primary-button" type="submit" :disabled="!canSend" title="发送">
            <Send :size="16" aria-hidden="true" />
          </button>
        </form>
      </main>

      <aside class="coding-pane runtime-pane">
        <div class="runtime-card">
          <div class="pane-heading compact">
            <Bot :size="16" aria-hidden="true" />
            <strong>SDK runtime</strong>
          </div>
          <dl>
            <dt>状态</dt>
            <dd>{{ runtime?.message ?? 'loading' }}</dd>
            <dt>Node</dt>
            <dd>{{ runtime?.nodeBin || 'node' }}</dd>
            <dt>Adapter</dt>
            <dd>{{ runtime?.adapterPath || '未找到' }}</dd>
          </dl>
        </div>

        <div class="runtime-card">
          <div class="pane-heading compact">
            <GitBranch :size="16" aria-hidden="true" />
            <strong>文件变更</strong>
            <button type="button" title="刷新状态" @click="refreshStatus">
              <RefreshCw :size="14" aria-hidden="true" />
            </button>
          </div>
          <p>
            {{
              fileStatus?.branch
                ? `branch ${fileStatus.branch}`
                : (fileStatus?.message ?? '暂无状态')
            }}
          </p>
          <ul>
            <li v-for="change in changedFiles" :key="`${change.status}_${change.path}`">
              <span>{{ change.status }}</span>
              <strong>{{ change.path }}</strong>
            </li>
          </ul>
        </div>

        <div class="runtime-card command-card">
          <div class="pane-heading compact">
            <Terminal :size="16" aria-hidden="true" />
            <strong>命令日志</strong>
            <Loader2 v-if="streaming" class="spin" :size="14" aria-hidden="true" />
          </div>
          <ol>
            <li v-for="log in commandLogs" :key="log.id">
              <strong>{{ log.command }}</strong>
              <pre>{{ log.output }}</pre>
            </li>
          </ol>
        </div>
      </aside>
    </section>
  </section>
</template>

<style scoped>
.coding-agent-workspace {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 14px;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  background:
    url('/aios/resource-orbit-banner.png') top right / min(36vw, 460px) auto no-repeat,
    rgba(255, 255, 255, 0.78);
  padding: 18px;
}

.coding-topbar,
.coding-controls,
.pane-heading,
.coding-composer,
.runtime-pill,
.coding-empty-state button {
  display: flex;
  align-items: center;
}

.coding-topbar {
  gap: 12px;
  min-width: 0;
}

.coding-title {
  min-width: 170px;
}

.coding-title h2 {
  margin: 0;
}

.coding-controls {
  flex: 1;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  min-width: 0;
}

.segmented-control {
  display: inline-grid;
  grid-auto-flow: column;
  gap: 4px;
  border: 1px solid rgba(148, 163, 184, 0.26);
  border-radius: 8px;
  padding: 4px;
  background: rgba(248, 250, 252, 0.78);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.segmented-control button,
.pane-heading button,
.file-list button,
.coding-composer button {
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  cursor: pointer;
}

.segmented-control button {
  min-height: 30px;
  padding: 0 10px;
  color: var(--muted);
}

.segmented-control button.active {
  border-color: rgba(37, 99, 235, 0.28);
  background: rgba(239, 246, 255, 0.92);
  color: var(--text);
  box-shadow: 0 6px 14px rgba(37, 99, 235, 0.1);
}

.coding-controls select,
.coding-input,
.coding-composer textarea {
  border: 1px solid rgba(148, 163, 184, 0.26);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.86);
  color: var(--text);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.coding-controls select {
  max-width: 220px;
  min-height: 38px;
  padding: 0 10px;
}

.runtime-pill {
  gap: 6px;
  min-height: 34px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 999px;
  padding: 0 10px;
  color: var(--muted);
  background: rgba(248, 250, 252, 0.76);
}

.runtime-pill.ready {
  border-color: rgba(125, 242, 176, 0.35);
  color: var(--green);
  background: rgba(236, 253, 245, 0.82);
}

.runtime-pill.error {
  border-color: rgba(255, 107, 154, 0.35);
  color: var(--danger);
  background: rgba(254, 242, 242, 0.82);
}

.coding-empty-state {
  display: grid;
  place-items: center;
  align-content: center;
  gap: 10px;
  min-height: 420px;
  color: var(--muted);
  text-align: center;
}

.coding-empty-state h3,
.coding-empty-state p {
  margin: 0;
}

.coding-grid {
  display: grid;
  grid-template-columns: minmax(240px, 0.85fr) minmax(420px, 1.45fr) minmax(260px, 0.9fr);
  gap: 12px;
  min-width: 0;
  min-height: 0;
}

.coding-pane {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.68);
  box-shadow: var(--shadow-soft);
  backdrop-filter: var(--glass-blur);
}

.file-pane,
.runtime-pane,
.chat-pane {
  display: grid;
  gap: 10px;
  padding: 12px;
}

.file-pane {
  grid-template-rows: auto auto minmax(0, 1fr) minmax(140px, 0.65fr);
}

.chat-pane {
  grid-template-rows: minmax(0, 1fr) auto auto;
}

.runtime-pane {
  align-content: start;
  overflow: auto;
}

.pane-heading {
  gap: 8px;
  min-width: 0;
}

.pane-heading.compact {
  min-height: 28px;
}

.pane-heading strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pane-heading button {
  margin-left: auto;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-color: rgba(148, 163, 184, 0.24);
  background: rgba(248, 250, 252, 0.76);
}

.coding-input {
  width: 100%;
  min-height: 36px;
  padding: 0 10px;
}

.file-list {
  min-height: 0;
  overflow: auto;
  display: grid;
  align-content: start;
  gap: 4px;
}

.file-list button {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr) auto;
  gap: 6px;
  min-height: 30px;
  padding: 5px 6px;
  color: var(--muted);
  text-align: left;
}

.file-list button:hover,
.file-list button.active {
  border-color: rgba(124, 58, 237, 0.18);
  background: rgba(237, 231, 255, 0.62);
  color: var(--text);
}

.file-list strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.file-preview {
  min-height: 0;
  overflow: hidden;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
  padding-top: 8px;
}

.file-preview pre,
.command-card pre {
  margin: 0;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  color: #334155;
  font-size: 12px;
  line-height: 1.55;
}

.file-preview pre {
  max-height: calc(100% - 34px);
}

.file-preview p,
.runtime-card p,
.coding-message p {
  margin: 0;
  color: var(--muted);
}

.message-stream {
  display: grid;
  align-content: start;
  gap: 10px;
  min-height: 0;
  overflow: auto;
}

.coding-message {
  display: grid;
  gap: 4px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.76);
  box-shadow: 0 8px 18px rgba(100, 116, 139, 0.08);
}

.coding-message.user {
  margin-left: 32px;
  border-color: rgba(37, 99, 235, 0.18);
  background: rgba(239, 246, 255, 0.82);
}

.coding-message.assistant {
  margin-right: 32px;
}

.coding-message.system {
  border-style: dashed;
}

.coding-message span {
  color: var(--purple);
  font-size: 12px;
  font-weight: 800;
}

.coding-message.error {
  border-color: rgba(255, 107, 154, 0.42);
}

.coding-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--danger);
  font-size: 13px;
}

.coding-composer {
  gap: 8px;
}

.coding-composer textarea {
  flex: 1;
  min-width: 0;
  resize: none;
  padding: 10px;
}

.coding-composer button {
  width: 42px;
  height: 42px;
  justify-content: center;
  border-color: rgba(148, 163, 184, 0.24);
  background: rgba(248, 250, 252, 0.82);
}

.runtime-card {
  display: grid;
  gap: 8px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  padding-bottom: 12px;
}

.runtime-card dl {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr);
  gap: 6px;
  margin: 0;
  font-size: 12px;
}

.runtime-card dd {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
}

.runtime-card ul,
.runtime-card ol {
  display: grid;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.runtime-card li {
  display: grid;
  gap: 3px;
  min-width: 0;
  color: var(--muted);
  font-size: 12px;
}

.runtime-card li span {
  color: var(--amber);
  font-weight: 800;
}

.runtime-card li strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spin {
  animation: coding-spin 900ms linear infinite;
}

@keyframes coding-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
