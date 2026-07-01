<script setup lang="ts">
import { computed } from 'vue'
import {
  Bot,
  Copy,
  DatabaseZap,
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
  providerCapabilityOptions,
  providerTemplateOptions,
  providerTypeOptions,
  useAppShellStore,
  type ProviderTemplateId
} from '../stores/app-shell'
import type { ProviderStatus, ProviderType } from '../../../shared/dreamworker-api'

const appShell = useAppShellStore()

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

const profileModelOptions = computed(() => {
  const providerModels =
    appShell.providers.find((provider) => provider.providerId === appShell.profileDraft.providerId)
      ?.availableModels ?? []
  const currentModel = appShell.profileDraft.model.trim()
  if (currentModel && !providerModels.includes(currentModel)) {
    return [currentModel, ...providerModels]
  }
  return providerModels
})

function providerInitial(providerName: string): string {
  return providerName.trim().slice(0, 1).toUpperCase() || 'AI'
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

function splitLines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

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
</script>

<template>
  <section class="resource-page panel-surface">
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
              <img :src="providerLogoSrc[provider.providerType]" alt="" />
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
                <select v-model="appShell.providerDraft.defaultModel">
                  <option
                    v-for="model in splitLines(appShell.providerDraft.availableModelsText)"
                    :key="model"
                    :value="model"
                  >
                    {{ model }}
                  </option>
                </select>
              </label>
              <label>
                API Key
                <input
                  v-model="appShell.providerDraft.apiKey"
                  type="password"
                  placeholder="留空保留当前密钥，保存后只显示脱敏值"
                />
              </label>
            </div>

            <label class="provider-model-textarea">
              可用模型清单
              <textarea v-model="appShell.providerDraft.availableModelsText" />
            </label>
          </section>

          <aside class="provider-model-column" aria-label="能力与模型">
            <div class="provider-column-heading">
              <div class="section-title">
                <DatabaseZap :size="16" aria-hidden="true" />
                <span>能力与模型</span>
              </div>
              <strong
                >{{ splitLines(appShell.providerDraft.availableModelsText).length }} 个模型</strong
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
              <article
                v-for="model in splitLines(appShell.providerDraft.availableModelsText)"
                :key="model"
              >
                <DatabaseZap :size="16" aria-hidden="true" />
                <span>{{ model }}</span>
              </article>
            </section>
          </aside>
        </div>
      </form>
    </section>

    <section v-else-if="appShell.activeResourceTab === 'profiles'" class="resource-grid">
      <aside class="resource-list" aria-label="模型配置">
        <button type="button" class="list-row create-row" @click="appShell.newProfileDraft()">
          <strong>新增模型配置</strong>
          <span>绑定服务商、模型、温度和上下文窗口</span>
        </button>
        <button
          v-for="profile in appShell.profiles"
          :key="profile.profileId"
          class="list-row"
          :class="{ active: profile.profileId === appShell.activeProfileId }"
          type="button"
          @click="appShell.selectProfile(profile.profileId)"
        >
          <strong>{{ profile.displayName }}</strong>
          <span>{{ profile.providerId }} / {{ profile.model }}</span>
          <span>{{ profile.responseFormat }} / {{ profile.toolMode }}</span>
        </button>
      </aside>

      <form class="editor-card" @submit.prevent="appShell.saveProfileDraft()">
        <div class="editor-toolbar">
          <div class="section-title">
            <DatabaseZap :size="18" aria-hidden="true" />
            <span>模型配置</span>
          </div>
          <div class="horizontal-actions">
            <button type="button" @click="appShell.newProfileDraft()">
              <Plus :size="15" aria-hidden="true" />
              新增
            </button>
            <button class="primary-button" type="submit">
              <Save :size="15" aria-hidden="true" />
              保存
            </button>
            <button class="danger-button" type="button" @click="appShell.deleteActiveProfile()">
              <Trash2 :size="15" aria-hidden="true" />
              删除
            </button>
          </div>
        </div>
        <div class="form-grid two">
          <label>
            配置 ID
            <input v-model="appShell.profileDraft.profileId" />
          </label>
          <label>
            名称
            <input v-model="appShell.profileDraft.displayName" />
          </label>
          <label>
            服务商
            <select v-model="appShell.profileDraft.providerId">
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
            <input v-model="appShell.profileDraft.model" list="profile-model-options" />
            <datalist id="profile-model-options">
              <option v-for="model in profileModelOptions" :key="model" :value="model" />
            </datalist>
          </label>
          <label>
            温度
            <input
              v-model.number="appShell.profileDraft.temperature"
              type="number"
              min="0"
              max="2"
              step="0.1"
            />
          </label>
          <label>
            最大输出
            <input v-model.number="appShell.profileDraft.maxTokens" type="number" min="1" />
          </label>
          <label>
            上下文窗口
            <input v-model.number="appShell.profileDraft.contextWindow" type="number" min="1024" />
          </label>
          <label>
            超时毫秒
            <input v-model.number="appShell.profileDraft.timeoutMs" type="number" min="1000" />
          </label>
          <label>
            输出格式
            <select v-model="appShell.profileDraft.responseFormat">
              <option value="text">文本</option>
              <option value="json_object">JSON 对象</option>
              <option value="json_schema">JSON Schema</option>
            </select>
          </label>
          <label>
            工具模式
            <select v-model="appShell.profileDraft.toolMode">
              <option value="none">不调用</option>
              <option value="auto">自动</option>
              <option value="required">必须调用</option>
            </select>
          </label>
        </div>
        <label>
          用途
          <textarea v-model="appShell.profileDraft.purpose" />
        </label>
        <label class="check-row">
          <input v-model="appShell.profileDraft.enabled" type="checkbox" />
          启用
        </label>
      </form>
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
          <span>{{ agent.modelProfileId }}</span>
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
            模型配置
            <select v-model="appShell.agentDraft.modelProfileId">
              <option
                v-for="profile in appShell.profiles"
                :key="profile.profileId"
                :value="profile.profileId"
              >
                {{ profile.displayName }}
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
