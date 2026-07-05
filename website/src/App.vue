<script setup lang="ts">
import type { Component } from 'vue'
import {
  ArrowRight,
  BadgeCheck,
  BrainCircuit,
  Boxes,
  ChartNoAxesCombined,
  CheckCircle2,
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
const windowsDownloadUrl = `https://github.com/zhongxinqiao06-spec/DreamWorker/releases/download/${releaseTag}/DreamWorker%20Setup%200.1.0.exe`

const navItems = [
  { label: '定位', href: '#position' },
  { label: '工作流', href: '#workflow' },
  { label: '能力矩阵', href: '#capabilities' },
  { label: '对比', href: '#compare' },
  { label: '架构', href: '#architecture' },
  { label: '9Router', href: '#router' }
]

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

const workflowStages = [
  { step: '01', icon: ScanSearch, title: '机会扫描', desc: '把灵感变成可验证问题，找到用户、痛点、场景和证据缺口。' },
  { step: '02', icon: ChartNoAxesCombined, title: '竞品地图', desc: '区分直接竞品、替代方案和手工流程，形成定位差异。' },
  { step: '03', icon: BadgeCheck, title: 'PRD 与门禁', desc: '生成需求、范围、验收标准和继续/暂停/转向的决策门。' },
  { step: '04', icon: CircuitBoard, title: '架构蓝图', desc: '拆模块、数据契约、运行时边界、安全策略和工程切片。' },
  { step: '05', icon: Code2, title: '多智能体执行', desc: '让产品、架构、开发、评估、增长角色围绕同一个项目协作。' },
  { step: '06', icon: Rocket, title: '发布复盘', desc: '落地发布清单、渠道动作、反馈循环和下一轮迭代计划。' }
]

const agentRoles = [
  { title: '机会侦察员', desc: '分析目标人群、痛点、替代方案和机会窗口。', tags: ['探索机会', 'explore', 'project'] },
  { title: '用户研究员', desc: '沉淀用户画像、访谈问题、使用场景和证据缺口。', tags: ['用户画像', 'explore', 'research'] },
  { title: '客群分析师', desc: '拆解 ICP、购买动机、预算和触达方式。', tags: ['客群细分', 'explore', 'sales'] },
  { title: '竞品分析师', desc: '比较竞品定位、功能、价格、渠道和差异化。', tags: ['竞品分析', 'explore', 'market'] },
  { title: '产品设计师', desc: '把验证后的机会转成信息架构、用户路径和 PRD。', tags: ['产品定义', 'product', 'prd'] },
  { title: '原型设计师', desc: '输出关键页面、状态、交互说明和 demo flow。', tags: ['原型方案', 'product', 'prototype'] },
  { title: '系统架构师', desc: '设计边界、数据流、模块契约和安全策略。', tags: ['系统设计', 'development', 'architecture'] },
  { title: '技术栈顾问', desc: '评估框架、存储、模型、部署和成本。', tags: ['技术选型', 'development', 'stack'] },
  { title: '开发编排员', desc: '把蓝图拆成可验收 PR、任务切片和测试门禁。', tags: ['开发计划', 'development', 'delivery'] },
  { title: '质量评估员', desc: '检查产物完整度、证据质量、幻觉风险和下一步行动。', tags: ['质量评估', 'qa', 'review'] },
  { title: '发布增长官', desc: '规划落地页、发布渠道、素材和反馈回收。', tags: ['发布计划', 'sales', 'growth'] },
  { title: '复盘运营官', desc: '把反馈、数据和经验沉淀成下一轮工作流。', tags: ['复盘迭代', 'ops', 'learning'] }
]

const skillPlan = [
  '项目任务定义',
  '机会与证据扫描',
  '用户画像生成',
  '竞品地图拆解',
  'PRD 与范围门禁',
  '原型说明生成',
  '架构蓝图拆分',
  '开发任务切片',
  '代码审查清单',
  '测试验收规划',
  '发布素材清单',
  '复盘迭代报告'
]

const toolGroups = [
  {
    title: '研究与证据',
    tools: ['Web 搜索', '竞品抓取', '访谈提纲', '证据评分', '用户画像表', '机会评分']
  },
  {
    title: '项目与文档',
    tools: ['文件读写', '目录扫描', 'Markdown 产物', 'Excel 表格', 'Word 文档', 'Mermaid 图']
  },
  {
    title: '工程与运行',
    tools: ['Git 状态', 'Shell 执行', '包管理', '单测运行', '代码索引', 'API 调试']
  },
  {
    title: '模型与交付',
    tools: ['模型路由检测', 'MCP 发现', 'Token 统计', '截图采集', '发布清单', '复盘报告']
  }
]

const competitorRows = [
  {
    product: 'Cherry Studio',
    positioning: '全能 AI 工作站：多模型对话、知识库、绘图、翻译、MCP 与高度自定义。',
    dreamworker: 'DreamWorker 更聚焦“做项目”：把机会分析、PRD、架构、开发、评估、发布做成连续工作流和可复用 Skill。'
  },
  {
    product: 'WorkBuddy',
    positioning: '场景化办公 AI 套件：专家团队、技能市场、办公生态连接和远程执行。',
    dreamworker: 'DreamWorker 更适合创造者与开发团队：本地优先桌面 AI OS、项目级上下文、可编排多智能体和开放插件运行时。'
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
    details: ['Windows / macOS / Linux 一套桌面体验', 'Main / Preload / Renderer 分层隔离', '本地 token 与 Engine 进程不暴露给 UI']
  },
  {
    icon: Layers3,
    title: 'Vue + Vite 前端',
    details: ['贴近国内开发生态，招人和维护成本更友好', '组件清晰、状态可读、迭代速度快', '适合长期沉淀工作台与资源中心']
  },
  {
    icon: Cpu,
    title: 'Go 自研 AI OS 后端',
    details: ['适合长驻本地 daemon、并发流式任务和单文件分发', '把模型、Agent、Skill、Tool、MCP、项目空间统一成运行时', '比纯 Node 后端更容易控制资源边界和发布体积']
  },
  {
    icon: Puzzle,
    title: 'Node Extension Bridge',
    details: ['保留 Node.js SDK 生态的接入便利', '第三方 SDK 可作为扩展插件导入', 'Go 负责核心调度，Node 负责生态适配']
  }
]

const routerFeatures = [
  { icon: Route, title: '一个 OpenAI-compatible 端点', desc: '把 Codex CLI、Claude Code、Cursor、Cline 等工具指向本地路由。' },
  { icon: Network, title: '免费 / 低成本 / 订阅三层路由', desc: '按可用性、成本和额度做 fallback，减少频繁换 key 和换模型。' },
  { icon: Zap, title: '工具输出 token 优化', desc: '9Router 的 RTK / Caveman 能压缩工具结果和回答冗余，适合编码场景。' },
  { icon: ShieldCheck, title: '先本地接入再扩展', desc: 'DreamWorker 首批把 9Router 作为免费路由入口，后续沉淀为资源中心的一键配置。' }
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
  <div class="site-shell">
    <header class="top-nav">
      <a class="brand" href="#top" aria-label="DreamWorker 首页">
        <img src="/images/brand-mark.png" alt="" />
        <span>DreamWorker</span>
      </a>
      <nav aria-label="主导航">
        <a v-for="item in navItems" :key="item.href" :href="item.href">{{ item.label }}</a>
      </nav>
      <a class="nav-action" :href="windowsDownloadUrl" target="_blank" rel="noreferrer">
        <Download :size="17" />
        <span>Windows 体验版</span>
      </a>
    </header>

    <main id="top">
      <section class="hero-section">
        <img class="hero-orbit" src="/images/glass-orbit-hero.png" alt="" />
        <img class="hero-product" src="/images/dreamworker-home-screenshot.png" alt="DreamWorker AI 工作台首页真实界面截图" />
        <div class="hero-copy">
          <p class="eyebrow">
            <Sparkles :size="16" />
            本地优先的 AI OS + 项目孵化工作台
          </p>
          <h1>DreamWorker</h1>
          <p class="hero-lede">
            帮助每一个有梦想的人，把想做的项目从灵感推进到真实发布。不是再造一个聊天框，而是把真实做项目的流程做成软件内工作流、Skill 和多智能体协作。
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

      <section id="position" class="section section-tight">
        <div class="section-heading">
          <p class="eyebrow">
            <BrainCircuit :size="16" />
            产品定位
          </p>
          <h2>把“会用 AI”升级成“能做成项目”</h2>
          <p>
            大多数 AI 工具让你更快获得回答，DreamWorker 想解决更难的一步：让回答进入项目推进链路，变成可验证的计划、产物和下一步行动。
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

      <section id="workflow" class="section workflow-section">
        <div class="section-heading">
          <p class="eyebrow">
            <Workflow :size="16" />
            内置项目流程
          </p>
          <h2>从想法到发布，由 Agent、Skill 和 Tool 连续接力</h2>
          <p>
            DreamWorker 的核心优势，是把真实项目孵化路径内置进产品：Agent 负责角色判断，Skill 负责可复用流程，Tool 负责真实执行。
          </p>
        </div>
        <div class="workflow-track">
          <article v-for="stage in workflowStages" :key="stage.step" class="workflow-card">
            <span>{{ stage.step }}</span>
            <component :is="stage.icon" :size="23" />
            <h3>{{ stage.title }}</h3>
            <p>{{ stage.desc }}</p>
          </article>
        </div>
      </section>

      <section id="capabilities" class="section capability-section">
        <div class="section-heading">
          <p class="eyebrow">
            <UsersRound :size="16" />
            Agent / Skill / Tool
          </p>
          <h2>规划内置 12 个 Agent、12+ Skill 和 20+ 工具</h2>
          <p>
            首页和资源中心会围绕这三层展开：用户看到的不是散装提示词，而是一组能协作、能调用工具、能沉淀产物的项目团队。
          </p>
        </div>

        <div class="capability-summary-grid" aria-label="DreamWorker 能力规划概览">
          <article>
            <strong>12</strong>
            <span>项目 Agent</span>
            <p>覆盖探索、产品、架构、开发、评估、发布与复盘角色。</p>
          </article>
          <article>
            <strong>12+</strong>
            <span>内置 Skill</span>
            <p>把每个阶段的输入、产出、质量门禁和下一步动作标准化。</p>
          </article>
          <article>
            <strong>20+</strong>
            <span>工具与 MCP</span>
            <p>让搜索、文件、代码、测试、模型路由和发布交付都能被调用。</p>
          </article>
        </div>

        <div class="agent-role-grid" aria-label="规划内置 Agent 角色">
          <article v-for="role in agentRoles" :key="role.title" class="agent-role-card">
            <span>agent</span>
            <h3>{{ role.title }}</h3>
            <p>{{ role.desc }}</p>
            <div>
              <small v-for="tag in role.tags" :key="tag">{{ tag }}</small>
            </div>
          </article>
        </div>

        <div class="skill-tool-layout">
          <article class="skill-plan-card">
            <div>
              <p class="eyebrow">
                <Workflow :size="16" />
                12+ Skill
              </p>
              <h3>把项目方法论做成可复用工作流</h3>
            </div>
            <ol>
              <li v-for="(skill, index) in skillPlan" :key="skill">
                <span>{{ String(index + 1).padStart(2, '0') }}</span>
                <strong>{{ skill }}</strong>
              </li>
            </ol>
          </article>

          <article class="tool-plan-card">
            <div>
              <p class="eyebrow">
                <Wrench :size="16" />
                20+ Tool
              </p>
              <h3>让 Agent 能查、能写、能跑、能交付</h3>
            </div>
            <div class="tool-group-grid">
              <section v-for="group in toolGroups" :key="group.title">
                <h4>{{ group.title }}</h4>
                <div>
                  <span v-for="tool in group.tools" :key="tool">{{ tool }}</span>
                </div>
              </section>
            </div>
          </article>
        </div>
      </section>

      <section id="compare" class="section compare-section">
        <div class="section-heading">
          <p class="eyebrow">
            <Boxes :size="16" />
            竞品定位
          </p>
          <h2>不是取代所有 AI 工具，而是补上“项目落地层”</h2>
          <p>
            Cherry Studio、WorkBuddy 都有成熟方向。DreamWorker 的差异，是把项目生命周期、运行时和可扩展 Skill 绑定在一起。
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

      <section id="architecture" class="section architecture-section">
        <div class="section-heading">
          <p class="eyebrow">
            <CircuitBoard :size="16" />
            整体架构
          </p>
          <h2>Electron + Vue + Go Engine，兼容生态但核心可控</h2>
          <p>
            前端贴近国内工程习惯，后端用 Go 承担长驻 AI OS 运行时，同时通过 Node Extension Bridge 接入市场上丰富的 Node.js SDK。
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
              <strong>Go Engine Runtime API / SSE / Workspace Store</strong>
            </div>
            <div class="stack-layer">
              <span>能力层</span>
              <strong>Agent / Skill / Tool / MCP / Node Extensions</strong>
            </div>
          </div>
          <div class="architecture-grid">
            <article v-for="card in architectureCards" :key="card.title" class="feature-card compact">
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

      <section id="router" class="section router-section">
        <img class="router-bg" src="/images/resource-orbit-banner.png" alt="" />
        <div class="section-heading">
          <p class="eyebrow">
            <Globe2 :size="16" />
            首批模型入口
          </p>
          <h2>9Router 免费路由，一键接入多种编码模型链路</h2>
          <p>
            DreamWorker 首先提供 9Router 接入：把 Codex CLI、Claude Code、OpenAI / Anthropic 兼容格式和多种免费模型来源汇成一个本地路由入口。
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
            DreamWorker 的长期方向不是封闭模板，而是让用户把自己的经验、行业知识和项目方法沉淀成可编排能力。
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
        <img src="/images/dreamworker-home-screenshot.png" alt="DreamWorker AI 工作台首页界面截图" />
        <div>
          <p class="eyebrow">
            <Wrench :size="16" />
            为创造者而建
          </p>
          <h2>从今天开始，把你的项目流程交给 DreamWorker 管起来。</h2>
          <p>
            先让 AI 帮你想清楚方向，再让 Skill 帮你形成产物，最后让多智能体把项目推进到能发布、能验证、能复盘。
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
        <a v-for="link in sourceLinks" :key="link.url" :href="link.url" target="_blank" rel="noreferrer">
          {{ link.label }}
          <ChevronRight :size="14" />
        </a>
      </div>
    </footer>
  </div>
</template>
