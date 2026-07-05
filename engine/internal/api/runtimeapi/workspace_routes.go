package runtimeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/workspace"
)

type workspaceHandler struct {
	token string
	store *workspace.Store
}

func RegisterWorkspaceRoutes(mux *http.ServeMux, token string, store *workspace.Store) {
	handler := workspaceHandler{token: token, store: store}
	registerGet(mux, handler, "/models/providers", store.ListProviders)
	registerPost(mux, handler, "/models/providers/save", store.SaveProvider)
	registerPostID(mux, handler, "/models/providers/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteProvider(request.ProviderID)
	})
	registerPostID(mux, handler, "/models/providers/test", func(request workspace.IDRequest) (workspace.TestResult, *workspace.AppError) {
		return store.TestProvider(request.ProviderID)
	})
	registerPostID(mux, handler, "/models/providers/refresh-models", func(request workspace.IDRequest) (workspace.SafeModelProvider, *workspace.AppError) {
		return store.RefreshProviderModels(request.ProviderID)
	})
	registerGet(mux, handler, "/models/profiles", store.ListProfiles)
	registerPost(mux, handler, "/models/profiles/save", store.SaveProfile)
	registerPostID(mux, handler, "/models/profiles/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteProfile(request.ProfileID)
	})

	registerGet(mux, handler, "/settings", store.GetSettings)
	registerPost(mux, handler, "/settings/update", store.UpdateSettings)
	registerPost(mux, handler, "/settings/reset-extension", store.ResetExtensionSettings)

	registerGet(mux, handler, "/extensions", store.ListExtensions)
	registerPost(mux, handler, "/extensions/status", store.GetExtensionStatus)
	registerPost(mux, handler, "/extensions/detect", store.DetectExtension)
	registerPost(mux, handler, "/extensions/install", store.InstallExtension)
	registerPost(mux, handler, "/extensions/start", store.StartExtension)
	registerPost(mux, handler, "/extensions/stop", store.StopExtension)
	registerPost(mux, handler, "/extensions/restart", store.RestartExtension)
	registerPost(mux, handler, "/extensions/test", store.TestExtension)
	registerPost(mux, handler, "/extensions/refresh-models", store.RefreshExtensionModels)
	registerPost(mux, handler, "/extensions/verify-streaming", store.VerifyExtensionStreaming)
	registerPost(mux, handler, "/extensions/logs/tail", store.TailExtensionLogs)
	registerPost(mux, handler, "/extensions/logs/clear", store.ClearExtensionLogs)

	registerGet(mux, handler, "/agents", store.ListAgents)
	registerPostID(mux, handler, "/agents/get", func(request workspace.IDRequest) (workspace.AgentConfig, *workspace.AppError) {
		return store.GetAgent(request.AgentID)
	})
	registerPost(mux, handler, "/agents/save", store.SaveAgent)
	registerPostID(mux, handler, "/agents/duplicate", func(request workspace.IDRequest) (workspace.AgentConfig, *workspace.AppError) {
		return store.DuplicateAgent(request.AgentID)
	})
	registerPostID(mux, handler, "/agents/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteAgent(request.AgentID)
	})

	registerGet(mux, handler, "/skills", store.ListSkills)
	registerPostID(mux, handler, "/skills/get", func(request workspace.IDRequest) (workspace.SkillConfig, *workspace.AppError) {
		return store.GetSkill(request.SkillID)
	})
	registerPost(mux, handler, "/skills/save", store.SaveSkill)
	registerPostID(mux, handler, "/skills/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteSkill(request.SkillID)
	})

	registerGet(mux, handler, "/tools", store.ListTools)
	registerPostID(mux, handler, "/tools/get", func(request workspace.IDRequest) (workspace.ToolConfig, *workspace.AppError) {
		return store.GetTool(request.ToolID)
	})
	registerPost(mux, handler, "/tools/save", store.SaveTool)
	registerPostID(mux, handler, "/tools/set-enabled", func(request workspace.IDRequest) (workspace.ToolConfig, *workspace.AppError) {
		return store.SetToolEnabled(request.ToolID, request.Enabled)
	})
	registerPostID(mux, handler, "/tools/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteTool(request.ToolID)
	})

	registerGet(mux, handler, "/mcp/servers", store.ListMCPServers)
	registerPost(mux, handler, "/mcp/servers/save", store.SaveMCPServer)
	registerPostID(mux, handler, "/mcp/servers/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteMCPServer(request.ServerID)
	})
	registerPostID(mux, handler, "/mcp/servers/test", func(request workspace.IDRequest) (workspace.TestResult, *workspace.AppError) {
		return store.TestMCPServer(request.ServerID)
	})
	registerPostID(mux, handler, "/mcp/servers/refresh-tools", func(request workspace.IDRequest) ([]workspace.ToolConfig, *workspace.AppError) {
		return store.RefreshMCPTools(request.ServerID)
	})

	registerGet(mux, handler, "/projects", store.ListProjects)
	registerPost(mux, handler, "/projects/create", store.CreateProject)
	registerPostID(mux, handler, "/projects/get", func(request workspace.IDRequest) (workspace.Project, *workspace.AppError) {
		return store.GetProject(request.ProjectID)
	})
	registerPost(mux, handler, "/projects/update", store.UpdateProject)
	registerPostID(mux, handler, "/projects/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteProject(request.ProjectID)
	})
	registerPostID(mux, handler, "/projects/local-directory/validate", func(request workspace.IDRequest) (workspace.ProjectDirectoryCheck, *workspace.AppError) {
		return store.ValidateLocalDirectory(request.ProjectID)
	})
	registerPostID(mux, handler, "/projects/local-directory/initialize", func(request workspace.IDRequest) (workspace.ProjectDirectoryCheck, *workspace.AppError) {
		return store.InitializeLocalDirectory(request.ProjectID)
	})
	registerPostID(mux, handler, "/projects/export-manifest", func(request workspace.IDRequest) (workspace.ProjectManifestExport, *workspace.AppError) {
		return store.ExportProjectManifest(request.ProjectID)
	})
	registerPostID(mux, handler, "/projects/modules", func(request workspace.IDRequest) ([]workspace.ProjectModule, *workspace.AppError) {
		return store.ListProjectModules(request.ProjectID)
	})
	registerPost(mux, handler, "/projects/modules/get", store.GetProjectModule)
	registerPost(mux, handler, "/projects/modules/update-config", store.UpdateProjectModuleConfig)
	registerPost(mux, handler, "/projects/requirements/import-files", store.ImportRequirementFiles)
	registerPostID(mux, handler, "/projects/requirements/sources", func(request workspace.IDRequest) (workspace.RequirementSourcesResult, *workspace.AppError) {
		return store.ListRequirementSources(request.ProjectID)
	})
	registerPost(mux, handler, "/projects/requirements/preview-source", func(request workspace.PreviewRequirementSourceInput) (workspace.RequirementSourcePreviewResult, *workspace.AppError) {
		return store.PreviewRequirementSource(context.Background(), request)
	})
	registerPost(mux, handler, "/projects/requirements/run", func(request workspace.RunRequirementAnalysisInput) (workspace.RequirementAnalysisRun, *workspace.AppError) {
		return store.RunRequirementAnalysis(context.Background(), request)
	})

	registerGet(mux, handler, "/chat/sessions", store.ListChatSessions)
	registerPost(mux, handler, "/chat/sessions/create", store.CreateChatSession)
	registerPost(mux, handler, "/chat/sessions/update", store.UpdateChatSession)
	registerPostID(mux, handler, "/chat/messages", func(request workspace.IDRequest) ([]workspace.ChatMessage, *workspace.AppError) {
		return store.ListChatMessages(request.SessionID)
	})
	registerChatStream(mux, handler, "/chat/messages/stream")
	registerPost(mux, handler, "/chat/messages/cancel", store.CancelChatStream)
	registerPost(mux, handler, "/chat/messages/send", store.SendChatMessage)
	registerPost(mux, handler, "/chat/images/generate", func(request workspace.GenerateChatImageInput) (workspace.ChatTurnResult, *workspace.AppError) {
		return store.GenerateChatImage(context.Background(), request)
	})
	registerPostID(mux, handler, "/chat/sessions/delete", func(request workspace.IDRequest) (workspace.DeleteResult, *workspace.AppError) {
		return store.DeleteChatSession(request.SessionID)
	})

	registerGet(mux, handler, "/coding/engines", store.ListCodingEngines)
	registerPost(mux, handler, "/coding/sessions/create", store.CreateCodingSession)
	registerPost(mux, handler, "/coding/sessions/get", store.GetCodingSession)
	registerPost(mux, handler, "/coding/files/list", store.ListCodingFiles)
	registerPost(mux, handler, "/coding/files/read", store.ReadCodingFile)
	registerPost(mux, handler, "/coding/files/status", store.CodingFileStatus)
	registerCodingStream(mux, handler, "/coding/turns/stream")
	registerPost(mux, handler, "/coding/turns/cancel", store.CancelCodingTurn)
}

func registerChatStream(mux *http.ServeMux, h workspaceHandler, path string) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if !h.guard(w, r, http.MethodPost) {
			return
		}
		request, ok := decodeWorkspaceRequest[workspace.SendChatMessageInput](w, r)
		if !ok {
			return
		}
		events, appErr := h.store.StreamChatMessage(r.Context(), request)
		if appErr != nil {
			writeWorkspaceError(w, appErr)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "STREAM_UNSUPPORTED", "streaming is not supported by this response writer", "retry the request")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		for event := range events {
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	})
}

func registerCodingStream(mux *http.ServeMux, h workspaceHandler, path string) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if !h.guard(w, r, http.MethodPost) {
			return
		}
		request, ok := decodeWorkspaceRequest[workspace.CodingTurnInput](w, r)
		if !ok {
			return
		}
		events, appErr := h.store.StreamCodingTurn(r.Context(), request)
		if appErr != nil {
			writeWorkspaceError(w, appErr)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "STREAM_UNSUPPORTED", "streaming is not supported by this response writer", "retry the request")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		for event := range events {
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	})
}

func registerGet[T any](mux *http.ServeMux, h workspaceHandler, path string, action func() T) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if !h.guard(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, http.StatusOK, action())
	})
}

func registerPost[Req any, Resp any](
	mux *http.ServeMux,
	h workspaceHandler,
	path string,
	action func(Req) (Resp, *workspace.AppError),
) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if !h.guard(w, r, http.MethodPost) {
			return
		}
		request, ok := decodeWorkspaceRequest[Req](w, r)
		if !ok {
			return
		}
		response, appErr := action(request)
		if appErr != nil {
			writeWorkspaceError(w, appErr)
			return
		}
		writeJSON(w, http.StatusOK, response)
	})
}

func registerPostID[Resp any](
	mux *http.ServeMux,
	h workspaceHandler,
	path string,
	action func(workspace.IDRequest) (Resp, *workspace.AppError),
) {
	registerPost(mux, h, path, action)
}

func (h workspaceHandler) guard(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "当前接口不支持这个 HTTP 方法。", "请通过 Electron Main 使用正确的 typed API。")
		return false
	}
	if !hasValidToken(r, h.token) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "缺少有效的本地引擎访问令牌。", "请通过 Electron Main 持有的本地令牌重新调用。")
		return false
	}
	return true
}

func decodeWorkspaceRequest[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var request T
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "请求 JSON 格式无效。", "请检查 typed API 入参。")
		return request, false
	}
	return request, true
}

func writeWorkspaceError(w http.ResponseWriter, appErr *workspace.AppError) {
	writeError(w, appErr.Status, appErr.Code, appErr.Message, appErr.UserAction)
}
