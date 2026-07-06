import { describe, expect, it } from 'vitest'
import {
  CONTRACT_SCHEMA_VERSION,
  createEngineNotConnectedResponse,
  type RuntimePingResponse,
  type SafeModelProvider
} from './dreamworker-api'
import type {
  AgentSpec,
  AgentTask,
  ArtifactMetadata,
  DreamWorkerError,
  EventEnvelope
} from './generated/contracts'

describe('generated contract types', () => {
  it('keeps runtime.ping responses versioned', () => {
    const response: RuntimePingResponse = {
      schema_version: CONTRACT_SCHEMA_VERSION,
      ok: true,
      engineVersion: '0.1.0',
      trace_id: 'tr_contract'
    }

    expect(response).toEqual({
      schema_version: '0.1',
      ok: true,
      engineVersion: '0.1.0',
      trace_id: 'tr_contract'
    })
  })

  it('creates a failure response backed by the generated error envelope', () => {
    const response = createEngineNotConnectedResponse('tr_failure')

    expect(response.schema_version).toBe('0.1')
    expect(response.ok).toBe(false)
    if (!response.ok) {
      const error: DreamWorkerError = response.error
      expect(error).toEqual({
        code: 'ENGINE_NOT_CONNECTED',
        message: 'Main Runtime 尚未连接。',
        recoverable: true,
        user_action: '等待本地 Runtime 启动后重试。',
        trace_id: 'tr_failure'
      })
    }
  })

  it('types the MVP event and artifact envelopes used by cross-process contracts', () => {
    const event: EventEnvelope = {
      event_id: 'evt_contract',
      schema_version: '0.1',
      trace_id: 'tr_contract',
      mission_id: 'msn_contract',
      run_id: 'run_contract',
      actor: 'orchestrator',
      timestamp: '2026-06-30T00:00:00Z',
      type: 'mission.created',
      payload: {
        title: 'AI 项目孵化器'
      }
    }

    const artifact: ArtifactMetadata = {
      schema_version: '0.1',
      artifact_id: 'art_contract',
      mission_id: 'msn_contract',
      run_id: 'run_contract',
      kind: 'dream_brief',
      title: 'Dream Brief',
      version: 1,
      uri: 'artifact://msn_contract/dream_brief.md',
      content_type: 'text/markdown'
    }

    expect(event.schema_version).toBe(CONTRACT_SCHEMA_VERSION)
    expect(artifact.schema_version).toBe(CONTRACT_SCHEMA_VERSION)
  })

  it('types the PR-06 agent and task contracts', () => {
    const agent: AgentSpec = {
      schema_version: '0.1',
      id: 'product_analyst',
      role: 'Analyze users and MVP scope.',
      input_schema: {},
      output_schema: {},
      allowed_capabilities: ['cap_model_generate_stub', 'cap_artifact_write'],
      default_model_profile: 'stub_reasoning_light',
      budget: { max_tokens: 20000, max_cost_usd: 0 },
      timeout: '120s',
      approval_policy: 'on_risk',
      expected_artifacts: ['dream_brief.md'],
      prompt_ref: {
        prompt_id: 'prm_product_analyst',
        prompt_version: 'v1',
        agent_id: 'product_analyst'
      }
    }

    const task: AgentTask = {
      schema_version: '0.1',
      task_id: 'tsk_discover_brief',
      stage: 'Discover',
      goal: 'Generate Dream Brief.',
      assigned_agent: agent.id,
      required_capabilities: ['cap_model_generate_stub', 'cap_artifact_write'],
      expected_artifacts: ['dream_brief.md'],
      depends_on: [],
      budget: { max_tokens: 4000, max_cost_usd: 0 },
      status: 'pending',
      trace_id: 'tr_contract'
    }

    expect(task.assigned_agent).toBe(agent.id)
    expect(agent.prompt_ref.prompt_version).toBe('v1')
  })

  it('types safe model providers without raw api keys', () => {
    const provider: SafeModelProvider = {
      providerId: 'provider_deepseek',
      providerType: 'deepseek',
      displayName: 'DeepSeek 兼容服务',
      baseURL: 'https://api.deepseek.com',
      organization: null,
      project: null,
      defaultModel: 'deepseek-v4-flash',
      availableModels: ['deepseek-v4-flash'],
      enabled: true,
      status: 'connected',
      capabilities: ['chat', 'tools', 'json_schema'],
      supportsStreaming: true,
      healthStatus: 'connected',
      modelCount: 1,
      latencyMs: 18,
      lastDiscoveryAt: null,
      lastStreamAt: null,
      lastErrorCode: null,
      streamingVerified: true,
      hasApiKey: true,
      maskedKey: 'sk-b...4f3c',
      lastTestedAt: '2026-07-01T00:00:00Z',
      lastError: null,
      createdAt: '2026-07-01T00:00:00Z',
      updatedAt: '2026-07-01T00:00:00Z'
    }

    expect(JSON.stringify(provider)).not.toContain('apiKey')
  })
})
