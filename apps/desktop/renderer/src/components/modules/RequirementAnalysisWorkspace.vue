<script setup lang="ts">
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  CircleDot,
  Database,
  FileCheck2,
  FileSearch,
  FileSpreadsheet,
  FileText,
  FolderOpen,
  ListChecks,
  Play,
  RefreshCw,
  Sparkles,
  Upload
} from 'lucide-vue-next'
import { computed, watch } from 'vue'
import type { ProjectSubmodule, RequirementSource } from '../../../../shared/dreamworker-api'
import { useAppShellStore } from '../../stores/app-shell'
import { statusLabel } from '../../stores/workspace-navigation'
import ProjectContextPanel from './ProjectContextPanel.vue'

type PipelineStep = {
  readonly phase: string
  readonly title: string
  readonly detail: string
  readonly status: 'ready' | 'running' | 'completed'
}

const appShell = useAppShellStore()

const fallbackSubmodule: ProjectSubmodule = {
  projectId: '',
  moduleId: 'product',
  submoduleId: 'requirement_analysis',
  displayName: '需求分析',
  status: 'ready',
  summary: '根据探索结果或用户上传的项目要求文件，抽取功能清单并生成需求规格说明。',
  defaultAgents: ['agent_product_designer', 'agent_evaluator'],
  enabledSkills: ['skill_prd_draft'],
  enabledTools: ['tool_model_generate_stub', 'tool_artifact_write'],
  outputArtifacts: ['feature_list.xlsx', 'requirements_spec.docx', 'requirements_analysis.json'],
  nextBestAction: '导入需求文件或选择探索产物后运行分析。',
  config: { stage: 'Analyze' }
}

const submodule = computed(() => appShell.requirementAnalysisSubmodule ?? fallbackSubmodule)
const selectedSources = computed(() => appShell.selectedRequirementSources)
const outputFiles = computed(() => appShell.requirementAnalysisRun?.outputFiles ?? [])
const sourcePreview = computed(() => appShell.requirementSourcePreview)
const featurePreview = computed(
  () => appShell.requirementAnalysisRun?.analysis.features.slice(0, 7) ?? []
)
const warningItems = computed(() => appShell.requirementAnalysisRun?.warnings ?? [])
const canRun = computed(
  () => selectedSources.value.length > 0 && !appShell.requirementAnalysisLoading
)
const selectedImportedFiles = computed(
  () => selectedSources.value.filter((source) => source.kind === 'imported_file').length
)
const runStatusLabel = computed(() => {
  if (appShell.requirementAnalysisLoading) {
    return '运行中'
  }
  return appShell.requirementAnalysisRun ? '已生成文档' : '待运行'
})
const pipelineSteps = computed<PipelineStep[]>(() => {
  const completed = Boolean(appShell.requirementAnalysisRun)
  const running = appShell.requirementAnalysisLoading
  return [
    {
      phase: 'INPUT',
      title: '收集需求来源',
      detail: '项目描述、探索产物、用户上传 Word/PDF',
      status: appShell.requirementSources.length > 0 ? 'completed' : 'ready'
    },
    {
      phase: 'PARSE',
      title: 'MinerU 解析文件',
      detail: '对选中的 Word/PDF 提取版面、文本和结构',
      status: completed ? 'completed' : running && selectedImportedFiles.value > 0 ? 'running' : 'ready'
    },
    {
      phase: 'ANALYZE',
      title: '抽取功能清单',
      detail: '识别角色、场景、功能、验收标准和依赖',
      status: completed ? 'completed' : running ? 'running' : 'ready'
    },
    {
      phase: 'WRITE',
      title: '写入交付文档',
      detail: '生成 feature_list.xlsx 与 requirements_spec.docx',
      status: completed ? 'completed' : 'ready'
    }
  ]
})

watch(
  () =>
    [
      appShell.activeProjectId,
      appShell.activeSubmoduleDetail?.moduleId,
      appShell.activeSubmoduleDetail?.submoduleId
    ] as const,
  ([projectId, moduleId, submoduleId]) => {
    if (projectId && moduleId === 'product' && submoduleId === 'requirement_analysis') {
      void appShell.loadRequirementSources()
    }
  },
  { immediate: true }
)

function sourceKindLabel(kind: RequirementSource['kind']): string {
  const labels: Record<RequirementSource['kind'], string> = {
    project_description: '项目描述',
    imported_file: '导入文件',
    explore_artifact: '探索产物'
  }
  return labels[kind]
}

function outputKindLabel(kind: string): string {
  const labels: Record<string, string> = {
    feature_excel: '功能清单 Excel',
    requirements_word: '需求规格 Word',
    analysis_json: '结构化 JSON'
  }
  return labels[kind] ?? kind
}

function sourceMeta(source: RequirementSource): string {
  const count = source.charCount > 0 ? `${source.charCount} 字` : source.mimeType
  return `${sourceKindLabel(source.kind)} · ${count}`
}

function parserLabel(parser: string | undefined): string {
  if (parser === 'mineru') {
    return 'MinerU'
  }
  if (parser === 'direct_text') {
    return '直接文本'
  }
  return parser ?? '暂无'
}

function formatTime(value: string | undefined): string {
  if (!value) {
    return '暂无'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

function stepText(status: PipelineStep['status']): string {
  if (status === 'completed') {
    return '完成'
  }
  if (status === 'running') {
    return '运行中'
  }
  return '待运行'
}
</script>

<template>
  <section class="workspace-layout module-workspace-layout explore-submodule-layout">
    <ProjectContextPanel />

    <section class="module-center panel-surface explore-submodule-center" aria-label="需求分析模块">
      <div class="submodule-detail-header">
        <button
          class="icon-button small"
          type="button"
          title="返回产品模块"
          @click="appShell.leaveSubmoduleDetail()"
        >
          <ArrowLeft :size="16" aria-hidden="true" />
        </button>
        <div>
          <p class="eyebrow">产品模块 / {{ submodule.displayName }}</p>
          <h2>{{ submodule.displayName }}</h2>
          <p>{{ submodule.summary }}</p>
        </div>
        <div class="submodule-detail-actions">
          <button type="button" title="刷新需求来源" @click="appShell.loadRequirementSources()">
            <RefreshCw :size="15" aria-hidden="true" />
            刷新
          </button>
          <button type="button" title="导入 Word 或 PDF 需求文件" @click="appShell.importRequirementFiles()">
            <Upload :size="15" aria-hidden="true" />
            导入
          </button>
          <button
            class="primary-button"
            type="button"
            :disabled="!canRun"
            title="生成功能清单与需求规格说明"
            @click="appShell.runRequirementAnalysis()"
          >
            <Play :size="15" aria-hidden="true" />
            {{ appShell.requirementAnalysisLoading ? '生成中' : '生成文档' }}
          </button>
        </div>
      </div>

      <div class="context-pills explore-runtime-pills">
        <span>{{ statusLabel(submodule.status) }}</span>
        <span>{{ submodule.config.stage ?? 'Analyze' }}</span>
        <span>MinerU 内置解析</span>
        <span>{{ runStatusLabel }}</span>
      </div>

      <section class="explore-runtime-grid requirement-runtime-grid">
        <article class="explore-control-panel">
          <div class="section-heading-row">
            <Database :size="17" aria-hidden="true" />
            <div>
              <p class="eyebrow">Requirement Sources</p>
              <h3>需求来源</h3>
            </div>
          </div>
          <dl>
            <div>
              <dt>项目</dt>
              <dd>{{ appShell.activeProject?.title ?? '暂无项目' }}</dd>
            </div>
            <div>
              <dt>目录</dt>
              <dd>{{ appShell.activeProject?.localDirectoryStatus ?? 'not_set' }}</dd>
            </div>
            <div>
              <dt>解析器</dt>
              <dd>CLI / Open API</dd>
            </div>
          </dl>
          <textarea
            v-model="appShell.requirementAnalysisPrompt"
            placeholder="补充本轮需求分析的偏好、约束或重点，例如优先首版、拆分后台管理、保留合规要求。"
            aria-label="需求分析补充要求"
          />
          <div v-if="appShell.requirementSources.length" class="source-toggle-grid requirement-source-grid">
            <button
              v-for="source in appShell.requirementSources"
              :key="source.sourceId"
              type="button"
              :class="{
                active: appShell.selectedRequirementSourceIds.includes(source.sourceId),
                previewing: sourcePreview?.source.sourceId === source.sourceId
              }"
              @click="appShell.selectRequirementSourceForPreview(source.sourceId)"
            >
              <FileText :size="14" aria-hidden="true" />
              <span>{{ source.fileName }}</span>
              <small>{{ sourceMeta(source) }}</small>
            </button>
          </div>
          <div v-else class="opportunity-empty-result">
            <Upload :size="17" aria-hidden="true" />
            <span>等待项目描述、探索产物或导入需求文件</span>
          </div>
        </article>

        <div class="explore-analysis-column">
          <article class="explore-run-panel requirement-preview-panel">
            <div class="section-heading-row">
              <FileSearch :size="17" aria-hidden="true" />
              <div>
                <p class="eyebrow">MinerU Preview</p>
                <h3>解析预览</h3>
              </div>
            </div>
            <div v-if="appShell.requirementSourcePreviewLoading" class="opportunity-empty-result">
              <RefreshCw :size="17" aria-hidden="true" />
              <span>解析中</span>
            </div>
            <div v-else-if="sourcePreview" class="requirement-preview-body">
              <dl class="requirement-preview-meta">
                <div>
                  <dt>文件</dt>
                  <dd>{{ sourcePreview.source.fileName }}</dd>
                </div>
                <div>
                  <dt>解析</dt>
                  <dd>{{ parserLabel(sourcePreview.parser) }}</dd>
                </div>
                <div>
                  <dt>字数</dt>
                  <dd>{{ sourcePreview.charCount }}</dd>
                </div>
              </dl>
              <pre>{{ sourcePreview.content }}</pre>
              <small v-if="sourcePreview.truncated">已截取前 24000 字用于界面预览</small>
            </div>
            <div v-else class="opportunity-empty-result">
              <FileText :size="17" aria-hidden="true" />
              <span>选择来源后显示解析文本</span>
            </div>
          </article>

          <article class="explore-run-panel">
            <div class="section-heading-row">
              <Sparkles :size="17" aria-hidden="true" />
              <div>
                <p class="eyebrow">Run Pipeline</p>
                <h3>分析管线</h3>
              </div>
            </div>
            <div class="explore-runtime-summary">
              <div>
                <Database :size="16" aria-hidden="true" />
                <strong>{{ selectedSources.length }}</strong>
                <span>Sources</span>
              </div>
              <div>
                <ListChecks :size="16" aria-hidden="true" />
                <strong>{{ appShell.requirementAnalysisRun?.featureCount ?? 0 }}</strong>
                <span>Features</span>
              </div>
              <div>
                <FileCheck2 :size="16" aria-hidden="true" />
                <strong>{{ outputFiles.length }}</strong>
                <span>Outputs</span>
              </div>
            </div>
            <ol class="runtime-steps explore-steps">
              <li v-for="step in pipelineSteps" :key="step.phase" :data-status="step.status">
                <CheckCircle2 v-if="step.status === 'completed'" :size="15" aria-hidden="true" />
                <CircleDot v-else :size="15" aria-hidden="true" />
                <div>
                  <strong>{{ step.phase }} · {{ step.title }}</strong>
                  <small>{{ step.detail }}</small>
                </div>
                <span>{{ stepText(step.status) }}</span>
              </li>
            </ol>
          </article>

          <article class="comparison-panel">
            <div class="section-heading-row">
              <FileSpreadsheet :size="17" aria-hidden="true" />
              <div>
                <p class="eyebrow">Artifacts</p>
                <h3>交付文件</h3>
              </div>
            </div>
            <div v-if="outputFiles.length" class="artifact-slot-list requirement-output-list">
              <button v-for="file in outputFiles" :key="file.relativePath" type="button">
                <FileText :size="15" aria-hidden="true" />
                <span>{{ outputKindLabel(file.kind) }}</span>
                <small>{{ file.relativePath }}</small>
              </button>
            </div>
            <div v-else class="opportunity-empty-result">
              <FileCheck2 :size="17" aria-hidden="true" />
              <span>运行后生成 Excel、Word 和 JSON 产物</span>
            </div>
          </article>
        </div>

        <article class="explore-run-panel">
          <div class="section-heading-row">
            <ListChecks :size="17" aria-hidden="true" />
            <div>
              <p class="eyebrow">Feature List</p>
              <h3>功能清单预览</h3>
            </div>
          </div>
          <p class="runtime-save-line">
            {{
              appShell.requirementAnalysisRun
                ? appShell.requirementAnalysisRun.analysis.summary
                : '等待需求分析生成结构化功能项。'
            }}
          </p>
          <div class="explore-insight-list requirement-feature-list">
            <div v-if="featurePreview.length === 0" class="explore-empty-result">
              <Sparkles :size="17" aria-hidden="true" />
              <span>暂无功能项</span>
            </div>
            <article v-for="feature in featurePreview" :key="feature.featureId">
              <strong>{{ feature.featureId }} · {{ feature.name }}</strong>
              <span>{{ feature.description }}</span>
              <small>{{ feature.module }} · {{ feature.priority }} · {{ feature.type }}</small>
            </article>
          </div>
        </article>
      </section>
    </section>

    <aside class="right-panel explore-runtime-side" aria-label="需求分析运行配置">
      <section class="inspector-card">
        <p class="eyebrow">Runtime</p>
        <h3>需求分析运行</h3>
        <dl>
          <div>
            <dt>来源</dt>
            <dd>{{ appShell.requirementSources.length }}</dd>
          </div>
          <div>
            <dt>已选</dt>
            <dd>{{ selectedSources.length }}</dd>
          </div>
          <div>
            <dt>更新时间</dt>
            <dd>{{ formatTime(appShell.requirementAnalysisRun?.createdAt) }}</dd>
          </div>
          <div>
            <dt>Trace</dt>
            <dd>{{ appShell.requirementAnalysisRun?.traceId ?? '暂无' }}</dd>
          </div>
        </dl>
      </section>

      <section class="inspector-card">
        <div class="section-heading-row">
          <FileText :size="16" aria-hidden="true" />
          <div>
            <p class="eyebrow">MinerU</p>
            <h3>文档解析</h3>
          </div>
        </div>
        <p>选中 Word/PDF 时优先调用本地 MinerU CLI；系统未安装时自动使用官方 Go SDK 的 Open API。纯项目描述或探索产物可直接分析。</p>
      </section>

      <section class="inspector-card">
        <div class="section-heading-row">
          <AlertTriangle :size="16" aria-hidden="true" />
          <div>
            <p class="eyebrow">Risk</p>
            <h3>风险与待确认</h3>
          </div>
        </div>
        <div class="tag-list">
          <span v-for="warning in warningItems" :key="warning">{{ warning }}</span>
          <span v-for="risk in appShell.requirementAnalysisRun?.analysis.risks ?? []" :key="risk">
            {{ risk }}
          </span>
          <span v-if="warningItems.length === 0 && !appShell.requirementAnalysisRun">
            等待运行结果
          </span>
        </div>
      </section>

      <section class="inspector-card">
        <p class="eyebrow">Next Action</p>
        <h3>{{ submodule.nextBestAction }}</h3>
        <button type="button" class="full-width-button" @click="appShell.openActiveProjectDirectory()">
          <FolderOpen :size="15" aria-hidden="true" />
          打开项目目录
        </button>
      </section>
    </aside>
  </section>
</template>

