<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  AtSign,
  Bot,
  BrainCircuit,
  FolderKanban,
  Globe2,
  ListChecks,
  Paperclip,
  Plus,
  Send,
  Sparkles,
  Square,
  WandSparkles,
  Wrench
} from 'lucide-vue-next'
import { useAppShellStore } from '../stores/app-shell'
import { isNearScrollBottom } from '../utils/chat-scroll'
import type {
  ChatMessage,
  ChatModelUsage,
  ModelProfile,
  ProviderType,
  SafeModelProvider,
  SkillRuntimeDescriptor,
  ToolRuntimeDescriptor
} from '../../../shared/dreamworker-api'

const appShell = useAppShellStore()
const messageListRef = ref<HTMLElement | null>(null)
const shouldStickToThreadBottom = ref(true)
let scrollFrameId = 0
let scrollSyncTimer: number | undefined
let suppressScrollSync = false
let previousScrollMessageCount = 0
let previousScrollSessionId = ''

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

const activeAgentName = computed(() => appShell.activeAgent?.displayName ?? '通用助手')
const activeProjectTitle = computed(() => {
  if (!appShell.activeChatProjectId) {
    return '未绑定项目'
  }
  return (
    appShell.projects.find((project) => project.projectId === appShell.activeChatProjectId)
      ?.title ?? '未绑定项目'
  )
})
const runtimeStateText = computed(() => {
  if (appShell.activeSessionStreaming) {
    return '模型思考中'
  }
  return attemptStatusText(appShell.chatRuntimeAttemptStatus)
})
const activeModelLabel = computed(() => {
  const provider = appShell.providers.find(
    (item) => item.providerId === appShell.activeChatProviderId
  )
  return `${provider?.displayName ?? '模型服务'} / ${appShell.activeChatModel}`
})
const activeSessionTitle = computed(() => appShell.activeChatSession?.title ?? '新的 Agent 对话')
const chatThreadScrollSignature = computed(() =>
  appShell.chatMessages
    .map(
      (message) =>
        `${message.messageId}:${message.status}:${message.content.length}:${
          appShell.chatReasoningByMessage[message.messageId]?.length ?? 0
        }`
    )
    .join('|')
)
const quickPrompts = [
  '把这个 AI 产品想法拆成机会、风险和下一步验证计划。',
  '根据当前项目，生成一版 PRD 大纲和关键页面清单。',
  '检查当前资源配置，告诉我模型、Agent、工具还有哪些缺口。'
]

function messageStatusText(status: string): string {
  const map: Record<string, string> = {
    streaming: '生成中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return map[status] ?? status
}

function finishReasonText(reason: string): string {
  const map: Record<string, string> = {
    stop: '自然结束',
    length: '长度截断',
    tool_calls: '工具调用',
    streaming: '生成中',
    cancelled: '已取消',
    error: '异常'
  }
  return map[reason] ?? reason
}

function attemptStatusText(status: string): string {
  const map: Record<string, string> = {
    ready: '就绪',
    streaming: '流式生成中',
    blocked: '受阻',
    failed: '失败',
    cancelled: '已取消',
    completed: '已完成'
  }
  return map[status] ?? status
}

function plannerStrategyText(strategy: string | undefined): string {
  const map: Record<string, string> = {
    'plan-execute': '计划后执行',
    react: '边想边做',
    manual: '手动推进'
  }
  return strategy ? (map[strategy] ?? strategy) : '暂无'
}

function memoryScopeText(scope: string | undefined): string {
  const map: Record<string, string> = {
    short_term: '短期',
    project: '项目',
    semantic: '语义'
  }
  return scope ? (map[scope] ?? scope) : '暂无'
}

function riskLevelText(risk: string): string {
  const map: Record<string, string> = {
    low: '低风险',
    medium: '中风险',
    high: '高风险',
    critical: '关键风险'
  }
  return map[risk] ?? risk
}

function toolStatusText(status: string): string {
  const map: Record<string, string> = {
    preview: '预览',
    pending_approval: '待审批',
    running: '运行中',
    completed: '已完成',
    blocked: '已阻止'
  }
  return map[status] ?? status
}

function messageAuthor(role: string): string {
  if (role === 'user') {
    return '你'
  }
  if (role === 'assistant') {
    return activeAgentName.value
  }
  return '系统'
}

function activeProviderForMessage(message: ChatMessage): SafeModelProvider | undefined {
  if (message.providerId) {
    return appShell.providers.find((provider) => provider.providerId === message.providerId)
  }
  const profile = modelProfileForMessage(message)
  if (profile) {
    return appShell.providers.find((provider) => provider.providerId === profile.providerId)
  }
  if (appShell.chatRuntimeProvider) {
    return appShell.providers.find(
      (provider) => provider.providerId === appShell.chatRuntimeProvider
    )
  }
  return appShell.activeProvider
}

function modelProfileForMessage(message: ChatMessage): ModelProfile | undefined {
  return (
    appShell.profiles.find((profile) => profile.profileId === message.model) ??
    appShell.profiles.find((profile) => profile.model === message.model) ??
    appShell.profiles.find((profile) => profile.profileId === appShell.activeChatModelProfileId)
  )
}

function providerLogoForMessage(message: ChatMessage): string {
  const provider = activeProviderForMessage(message)
  return provider ? providerLogoSrc[provider.providerType] : '/provider-icons/openai.svg'
}

function messageProviderName(message: ChatMessage): string {
  return (
    activeProviderForMessage(message)?.displayName ??
    message.providerId ??
    appShell.chatRuntimeProvider ??
    '模型服务'
  )
}

function messageModelName(message: ChatMessage): string {
  const profile = modelProfileForMessage(message)
  if (profile?.model) {
    return profile.model
  }
  return message.model || appShell.chatRuntimeModel || appShell.activeChatModelProfileId
}

function messageRuntimeLabel(message: ChatMessage): string {
  if (message.role !== 'assistant') {
    return `${messageStatusText(message.status)} / ${finishReasonText(message.finishReason || 'streaming')}`
  }
  return `${messageProviderName(message)} / ${messageModelName(message)}`
}

function parseRuntimeSummary(summary: string): Record<string, string> {
  return summary.split('/').reduce<Record<string, string>>((result, part) => {
    const [rawKey = '', ...rawValue] = part.trim().split('=')
    const key = rawKey.trim()
    const value = rawValue.join('=').trim()
    if (key && value) {
      result[key] = value
    }
    return result
  }, {})
}

function formatDuration(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) {
    return '0 ms'
  }
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(1)} s`
  }
  return `${Math.round(ms)} ms`
}

function messageUsageText(usage: ChatModelUsage | null): string {
  if (!usage) {
    return 'Token 统计等待中'
  }
  return `花费 ${usage.totalTokens} token · 输入 ${usage.inputTokens} / 输出 ${usage.outputTokens}`
}

function isLatestAssistantMessage(message: ChatMessage): boolean {
  const latest = [...appShell.chatMessages].reverse().find((item) => item.role === 'assistant')
  return latest?.messageId === message.messageId
}

function messageLatencyText(message: ChatMessage): string {
  if (message.status === 'streaming') {
    return '正在执行'
  }
  if (!isLatestAssistantMessage(message) || appShell.chatRuntimeLatencyMs <= 0) {
    return ''
  }
  return `耗时 ${formatDuration(appShell.chatRuntimeLatencyMs)}`
}

function messageExecutionText(message: ChatMessage): string {
  const summary = parseRuntimeSummary(message.runtimeSummary)
  const planner = plannerStrategyText(summary.Planner || appShell.activeAgent?.planner.strategy)
  const memory = memoryScopeText(summary.Memory || appShell.activeAgent?.memoryScope)
  const timeoutValue = Number((summary.Timeout || '').replace(/[^\d.]/g, ''))
  const timeout = timeoutValue > 0 ? ` / 超时 ${formatDuration(timeoutValue)}` : ''
  const finishReason = message.finishReason
    ? finishReasonText(message.finishReason)
    : messageStatusText(message.status)
  return `执行：${planner} / ${memory} / ${finishReason}${timeout}`
}

function runtimeSelectionForMessage(message: ChatMessage) {
  return isLatestAssistantMessage(message) ? appShell.chatRuntimeSelection : null
}

function selectedSkills(message: ChatMessage): readonly SkillRuntimeDescriptor[] {
  return runtimeSelectionForMessage(message)?.skills ?? []
}

function selectedTools(message: ChatMessage): readonly ToolRuntimeDescriptor[] {
  return runtimeSelectionForMessage(message)?.tools ?? []
}

function reasoningText(message: ChatMessage): string {
  return appShell.chatReasoningByMessage[message.messageId]?.trim() ?? ''
}

function thinkingDetailSummary(message: ChatMessage): string {
  const selection = runtimeSelectionForMessage(message)
  if (selection?.summary) {
    return selection.summary
  }
  return `${selectedSkills(message).length} 个 Skill / ${selectedTools(message).length} 个工具`
}

function toolCallSummary(): string {
  if (appShell.chatToolCalls.length === 0) {
    return '暂无实际工具调用'
  }
  return appShell.chatToolCalls
    .map((call) => `${call.displayName}：${toolStatusText(call.status)}`)
    .join('、')
}

function showModelThinking(message: ChatMessage): boolean {
  return message.role === 'assistant' && message.status === 'streaming' && !message.content.trim()
}

function messageInitial(role: string): string {
  if (role === 'user') {
    return '你'
  }
  if (role === 'assistant') {
    return 'AI'
  }
  return 'SYS'
}

function sessionAgentName(agentId: string): string {
  return appShell.agents.find((agent) => agent.agentId === agentId)?.displayName ?? agentId
}

function sessionModelName(session: {
  modelProfileId: string
  providerId?: string
  model?: string
}): string {
  if (session.providerId || session.model) {
    const provider = appShell.providers.find((item) => item.providerId === session.providerId)
    return `${provider?.displayName ?? session.providerId ?? '模型服务'} / ${session.model ?? ''}`
  }
  const profile = appShell.profiles.find((item) => item.profileId === session.modelProfileId)
  return profile ? `${profile.providerId} / ${profile.model}` : session.modelProfileId
}

function formatSessionTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '刚刚'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

function traceText(traceId: string): string {
  return traceId ? `追踪 ${traceId}` : '等待追踪'
}

function useQuickPrompt(prompt: string): void {
  appShell.chatDraft = prompt
}

function clearScrollSyncTimer(): void {
  if (!scrollSyncTimer) {
    return
  }
  window.clearTimeout(scrollSyncTimer)
  scrollSyncTimer = undefined
}

function threadIsNearBottom(): boolean {
  const element = messageListRef.value
  if (!element) {
    return true
  }
  return isNearScrollBottom({
    scrollTop: element.scrollTop,
    scrollHeight: element.scrollHeight,
    clientHeight: element.clientHeight
  })
}

function syncThreadStickiness(): void {
  if (suppressScrollSync) {
    return
  }
  shouldStickToThreadBottom.value = threadIsNearBottom()
}

function handleThreadScroll(): void {
  syncThreadStickiness()
}

async function scrollThreadToBottom(force = false): Promise<void> {
  const shouldScroll = force || shouldStickToThreadBottom.value
  if (!shouldScroll) {
    return
  }
  await nextTick()
  window.cancelAnimationFrame(scrollFrameId)
  scrollFrameId = window.requestAnimationFrame(() => {
    const element = messageListRef.value
    if (!element) {
      return
    }
    suppressScrollSync = true
    element.scrollTo({ top: element.scrollHeight, behavior: 'auto' })
    clearScrollSyncTimer()
    scrollSyncTimer = window.setTimeout(() => {
      suppressScrollSync = false
      shouldStickToThreadBottom.value = threadIsNearBottom()
      scrollSyncTimer = undefined
    }, 80)
  })
}

watch(
  () => appShell.activeChatSessionId,
  () => {
    previousScrollSessionId = appShell.activeChatSessionId
    previousScrollMessageCount = appShell.chatMessages.length
    shouldStickToThreadBottom.value = true
    void scrollThreadToBottom(true)
  },
  { flush: 'post' }
)

watch(
  chatThreadScrollSignature,
  () => {
    const sessionChanged = previousScrollSessionId !== appShell.activeChatSessionId
    const messageCountChanged = previousScrollMessageCount !== appShell.chatMessages.length
    previousScrollSessionId = appShell.activeChatSessionId
    previousScrollMessageCount = appShell.chatMessages.length
    if (sessionChanged || messageCountChanged) {
      shouldStickToThreadBottom.value = true
    }
    void scrollThreadToBottom(sessionChanged || messageCountChanged)
  },
  { flush: 'post' }
)

onMounted(() => {
  previousScrollSessionId = appShell.activeChatSessionId
  previousScrollMessageCount = appShell.chatMessages.length
  void scrollThreadToBottom(true)
})

onBeforeUnmount(() => {
  window.cancelAnimationFrame(scrollFrameId)
  clearScrollSyncTimer()
})
</script>

<template>
  <section class="workspace-layout chat-layout">
    <aside class="chat-session-rail sub-rail" aria-label="会话列表">
      <div class="chat-rail-header">
        <div>
          <p class="eyebrow">会话</p>
          <h2>Agent 对话</h2>
        </div>
        <button
          class="icon-button small"
          type="button"
          title="新建对话"
          aria-label="新建对话"
          @click="appShell.createChatSession()"
        >
          <Plus :size="16" aria-hidden="true" />
        </button>
      </div>

      <div class="chat-session-list">
        <button
          v-for="session in appShell.chatSessions"
          :key="session.sessionId"
          class="chat-session-row"
          :class="{ active: session.sessionId === appShell.activeChatSessionId }"
          :aria-current="session.sessionId === appShell.activeChatSessionId ? 'true' : undefined"
          type="button"
          @click="appShell.selectChatSession(session.sessionId)"
        >
          <span class="session-dot" aria-hidden="true">
            <Bot :size="17" />
          </span>
          <span class="session-main">
            <strong>{{ session.title }}</strong>
            <small
              >{{ session.messageCount }} 条消息 / {{ sessionAgentName(session.agentId) }}</small
            >
          </span>
          <time :datetime="session.updatedAt">{{ formatSessionTime(session.updatedAt) }}</time>
          <small class="session-model">{{ sessionModelName(session) }}</small>
        </button>
      </div>
    </aside>

    <section class="chat-center panel-surface" aria-label="Agent 对话">
      <header class="chat-main-header">
        <div class="chat-agent-title">
          <span class="chat-agent-avatar" aria-hidden="true">
            <Bot :size="21" />
          </span>
          <div>
            <p class="eyebrow">当前助手</p>
            <h2>{{ activeAgentName }}</h2>
            <small
              >{{ activeSessionTitle }} / {{ activeProjectTitle }} / {{ activeModelLabel }}</small
            >
          </div>
        </div>
        <div class="chat-header-actions">
          <span class="runtime-pill" :data-streaming="appShell.activeSessionStreaming">
            <Sparkles :size="14" aria-hidden="true" />
            {{ runtimeStateText }}
          </span>
          <button
            class="icon-button small"
            type="button"
            title="资源配置"
            aria-label="资源配置"
            @click="appShell.setPrimary('resources')"
          >
            <Wrench :size="15" aria-hidden="true" />
          </button>
        </div>
      </header>

      <div class="chat-binding-bar" aria-label="对话绑定">
        <label>
          Agent
          <select
            :value="appShell.activeAgentId"
            :disabled="appShell.activeSessionStreaming"
            @change="appShell.setActiveChatAgent(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="agent in appShell.agents" :key="agent.agentId" :value="agent.agentId">
              {{ agent.displayName }}
            </option>
          </select>
        </label>
        <label>
          服务商
          <select
            :value="appShell.activeChatProviderId"
            :disabled="appShell.activeSessionStreaming"
            @change="appShell.setActiveChatProvider(($event.target as HTMLSelectElement).value)"
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
            :value="appShell.activeChatModel"
            :disabled="appShell.activeSessionStreaming"
            @change="appShell.setActiveChatModel(($event.target as HTMLSelectElement).value)"
          >
            <option
              v-for="model in appShell.modelsForProvider(appShell.activeChatProviderId)"
              :key="model"
              :value="model"
            >
              {{ model }}
            </option>
          </select>
        </label>
        <label>
          项目
          <select
            :value="appShell.activeChatProjectId"
            :disabled="appShell.activeSessionStreaming"
            @change="appShell.setActiveChatProject(($event.target as HTMLSelectElement).value)"
          >
            <option value="">不绑定项目</option>
            <option
              v-for="project in appShell.projects"
              :key="project.projectId"
              :value="project.projectId"
            >
              {{ project.title }}
            </option>
          </select>
        </label>
      </div>

      <div
        ref="messageListRef"
        class="message-list chat-thread"
        aria-label="消息列表"
        @scroll.passive="handleThreadScroll"
      >
        <article v-if="appShell.chatMessages.length === 0" class="empty-message chat-empty-state">
          <WandSparkles :size="24" aria-hidden="true" />
          <h3>直接开始一次工作对话</h3>
          <p>当前 Agent 会带上模型、项目和运行时上下文，右侧同步展示计划、工具和执行状态。</p>
          <div class="quick-prompt-row" aria-label="快捷提示">
            <button
              v-for="prompt in quickPrompts"
              :key="prompt"
              type="button"
              @click="useQuickPrompt(prompt)"
            >
              {{ prompt }}
            </button>
          </div>
        </article>
        <article
          v-for="message in appShell.chatMessages"
          :key="message.messageId"
          class="chat-message-row"
          :data-role="message.role"
          :data-status="message.status"
        >
          <span
            class="message-avatar"
            :class="{ 'provider-avatar': message.role === 'assistant' }"
            :data-role="message.role"
            aria-hidden="true"
          >
            <img
              v-if="message.role === 'assistant'"
              :src="providerLogoForMessage(message)"
              alt=""
            />
            <span v-else>{{ messageInitial(message.role) }}</span>
          </span>
          <div class="chat-message">
            <header>
              <strong>{{ messageAuthor(message.role) }}</strong>
              <small>{{ messageRuntimeLabel(message) }}</small>
            </header>
            <div v-if="showModelThinking(message)" class="model-thinking">
              <img :src="providerLogoForMessage(message)" alt="" />
              <span>{{ messageProviderName(message) }} 正在思考</span>
              <span class="thinking-dots" aria-hidden="true">
                <i></i>
                <i></i>
                <i></i>
              </span>
            </div>
            <p v-else>
              {{ message.content || (message.status === 'streaming' ? '正在生成...' : '暂无内容') }}
            </p>
            <details v-if="message.role === 'assistant'" class="thinking-detail-card">
              <summary>
                <span class="thinking-summary-main">
                  <img :src="providerLogoForMessage(message)" alt="" />
                  <strong>{{
                    message.status === 'streaming' ? '查看模型思考' : '思考与执行详情'
                  }}</strong>
                </span>
                <small>{{ thinkingDetailSummary(message) }}</small>
              </summary>
              <div class="thinking-detail-body">
                <section>
                  <strong>模型思考</strong>
                  <p>
                    {{
                      reasoningText(message) ||
                      '当前模型未返回可展示的推理内容，已展示思考状态和执行轨迹。'
                    }}
                  </p>
                </section>
                <section>
                  <strong>Skill 选择</strong>
                  <div v-if="selectedSkills(message).length > 0" class="detail-chip-list">
                    <span v-for="skill in selectedSkills(message)" :key="skill.skillId">
                      {{ skill.displayName || skill.skillId }}
                    </span>
                  </div>
                  <p v-else>本轮没有命中可用 Skill。</p>
                </section>
                <section>
                  <strong>工具选择</strong>
                  <div v-if="selectedTools(message).length > 0" class="detail-chip-list">
                    <span v-for="tool in selectedTools(message)" :key="tool.toolId">
                      {{ tool.displayName || tool.toolId }} / {{ riskLevelText(tool.riskLevel) }}
                    </span>
                  </div>
                  <p v-else>本轮没有可用工具。</p>
                </section>
                <section>
                  <strong>执行摘要</strong>
                  <p>{{ messageExecutionText(message) }} / {{ toolCallSummary() }}</p>
                </section>
              </div>
            </details>
            <footer
              :class="{ 'message-runtime-footer': message.role === 'assistant' }"
              aria-label="消息运行信息"
            >
              <template v-if="message.role === 'assistant'">
                <span>{{ messageUsageText(message.usage) }}</span>
                <span v-if="messageLatencyText(message)">{{ messageLatencyText(message) }}</span>
                <span>{{ messageExecutionText(message) }}</span>
                <span>{{ traceText(message.trace_id) }}</span>
              </template>
              <small v-else>{{ traceText(message.trace_id) }}</small>
            </footer>
          </div>
        </article>
      </div>

      <form class="composer chat-composer" @submit.prevent="appShell.sendChatMessage()">
        <p v-if="appShell.chatStreamError" class="stream-error">{{ appShell.chatStreamError }}</p>
        <textarea
          v-model="appShell.chatDraft"
          aria-label="输入消息"
          placeholder="输入问题、产品想法或下一步指令，Ctrl / ⌘ + Enter 发送"
          :disabled="appShell.activeSessionStreaming"
          @keydown.ctrl.enter.prevent="appShell.sendChatMessage()"
          @keydown.meta.enter.prevent="appShell.sendChatMessage()"
        />
        <div class="composer-toolbar">
          <div class="composer-tools" aria-label="对话工具">
            <button
              type="button"
              title="项目上下文"
              aria-label="项目上下文"
              @click="appShell.setPrimary('projects')"
            >
              <FolderKanban :size="16" aria-hidden="true" />
            </button>
            <button
              type="button"
              title="资源配置"
              aria-label="资源配置"
              @click="appShell.setPrimary('resources')"
            >
              <Wrench :size="16" aria-hidden="true" />
            </button>
            <button type="button" title="引用资源" aria-label="引用资源" disabled>
              <AtSign :size="16" aria-hidden="true" />
            </button>
            <button type="button" title="附件暂未开放" aria-label="附件暂未开放" disabled>
              <Paperclip :size="16" aria-hidden="true" />
            </button>
            <button type="button" title="联网暂未开放" aria-label="联网暂未开放" disabled>
              <Globe2 :size="16" aria-hidden="true" />
            </button>
          </div>
          <div class="composer-actions">
            <button
              v-if="
                !appShell.activeSessionStreaming &&
                (appShell.chatRuntimeAttemptStatus === 'failed' ||
                  appShell.chatRuntimeAttemptStatus === 'cancelled')
              "
              type="button"
              @click="appShell.retryLastChatMessage()"
            >
              重试
            </button>
            <button
              v-if="appShell.activeSessionStreaming"
              class="danger-button"
              type="button"
              @click="appShell.cancelChatStream()"
            >
              <Square :size="15" aria-hidden="true" />
              停止
            </button>
            <button
              v-else
              class="primary-button"
              type="submit"
              :disabled="
                appShell.chatSending ||
                Boolean(appShell.composerDisabledReason) ||
                !appShell.chatDraft.trim()
              "
            >
              <Send :size="16" aria-hidden="true" />
              发送
            </button>
          </div>
        </div>
        <small v-if="appShell.composerDisabledReason" class="disabled-reason">
          {{ appShell.composerDisabledReason }}
        </small>
      </form>
    </section>

    <aside class="chat-runtime-panel right-panel" aria-label="Agent 运行摘要">
      <section class="inspector-card">
        <div class="section-title">
          <BrainCircuit :size="16" aria-hidden="true" />
          <span>Agent</span>
        </div>
        <h3>{{ activeAgentName }}</h3>
        <p>{{ appShell.activeAgent?.description }}</p>
        <dl>
          <div>
            <dt>角色</dt>
            <dd>{{ appShell.activeAgent?.role }}</dd>
          </div>
          <div>
            <dt>模型</dt>
            <dd>{{ activeModelLabel }}</dd>
          </div>
          <div>
            <dt>Skill</dt>
            <dd>{{ appShell.activeAgent?.enabledSkills.length ?? 0 }}</dd>
          </div>
          <div>
            <dt>工具</dt>
            <dd>{{ appShell.activeAgent?.enabledTools.length ?? 0 }}</dd>
          </div>
          <div>
            <dt>规划器</dt>
            <dd>{{ plannerStrategyText(appShell.activeAgent?.planner.strategy) }}</dd>
          </div>
          <div>
            <dt>记忆</dt>
            <dd>{{ memoryScopeText(appShell.activeAgent?.memoryScope) }}</dd>
          </div>
        </dl>
      </section>

      <section class="inspector-card runtime-card">
        <div class="section-title">
          <ListChecks :size="16" aria-hidden="true" />
          <span>运行状态</span>
        </div>
        <dl class="runtime-metrics">
          <div>
            <dt>服务商</dt>
            <dd>{{ appShell.chatRuntimeProvider || appShell.activeProvider?.providerId }}</dd>
          </div>
          <div>
            <dt>模型</dt>
            <dd>{{ appShell.chatRuntimeModel || activeModelLabel }}</dd>
          </div>
          <div>
            <dt>延迟</dt>
            <dd>{{ appShell.chatRuntimeLatencyMs }} ms</dd>
          </div>
          <div>
            <dt>结束原因</dt>
            <dd>
              {{
                appShell.chatRuntimeFinishReason
                  ? finishReasonText(appShell.chatRuntimeFinishReason)
                  : attemptStatusText(appShell.chatRuntimeAttemptStatus)
              }}
            </dd>
          </div>
          <div>
            <dt>上下文</dt>
            <dd>
              {{ appShell.chatContextBudget.estimatedTokens }} /
              {{ appShell.chatContextBudget.inputBudgetTokens }}
            </dd>
          </div>
          <div>
            <dt>摘要</dt>
            <dd>
              {{
                appShell.chatContextBudget.compacted
                  ? `已压缩 ${appShell.chatContextBudget.compactedCount} 条`
                  : '就绪'
              }}
            </dd>
          </div>
          <div>
            <dt>工具</dt>
            <dd>{{ toolStatusText(appShell.chatRuntimeToolState) }}</dd>
          </div>
        </dl>
        <ol class="runtime-steps">
          <li
            v-for="step in appShell.chatExecutionSteps"
            :key="step.stepId"
            :data-status="step.status"
          >
            <strong>{{ step.phase }}</strong>
            <span>{{ step.title }}</span>
            <small>{{ step.summary }}</small>
          </li>
        </ol>
        <p v-if="appShell.chatExecutionSteps.length === 0">发送消息后显示执行阶段。</p>
      </section>

      <section class="inspector-card runtime-card">
        <div class="section-title">
          <Wrench :size="16" aria-hidden="true" />
          <span>工具调用</span>
        </div>
        <div v-if="appShell.chatToolCalls.length > 0" class="tool-call-list">
          <article v-for="call in appShell.chatToolCalls" :key="call.callId">
            <strong>{{ call.displayName }}</strong>
            <span>{{ riskLevelText(call.riskLevel) }} / {{ toolStatusText(call.status) }}</span>
            <p>{{ call.summary }}</p>
          </article>
        </div>
        <p v-else>暂无工具调用。</p>
      </section>

      <section class="inspector-card">
        <p class="eyebrow">下一步</p>
        <h3>沉淀到项目空间</h3>
        <p>聊清楚的问题可以继续进入探索、产品、开发、销售模块，形成可跟踪的工作闭环。</p>
        <div class="vertical-actions">
          <button type="button" @click="appShell.setPrimary('projects')">查看项目空间</button>
          <button type="button" @click="appShell.setPrimary('resources')">调整 Agent 资源</button>
        </div>
      </section>
    </aside>
  </section>
</template>
