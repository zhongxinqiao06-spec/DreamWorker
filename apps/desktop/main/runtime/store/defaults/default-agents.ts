import type { JsonRecord } from '../../types'

export function defaultAgents(timestamp: string): Record<string, JsonRecord> {
  return {
    agent_general_assistant: {
      agentId: 'agent_general_assistant',
      displayName: '通用助手',
      role: '通用 Agent 聊天入口',
      description: '处理日常问答、上下文整理和轻量任务拆解。',
      systemPrompt: '你是 DreamWorker 的通用助手，优先用中文清晰回答。',
      modelProfileId: 'profile_fast',
      providerId: 'provider_deepseek',
      model: process.env.DEEPSEEK_FAST_MODEL || process.env.DEEPSEEK_MODEL || 'deepseek-v4-flash',
      enabledSkills: ['skill_opportunity_scan'],
      enabledTools: ['tool_model_generate_stub', 'tool_human_input'],
      enabledMcpServers: [],
      runtimeConfig: { contextWindow: 128000, temperature: 0.4, maxTokens: 4096 },
      planner: { enabled: true, strategy: 'react' },
      executor: { timeoutMs: 120000, retryPolicy: 'none' },
      memoryScope: 'project',
      enabled: true,
      builtIn: true,
      createdAt: timestamp,
      updatedAt: timestamp
    }
  }
}
