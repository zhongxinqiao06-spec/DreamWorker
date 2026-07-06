<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import {
  ArrowLeft,
  Bot,
  CheckCircle2,
  Circle,
  CircleDot,
  Database,
  FileText,
  Play,
  RefreshCw,
  Save,
  SlidersHorizontal,
  Sparkles,
  Target,
  Workflow
} from 'lucide-vue-next'
import { computed, ref, watch } from 'vue'
import type { ProjectSubmodule } from '../../../../shared/dreamworker-api'
import { useAppShellStore } from '../../stores/app-shell'
import { statusLabel } from '../../stores/workspace-navigation'
import EChartPanel from './EChartPanel.vue'
import ProjectContextPanel from './ProjectContextPanel.vue'

type ExploreSubmoduleId = 'opportunity_radar' | 'user_persona' | 'competitor_map' | 'evidence_graph'

type ChartKind = 'radar' | 'bar' | 'scatter' | 'graph'

type SourceOption = {
  readonly id: string
  readonly label: string
  readonly detail: string
}

type PlanOption = {
  readonly id: string
  readonly title: string
  readonly goal: string
  readonly score: number
  readonly risk: string
  readonly cost: string
  readonly decision: string
}

type ArtifactOption = {
  readonly id: string
  readonly label: string
  readonly path: string
  readonly detail: string
}

type ExploreBlueprint = {
  readonly submoduleId: ExploreSubmoduleId
  readonly displayName: string
  readonly status: ProjectSubmodule['status']
  readonly summary: string
  readonly nextBestAction: string
  readonly stage: string
  readonly chartKind: ChartKind
  readonly chartTitle: string
  readonly defaultBrief: string
  readonly agent: string
  readonly skill: string
  readonly tool: string
  readonly sources: readonly SourceOption[]
  readonly plans: readonly PlanOption[]
  readonly artifacts: readonly ArtifactOption[]
  readonly insights: readonly string[]
}

type RuntimeStep = {
  readonly phase: string
  readonly title: string
  readonly detail: string
  readonly status: 'ready' | 'running' | 'completed'
}

type InsightItem = {
  readonly id: string
  readonly title: string
  readonly detail: string
}

const appShell = useAppShellStore()

const exploreSubmoduleIds = new Set<ExploreSubmoduleId>([
  'opportunity_radar',
  'user_persona',
  'competitor_map',
  'evidence_graph'
])

const blueprints: Record<ExploreSubmoduleId, ExploreBlueprint> = {
  opportunity_radar: {
    submoduleId: 'opportunity_radar',
    displayName: '机会雷达',
    status: 'ready',
    summary: '扫描用户痛点、市场窗口和可验证机会。',
    nextBestAction: '先生成机会清单，再挑选高置信假设。',
    stage: 'Discover',
    chartKind: 'radar',
    chartTitle: '机会适配度雷达',
    defaultBrief: '面向独立开发者，寻找可在 2-4 周内验证、能沉淀产品资产的 AI 工作台机会。',
    agent: 'agent_opportunity_scout',
    skill: 'skill_opportunity_scan',
    tool: 'tool_web_search_stub',
    sources: [
      { id: 'project_brief', label: '项目简述', detail: '当前项目标题、描述和本地配置' },
      { id: 'pain_posts', label: '痛点样本', detail: '社区反馈、客服记录、聊天摘要' },
      { id: 'market_window', label: '市场窗口', detail: '趋势、政策、平台能力变化' },
      { id: 'competitor_signal', label: '竞品信号', detail: '同类产品定价、评论和路线图' }
    ],
    plans: [
      {
        id: 'narrow_probe',
        title: '窄场景探针',
        goal: '先围绕一个高频痛点做可验证假设。',
        score: 88,
        risk: '样本偏少',
        cost: '2 天',
        decision: '适合作为第一轮验证'
      },
      {
        id: 'workflow_bundle',
        title: '工作流组合',
        goal: '把探索、产品和开发串成单条闭环。',
        score: 81,
        risk: '范围容易膨胀',
        cost: '5 天',
        decision: '需要强约束 MVP 边界'
      },
      {
        id: 'platform_bet',
        title: '平台化下注',
        goal: '优先验证 Agent runtime 可配置能力。',
        score: 74,
        risk: '验证周期更长',
        cost: '8 天',
        decision: '保留为第二阶段方案'
      }
    ],
    artifacts: [
      {
        id: 'dream_brief',
        label: 'dream_brief.md',
        path: 'artifacts/explore/dream_brief.md',
        detail: '机会摘要、目标用户、问题边界'
      },
      {
        id: 'hypotheses',
        label: 'hypotheses.yaml',
        path: 'artifacts/explore/hypotheses.yaml',
        detail: '假设、验证方式、成功阈值'
      }
    ],
    insights: [
      '优先选择能在本地目录沉淀文件资产的工作流，后续可直接接入产品模块。',
      '机会假设需要同时写明 ICP、付费触发点和第一个可观测指标。',
      '若样本少于 12 条，产物必须标记为低置信并进入用户画像补样。'
    ]
  },
  user_persona: {
    submoduleId: 'user_persona',
    displayName: '用户画像',
    status: 'idle',
    summary: '把目标用户、场景、付费动机和反对理由结构化。',
    nextBestAction: '基于机会雷达结果补齐 ICP 和痛点证据。',
    stage: 'Discover',
    chartKind: 'bar',
    chartTitle: '画像信号强度',
    defaultBrief: '围绕独立开发者、AI 工具操盘手和小团队负责人整理可行动用户画像。',
    agent: 'agent_persona_architect',
    skill: 'skill_persona_synthesis',
    tool: 'tool_model_generate_stub',
    sources: [
      { id: 'opportunity_notes', label: '机会假设', detail: '来自机会雷达的高置信假设' },
      { id: 'interview_notes', label: '访谈笔记', detail: '用户访谈、问卷和聊天记录' },
      { id: 'behavior_log', label: '行为线索', detail: '任务频次、失败点和替代方案' },
      { id: 'buying_trigger', label: '付费触发', detail: '预算、采购阻力和决策角色' }
    ],
    plans: [
      {
        id: 'icp_first',
        title: 'ICP 优先',
        goal: '先收敛最可能付费的窄人群。',
        score: 84,
        risk: '忽略次级使用者',
        cost: '3 天',
        decision: '适合早期验证'
      },
      {
        id: 'scenario_map',
        title: '场景地图',
        goal: '按任务场景拆分画像和需求触发。',
        score: 79,
        risk: '需要更多样本',
        cost: '4 天',
        decision: '适合进入需求分析前'
      },
      {
        id: 'buyer_user_split',
        title: '买用分离',
        goal: '区分使用者、购买者和影响者。',
        score: 72,
        risk: 'B2B 复杂度上升',
        cost: '6 天',
        decision: '适合作为销售模块输入'
      }
    ],
    artifacts: [
      {
        id: 'personas',
        label: 'personas.json',
        path: 'artifacts/explore/personas.json',
        detail: '结构化画像、动机、反对理由'
      },
      {
        id: 'journey_map',
        label: 'journey_map.md',
        path: 'artifacts/explore/journey_map.md',
        detail: '任务旅程、痛点和关键时刻'
      }
    ],
    insights: [
      '画像必须能回答“谁今天就会用”，否则需要退回机会雷达缩窄场景。',
      '每个画像至少绑定一个可观察行为，而不是只写抽象身份标签。',
      '反对理由会直接影响产品模块的验收标准和定价假设。'
    ]
  },
  competitor_map: {
    submoduleId: 'competitor_map',
    displayName: '竞品地图',
    status: 'idle',
    summary: '整理替代方案、差异化空间和进入壁垒。',
    nextBestAction: '先确认竞品范围，再输出差异化判断。',
    stage: 'Discover',
    chartKind: 'scatter',
    chartTitle: '竞品差异化空间',
    defaultBrief: '对比 AI 项目工作台、通用 Agent 平台和开发者效率工具的替代关系。',
    agent: 'agent_competitor_mapper',
    skill: 'skill_competitor_landscape',
    tool: 'tool_web_search_stub',
    sources: [
      { id: 'direct_competitor', label: '直接竞品', detail: '同目标用户、同任务链路产品' },
      { id: 'alternative_flow', label: '替代流程', detail: '表格、文档、聊天工具等替代做法' },
      { id: 'pricing_page', label: '定价页面', detail: '价格、限额、套餐和试用策略' },
      { id: 'review_signal', label: '评论信号', detail: '用户评价、抱怨和迁移原因' }
    ],
    plans: [
      {
        id: 'matrix_scan',
        title: '矩阵扫描',
        goal: '快速建立价格、能力、用户群对比。',
        score: 82,
        risk: '信息可能过期',
        cost: '3 天',
        decision: '适合先出方向判断'
      },
      {
        id: 'job_to_be_done',
        title: '任务替代',
        goal: '按用户真实任务而不是品类划分竞品。',
        score: 86,
        risk: '需要访谈支撑',
        cost: '4 天',
        decision: '更适合作为产品定位输入'
      },
      {
        id: 'moat_probe',
        title: '壁垒探针',
        goal: '检查数据、集成、本地化和工作流壁垒。',
        score: 76,
        risk: '结论依赖长期验证',
        cost: '6 天',
        decision: '进入销售模块前再深化'
      }
    ],
    artifacts: [
      {
        id: 'competitor_matrix',
        label: 'competitor_matrix.xlsx',
        path: 'artifacts/explore/competitor_matrix.xlsx',
        detail: '竞品清单、能力评分、定价对比'
      },
      {
        id: 'positioning_notes',
        label: 'positioning_notes.md',
        path: 'artifacts/explore/positioning_notes.md',
        detail: '差异化定位、进入壁垒、风险'
      }
    ],
    insights: [
      '竞品地图不只比较产品，还要比较用户当前“凑合能用”的替代流程。',
      '差异化判断需要绑定可验证证据，避免只写营销式口号。',
      '若竞品密度高，应优先寻找集成、本地文件和团队协作边界。'
    ]
  },
  evidence_graph: {
    submoduleId: 'evidence_graph',
    displayName: '证据图谱',
    status: 'idle',
    summary: '把假设、证据、风险和下一步动作连成可审计图谱。',
    nextBestAction: '证据不足时返回 ask_user，不静默推进。',
    stage: 'Discover',
    chartKind: 'graph',
    chartTitle: '假设证据关系图',
    defaultBrief: '把探索阶段的机会、画像和竞品证据汇总成可追踪、可审计的判断图谱。',
    agent: 'agent_evidence_curator',
    skill: 'skill_evidence_graph',
    tool: 'tool_artifact_write',
    sources: [
      { id: 'hypotheses', label: '假设清单', detail: '机会雷达输出的待验证假设' },
      { id: 'persona_signal', label: '画像信号', detail: '用户画像中的行为和反对理由' },
      { id: 'competitor_signal', label: '竞品信号', detail: '竞品地图中的替代关系' },
      { id: 'decision_log', label: '决策日志', detail: '项目推进中的人工判断和追问' }
    ],
    plans: [
      {
        id: 'audit_first',
        title: '审计优先',
        goal: '先建立来源、置信度和判断链路。',
        score: 87,
        risk: '产物偏重结构',
        cost: '3 天',
        decision: '适合作为跨模块检查点'
      },
      {
        id: 'risk_first',
        title: '风险优先',
        goal: '优先暴露证据缺口和反例。',
        score: 83,
        risk: '可能阻塞推进',
        cost: '2 天',
        decision: '适合高不确定机会'
      },
      {
        id: 'action_first',
        title: '行动优先',
        goal: '把证据缺口转成下一步采集任务。',
        score: 78,
        risk: '图谱完整度下降',
        cost: '2 天',
        decision: '适合快速推进'
      }
    ],
    artifacts: [
      {
        id: 'evidence_graph',
        label: 'evidence_graph.json',
        path: 'artifacts/explore/evidence_graph.json',
        detail: '节点、边、证据来源、置信度'
      },
      {
        id: 'assumption_log',
        label: 'assumption_log.md',
        path: 'artifacts/explore/assumption_log.md',
        detail: '假设变更、证据缺口、追问项'
      }
    ],
    insights: [
      '证据图谱需要显示“不知道什么”，证据缺口比结论更重要。',
      '每条边都要能追溯到来源或人工判断，否则进入 ask_user。',
      '图谱输出应被产品需求分析读取，作为功能优先级和验收条件的证据。'
    ]
  }
}

const activeSubmoduleId = computed<ExploreSubmoduleId>(() => {
  const candidate = appShell.activeSubmoduleDetail?.submoduleId
  return exploreSubmoduleIds.has(candidate as ExploreSubmoduleId)
    ? (candidate as ExploreSubmoduleId)
    : 'opportunity_radar'
})

const module = computed(() =>
  appShell.projectModules.find((projectModule) => projectModule.moduleId === 'explore')
)
const moduleConfig = computed(() => module.value?.config ?? {})
const blueprint = computed(() => blueprints[activeSubmoduleId.value])
const submodule = computed<ProjectSubmodule>(() => {
  const configured = module.value?.submodules.find(
    (item) => item.submoduleId === activeSubmoduleId.value
  )
  if (configured) {
    return configured
  }
  const current = blueprint.value
  return {
    projectId: appShell.activeProjectId,
    moduleId: 'explore',
    submoduleId: current.submoduleId,
    displayName: current.displayName,
    status: current.status,
    summary: current.summary,
    defaultAgents: [current.agent],
    enabledSkills: [current.skill],
    enabledTools: [current.tool, 'tool_model_generate_stub'],
    outputArtifacts: current.artifacts.map((artifact) => artifact.label),
    nextBestAction: current.nextBestAction,
    config: { stage: current.stage }
  }
})

const scanBrief = ref('')
const selectedSourceIds = ref<string[]>([])
const selectedArtifactId = ref('')
const selectedPlanId = ref('')
const runStarted = ref(false)
const saving = ref(false)
const previewInsights = ref<InsightItem[]>([])

const selectedSources = computed(() =>
  blueprint.value.sources.filter((source) => selectedSourceIds.value.includes(source.id))
)
const selectedArtifact = computed(
  () =>
    blueprint.value.artifacts.find((artifact) => artifact.id === selectedArtifactId.value) ??
    blueprint.value.artifacts[0]
)
const selectedPlan = computed(
  () =>
    blueprint.value.plans.find((plan) => plan.id === selectedPlanId.value) ??
    blueprint.value.plans[0]
)
const savedAt = computed(() => readConfigString('savedAt') || '暂无')
const runStatusLabel = computed(() => (runStarted.value ? '已生成预览' : '待运行'))
const chartHeight = computed(() => (blueprint.value.chartKind === 'graph' ? 340 : 292))
const sourceSummary = computed(() =>
  selectedSources.value.length > 0 ? `${selectedSources.value.length} 个信号源` : '未选择信号源'
)
const runtimeSteps = computed<RuntimeStep[]>(() => [
  {
    phase: 'INPUT',
    title: '收集输入信号',
    detail: '读取项目简述、探索产物和人工补充内容。',
    status: selectedSources.value.length > 0 ? 'completed' : 'ready'
  },
  {
    phase: 'PLAN',
    title: '选择运行方案',
    detail: selectedPlan.value?.goal ?? '选择一个验证方案。',
    status: selectedPlanId.value ? 'completed' : 'ready'
  },
  {
    phase: 'ANALYZE',
    title: '生成结构化判断',
    detail: '按子模块目标生成图表、比较方案和关键洞察。',
    status: runStarted.value ? 'completed' : selectedSources.value.length > 0 ? 'running' : 'ready'
  },
  {
    phase: 'WRITE',
    title: '沉淀产物槽',
    detail: selectedArtifact.value?.path ?? '等待选择产物。',
    status: runStarted.value ? 'completed' : 'ready'
  }
])
const chartOption = computed<EChartsOption>(() =>
  createChartOption(blueprint.value.chartKind, selectedPlan.value?.score ?? 80)
)

watch(
  [activeSubmoduleId, moduleConfig],
  () => {
    hydrateFromConfig()
  },
  { immediate: true }
)

function configKey(suffix: string): string {
  return `${activeSubmoduleId.value}_${suffix}`
}

function readConfigString(suffix: string): string {
  const value = moduleConfig.value[configKey(suffix)]
  return typeof value === 'string' ? value : ''
}

function hydrateFromConfig(): void {
  const current = blueprint.value
  scanBrief.value = readConfigString('brief') || current.defaultBrief

  const savedSources = readConfigString('sources')
    .split('|')
    .map((item) => item.trim())
    .filter(Boolean)
  const validSourceIds = new Set(current.sources.map((source) => source.id))
  const validSavedSources = savedSources.filter((sourceId) => validSourceIds.has(sourceId))
  selectedSourceIds.value =
    validSavedSources.length > 0
      ? validSavedSources
      : current.sources.slice(0, 3).map((source) => source.id)

  const savedArtifact = readConfigString('artifact')
  selectedArtifactId.value = current.artifacts.some((artifact) => artifact.id === savedArtifact)
    ? savedArtifact
    : (current.artifacts[0]?.id ?? '')

  const savedPlan = readConfigString('plan')
  selectedPlanId.value = current.plans.some((plan) => plan.id === savedPlan)
    ? savedPlan
    : (current.plans[0]?.id ?? '')

  runStarted.value = readConfigString('runStatus') === 'completed'
  previewInsights.value = runStarted.value ? buildInsights(current) : []
}

function toggleSource(sourceId: string): void {
  selectedSourceIds.value = selectedSourceIds.value.includes(sourceId)
    ? selectedSourceIds.value.filter((id) => id !== sourceId)
    : [...selectedSourceIds.value, sourceId]
}

function resetDraft(): void {
  const current = blueprint.value
  scanBrief.value = current.defaultBrief
  selectedSourceIds.value = current.sources.slice(0, 3).map((source) => source.id)
  selectedArtifactId.value = current.artifacts[0]?.id ?? ''
  selectedPlanId.value = current.plans[0]?.id ?? ''
  runStarted.value = false
  previewInsights.value = []
}

function runPreview(): void {
  runStarted.value = true
  previewInsights.value = buildInsights(blueprint.value)
}

async function saveConfig(): Promise<void> {
  saving.value = true
  try {
    await appShell.updateProjectModuleConfig('explore', {
      ...moduleConfig.value,
      [configKey('brief')]: scanBrief.value.trim(),
      [configKey('sources')]: selectedSourceIds.value.join('|'),
      [configKey('artifact')]: selectedArtifactId.value,
      [configKey('plan')]: selectedPlanId.value,
      [configKey('runStatus')]: runStarted.value ? 'completed' : 'ready',
      [configKey('savedAt')]: new Date().toISOString()
    })
  } finally {
    saving.value = false
  }
}

function buildInsights(current: ExploreBlueprint): InsightItem[] {
  return current.insights.map((detail, index) => ({
    id: `${current.submoduleId}_${index}`,
    title: `${current.displayName}洞察 ${index + 1}`,
    detail
  }))
}

function stepText(status: RuntimeStep['status']): string {
  if (status === 'completed') {
    return '完成'
  }
  if (status === 'running') {
    return '待生成'
  }
  return '待准备'
}

function formatTime(value: string): string {
  if (!value || value === '暂无') {
    return '暂无'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function createChartOption(kind: ChartKind, planScore: number): EChartsOption {
  if (kind === 'bar') {
    return {
      color: ['#7c5cff'],
      grid: { left: 36, right: 18, top: 28, bottom: 34 },
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: ['痛点', '频次', '预算', '触达', '反对'],
        axisLine: { lineStyle: { color: '#dbe3f4' } },
        axisLabel: { color: '#687894', fontWeight: 700 }
      },
      yAxis: {
        type: 'value',
        max: 100,
        splitLine: { lineStyle: { color: 'rgba(160,129,255,0.16)' } },
        axisLabel: { color: '#687894' }
      },
      series: [
        {
          type: 'bar',
          barWidth: 22,
          data: [82, 76, planScore, 68, 55],
          itemStyle: { borderRadius: [6, 6, 0, 0] }
        }
      ]
    }
  }

  if (kind === 'scatter') {
    return {
      color: ['#7c5cff', '#12b981', '#f59e0b'],
      grid: { left: 42, right: 22, top: 26, bottom: 42 },
      tooltip: {
        trigger: 'item',
        formatter: '{b}<br/>差异化 {c0}<br/>进入壁垒 {c1}'
      },
      xAxis: {
        type: 'value',
        name: '差异化',
        max: 100,
        splitLine: { lineStyle: { color: 'rgba(160,129,255,0.14)' } },
        axisLabel: { color: '#687894' }
      },
      yAxis: {
        type: 'value',
        name: '壁垒',
        max: 100,
        splitLine: { lineStyle: { color: 'rgba(160,129,255,0.14)' } },
        axisLabel: { color: '#687894' }
      },
      series: [
        {
          type: 'scatter',
          symbolSize: 28,
          label: { show: true, formatter: '{b}', color: '#172033', fontWeight: 800 },
          data: [
            { name: '通用聊天', value: [34, 38] },
            { name: 'Agent 平台', value: [62, 66] },
            { name: '项目工作台', value: [planScore, 74] },
            { name: '文档流程', value: [44, 52] }
          ]
        }
      ]
    }
  }

  if (kind === 'graph') {
    return {
      animation: false,
      tooltip: { trigger: 'item' },
      series: [
        {
          type: 'graph',
          layout: 'none',
          roam: false,
          draggable: false,
          left: 12,
          right: 12,
          top: 12,
          bottom: 12,
          label: { show: true, color: '#172033', fontWeight: 900 },
          edgeSymbol: ['none', 'arrow'],
          edgeSymbolSize: [0, 8],
          lineStyle: { color: 'rgba(124,92,255,0.42)', width: 2, curveness: 0.08 },
          emphasis: { focus: 'adjacency' },
          data: [
            { name: '机会假设', x: 70, y: 155, symbolSize: 62, itemStyle: { color: '#7c5cff' } },
            { name: '用户证据', x: 230, y: 70, symbolSize: 48, itemStyle: { color: '#23c3ff' } },
            { name: '竞品证据', x: 236, y: 238, symbolSize: 48, itemStyle: { color: '#12b981' } },
            { name: '风险缺口', x: 390, y: 94, symbolSize: 46, itemStyle: { color: '#f59e0b' } },
            { name: '下一步动作', x: 412, y: 226, symbolSize: 54, itemStyle: { color: '#ef4444' } }
          ],
          links: [
            { source: '机会假设', target: '用户证据' },
            { source: '机会假设', target: '竞品证据' },
            { source: '用户证据', target: '风险缺口' },
            { source: '竞品证据', target: '风险缺口' },
            { source: '风险缺口', target: '下一步动作' }
          ]
        }
      ]
    }
  }

  return {
    color: ['#7c5cff'],
    tooltip: {},
    radar: {
      radius: '62%',
      indicator: [
        { name: '痛点强度', max: 100 },
        { name: '市场窗口', max: 100 },
        { name: '付费意愿', max: 100 },
        { name: '可验证性', max: 100 },
        { name: '壁垒', max: 100 }
      ],
      splitNumber: 4,
      splitLine: { lineStyle: { color: 'rgba(160,129,255,0.2)' } },
      splitArea: { areaStyle: { color: ['rgba(124,92,255,0.04)', 'rgba(124,92,255,0.09)'] } },
      axisName: { color: '#687894', fontWeight: 800 }
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: [88, 72, planScore, 84, 64],
            name: '机会适配度',
            areaStyle: { color: 'rgba(124,92,255,0.18)' },
            lineStyle: { color: '#7c5cff', width: 3 },
            symbolSize: 5
          }
        ]
      }
    ]
  }
}
</script>

<template>
  <section class="workspace-layout module-workspace-layout explore-submodule-layout">
    <ProjectContextPanel />

    <section
      class="module-center panel-surface explore-submodule-center"
      :aria-label="`${submodule.displayName}详情页`"
    >
      <div class="submodule-detail-header">
        <button
          class="icon-button small"
          type="button"
          title="返回探索模块"
          @click="appShell.leaveSubmoduleDetail()"
        >
          <ArrowLeft :size="16" aria-hidden="true" />
        </button>
        <div>
          <p class="eyebrow">探索模块 / {{ submodule.displayName }}</p>
          <h2>{{ submodule.displayName }}</h2>
          <p>{{ submodule.summary }}</p>
        </div>
        <div class="submodule-detail-actions">
          <button type="button" title="重置当前草稿" @click="resetDraft">
            <RefreshCw :size="15" aria-hidden="true" />
            重置
          </button>
          <button type="button" title="生成当前子模块预览" @click="runPreview">
            <Play :size="15" aria-hidden="true" />
            运行预览
          </button>
          <button
            class="primary-button"
            type="button"
            :disabled="saving"
            title="保存探索子模块配置"
            @click="saveConfig"
          >
            <Save :size="15" aria-hidden="true" />
            {{ saving ? '保存中' : '保存配置' }}
          </button>
        </div>
      </div>

      <div class="context-pills explore-runtime-pills">
        <span>{{ statusLabel(submodule.status) }}</span>
        <span>{{ submodule.config.stage ?? blueprint.stage }}</span>
        <span>{{ sourceSummary }}</span>
        <span>{{ runStatusLabel }}</span>
      </div>

      <section class="explore-runtime-grid">
        <article class="explore-control-panel">
          <div class="section-heading-row">
            <Database :size="17" aria-hidden="true" />
            <div>
              <p class="eyebrow">Runtime Input</p>
              <h3>运行输入</h3>
            </div>
          </div>

          <dl>
            <div>
              <dt>项目</dt>
              <dd>{{ appShell.activeProject?.title ?? '暂无项目' }}</dd>
            </div>
            <div>
              <dt>本地目录</dt>
              <dd>{{ appShell.activeProject?.localRootPath ?? '未设置' }}</dd>
            </div>
            <div>
              <dt>保存时间</dt>
              <dd>{{ formatTime(savedAt) }}</dd>
            </div>
          </dl>

          <label class="field-label" for="explore-scan-brief">探索说明</label>
          <textarea id="explore-scan-brief" v-model="scanBrief" />

          <div class="section-heading-row compact">
            <SlidersHorizontal :size="16" aria-hidden="true" />
            <div>
              <p class="eyebrow">Signals</p>
              <h3>信号源</h3>
            </div>
          </div>
          <div class="source-toggle-grid">
            <button
              v-for="source in blueprint.sources"
              :key="source.id"
              type="button"
              :class="{ active: selectedSourceIds.includes(source.id) }"
              @click="toggleSource(source.id)"
            >
              <CheckCircle2
                v-if="selectedSourceIds.includes(source.id)"
                :size="16"
                aria-hidden="true"
              />
              <Circle v-else :size="16" aria-hidden="true" />
              <span>{{ source.label }}</span>
              <small>{{ source.detail }}</small>
            </button>
          </div>
        </article>

        <div class="explore-analysis-column">
          <EChartPanel
            :title="blueprint.chartTitle"
            :caption="selectedPlan?.title ?? ''"
            :option="chartOption"
            :height="chartHeight"
          />

          <article class="comparison-panel">
            <div class="section-heading-row">
              <Target :size="17" aria-hidden="true" />
              <div>
                <p class="eyebrow">Plan Compare</p>
                <h3>多方案对比</h3>
              </div>
            </div>
            <div class="plan-comparison-grid">
              <button
                v-for="plan in blueprint.plans"
                :key="plan.id"
                type="button"
                :class="{ active: selectedPlanId === plan.id }"
                @click="selectedPlanId = plan.id"
              >
                <span>{{ plan.title }}</span>
                <strong>{{ plan.score }}</strong>
                <small>{{ plan.goal }}</small>
                <em>{{ plan.decision }}</em>
              </button>
            </div>
          </article>
        </div>

        <article class="explore-run-panel">
          <div class="section-heading-row">
            <Workflow :size="17" aria-hidden="true" />
            <div>
              <p class="eyebrow">Agent Runtime</p>
              <h3>运行链路</h3>
            </div>
          </div>

          <div class="explore-runtime-summary">
            <div>
              <Bot :size="16" aria-hidden="true" />
              <strong>1</strong>
              <span>Agent</span>
            </div>
            <div>
              <Sparkles :size="16" aria-hidden="true" />
              <strong>1</strong>
              <span>Skill</span>
            </div>
            <div>
              <FileText :size="16" aria-hidden="true" />
              <strong>{{ blueprint.artifacts.length }}</strong>
              <span>产物</span>
            </div>
          </div>

          <ol class="runtime-steps explore-steps">
            <li v-for="step in runtimeSteps" :key="step.phase" :data-status="step.status">
              <CheckCircle2 v-if="step.status === 'completed'" :size="16" aria-hidden="true" />
              <CircleDot v-else :size="16" aria-hidden="true" />
              <div>
                <strong>{{ step.title }}</strong>
                <span>{{ step.detail }}</span>
              </div>
              <small>{{ stepText(step.status) }}</small>
            </li>
          </ol>

          <div class="artifact-slot-list">
            <button
              v-for="artifact in blueprint.artifacts"
              :key="artifact.id"
              type="button"
              :class="{ active: selectedArtifactId === artifact.id }"
              @click="selectedArtifactId = artifact.id"
            >
              <FileText :size="16" aria-hidden="true" />
              <span>{{ artifact.label }}</span>
              <small>{{ artifact.detail }}</small>
            </button>
          </div>
        </article>
      </section>
    </section>

    <aside class="right-panel explore-runtime-side" :aria-label="`${submodule.displayName}配置`">
      <article class="inspector-card">
        <p class="eyebrow">当前子模块</p>
        <h3>{{ submodule.displayName }}</h3>
        <p>{{ submodule.nextBestAction }}</p>
        <dl>
          <div>
            <dt>Agent</dt>
            <dd>{{ blueprint.agent }}</dd>
          </div>
          <div>
            <dt>Skill</dt>
            <dd>{{ blueprint.skill }}</dd>
          </div>
          <div>
            <dt>Tool</dt>
            <dd>{{ blueprint.tool }}</dd>
          </div>
        </dl>
      </article>

      <article class="inspector-card">
        <p class="eyebrow">产物槽</p>
        <h3>{{ selectedArtifact?.label }}</h3>
        <p>{{ selectedArtifact?.path }}</p>
      </article>

      <article class="inspector-card">
        <p class="eyebrow">运行洞察</p>
        <h3>{{ runStarted ? '已生成' : '等待预览' }}</h3>
        <div v-if="previewInsights.length" class="explore-insight-list">
          <article v-for="insight in previewInsights" :key="insight.id">
            <strong>{{ insight.title }}</strong>
            <span>{{ insight.detail }}</span>
          </article>
        </div>
        <div v-else class="explore-empty-result">
          <Sparkles :size="16" aria-hidden="true" />
          <span>运行预览后会生成当前子模块的关键判断。</span>
        </div>
      </article>
    </aside>
  </section>
</template>
