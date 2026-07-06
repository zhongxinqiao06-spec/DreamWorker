<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { Component } from 'vue'
import {
  ArrowRight,
  BadgeCheck,
  BrainCircuit,
  Boxes,
  ChartNoAxesCombined,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  CircuitBoard,
  Code2,
  Cpu,
  Download,
  GitBranch,
  Globe2,
  Layers3,
  Lightbulb,
  MonitorCog,
  Network,
  Puzzle,
  Rocket,
  Route,
  ScanSearch,
  ShieldCheck,
  Sparkles,
  Terminal,
  UsersRound,
  Workflow,
  Wrench,
  Zap
} from 'lucide-vue-next'

type IconComponent = Component
const releaseTag = 'v0.1.0-preview.1'
const windowsDownloadUrl = `https://github.com/zhongxinqiao06-spec/DreamWorker/releases/download/${releaseTag}/DreamWorker.Setup.0.1.0.exe`

const navItems = [
  { label: '首页', href: '#home' },
  { label: '定位', href: '#position' },
  { label: '工作流', href: '#workflow' },
  { label: '能力矩阵', href: '#capabilities' },
  { label: '对比', href: '#compare' },
  { label: '架构', href: '#architecture' },
  { label: '9Router', href: '#router' }
]
const topNavSectionIds = navItems.map((item) => item.href.slice(1))
const activeTopNavIndex = ref(0)
const topNavDirection = ref(1)
const topNavMotionClass = computed(() =>
  topNavDirection.value >= 0 ? 'page-motion-down' : 'page-motion-up'
)
let topNavObserver: IntersectionObserver | null = null
let topNavWheelLocked = false
let topNavWheelUnlockTimer: number | undefined
const topNavWheelOptions: AddEventListenerOptions = { passive: false }

const heroStats = [
  { value: '12', label: '规划内置 Agent' },
  { value: '12+', label: '项目推进 Skill' },
  { value: '20+', label: '工具与 MCP 能力' }
]

const positionCards: Array<{
  icon: IconComponent
  title: string
  text: string
}> = [
  {
    icon: Lightbulb,
    title: '为有想法的人准备',
    text: '从一句“我想做个项目”开始，把目标、假设、竞品、PRD、蓝图、开发、发布和复盘拆成连续动作。'
  },
  {
    icon: Workflow,
    title: '工作流不是提示词合集',
    text: '把真实项目推进方式沉淀为软件内 Workflow 与 Skill，让 AI 知道下一步该产出什么、如何检查、何时进入下个阶段。'
  },
  {
    icon: UsersRound,
    title: '多智能体协作有上下文',
    text: '后续支持用户自定义 Skill、编排多智能体，并把项目事实、资源、决策和产物持续留在同一个工作空间。'
  }
]

const moduleFlowStages: Array<{
  step: string
  module: string
  icon: IconComponent
  title: string
  desc: string
  image: string
  imageAlt: string
  outputs: string[]
}> = [
  {
    step: '01',
    module: '项目配置',
    icon: Boxes,
    title: '项目空间',
    desc: '项目名称、描述、本地目录、资源绑定、模块配置和运行策略统一落到 projectId。',
    image: '/images/module-project-config.png',
    imageAlt: 'DreamWorker 项目配置真实页面截图',
    outputs: ['project_001', 'module config', 'security policy']
  },
  {
    step: '02',
    module: '资源中心',
    icon: Network,
    title: '模型、Agent、Skill、Tool 和 MCP',
    desc: '统一管理工作台资源，模型服务商、Agent、Skill、工具和 MCP 都可在资源中心配置。',
    image: '/images/module-resource-center.png',
    imageAlt: 'DreamWorker 资源配置中心真实页面截图',
    outputs: ['4 providers', '12 Agents', '7 Skills', '6 Tools']
  },
  {
    step: '03',
    module: 'Agent 聊天',
    icon: UsersRound,
    title: '普通 Agent 工作台',
    desc: '围绕项目上下文选择 Agent、服务商、模型和项目，让日常问答能进入项目空间。',
    image: '/images/module-chat-agent.png',
    imageAlt: 'DreamWorker 普通 Agent 聊天工作台真实页面截图',
    outputs: ['context', 'token trace', 'runtime panel']
  },
  {
    step: '04',
    module: '探索模块',
    icon: ScanSearch,
    title: '机会雷达',
    desc: '扫描用户痛点、市场窗口和可验证机会，运行链路直接产出探索文档。',
    image: '/images/module-explore-opportunity.png',
    imageAlt: 'DreamWorker 探索模块机会雷达真实页面截图',
    outputs: ['dream_brief.md', 'hypotheses.yaml']
  },
  {
    step: '05',
    module: '产品模块',
    icon: BadgeCheck,
    title: '需求分析',
    desc: '根据探索结果或用户上传的项目要求文件，抽取功能清单并生成需求规格说明。',
    image: '/images/module-product-requirements.png',
    imageAlt: 'DreamWorker 产品模块需求分析真实页面截图',
    outputs: ['feature_list.xlsx', 'requirements_spec.docx', 'requirements_analysis.json']
  },
  {
    step: '06',
    module: '开发模块',
    icon: Code2,
    title: '编码 Agent',
    desc: '内置 Claude Agent、Codex、OpenCode 三种 SDK，提供文件树、编码对话和直接写入。',
    image: '/images/module-development-coding.png',
    imageAlt: 'DreamWorker 开发模块编码 Agent 真实页面截图',
    outputs: ['Claude Agent', 'Codex', 'OpenCode']
  }
]

const capabilityPages: Array<{
  id: string
  label: string
  icon: IconComponent
  badge: string
  title: string
  desc: string
  image: string
  imageAlt: string
  metrics: Array<{ value: string; label: string }>
  highlights: Array<{ title: string; text: string }>
}> = [
  {
    id: 'modules',
    label: '模块',
    icon: Boxes,
    badge: '已落地项目骨架',
    title: '四个项目模块，把流程固定在产品里',
    desc: '当前已经落下探索、产品、开发、销售四个模块，每个模块都有子模块、默认 Agent、Skill、Tool、产物和下一步动作。',
    image: '/images/module-project-config.png',
    imageAlt: 'DreamWorker 项目配置真实页面截图',
    metrics: [
      { value: '4', label: '项目模块' },
      { value: '17', label: '子模块' },
      { value: 'ready', label: '可继续扩展' }
    ],
    highlights: [
      { title: '探索模块', text: '机会雷达、用户画像、竞品地图、证据图谱。' },
      { title: '产品模块', text: '需求分析、PRD 草案、原型说明、蓝图画布。' },
      { title: '开发模块', text: '技术架构、成本评估、PR 拆分、测试门禁、编码 Agent。' },
      { title: '销售模块', text: '定位文案、落地页、发布计划、反馈循环。' }
    ]
  },
  {
    id: 'agents',
    label: 'Agent',
    icon: UsersRound,
    badge: '角色协作',
    title: 'Agent 按真实项目角色分工',
    desc: '不是让一个聊天机器人包办全部事情，而是让机会、产品、架构、开发、评估、销售角色围绕同一个项目上下文接力。',
    image: '/images/module-chat-agent.png',
    imageAlt: 'DreamWorker 普通 Agent 聊天工作台截图',
    metrics: [
      { value: '12', label: '规划角色' },
      { value: '11+', label: '模块默认挂载' },
      { value: '1', label: '通用助手入口' }
    ],
    highlights: [
      { title: '探索侧', text: '机会侦察、竞品分析、客群细分。' },
      { title: '产品侧', text: '产品设计、原型设计、质量评估。' },
      { title: '开发侧', text: '系统架构、技术栈顾问、开发编排。' },
      { title: '发布侧', text: '销售策略、演示设计、反馈复盘。' }
    ]
  },
  {
    id: 'skills',
    label: 'Skill',
    icon: Workflow,
    badge: '可复用工作流',
    title: 'Skill 负责把阶段动作标准化',
    desc: '已落地的 Skill 先覆盖机会扫描、竞品地图、PRD 草案、蓝图和发布计划，后续继续扩成用户可自定义的工作流包。',
    image: '/images/module-explore-opportunity.png',
    imageAlt: 'DreamWorker 探索模块机会雷达截图',
    metrics: [
      { value: '5', label: '已落地核心 Skill' },
      { value: '12+', label: '规划扩展' },
      { value: 'SOP', label: '后续可自定义' }
    ],
    highlights: [
      { title: 'skill_opportunity_scan', text: '从想法生成机会、用户痛点和证据缺口。' },
      { title: 'skill_competitor_map', text: '把竞品、替代方案和定位差异整理成判断。' },
      { title: 'skill_prd_draft', text: '从探索结论和需求文件进入 PRD 与规格说明。' },
      { title: 'skill_blueprint / launch_plan', text: '把工程蓝图和发布动作接到后续模块。' }
    ]
  },
  {
    id: 'tools',
    label: 'Tool',
    icon: Wrench,
    badge: '真实执行层',
    title: 'Tool 负责查、写、跑和交付',
    desc: '当前工具先覆盖模型生成、Web 搜索、产物写入、人类输入、编码 Agent 文件树与三引擎 SDK，后续继续接 MCP 和插件工具。',
    image: '/images/module-resource-center.png',
    imageAlt: 'DreamWorker 资源配置中心截图',
    metrics: [
      { value: '20+', label: '工具规划' },
      { value: 'MCP', label: '可扩展入口' },
      { value: '3', label: '编码 SDK 引擎' }
    ],
    highlights: [
      { title: '研究工具', text: 'Web 搜索、模型生成、证据整理。' },
      { title: '产物工具', text: 'Markdown、Excel、Word、JSON、YAML 写入。' },
      { title: '工程工具', text: '文件树、直接写入、Shell、测试和包管理规划。' },
      { title: '扩展工具', text: 'MCP 发现、模型路由、插件化 SDK 导入。' }
    ]
  }
]

const activeWorkflowIndex = ref(0)
const workflowDirection = ref(1)
const activeWorkflowStage = computed(
  () => moduleFlowStages[activeWorkflowIndex.value] ?? moduleFlowStages[0]
)
const workflowTransitionName = computed(() =>
  workflowDirection.value >= 0 ? 'workflow-next' : 'workflow-prev'
)
const activeCapabilityIndex = ref(0)
const capabilityDirection = ref(1)
const capabilityAutoplayPaused = ref(false)
let capabilityAutoplayTimer: number | undefined
const activeCapabilityPage = computed(
  () => capabilityPages[activeCapabilityIndex.value] ?? capabilityPages[0]
)
const capabilityTransitionName = computed(() =>
  capabilityDirection.value >= 0 ? 'capability-next' : 'capability-prev'
)

function showWorkflowStage(index: number): void {
  if (index === activeWorkflowIndex.value) {
    return
  }
  workflowDirection.value = index > activeWorkflowIndex.value ? 1 : -1
  activeWorkflowIndex.value = (index + moduleFlowStages.length) % moduleFlowStages.length
}

function shiftWorkflowStage(offset: number): void {
  workflowDirection.value = offset >= 0 ? 1 : -1
  activeWorkflowIndex.value =
    (activeWorkflowIndex.value + offset + moduleFlowStages.length) % moduleFlowStages.length
}

function handleWorkflowKeydown(event: KeyboardEvent): void {
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    shiftWorkflowStage(1)
  }
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    shiftWorkflowStage(-1)
  }
  if (event.key === 'Home') {
    event.preventDefault()
    showWorkflowStage(0)
  }
  if (event.key === 'End') {
    event.preventDefault()
    showWorkflowStage(moduleFlowStages.length - 1)
  }
}

function showCapabilityPage(index: number, resetAutoplay = false): void {
  if (index === activeCapabilityIndex.value) {
    return
  }
  capabilityDirection.value = index > activeCapabilityIndex.value ? 1 : -1
  activeCapabilityIndex.value = index
  if (resetAutoplay) {
    startCapabilityAutoplay()
  }
}

function shiftCapabilityPage(offset: number, resetAutoplay = false): void {
  capabilityDirection.value = offset >= 0 ? 1 : -1
  activeCapabilityIndex.value =
    (activeCapabilityIndex.value + offset + capabilityPages.length) % capabilityPages.length
  if (resetAutoplay) {
    startCapabilityAutoplay()
  }
}

function pauseCapabilityAutoplay(): void {
  capabilityAutoplayPaused.value = true
}

function resumeCapabilityAutoplay(): void {
  capabilityAutoplayPaused.value = false
}

function startCapabilityAutoplay(): void {
  window.clearInterval(capabilityAutoplayTimer)
  capabilityAutoplayTimer = window.setInterval(() => {
    if (!capabilityAutoplayPaused.value) {
      shiftCapabilityPage(1)
    }
  }, 5200)
}

function handleCapabilityKeydown(event: KeyboardEvent): void {
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    shiftCapabilityPage(1, true)
  }
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    shiftCapabilityPage(-1, true)
  }
  if (event.key === 'Home') {
    event.preventDefault()
    showCapabilityPage(0, true)
  }
  if (event.key === 'End') {
    event.preventDefault()
    showCapabilityPage(capabilityPages.length - 1, true)
  }
}

function normalizeTopNavIndex(index: number): number {
  return (index + navItems.length) % navItems.length
}

function focusTopNavTab(index: number): void {
  const tab = document.getElementById(`top-nav-tab-${index}`)
  tab?.focus({ preventScroll: true })
}

function showTopNavSection(index: number): void {
  const nextIndex = normalizeTopNavIndex(index)
  topNavDirection.value = index >= activeTopNavIndex.value ? 1 : -1
  activeTopNavIndex.value = nextIndex
  const section = document.getElementById(topNavSectionIds[nextIndex] ?? '')
  section?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  focusTopNavTab(nextIndex)
}

function handleTopNavClick(index: number): void {
  topNavDirection.value = index >= activeTopNavIndex.value ? 1 : -1
  activeTopNavIndex.value = index
}

function handleTopNavKeydown(event: KeyboardEvent): void {
  if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
    event.preventDefault()
    showTopNavSection(activeTopNavIndex.value + 1)
  }
  if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
    event.preventDefault()
    showTopNavSection(activeTopNavIndex.value - 1)
  }
  if (event.key === 'Home') {
    event.preventDefault()
    showTopNavSection(0)
  }
  if (event.key === 'End') {
    event.preventDefault()
    showTopNavSection(navItems.length - 1)
  }
}

function isTypingTarget(target: EventTarget | null): boolean {
  const element = target instanceof HTMLElement ? target : null
  if (!element) {
    return false
  }
  const tagName = element.tagName.toLowerCase()
  return (
    element.isContentEditable ||
    tagName === 'input' ||
    tagName === 'textarea' ||
    tagName === 'select'
  )
}

function handleGlobalTopNavKeydown(event: KeyboardEvent): void {
  if (
    event.defaultPrevented ||
    event.altKey ||
    event.ctrlKey ||
    event.metaKey ||
    isTypingTarget(event.target)
  ) {
    return
  }
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    showTopNavSection(activeTopNavIndex.value + 1)
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    showTopNavSection(activeTopNavIndex.value - 1)
  }
}

function unlockTopNavWheel(): void {
  topNavWheelLocked = false
}

function handleGlobalTopNavWheel(event: WheelEvent): void {
  if (
    event.defaultPrevented ||
    event.altKey ||
    event.ctrlKey ||
    event.metaKey ||
    isTypingTarget(event.target)
  ) {
    return
  }
  if (Math.abs(event.deltaY) < 18 || Math.abs(event.deltaY) < Math.abs(event.deltaX)) {
    return
  }

  event.preventDefault()
  if (topNavWheelLocked) {
    return
  }

  topNavWheelLocked = true
  showTopNavSection(activeTopNavIndex.value + (event.deltaY > 0 ? 1 : -1))
  window.clearTimeout(topNavWheelUnlockTimer)
  topNavWheelUnlockTimer = window.setTimeout(unlockTopNavWheel, 760)
}

function observeTopNavSections(): void {
  if (typeof IntersectionObserver === 'undefined') {
    return
  }
  const sections = topNavSectionIds
    .map((id) => document.getElementById(id))
    .filter((section): section is HTMLElement => Boolean(section))

  topNavObserver = new IntersectionObserver(
    (entries) => {
      const activeEntry = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0]
      if (!activeEntry) {
        return
      }
      const nextIndex = topNavSectionIds.indexOf(activeEntry.target.id)
      if (nextIndex >= 0) {
        activeTopNavIndex.value = nextIndex
      }
    },
    {
      rootMargin: '-28% 0px -58% 0px',
      threshold: [0.15, 0.35, 0.6]
    }
  )

  for (const section of sections) {
    topNavObserver.observe(section)
  }
}

onMounted(() => {
  startCapabilityAutoplay()
  observeTopNavSections()
  window.addEventListener('keydown', handleGlobalTopNavKeydown)
  window.addEventListener('wheel', handleGlobalTopNavWheel, topNavWheelOptions)
})

onBeforeUnmount(() => {
  window.clearInterval(capabilityAutoplayTimer)
  window.clearTimeout(topNavWheelUnlockTimer)
  topNavObserver?.disconnect()
  window.removeEventListener('keydown', handleGlobalTopNavKeydown)
  window.removeEventListener('wheel', handleGlobalTopNavWheel, topNavWheelOptions)
})

const competitorRows = [
  {
    product: 'Cherry Studio',
    positioning: '全能 AI 工作站：多模型对话、知识库、绘图、翻译、MCP 与高度自定义。',
    dreamworker:
      'DreamWorker 更聚焦“做项目”：把机会分析、PRD、架构、开发、评估、发布做成连续工作流和可复用 Skill。'
  },
  {
    product: 'WorkBuddy',
    positioning: '场景化办公 AI 套件：专家团队、技能市场、办公生态连接和远程执行。',
    dreamworker:
      'DreamWorker 更适合创造者与开发团队：本地优先桌面 AI OS、项目级上下文、可编排多智能体和开放插件运行时。'
  },
  {
    product: '通用 Agent Builder',
    positioning: '提供可视化编排或聊天式自动化，但常缺少真实项目生命周期约束。',
    dreamworker: 'DreamWorker 把流程、产物、门禁和资源绑定到项目空间，减少“会聊但不落地”的断层。'
  }
]

const architectureCards: Array<{
  icon: IconComponent
  title: string
  details: string[]
}> = [
  {
    icon: MonitorCog,
    title: 'Electron 多端桌面壳',
    details: [
      'Windows / macOS / Linux 一套桌面体验',
      'Main / Preload / Renderer 分层隔离',
      '本地 Runtime 能力不暴露给 UI'
    ]
  },
  {
    icon: Layers3,
    title: 'Vue + Vite 前端',
    details: [
      '贴近国内开发生态，招人和维护成本更友好',
      '组件清晰、状态可读、迭代速度快',
      '适合长期沉淀工作台与资源中心'
    ]
  },
  {
    icon: Cpu,
    title: 'Node Main Runtime 本地运行时',
    details: [
      'TypeScript 服务层随桌面主包发布',
      '按 bootstrap、router、kernel、services、store 分层',
      '生产路径不再依赖 Go Engine 或本机 HTTP 中转'
    ]
  },
  {
    icon: Puzzle,
    title: '内置 SDK 能力层',
    details: [
      'Claude Agent、Codex、OpenCode SDK 随安装包预置',
      'Coding Agent 固化为 Runtime service，文件根限定 workspace/code',
      'OpenCode server、session、event 和 auth 由 Node Runtime 托管'
    ]
  }
]

const routerFeatures = [
  {
    icon: Route,
    title: '一个 OpenAI-compatible 端点',
    desc: '把 Codex CLI、Claude Code、Cursor、Cline 等工具指向本地路由。'
  },
  {
    icon: Network,
    title: '免费 / 低成本 / 订阅三层路由',
    desc: '按可用性、成本和额度做 fallback，减少频繁换 key 和换模型。'
  },
  {
    icon: Zap,
    title: '工具输出 token 优化',
    desc: '9Router 的 RTK / Caveman 能压缩工具结果和回答冗余，适合编码场景。'
  },
  {
    icon: ShieldCheck,
    title: '先本地接入再扩展',
    desc: 'DreamWorker 首批把 9Router 作为免费路由入口，后续沉淀为资源中心的一键配置。'
  }
]

const roadmapItems = [
  '用户自定义工作流：从内置 Skill 走向个人方法论和团队 SOP。',
  'Skill 市场与安装器：把可复用能力沉淀为版本化包。',
  '多智能体编排：产品、研发、评估、增长角色围绕项目产物协同。',
  '项目知识图谱：让需求、证据、决策、代码和发布反馈形成持续记忆。',
  '团队工作区：从本地个人 AI OS 扩展到可审计的团队协作。'
]

const sourceLinks = [
  { label: 'Cherry Studio 项目简介', url: 'https://docs.cherry-ai.com/docs/zhong-wen-fan-ti' },
  { label: 'WorkBuddy 官方页面', url: 'https://www.tencentcloud.com/act/pro/workbuddy' },
  { label: '腾讯 WorkBuddy 发布报道', url: 'https://www.tencent.com/en-us/articles/2202350.html' },
  { label: '9Router 官方网站', url: 'https://9router.com/' }
]
</script>

<template>
  <div class="site-shell" :class="topNavMotionClass">
    <header class="top-nav">
      <a class="brand" href="#home" aria-label="DreamWorker 首页">
        <img src="/images/brand-mark.png" alt="" />
        <span>DreamWorker</span>
      </a>
      <nav role="tablist" aria-label="主导航" @keydown="handleTopNavKeydown">
        <a
          v-for="(item, index) in navItems"
          :id="`top-nav-tab-${index}`"
          :key="item.href"
          :href="item.href"
          role="tab"
          :aria-selected="index === activeTopNavIndex"
          :aria-controls="topNavSectionIds[index]"
          :tabindex="index === activeTopNavIndex ? 0 : -1"
          @click="handleTopNavClick(index)"
        >
          {{ item.label }}
        </a>
      </nav>
      <a class="nav-action" :href="windowsDownloadUrl" target="_blank" rel="noreferrer">
        <Download :size="17" />
        <span>Windows 体验版</span>
      </a>
    </header>

    <main id="top">
      <section
        id="home"
        class="hero-section nav-page-section hero-nav-page"
        :class="{ 'is-active-page': activeTopNavIndex === 0 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-0"
      >
        <img class="hero-orbit" src="/images/glass-orbit-hero.png" alt="" />
        <img
          class="hero-product"
          src="/images/dreamworker-home-screenshot.png"
          alt="DreamWorker AI 工作台首页真实界面截图"
        />
        <div class="hero-copy">
          <p class="eyebrow">
            <Sparkles :size="16" />
            本地优先的 AI OS + 项目孵化工作台
          </p>
          <h1>DreamWorker</h1>
          <p class="hero-lede">
            帮助每一个有梦想的人，把想做的项目从灵感推进到真实发布。不是再造一个聊天框，而是把真实做项目的流程做成软件内工作流、Skill
            和多智能体协作。
          </p>
          <div class="hero-actions">
            <a class="primary-link" :href="windowsDownloadUrl" target="_blank" rel="noreferrer">
              <Download :size="18" />
              <span>下载 Windows 体验版</span>
              <ArrowRight :size="17" />
            </a>
            <a class="secondary-link" href="#workflow">
              <Rocket :size="18" />
              <span>查看项目工作流</span>
            </a>
            <a class="secondary-link" href="#compare">
              <ChartNoAxesCombined :size="18" />
              <span>看竞品差异</span>
            </a>
          </div>
          <div class="hero-stats" aria-label="DreamWorker 概览数据">
            <article v-for="stat in heroStats" :key="stat.label">
              <strong>{{ stat.value }}</strong>
              <span>{{ stat.label }}</span>
            </article>
          </div>
        </div>
      </section>

      <section
        id="position"
        class="section section-tight nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 1 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-1"
      >
        <div class="section-heading">
          <p class="eyebrow">
            <BrainCircuit :size="16" />
            产品定位
          </p>
          <h2>把“会用 AI”升级成“能做成项目”</h2>
          <p>
            大多数 AI 工具让你更快获得回答，DreamWorker
            想解决更难的一步：让回答进入项目推进链路，变成可验证的计划、产物和下一步行动。
          </p>
        </div>
        <div class="position-grid">
          <article v-for="card in positionCards" :key="card.title" class="feature-card">
            <component :is="card.icon" :size="24" />
            <h3>{{ card.title }}</h3>
            <p>{{ card.text }}</p>
          </article>
        </div>
      </section>

      <section
        id="workflow"
        class="section workflow-section nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 2 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-2"
      >
        <div class="section-heading">
          <p class="eyebrow">
            <Workflow :size="16" />
            真实项目模块
          </p>
          <h2>从想法到发布，把软件流程交给用户</h2>
          <p>
            传统通用 Agent
            更像一个会聊天的助手：它能回答问题，但项目怎么拆、下一步做什么、产出如何验收，仍然要用户自己组织。
            DreamWorker
            的优势是把真实做项目的软件流程做进产品里，从机会探索、需求沉淀、PRD、架构、开发到发布复盘，
            让用户拿到一套可以一步步推进、可以复用、也可以继续扩展的工作流。
          </p>
        </div>
        <div class="workflow-carousel" @keydown="handleWorkflowKeydown">
          <div class="workflow-carousel-head">
            <div class="workflow-pager" aria-label="项目模块切换">
              <button type="button" aria-label="上一个模块" @click="shiftWorkflowStage(-1)">
                <ChevronLeft :size="18" />
              </button>
              <span>{{ activeWorkflowIndex + 1 }} / {{ moduleFlowStages.length }}</span>
              <button type="button" aria-label="下一个模块" @click="shiftWorkflowStage(1)">
                <ChevronRight :size="18" />
              </button>
            </div>
            <div class="workflow-dots" role="tablist" aria-label="项目模块页签">
              <button
                v-for="(stage, index) in moduleFlowStages"
                :id="`workflow-tab-${stage.step}`"
                :key="stage.step"
                type="button"
                role="tab"
                :aria-selected="index === activeWorkflowIndex"
                :aria-controls="`workflow-panel-${stage.step}`"
                :tabindex="index === activeWorkflowIndex ? 0 : -1"
                @click="showWorkflowStage(index)"
              >
                <span>{{ stage.step }}</span>
                {{ stage.module }}
              </button>
            </div>
          </div>

          <Transition :name="workflowTransitionName" mode="out-in">
            <article
              :id="`workflow-panel-${activeWorkflowStage.step}`"
              :key="activeWorkflowStage.step"
              class="workflow-page"
              role="tabpanel"
              :aria-labelledby="`workflow-tab-${activeWorkflowStage.step}`"
            >
              <div class="workflow-page-copy">
                <p class="eyebrow">
                  <component :is="activeWorkflowStage.icon" :size="16" />
                  {{ activeWorkflowStage.step }} / {{ activeWorkflowStage.module }}
                </p>
                <h3>{{ activeWorkflowStage.title }}</h3>
                <p>{{ activeWorkflowStage.desc }}</p>
                <div class="workflow-output-list workflow-page-outputs">
                  <small v-for="output in activeWorkflowStage.outputs" :key="output">{{
                    output
                  }}</small>
                </div>
              </div>

              <figure class="workflow-page-visual">
                <img :src="activeWorkflowStage.image" :alt="activeWorkflowStage.imageAlt" />
                <figcaption>{{ activeWorkflowStage.imageAlt }}</figcaption>
              </figure>
            </article>
          </Transition>
        </div>
      </section>

      <section
        id="capabilities"
        class="section capability-section nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 3 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-3"
      >
        <div
          class="capability-shell"
          @mouseenter="pauseCapabilityAutoplay"
          @mouseleave="resumeCapabilityAutoplay"
          @focusin="pauseCapabilityAutoplay"
          @focusout="resumeCapabilityAutoplay"
        >
          <div class="capability-shell-heading">
            <div class="section-heading">
              <p class="eyebrow">
                <UsersRound :size="16" />
                Agent / Skill / Tool
              </p>
              <h2>能力矩阵拆成页，按模块轻量浏览</h2>
              <p>一次只看一层能力：先看真实模块，再看 Agent、Skill、Tool 如何接力。</p>
            </div>
            <div class="capability-pager" aria-label="能力页切换">
              <button type="button" aria-label="上一页" @click="shiftCapabilityPage(-1, true)">
                <ChevronLeft :size="18" />
              </button>
              <span>{{ activeCapabilityIndex + 1 }} / {{ capabilityPages.length }}</span>
              <button type="button" aria-label="下一页" @click="shiftCapabilityPage(1, true)">
                <ChevronRight :size="18" />
              </button>
            </div>
          </div>

          <div
            class="capability-tabs"
            role="tablist"
            aria-label="能力页签"
            @keydown="handleCapabilityKeydown"
          >
            <button
              v-for="(page, index) in capabilityPages"
              :id="`capability-tab-${page.id}`"
              :key="page.id"
              type="button"
              role="tab"
              :aria-selected="index === activeCapabilityIndex"
              :aria-controls="`capability-panel-${page.id}`"
              :tabindex="index === activeCapabilityIndex ? 0 : -1"
              @click="showCapabilityPage(index, true)"
            >
              <component :is="page.icon" :size="16" />
              <span>{{ page.label }}</span>
            </button>
          </div>

          <div class="capability-autoplay-track" aria-hidden="true">
            <span
              :key="activeCapabilityPage.id"
              :class="{ paused: capabilityAutoplayPaused }"
            ></span>
          </div>

          <Transition :name="capabilityTransitionName" mode="out-in">
            <article
              :id="`capability-panel-${activeCapabilityPage.id}`"
              :key="activeCapabilityPage.id"
              class="capability-page"
              role="tabpanel"
              :aria-labelledby="`capability-tab-${activeCapabilityPage.id}`"
            >
              <div class="capability-page-copy">
                <p class="eyebrow">
                  <component :is="activeCapabilityPage.icon" :size="16" />
                  {{ activeCapabilityPage.badge }}
                </p>
                <h3>{{ activeCapabilityPage.title }}</h3>
                <p>{{ activeCapabilityPage.desc }}</p>

                <div class="capability-metrics">
                  <article v-for="metric in activeCapabilityPage.metrics" :key="metric.label">
                    <strong>{{ metric.value }}</strong>
                    <span>{{ metric.label }}</span>
                  </article>
                </div>

                <div class="capability-highlights">
                  <section v-for="item in activeCapabilityPage.highlights" :key="item.title">
                    <CheckCircle2 :size="16" />
                    <div>
                      <h4>{{ item.title }}</h4>
                      <p>{{ item.text }}</p>
                    </div>
                  </section>
                </div>
              </div>

              <figure class="capability-visual">
                <img :src="activeCapabilityPage.image" :alt="activeCapabilityPage.imageAlt" />
                <figcaption>{{ activeCapabilityPage.imageAlt }}</figcaption>
              </figure>
            </article>
          </Transition>
        </div>
      </section>

      <section
        id="compare"
        class="section compare-section nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 4 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-4"
      >
        <div class="section-heading">
          <p class="eyebrow">
            <Boxes :size="16" />
            竞品定位
          </p>
          <h2>不是取代所有 AI 工具，而是补上“项目落地层”</h2>
          <p>
            Cherry Studio、WorkBuddy 都有成熟方向。DreamWorker
            的差异，是把项目生命周期、运行时和可扩展 Skill 绑定在一起。
          </p>
        </div>
        <div class="compare-table" role="table" aria-label="DreamWorker 与竞品定位差异">
          <div class="compare-head" role="row">
            <span role="columnheader">产品</span>
            <span role="columnheader">主要定位</span>
            <span role="columnheader">DreamWorker 的切入点</span>
          </div>
          <article v-for="row in competitorRows" :key="row.product" class="compare-row" role="row">
            <strong role="cell">{{ row.product }}</strong>
            <p role="cell">{{ row.positioning }}</p>
            <p role="cell">{{ row.dreamworker }}</p>
          </article>
        </div>
      </section>

      <section
        id="architecture"
        class="section architecture-section nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 5 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-5"
      >
        <div class="section-heading">
          <p class="eyebrow">
            <CircuitBoard :size="16" />
            整体架构
          </p>
          <h2>Electron + Vue + Node Main Runtime，兼容生态且本地可控</h2>
          <p>
            前端贴近国内工程习惯，Node Main Runtime 作为 Electron Main 内嵌能力层承担长驻 AI OS
            运行时，直接调度 router、kernel、services、Workspace Store、Coding Agent 和 SDK
            能力；生产路径不再依赖 Go Engine、独立 engine 包或本机 HTTP 通信。
          </p>
        </div>
        <div class="architecture-layout">
          <div class="architecture-stack" aria-label="DreamWorker 架构层级">
            <div class="stack-layer">
              <span>桌面壳</span>
              <strong>Electron Main / Preload / Renderer</strong>
            </div>
            <div class="stack-layer">
              <span>交互层</span>
              <strong>Vue + Vite + Pinia 工作台</strong>
            </div>
            <div class="stack-layer featured">
              <span>AI OS</span>
              <strong>Node Main Runtime / Router / Kernel / Services / Store</strong>
            </div>
            <div class="stack-layer">
              <span>能力层</span>
              <strong>Coding Agent / OpenCode / Claude Agent / Codex / MCP</strong>
            </div>
          </div>
          <div class="architecture-grid">
            <article
              v-for="card in architectureCards"
              :key="card.title"
              class="feature-card compact"
            >
              <component :is="card.icon" :size="23" />
              <h3>{{ card.title }}</h3>
              <ul>
                <li v-for="detail in card.details" :key="detail">
                  <CheckCircle2 :size="16" />
                  <span>{{ detail }}</span>
                </li>
              </ul>
            </article>
          </div>
        </div>
      </section>

      <section
        id="router"
        class="section router-section nav-page-section"
        :class="{ 'is-active-page': activeTopNavIndex === 6 }"
        role="tabpanel"
        aria-labelledby="top-nav-tab-6"
      >
        <img class="router-bg" src="/images/resource-orbit-banner.png" alt="" />
        <div class="section-heading">
          <p class="eyebrow">
            <Globe2 :size="16" />
            首批模型入口
          </p>
          <h2>9Router 免费路由，一键接入多种编码模型链路</h2>
          <p>
            DreamWorker 首先提供 9Router 接入：把 Codex CLI、Claude Code、OpenAI / Anthropic
            兼容格式和多种免费模型来源汇成一个本地路由入口。
          </p>
        </div>
        <div class="router-panel">
          <div class="router-terminal" aria-label="9Router 快速接入命令示例">
            <header>
              <img src="/icons/9router.svg" alt="" />
              <strong>9Router bridge</strong>
            </header>
            <code>npm install -g 9router</code>
            <code>9router</code>
            <code>base_url = http://localhost:20128/v1</code>
            <code>DreamWorker → Resource Center → 一键检测</code>
          </div>
          <div class="router-feature-grid">
            <article v-for="item in routerFeatures" :key="item.title">
              <component :is="item.icon" :size="22" />
              <h3>{{ item.title }}</h3>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </div>
      </section>

      <section class="section roadmap-section">
        <div class="roadmap-copy">
          <p class="eyebrow">
            <GitBranch :size="16" />
            后续演进
          </p>
          <h2>让每个人都能拥有自己的 AI 项目操作系统</h2>
          <p>
            DreamWorker
            的长期方向不是封闭模板，而是让用户把自己的经验、行业知识和项目方法沉淀成可编排能力。
          </p>
        </div>
        <div class="roadmap-list">
          <article v-for="(item, index) in roadmapItems" :key="item">
            <span>{{ String(index + 1).padStart(2, '0') }}</span>
            <p>{{ item }}</p>
          </article>
        </div>
      </section>

      <section class="final-cta">
        <img
          src="/images/dreamworker-home-screenshot.png"
          alt="DreamWorker AI 工作台首页界面截图"
        />
        <div>
          <p class="eyebrow">
            <Wrench :size="16" />
            为创造者而建
          </p>
          <h2>从今天开始，把你的项目流程交给 DreamWorker 管起来。</h2>
          <p>
            先让 AI 帮你想清楚方向，再让 Skill
            帮你形成产物，最后让多智能体把项目推进到能发布、能验证、能复盘。
          </p>
          <a class="primary-link" :href="windowsDownloadUrl" target="_blank" rel="noreferrer">
            <Download :size="18" />
            <span>下载 Windows 体验版</span>
          </a>
        </div>
      </section>
    </main>

    <footer class="site-footer">
      <div>
        <img src="/images/brand-lockup.png" alt="DreamWorker" />
        <p>DreamWorker turns any idea into an executable agent-powered launch plan.</p>
      </div>
      <div class="source-links">
        <span>联网参考</span>
        <a
          v-for="link in sourceLinks"
          :key="link.url"
          :href="link.url"
          target="_blank"
          rel="noreferrer"
        >
          {{ link.label }}
          <ChevronRight :size="14" />
        </a>
      </div>
    </footer>
  </div>
</template>
