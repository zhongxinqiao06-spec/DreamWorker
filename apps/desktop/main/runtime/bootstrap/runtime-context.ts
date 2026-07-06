import type { CodingService } from '../coding/coding-service'
import type { RuntimeLifecycle } from '../kernel/runtime-lifecycle'
import type { AgentService } from '../services/agents/agent-service'
import type { McpService } from '../services/agents/mcp-service'
import type { SkillService } from '../services/agents/skill-service'
import type { ToolConfigService } from '../services/agents/tool-config-service'
import type { ChatService } from '../services/chat/chat-service'
import type { ChatStreamService } from '../services/chat/chat-stream-service'
import type { ExtensionService } from '../services/extensions/extension-service'
import type { ProfileService } from '../services/models/profile-service'
import type { ProviderService } from '../services/models/provider-service'
import type { ProjectDirectoryService } from '../services/projects/project-directory-service'
import type { ProjectModuleService } from '../services/projects/project-module-service'
import type { ProjectService } from '../services/projects/project-service'
import type { RequirementAnalysisService } from '../services/requirements/requirement-analysis-service'
import type { RequirementImportService } from '../services/requirements/requirement-import-service'
import type { RequirementPreviewService } from '../services/requirements/requirement-preview-service'
import type { SettingsService } from '../services/settings/settings-service'
import type { WorkspaceStore } from '../store/workspace-store'

export type RuntimeContext = {
  readonly store: WorkspaceStore
  readonly providers: ProviderService
  readonly profiles: ProfileService
  readonly settings: SettingsService
  readonly projects: ProjectService
  readonly projectModules: ProjectModuleService
  readonly projectDirectory: ProjectDirectoryService
  readonly requirementImports: RequirementImportService
  readonly requirementPreview: RequirementPreviewService
  readonly requirementAnalysis: RequirementAnalysisService
  readonly agents: AgentService
  readonly skills: SkillService
  readonly tools: ToolConfigService
  readonly mcp: McpService
  readonly extensions: ExtensionService
  readonly chatService: ChatService
  readonly coding: CodingService
  readonly chat: ChatStreamService
  readonly lifecycle: RuntimeLifecycle
}
