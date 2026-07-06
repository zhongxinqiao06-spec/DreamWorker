import { CodingService } from '../coding/coding-service'
import { RuntimeLifecycle } from '../kernel/runtime-lifecycle'
import { AgentService } from '../services/agents/agent-service'
import { McpService } from '../services/agents/mcp-service'
import { SkillService } from '../services/agents/skill-service'
import { ToolConfigService } from '../services/agents/tool-config-service'
import { ChatService } from '../services/chat/chat-service'
import { ChatStreamService } from '../services/chat/chat-stream-service'
import { ExtensionService } from '../services/extensions/extension-service'
import { ModelGateway } from '../services/models/model-gateway'
import { ProfileService } from '../services/models/profile-service'
import { ProviderService } from '../services/models/provider-service'
import { ProjectDirectoryService } from '../services/projects/project-directory-service'
import { ProjectModuleService } from '../services/projects/project-module-service'
import { ProjectService } from '../services/projects/project-service'
import { RequirementAnalysisService } from '../services/requirements/requirement-analysis-service'
import { RequirementImportService } from '../services/requirements/requirement-import-service'
import { RequirementPreviewService } from '../services/requirements/requirement-preview-service'
import { SettingsService } from '../services/settings/settings-service'
import { ChatRepository } from '../store/repositories/chat-repository'
import { ProjectModuleRepository } from '../store/repositories/project-module-repository'
import { AgentRepository } from '../store/repositories/agent-repository'
import { McpRepository } from '../store/repositories/mcp-repository'
import { ProfileRepository } from '../store/repositories/profile-repository'
import { ProviderRepository } from '../store/repositories/provider-repository'
import { ProjectRepository } from '../store/repositories/project-repository'
import { SettingsRepository } from '../store/repositories/settings-repository'
import { SkillRepository } from '../store/repositories/skill-repository'
import { ToolRepository } from '../store/repositories/tool-repository'
import { WorkspaceStore } from '../store/workspace-store'
import type { RuntimeContext } from './runtime-context'

export function createRuntimeContext(configDir?: string): RuntimeContext {
  const store = new WorkspaceStore(configDir)
  const providers = new ProviderService(new ProviderRepository(store))
  const profiles = new ProfileService(new ProfileRepository(store))
  const modelGateway = new ModelGateway(providers)
  const settings = new SettingsService(new SettingsRepository(store))
  const projectRepository = new ProjectRepository(store)
  const projectModuleRepository = new ProjectModuleRepository(store)
  const projectDirectory = new ProjectDirectoryService(projectRepository)
  const projects = new ProjectService(projectRepository, projectModuleRepository, projectDirectory)
  const projectModules = new ProjectModuleService(projects, projectModuleRepository)
  const requirementImports = new RequirementImportService(store, projectDirectory)
  const requirementPreview = new RequirementPreviewService(requirementImports)
  const requirementAnalysis = new RequirementAnalysisService(store, projects, projectDirectory)
  const agents = new AgentService(new AgentRepository(store))
  const skills = new SkillService(new SkillRepository(store))
  const tools = new ToolConfigService(new ToolRepository(store))
  const mcp = new McpService(new McpRepository(store))
  const extensions = new ExtensionService(settings, store.configDir)
  const chatService = new ChatService(new ChatRepository(store))
  const coding = new CodingService(store, modelGateway, projectDirectory)
  const chat = new ChatStreamService(chatService)
  const lifecycle = new RuntimeLifecycle()

  lifecycle.add({ stop: () => store.close() })
  lifecycle.add({ stop: () => coding.dispose() })

  return {
    store,
    providers,
    profiles,
    modelGateway,
    settings,
    projects,
    projectModules,
    projectDirectory,
    requirementImports,
    requirementPreview,
    requirementAnalysis,
    agents,
    skills,
    tools,
    mcp,
    extensions,
    chatService,
    coding,
    chat,
    lifecycle
  }
}
