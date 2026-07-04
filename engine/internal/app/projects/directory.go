package projects

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var projectDirectoryLayout = []string{
	".dreamworker",
	".dreamworker/runs",
	".dreamworker/logs",
	".dreamworker/cache",
	".dreamworker/indexes",
	"docs",
	"artifacts",
	"artifacts/explore",
	"artifacts/product",
	"artifacts/development",
	"artifacts/sales",
	"workspace",
	"workspace/imports",
	"workspace/exports",
	"workspace/temp",
	"source",
	"source/repo",
}

var projectDocumentStubs = map[string]string{
	"docs/dream_brief.md":            "# Dream Brief\n\n",
	"docs/research_pack.md":          "# Research Pack\n\n",
	"docs/prd.md":                    "# PRD\n\n",
	"docs/architecture_blueprint.md": "# Architecture Blueprint\n\n",
	"docs/launch_plan.md":            "# Launch Plan\n\n",
}

func (s *Store) ValidateLocalDirectory(projectID string) (ProjectDirectoryCheck, *AppError) {
	project, appErr := s.projectForDirectoryOperation(projectID)
	if appErr != nil {
		return ProjectDirectoryCheck{}, appErr
	}
	check := s.inspectProjectDirectory(project)
	if appErr := s.applyDirectoryCheck(check); appErr != nil {
		return ProjectDirectoryCheck{}, appErr
	}
	return check, nil
}

func (s *Store) InitializeLocalDirectory(projectID string) (ProjectDirectoryCheck, *AppError) {
	project, appErr := s.projectForDirectoryOperation(projectID)
	if appErr != nil {
		return ProjectDirectoryCheck{}, appErr
	}
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		check := s.inspectProjectDirectory(project)
		_ = s.applyDirectoryCheck(check)
		return check, BadRequest("LOCAL_DIRECTORY_NOT_SET", "项目尚未绑定本地目录。", "请先选择项目本地目录。")
	}
	root := filepath.Clean(*project.LocalRootPath)
	for _, relativePath := range projectDirectoryLayout {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(relativePath)), 0o755); err != nil {
			check := s.inspectProjectDirectory(project)
			_ = s.applyDirectoryCheck(check)
			return check, BadRequest("LOCAL_DIRECTORY_INIT_FAILED", "项目目录初始化失败。", "请确认目录可写后重试。")
		}
	}
	for relativePath, content := range projectDocumentStubs {
		path := filepath.Join(root, filepath.FromSlash(relativePath))
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			check := s.inspectProjectDirectory(project)
			_ = s.applyDirectoryCheck(check)
			return check, BadRequest("LOCAL_DIRECTORY_INIT_FAILED", "项目文档占位文件写入失败。", "请确认 docs 目录可写后重试。")
		}
	}
	if _, appErr := s.writeProjectManifestFiles(project); appErr != nil {
		check := s.inspectProjectDirectory(project)
		_ = s.applyDirectoryCheck(check)
		return check, appErr
	}
	check := s.inspectProjectDirectory(project)
	if appErr := s.applyDirectoryCheck(check); appErr != nil {
		return ProjectDirectoryCheck{}, appErr
	}
	return check, nil
}

func (s *Store) ExportProjectManifest(projectID string) (ProjectManifestExport, *AppError) {
	project, appErr := s.projectForDirectoryOperation(projectID)
	if appErr != nil {
		return ProjectManifestExport{}, appErr
	}
	manifest := s.projectManifest(project)
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		return ProjectManifestExport{
			ProjectID:     project.ProjectID,
			LocalRootPath: nil,
			ManifestPath:  nil,
			Manifest:      manifest,
		}, nil
	}
	manifestPath, appErr := s.writeProjectManifestFiles(project)
	if appErr != nil {
		return ProjectManifestExport{}, appErr
	}
	return ProjectManifestExport{
		ProjectID:     project.ProjectID,
		LocalRootPath: project.LocalRootPath,
		ManifestPath:  &manifestPath,
		Manifest:      manifest,
	}, nil
}

func (s *Store) projectForDirectoryOperation(projectID string) (Project, *AppError) {
	if projectID == "" {
		return Project{}, BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择项目后重试。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	project, ok := s.Projects[projectID]
	if !ok {
		return Project{}, NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	project = normalizeProject(project)
	s.Projects[projectID] = project
	return project, nil
}

func (s *Store) inspectProjectDirectory(project Project) ProjectDirectoryCheck {
	now := s.Now()
	check := ProjectDirectoryCheck{
		ProjectID:           project.ProjectID,
		LocalRootPath:       project.LocalRootPath,
		Status:              "not_set",
		LastCheckedAt:       now,
		RequiredDirectories: []ProjectDirectoryEntryCheck{},
		Message:             "项目尚未绑定本地目录。",
	}
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		return check
	}
	root := filepath.Clean(*project.LocalRootPath)
	check.LocalRootPath = &root
	info, err := os.Stat(root)
	if err != nil {
		check.Status = "missing"
		check.Message = "本地目录不存在。"
		return check
	}
	if !info.IsDir() {
		check.Status = "invalid"
		check.Message = "本地路径不是目录。"
		return check
	}
	check.Exists = true
	check.Readable = directoryReadable(root)
	check.Writable = directoryWritable(root)
	check.RequiredDirectories = inspectRequiredDirectories(root)
	check.DreamworkerInitialized = directoryExists(filepath.Join(root, ".dreamworker"))
	if !check.Readable || !check.Writable {
		check.Status = "permission_denied"
		check.Message = "本地目录读写权限不足。"
		return check
	}
	if !check.DreamworkerInitialized || !allRequiredDirectoriesExist(check.RequiredDirectories) {
		check.Status = "invalid"
		check.Message = "本地目录尚未初始化 DreamWorker 项目结构。"
		return check
	}
	check.Status = "valid"
	check.Message = "本地目录可用，项目结构完整。"
	return check
}

func (s *Store) applyDirectoryCheck(check ProjectDirectoryCheck) *AppError {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	project, ok := s.Projects[check.ProjectID]
	if !ok {
		return nil
	}
	project = normalizeProject(project)
	project.LocalRootPath = check.LocalRootPath
	project.LocalDirectoryStatus = check.Status
	lastCheckedAt := check.LastCheckedAt
	project.LocalDirectoryLastCheckedAt = &lastCheckedAt
	project.UpdatedAt = s.Now()
	s.Projects[check.ProjectID] = project
	return s.PersistWorkspaceSnapshotLocked()
}

func (s *Store) writeProjectManifestFiles(project Project) (string, *AppError) {
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		return "", BadRequest("LOCAL_DIRECTORY_NOT_SET", "项目尚未绑定本地目录。", "请先选择项目本地目录。")
	}
	root := filepath.Clean(*project.LocalRootPath)
	metadataDir := filepath.Join(root, ".dreamworker")
	if err := os.MkdirAll(metadataDir, 0o755); err != nil {
		return "", BadRequest("LOCAL_DIRECTORY_WRITE_FAILED", "无法写入 .dreamworker 目录。", "请确认项目目录可写。")
	}
	projectPayload, err := json.MarshalIndent(s.projectManifest(project)["project"], "", "  ")
	if err != nil {
		return "", BadRequest("PROJECT_MANIFEST_FAILED", "项目配置序列化失败。", "请检查项目配置。")
	}
	if err := os.WriteFile(filepath.Join(metadataDir, "project.json"), projectPayload, 0o644); err != nil {
		return "", BadRequest("LOCAL_DIRECTORY_WRITE_FAILED", "project.json 写入失败。", "请确认项目目录可写。")
	}
	manifestPath := filepath.Join(metadataDir, "manifest.json")
	manifestPayload, err := json.MarshalIndent(s.projectManifest(project), "", "  ")
	if err != nil {
		return "", BadRequest("PROJECT_MANIFEST_FAILED", "项目 manifest 序列化失败。", "请检查项目配置。")
	}
	if err := os.WriteFile(manifestPath, manifestPayload, 0o644); err != nil {
		return "", BadRequest("LOCAL_DIRECTORY_WRITE_FAILED", "manifest.json 写入失败。", "请确认项目目录可写。")
	}
	return manifestPath, nil
}

func (s *Store) writeProjectManifestFilesIfInitialized(project Project) *AppError {
	if project.LocalRootPath == nil || strings.TrimSpace(*project.LocalRootPath) == "" {
		return nil
	}
	root := filepath.Clean(*project.LocalRootPath)
	if !directoryExists(filepath.Join(root, ".dreamworker")) {
		return nil
	}
	_, appErr := s.writeProjectManifestFiles(project)
	return appErr
}

func (s *Store) projectManifest(project Project) map[string]any {
	project = normalizeProject(project)
	return map[string]any{
		"schemaVersion": "dreamworker.project.v1",
		"exportedAt":    s.Now(),
		"project": map[string]any{
			"projectId":             project.ProjectID,
			"title":                 project.Title,
			"description":           project.Description,
			"status":                project.Status,
			"localRootPath":         project.LocalRootPath,
			"localDirectoryStatus":  project.LocalDirectoryStatus,
			"defaultModelProfileId": project.DefaultModelProfileID,
			"defaultRouteProfileId": project.DefaultRouteProfileID,
			"enabledAgents":         project.EnabledAgents,
			"enabledSkills":         project.EnabledSkills,
			"enabledTools":          project.EnabledTools,
			"enabledMcpServers":     project.EnabledMCPServers,
			"moduleConfigs":         project.ModuleConfigs,
			"memoryConfig":          project.MemoryConfig,
			"runPolicy":             project.RunPolicy,
			"securityPolicy":        project.SecurityPolicy,
			"createdAt":             project.CreatedAt,
			"updatedAt":             project.UpdatedAt,
		},
		"directories": projectDirectoryLayout,
	}
}

func inspectRequiredDirectories(root string) []ProjectDirectoryEntryCheck {
	result := make([]ProjectDirectoryEntryCheck, 0, len(projectDirectoryLayout))
	for _, relativePath := range projectDirectoryLayout {
		result = append(result, ProjectDirectoryEntryCheck{
			Path:   relativePath,
			Exists: directoryExists(filepath.Join(root, filepath.FromSlash(relativePath))),
		})
	}
	return result
}

func allRequiredDirectoriesExist(entries []ProjectDirectoryEntryCheck) bool {
	for _, entry := range entries {
		if !entry.Exists {
			return false
		}
	}
	return true
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func directoryReadable(path string) bool {
	_, err := os.ReadDir(path)
	return err == nil
}

func directoryWritable(path string) bool {
	file, err := os.CreateTemp(path, ".dreamworker-write-test-*")
	if err != nil {
		return false
	}
	name := file.Name()
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(name)
		return false
	}
	if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
