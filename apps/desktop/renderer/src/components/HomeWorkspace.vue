<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import {
  ArrowRight,
  Bot,
  Boxes,
  BrainCircuit,
  CheckCircle2,
  Download,
  Gauge,
  MessageSquareText,
  Network,
  Play,
  Rocket,
  Route,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap
} from 'lucide-vue-next'
import { useAppShellStore, type PrimaryNavId } from '../stores/app-shell'

type ChartInstance = ReturnType<typeof echarts.init>

const PLANNED_AGENT_COUNT = 12
const PLANNED_SKILL_COUNT = 12
const PLANNED_TOOL_COUNT = 20
const WINDOWS_PREVIEW_DOWNLOAD_URL =
  'https://github.com/zhongxinqiao06-spec/DreamWorker/releases/download/v0.1.0-preview.1/DreamWorker.Setup.0.1.0.exe'

const appShell = useAppShellStore()
const tokenChartRef = ref<HTMLDivElement | null>(null)
const capabilityChartRef = ref<HTMLDivElement | null>(null)
const providerChartRef = ref<HTMLDivElement | null>(null)
let tokenChart: ChartInstance | null = null
let capabilityChart: ChartInstance | null = null
let providerChart: ChartInstance | null = null
let chartResizeObserver: ResizeObserver | null = null

const enabledProviders = computed(() => appShell.providers.filter((provider) => provider.enabled))
const connectedProviders = computed(() =>
  appShell.providers.filter(
    (provider) => provider.status === 'connected' || provider.healthStatus === 'connected'
  )
)
const uniqueModels = computed(() => {
  const models = new Set<string>()
  for (const provider of appShell.providers) {
    for (const model of provider.availableModels) {
      models.add(model)
    }
  }
  for (const model of appShell.nineRouterStatus?.models ?? []) {
    models.add(model)
  }
  return [...models]
})
const totalUsage = computed(() =>
  appShell.chatMessages.reduce(
    (usage, message) => ({
      inputTokens: usage.inputTokens + (message.usage?.inputTokens ?? 0),
      outputTokens: usage.outputTokens + (message.usage?.outputTokens ?? 0),
      totalTokens: usage.totalTokens + (message.usage?.totalTokens ?? 0),
      costUsd: usage.costUsd + (message.usage?.costUsd ?? 0)
    }),
    { inputTokens: 0, outputTokens: 0, totalTokens: 0, costUsd: 0 }
  )
)
const contextUsedPercent = computed(() => {
  const budget = appShell.chatContextBudget
  if (!budget.contextWindow) {
    return 0
  }
  return Math.min(100, Math.round((budget.estimatedTokens / budget.contextWindow) * 100))
})
const activeProjectTitle = computed(() => appShell.activeProject?.title ?? '还没有选择项目')
const activeProjectStatus = computed(() => {
  const status = appShell.activeProject?.status
  if (status === 'active') {
    return '推进中'
  }
  if (status === 'paused') {
    return '已暂停'
  }
  if (status === 'archived') {
    return '已归档'
  }
  return '待创建'
})
const modelTicker = computed(() =>
  (appShell.nineRouterStatus?.models.length
    ? appShell.nineRouterStatus.models
    : uniqueModels.value
  ).slice(0, 7)
)
const runtimeLevel = computed(() => {
  if (appShell.runtimePing.status === 'ready') {
    return 'ready'
  }
  if (appShell.runtimePing.status === 'checking') {
    return 'checking'
  }
  return 'attention'
})

const metricCards = computed(() => [
  {
    label: '可用模型',
    value: uniqueModels.value.length || appShell.profiles.length,
    detail: `${enabledProviders.value.length} 个服务商已启用`,
    tone: 'blue'
  },
  {
    label: '本轮 Token',
    value: totalUsage.value.totalTokens || appShell.chatContextBudget.estimatedTokens,
    detail: `输入 ${totalUsage.value.inputTokens} / 输出 ${totalUsage.value.outputTokens}`,
    tone: 'purple'
  },
  {
    label: '能力规划',
    value: PLANNED_AGENT_COUNT + PLANNED_SKILL_COUNT + PLANNED_TOOL_COUNT,
    detail: `${PLANNED_AGENT_COUNT} Agent · ${PLANNED_SKILL_COUNT}+ Skill · ${PLANNED_TOOL_COUNT}+ 工具`,
    tone: 'green'
  },
  {
    label: '路由模型',
    value: appShell.nineRouterStatus?.modelCount ?? 0,
    detail: appShell.nineRouterStatus?.healthStatus === 'connected' ? '9Router 在线' : '等待接入',
    tone: 'amber'
  }
])

const guideCards: {
  step: string
  title: string
  desc: string
  action: string
  target: PrimaryNavId
}[] = [
  {
    step: '01',
    title: '定义项目任务',
    desc: '把灵感写成目标、用户、约束和成功信号，避免 AI 只是在泛泛聊天。',
    action: '进入项目',
    target: 'projects'
  },
  {
    step: '02',
    title: '跑机会与证据',
    desc: '用探索模块扫描用户、竞品、风险和证据缺口，先确认值得做什么。',
    action: '开始探索',
    target: 'explore'
  },
  {
    step: '03',
    title: '沉淀 PRD 与蓝图',
    desc: '把需求、MVP、架构和验收门禁变成后续多智能体执行的事实源。',
    action: '进入产品',
    target: 'product'
  },
  {
    step: '04',
    title: '交给智能体执行',
    desc: '选择模型、Agent、Skill、Tool 和 MCP，让任务在可审计的管线里推进。',
    action: '配置资源',
    target: 'resources'
  }
]

const pipelineCards = computed(() => [
  {
    title: '当前项目',
    value: activeProjectTitle.value,
    hint: activeProjectStatus.value,
    icon: Rocket
  },
  {
    title: '运行时',
    value: appShell.runtimePing.headline,
    hint:
      appShell.runtimePing.errorCode === '暂无'
        ? 'typed preload 已就绪'
        : appShell.runtimePing.errorCode,
    icon: ShieldCheck
  },
  {
    title: '上下文预算',
    value: `${contextUsedPercent.value}%`,
    hint: `${appShell.chatContextBudget.estimatedTokens} / ${appShell.chatContextBudget.contextWindow || 0} tokens`,
    icon: Gauge
  }
])

const tokenTrend = computed(() => {
  const usageRows = appShell.chatMessages
    .map((message, index) => ({
      name: `M${index + 1}`,
      input: message.usage?.inputTokens ?? 0,
      output: message.usage?.outputTokens ?? 0
    }))
    .filter((row) => row.input + row.output > 0)
    .slice(-6)

  if (usageRows.length > 0) {
    return usageRows
  }

  const estimatedTokens = totalUsage.value.totalTokens || appShell.chatContextBudget.estimatedTokens
  const base = Math.max(estimatedTokens, 1600)
  return [
    { name: '定位', input: Math.round(base * 0.12), output: Math.round(base * 0.07) },
    { name: '探索', input: Math.round(base * 0.18), output: Math.round(base * 0.1) },
    { name: '产品', input: Math.round(base * 0.2), output: Math.round(base * 0.13) },
    { name: '开发', input: Math.round(base * 0.24), output: Math.round(base * 0.16) },
    { name: '交付', input: Math.round(base * 0.17), output: Math.round(base * 0.11) }
  ]
})

const capabilityMix = computed(() => {
  const rows = [
    { name: 'Agent', value: Math.max(appShell.agents.length, PLANNED_AGENT_COUNT) },
    { name: 'Skill', value: Math.max(appShell.skills.length, PLANNED_SKILL_COUNT) },
    { name: 'Tool', value: Math.max(appShell.tools.length, PLANNED_TOOL_COUNT) },
    { name: 'MCP', value: appShell.mcpServers.length }
  ]
  return rows.some((row) => row.value > 0) ? rows : [{ name: '待配置', value: 1 }]
})

const capabilityCards = computed(() => [
  {
    icon: Bot,
    value: `${Math.max(appShell.agents.length, PLANNED_AGENT_COUNT)}`,
    label: '规划 Agent',
    detail: '机会、用户、产品、架构、开发、评估、发布等角色'
  },
  {
    icon: Workflow,
    value: `${Math.max(appShell.skills.length, PLANNED_SKILL_COUNT)}+`,
    label: '项目 Skill',
    detail: 'PRD、原型、架构、任务切片、测试、复盘等流程'
  },
  {
    icon: Boxes,
    value: `${Math.max(appShell.tools.length, PLANNED_TOOL_COUNT)}+`,
    label: '工具能力',
    detail: '搜索、文件、Git、Shell、测试、截图、路由和 MCP'
  }
])

const capabilityLanes = [
  {
    label: 'Agent 队列',
    items: ['机会侦察员', '产品设计师', '系统架构师', '开发编排员']
  },
  {
    label: 'Skill 库',
    items: ['机会扫描', 'PRD 门禁', '架构蓝图', '发布复盘']
  },
  {
    label: 'Tool 工具箱',
    items: ['Web 搜索', '文件读写', 'Git / Shell', '模型路由']
  }
]

const providerModelBars = computed(() => {
  const rows = enabledProviders.value.slice(0, 6).map((provider) => ({
    name: provider.displayName,
    value: provider.modelCount || provider.availableModels.length
  }))
  if (rows.length > 0) {
    return rows
  }
  return [{ name: '等待接入', value: appShell.nineRouterStatus?.modelCount ?? 0 }]
})

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') {
    return fallback
  }
  const rootValue = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  const bodyValue = getComputedStyle(document.body).getPropertyValue(name).trim()
  return rootValue || bodyValue || fallback
}

function renderCharts(): void {
  const colors = {
    text: cssVar('--text', '#111827'),
    muted: cssVar('--muted', '#64748b'),
    line: 'rgba(148, 163, 184, 0.22)',
    purple: cssVar('--purple', '#7c3aed'),
    purpleSoft: 'rgba(124, 58, 237, 0.16)',
    blue: cssVar('--blue', '#2563eb'),
    cyan: cssVar('--cyan', '#0891b2'),
    green: cssVar('--green', '#16a34a'),
    amber: cssVar('--amber', '#d97706')
  }

  tokenChart?.setOption({
    color: [colors.purple, colors.cyan],
    grid: { top: 28, right: 10, bottom: 24, left: 36 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: colors.line,
      textStyle: { color: colors.text }
    },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 8,
      itemHeight: 8,
      textStyle: { color: colors.muted, fontSize: 11 }
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: tokenTrend.value.map((row) => row.name),
      axisTick: { show: false },
      axisLine: { lineStyle: { color: colors.line } },
      axisLabel: { color: colors.muted, fontSize: 10 }
    },
    yAxis: {
      type: 'value',
      splitNumber: 3,
      axisLabel: { color: colors.muted, fontSize: 10 },
      splitLine: { lineStyle: { color: colors.line } }
    },
    series: [
      {
        name: '输入',
        type: 'line',
        smooth: true,
        symbolSize: 6,
        areaStyle: { color: colors.purpleSoft },
        lineStyle: { width: 3 },
        data: tokenTrend.value.map((row) => row.input)
      },
      {
        name: '输出',
        type: 'line',
        smooth: true,
        symbolSize: 6,
        areaStyle: { color: 'rgba(8, 145, 178, 0.13)' },
        lineStyle: { width: 3 },
        data: tokenTrend.value.map((row) => row.output)
      }
    ]
  })

  capabilityChart?.setOption({
    color: [colors.purple, colors.blue, colors.green, colors.amber],
    tooltip: {
      trigger: 'item',
      confine: true,
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: colors.line,
      textStyle: { color: colors.text }
    },
    series: [
      {
        name: '能力',
        type: 'pie',
        radius: ['56%', '78%'],
        center: ['50%', '52%'],
        avoidLabelOverlap: true,
        label: {
          color: colors.muted,
          fontSize: 10,
          fontWeight: 700,
          formatter: '{b}\n{c}'
        },
        labelLine: { length: 8, length2: 6 },
        data: capabilityMix.value
      }
    ]
  })

  providerChart?.setOption({
    color: [colors.blue],
    grid: { top: 8, right: 30, bottom: 8, left: 86 },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      confine: true,
      backgroundColor: 'rgba(255, 255, 255, 0.96)',
      borderColor: colors.line,
      textStyle: { color: colors.text }
    },
    xAxis: {
      type: 'value',
      axisLabel: { show: false },
      axisTick: { show: false },
      axisLine: { show: false },
      splitLine: { lineStyle: { color: colors.line } }
    },
    yAxis: {
      type: 'category',
      data: providerModelBars.value.map((row) => row.name),
      axisTick: { show: false },
      axisLine: { show: false },
      axisLabel: { color: colors.muted, fontSize: 10, width: 78, overflow: 'truncate' }
    },
    series: [
      {
        name: '模型数',
        type: 'bar',
        barWidth: 12,
        label: { show: true, position: 'right', color: colors.text, fontWeight: 800 },
        itemStyle: {
          borderRadius: [0, 999, 999, 0],
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 1,
            y2: 0,
            colorStops: [
              { offset: 0, color: colors.purple },
              { offset: 1, color: colors.cyan }
            ]
          }
        },
        data: providerModelBars.value.map((row) => row.value)
      }
    ]
  })
}

function resizeCharts(): void {
  tokenChart?.resize()
  capabilityChart?.resize()
  providerChart?.resize()
}

function initializeCharts(): void {
  if (tokenChartRef.value && !tokenChart) {
    tokenChart = echarts.init(tokenChartRef.value)
  }
  if (capabilityChartRef.value && !capabilityChart) {
    capabilityChart = echarts.init(capabilityChartRef.value)
  }
  if (providerChartRef.value && !providerChart) {
    providerChart = echarts.init(providerChartRef.value)
  }

  chartResizeObserver = new ResizeObserver(resizeCharts)
  for (const element of [tokenChartRef.value, capabilityChartRef.value, providerChartRef.value]) {
    if (element) {
      chartResizeObserver.observe(element)
    }
  }
  renderCharts()
}

onMounted(() => {
  void nextTick(initializeCharts)
})

onBeforeUnmount(() => {
  chartResizeObserver?.disconnect()
  tokenChart?.dispose()
  capabilityChart?.dispose()
  providerChart?.dispose()
})

watch([tokenTrend, capabilityMix, providerModelBars], () => {
  void nextTick(renderCharts)
})

function openPrimary(target: PrimaryNavId): void {
  appShell.setPrimary(target)
}

function openProviders(): void {
  appShell.setPrimary('resources')
  appShell.setResourceTab('providers')
}

async function downloadWindowsPreview(): Promise<void> {
  try {
    const opened = await window.dreamworker?.system.openExternal(WINDOWS_PREVIEW_DOWNLOAD_URL)
    if (opened?.ok) {
      return
    }
  } catch {
    // Browser preview mode does not have Electron's typed preload bridge.
  }
  window.open(WINDOWS_PREVIEW_DOWNLOAD_URL, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <section class="home-workspace" aria-label="DreamWorker 首页">
    <section class="home-hero panel-surface">
      <div class="home-hero-copy">
        <p class="eyebrow">
          <Sparkles :size="14" aria-hidden="true" />
          AI OS 项目驾驶舱
        </p>
        <h2>把一个想法推进成能发布的项目。</h2>
        <div class="home-hero-actions">
          <button class="primary-button" type="button" @click="openPrimary('explore')">
            <Play :size="16" aria-hidden="true" />
            启动项目孵化
          </button>
          <button class="icon-text-button" type="button" @click="downloadWindowsPreview">
            <Download :size="16" aria-hidden="true" />
            Windows 体验版
          </button>
          <button class="icon-text-button" type="button" @click="openProviders">
            <Route :size="16" aria-hidden="true" />
            管理模型路由
          </button>
        </div>
      </div>

      <div class="home-orbit-console" aria-label="模型与运行时状态">
        <div class="orbit-core" :data-level="runtimeLevel">
          <BrainCircuit :size="44" aria-hidden="true" />
          <strong>AI OS</strong>
          <span>{{ connectedProviders.length }} / {{ appShell.providers.length }} 服务商在线</span>
        </div>
        <div class="orbit-chip chip-one">
          <Bot :size="16" aria-hidden="true" />
          {{ PLANNED_AGENT_COUNT }} Agents
        </div>
        <div class="orbit-chip chip-two">
          <Workflow :size="16" aria-hidden="true" />
          {{ PLANNED_SKILL_COUNT }}+ Skills
        </div>
        <div class="orbit-chip chip-three">
          <Zap :size="16" aria-hidden="true" />
          {{ PLANNED_TOOL_COUNT }}+ Tools
        </div>
      </div>
    </section>

    <section class="home-metric-grid" aria-label="首页指标">
      <article
        v-for="card in metricCards"
        :key="card.label"
        class="home-metric-card"
        :data-tone="card.tone"
      >
        <span>{{ card.label }}</span>
        <strong>{{ card.value.toLocaleString() }}</strong>
        <small>{{ card.detail }}</small>
      </article>
    </section>

    <section class="home-main-grid">
      <div class="home-left-stack">
        <div class="home-guide-panel panel-surface">
          <div class="panel-heading compact">
            <div>
              <p class="eyebrow">项目指引</p>
              <h2>今天从哪一步开始</h2>
            </div>
            <button
              class="icon-button"
              type="button"
              title="打开聊天工作台"
              @click="openPrimary('chat')"
            >
              <MessageSquareText :size="18" aria-hidden="true" />
            </button>
          </div>

          <div class="home-guide-list">
            <article v-for="card in guideCards" :key="card.step" class="home-guide-card">
              <span>{{ card.step }}</span>
              <div>
                <h3>{{ card.title }}</h3>
                <p>{{ card.desc }}</p>
              </div>
              <button type="button" @click="openPrimary(card.target)">
                {{ card.action }}
                <ArrowRight :size="14" aria-hidden="true" />
              </button>
            </article>
          </div>
        </div>

        <div class="home-agent-panel panel-surface">
          <div class="panel-heading compact">
            <div>
              <p class="eyebrow">能力编排</p>
              <h2>Agent / Skill / Tool</h2>
            </div>
            <Boxes :size="20" aria-hidden="true" />
          </div>

          <div class="home-agent-analytics">
            <div class="home-capability-grid">
              <article v-for="card in capabilityCards" :key="card.label">
                <component :is="card.icon" :size="18" aria-hidden="true" />
                <strong>{{ card.value }}</strong>
                <span>{{ card.label }}</span>
                <small>{{ card.detail }}</small>
              </article>
            </div>

            <div class="home-chart-card">
              <div class="home-chart-heading">
                <span>能力占比</span>
                <small>编排资源</small>
              </div>
              <div
                ref="capabilityChartRef"
                class="home-echart home-echart-donut"
                role="img"
                aria-label="能力编排占比图"
              />
            </div>
          </div>

          <div class="home-capability-lanes" aria-label="规划能力示例">
            <article v-for="lane in capabilityLanes" :key="lane.label">
              <strong>{{ lane.label }}</strong>
              <div>
                <span v-for="item in lane.items" :key="item">{{ item }}</span>
              </div>
            </article>
          </div>

          <div class="home-next-action">
            <CheckCircle2 :size="18" aria-hidden="true" />
            <p>
              {{
                appShell.activeModule?.nextBestAction ||
                '选择一个项目模块，DreamWorker 会给出下一步动作。'
              }}
            </p>
            <button type="button" @click="openPrimary(appShell.activeModuleWorkspace ?? 'explore')">
              去执行
            </button>
          </div>
        </div>
      </div>

      <aside class="home-command-rail" aria-label="运行状态与模型图表">
        <div class="home-runtime-panel panel-surface">
          <div class="panel-heading compact">
            <div>
              <p class="eyebrow">运行管线</p>
              <h2>项目和上下文状态</h2>
            </div>
            <button
              class="icon-button"
              type="button"
              title="检查引擎"
              @click="appShell.checkRuntimePing()"
            >
              <ShieldCheck :size="18" aria-hidden="true" />
            </button>
          </div>

          <div class="home-pipeline-list">
            <article v-for="item in pipelineCards" :key="item.title">
              <component :is="item.icon" :size="20" aria-hidden="true" />
              <div>
                <span>{{ item.title }}</span>
                <strong>{{ item.value }}</strong>
                <small>{{ item.hint }}</small>
              </div>
            </article>
          </div>

          <div class="home-token-panel">
            <div>
              <span>Context</span>
              <strong>{{ contextUsedPercent }}%</strong>
            </div>
            <progress :value="contextUsedPercent" max="100" aria-label="上下文 token 使用率" />
            <p>
              本轮会话累计 {{ totalUsage.totalTokens.toLocaleString() }} tokens，估算成本 ${{
                totalUsage.costUsd.toFixed(4)
              }}。
            </p>
          </div>

          <div class="home-chart-card">
            <div class="home-chart-heading">
              <span>Token 趋势</span>
              <small>输入 / 输出</small>
            </div>
            <div ref="tokenChartRef" class="home-echart" role="img" aria-label="Token 使用趋势图" />
          </div>
        </div>

        <div class="home-model-panel panel-surface">
          <div class="panel-heading compact">
            <div>
              <p class="eyebrow">可用模型</p>
              <h2>路由与模型清单</h2>
            </div>
            <Network :size="20" aria-hidden="true" />
          </div>

          <div class="home-model-marquee">
            <span v-for="model in modelTicker" :key="model">{{ model }}</span>
            <span v-if="modelTicker.length === 0">等待模型发现</span>
          </div>

          <div class="home-chart-card">
            <div class="home-chart-heading">
              <span>模型分布</span>
              <small>服务商模型数</small>
            </div>
            <div
              ref="providerChartRef"
              class="home-echart home-echart-bars"
              role="img"
              aria-label="服务商模型数量图"
            />
          </div>

          <div class="home-provider-list">
            <article v-for="provider in enabledProviders.slice(0, 5)" :key="provider.providerId">
              <div>
                <strong>{{ provider.displayName }}</strong>
                <small>{{
                  provider.defaultModel || provider.availableModels[0] || '未设置默认模型'
                }}</small>
              </div>
              <span :data-status="provider.healthStatus">{{ provider.modelCount }}</span>
            </article>
          </div>
        </div>
      </aside>
    </section>
  </section>
</template>
