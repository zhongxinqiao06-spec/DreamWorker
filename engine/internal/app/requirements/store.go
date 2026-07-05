package requirements

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/agentruntime"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/projects"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

const (
	requirementsImportDir = "workspace/imports/requirements"
	requirementsTempDir   = "workspace/temp/requirements"
	requirementsOutputDir = "artifacts/product"
	maxRequirementBytes   = 30 * 1024 * 1024
	maxSourceChars        = 120000
	maxPreviewChars       = 24000
)

var supportedRequirementExts = map[string]string{
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".pdf":  "application/pdf",
}

func (s *Store) ImportRequirementFiles(input ImportRequirementFilesInput) (RequirementImportResult, *AppError) {
	project, root, appErr := s.projectAndReadyRoot(input.ProjectID)
	if appErr != nil {
		return RequirementImportResult{}, appErr
	}
	if len(input.FilePaths) == 0 {
		return RequirementImportResult{}, BadRequest("REQUIREMENT_FILE_NOT_SELECTED", "未选择需求文件。", "请选择 .docx 或 .pdf 文件。")
	}
	runID := s.nextPublicID("req")
	targetDir := filepath.Join(root, filepath.FromSlash(requirementsImportDir), runID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return RequirementImportResult{}, BadRequest("REQUIREMENT_IMPORT_FAILED", "需求文件导入目录创建失败。", "请确认项目目录可写。")
	}
	sources := make([]RequirementSource, 0, len(input.FilePaths))
	for _, inputPath := range input.FilePaths {
		source, err := s.copyRequirementFile(root, targetDir, inputPath)
		if err != nil {
			return RequirementImportResult{}, BadRequest("REQUIREMENT_IMPORT_FAILED", err.Error(), "请确认文件格式和权限后重试。")
		}
		sources = append(sources, source)
	}
	return RequirementImportResult{
		ProjectID: project.ProjectID,
		RunID:     runID,
		Sources:   sources,
		Message:   fmt.Sprintf("已导入 %d 个需求文件。", len(sources)),
	}, nil
}

func (s *Store) ListRequirementSources(projectID string) (RequirementSourcesResult, *AppError) {
	project, root, appErr := s.projectAndReadyRoot(projectID)
	if appErr != nil {
		return RequirementSourcesResult{}, appErr
	}
	sources, err := s.collectRequirementSources(project, root)
	if err != nil {
		return RequirementSourcesResult{}, BadRequest("REQUIREMENT_SOURCE_SCAN_FAILED", "需求来源扫描失败。", "请确认项目目录可读后重试。")
	}
	return RequirementSourcesResult{ProjectID: project.ProjectID, Sources: sources}, nil
}

func (s *Store) PreviewRequirementSource(ctx context.Context, input PreviewRequirementSourceInput) (RequirementSourcePreviewResult, *AppError) {
	if ctx == nil {
		ctx = context.Background()
	}
	project, root, appErr := s.projectAndReadyRoot(input.ProjectID)
	if appErr != nil {
		return RequirementSourcePreviewResult{}, appErr
	}
	sourceID := strings.TrimSpace(input.SourceID)
	if sourceID == "" {
		return RequirementSourcePreviewResult{}, BadRequest("REQUIREMENT_SOURCE_NOT_SELECTED", "未选择要预览的需求来源。", "请选择一个项目描述、探索产物或已导入文件。")
	}
	allSources, err := s.collectRequirementSources(project, root)
	if err != nil {
		return RequirementSourcePreviewResult{}, BadRequest("REQUIREMENT_SOURCE_SCAN_FAILED", "需求来源扫描失败。", "请确认项目目录可读后重试。")
	}
	selected := selectRequirementSources(allSources, []string{sourceID})
	if len(selected) == 0 {
		return RequirementSourcePreviewResult{}, BadRequest("REQUIREMENT_SOURCE_NOT_FOUND", "需求来源不存在或已移动。", "请刷新需求来源后重新选择。")
	}
	runID := filepath.ToSlash(filepath.Join("preview", selected[0].SourceID))
	prepared, _, appErr := s.prepareSources(ctx, project, root, runID, selected)
	if appErr != nil {
		return RequirementSourcePreviewResult{}, appErr
	}
	source := prepared[0].RequirementSource
	content := strings.TrimSpace(prepared[0].Content)
	charCount := len([]rune(content))
	preview := limitRunes(content, maxPreviewChars)
	parser := "direct_text"
	if source.Kind == "imported_file" {
		parser = "mineru"
	}
	return RequirementSourcePreviewResult{
		ProjectID: project.ProjectID,
		Source:    source,
		Parser:    parser,
		Content:   preview,
		CharCount: charCount,
		Truncated: charCount > len([]rune(preview)),
		TraceID:   s.TraceID(),
		CreatedAt: s.Now(),
	}, nil
}

func (s *Store) RunRequirementAnalysis(ctx context.Context, input RunRequirementAnalysisInput) (RequirementAnalysisRun, *AppError) {
	if ctx == nil {
		ctx = context.Background()
	}
	project, root, appErr := s.projectAndReadyRoot(input.ProjectID)
	if appErr != nil {
		return RequirementAnalysisRun{}, appErr
	}
	allSources, err := s.collectRequirementSources(project, root)
	if err != nil {
		return RequirementAnalysisRun{}, BadRequest("REQUIREMENT_SOURCE_SCAN_FAILED", "需求来源扫描失败。", "请确认项目目录可读后重试。")
	}
	selected := selectRequirementSources(allSources, input.SourceIDs)
	if len(selected) == 0 {
		return RequirementAnalysisRun{}, BadRequest("REQUIREMENT_SOURCE_EMPTY", "没有可分析的需求来源。", "请先导入需求文件，或选择项目描述/探索产物。")
	}
	runID := s.nextPublicID("req")
	prepared, warnings, appErr := s.prepareSources(ctx, project, root, runID, selected)
	if appErr != nil {
		return RequirementAnalysisRun{}, appErr
	}
	analysis, analysisWarnings, appErr := s.analyzeRequirements(ctx, project, prepared, input.Prompt)
	if appErr != nil {
		return RequirementAnalysisRun{}, appErr
	}
	warnings = append(warnings, analysisWarnings...)
	outputFiles, err := s.writeRequirementOutputs(root, analysis)
	if err != nil {
		return RequirementAnalysisRun{}, BadRequest("REQUIREMENT_OUTPUT_WRITE_FAILED", "需求分析产物写入失败。", "请确认项目目录可写后重试。")
	}
	traceID := s.TraceID()
	result := RequirementAnalysisRun{
		RunID:        runID,
		ProjectID:    project.ProjectID,
		Status:       "completed",
		Sources:      selected,
		FeatureCount: len(analysis.Features),
		OutputFiles:  outputFiles,
		Warnings:     warnings,
		TraceID:      traceID,
		CreatedAt:    s.Now(),
		Analysis:     analysis,
	}
	s.persistRequirementRun(project.ProjectID, result)
	return result, nil
}

func (s *Store) projectAndReadyRoot(projectID string) (resources.Project, string, *AppError) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return resources.Project{}, "", BadRequest("BAD_REQUEST", "缺少 projectId。", "请选择项目后重试。")
	}
	check := s.inspectProjectDirectory(projectID)
	if check.Status != "valid" || check.LocalRootPath == nil {
		return resources.Project{}, "", BadRequest("PROJECT_DIRECTORY_NOT_READY", "项目本地目录尚未初始化。", "请先在项目页选择并初始化本地目录。")
	}
	s.Mu.Lock()
	project, ok := s.Projects[projectID]
	s.Mu.Unlock()
	if !ok {
		return resources.Project{}, "", NotFound("PROJECT_NOT_FOUND", "项目不存在。", "请刷新项目列表。")
	}
	return project, filepath.Clean(*check.LocalRootPath), nil
}

func (s *Store) inspectProjectDirectory(projectID string) resources.ProjectDirectoryCheck {
	check, appErr := projects.NewStore(s.Store).ValidateLocalDirectory(projectID)
	if appErr != nil {
		return resources.ProjectDirectoryCheck{ProjectID: projectID, Status: "invalid", Message: appErr.Message}
	}
	return check
}

func (s *Store) nextPublicID(prefix string) string {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.NextIDLocked(prefix)
}

func (s *Store) copyRequirementFile(root string, targetDir string, inputPath string) (RequirementSource, error) {
	inputPath = strings.TrimSpace(inputPath)
	if inputPath == "" {
		return RequirementSource{}, errors.New("需求文件路径为空。")
	}
	absolute, err := filepath.Abs(inputPath)
	if err != nil {
		return RequirementSource{}, fmt.Errorf("需求文件路径无效：%w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return RequirementSource{}, errors.New("需求文件不存在或无法读取。")
	}
	if info.IsDir() {
		return RequirementSource{}, errors.New("请选择文件，不能选择目录。")
	}
	if info.Size() > maxRequirementBytes {
		return RequirementSource{}, fmt.Errorf("需求文件过大，单个文件不能超过 %d MB。", maxRequirementBytes/1024/1024)
	}
	ext := strings.ToLower(filepath.Ext(absolute))
	mimeType, ok := supportedRequirementExts[ext]
	if !ok {
		return RequirementSource{}, errors.New("仅支持 .docx 和 .pdf 需求文件，旧 .doc 请先转成 .docx。")
	}
	fileName := safeFileName(filepath.Base(absolute))
	targetPath := uniquePath(filepath.Join(targetDir, fileName))
	if err := copyFile(absolute, targetPath); err != nil {
		return RequirementSource{}, errors.New("需求文件复制失败。")
	}
	relative, err := relativeProjectPath(root, targetPath)
	if err != nil {
		return RequirementSource{}, err
	}
	return RequirementSource{
		SourceID:     sourceIDForPath(relative),
		Kind:         "imported_file",
		FileName:     filepath.Base(targetPath),
		RelativePath: filepath.ToSlash(relative),
		AbsolutePath: targetPath,
		MimeType:     mimeType,
		CharCount:    0,
		ImportedAt:   s.Now(),
		Summary:      "用户导入的需求文件，运行分析时通过 MinerU 解析。",
	}, nil
}

func (s *Store) collectRequirementSources(project resources.Project, root string) ([]RequirementSource, error) {
	sources := []RequirementSource{}
	if strings.TrimSpace(project.Description) != "" {
		sources = append(sources, RequirementSource{
			SourceID:   "project_description",
			Kind:       "project_description",
			FileName:   "project_description.txt",
			MimeType:   "text/plain",
			CharCount:  len([]rune(project.Description)),
			ImportedAt: project.UpdatedAt,
			Summary:    trimSummary(project.Description),
		})
	}
	importRoot := filepath.Join(root, filepath.FromSlash(requirementsImportDir))
	if directoryExists(importRoot) {
		if err := filepath.WalkDir(importRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			ext := strings.ToLower(filepath.Ext(path))
			mimeType, ok := supportedRequirementExts[ext]
			if !ok {
				return nil
			}
			relative, relErr := relativeProjectPath(root, path)
			if relErr != nil {
				return relErr
			}
			info, statErr := entry.Info()
			importedAt := s.Now()
			if statErr == nil {
				importedAt = info.ModTime().Format(time.RFC3339)
			}
			sources = append(sources, RequirementSource{
				SourceID:     sourceIDForPath(relative),
				Kind:         "imported_file",
				FileName:     entry.Name(),
				RelativePath: filepath.ToSlash(relative),
				AbsolutePath: path,
				MimeType:     mimeType,
				ImportedAt:   importedAt,
				Summary:      "用户导入的需求文件，运行分析时通过 MinerU 解析。",
			})
			return nil
		}); err != nil {
			return nil, err
		}
	}
	for _, relativePath := range []string{
		"docs/dream_brief.md",
		"docs/research_pack.md",
		"artifacts/explore/dream_brief.md",
		"artifacts/explore/hypotheses.yaml",
		"artifacts/explore/research_pack.md",
		"artifacts/explore/competitor_map.md",
		"artifacts/explore/evidence_graph.yaml",
	} {
		path := filepath.Join(root, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(path)
		if err != nil || strings.TrimSpace(string(content)) == "" {
			continue
		}
		sources = append(sources, RequirementSource{
			SourceID:     sourceIDForPath(relativePath),
			Kind:         "explore_artifact",
			FileName:     filepath.Base(path),
			RelativePath: filepath.ToSlash(relativePath),
			AbsolutePath: path,
			MimeType:     mimeTypeForTextPath(path),
			CharCount:    len([]rune(string(content))),
			ImportedAt:   fileModTime(path, s.Now()),
			Summary:      trimSummary(string(content)),
		})
	}
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].Kind != sources[j].Kind {
			return sources[i].Kind < sources[j].Kind
		}
		return sources[i].RelativePath < sources[j].RelativePath
	})
	return sources, nil
}

func selectRequirementSources(all []RequirementSource, sourceIDs []string) []RequirementSource {
	if len(sourceIDs) == 0 {
		return all
	}
	wanted := map[string]bool{}
	for _, sourceID := range sourceIDs {
		wanted[strings.TrimSpace(sourceID)] = true
	}
	selected := []RequirementSource{}
	for _, source := range all {
		if wanted[source.SourceID] {
			selected = append(selected, source)
		}
	}
	return selected
}

func (s *Store) prepareSources(ctx context.Context, project resources.Project, root string, runID string, selected []RequirementSource) ([]preparedRequirementSource, []string, *AppError) {
	prepared := make([]preparedRequirementSource, 0, len(selected))
	warnings := []string{}
	for _, source := range selected {
		switch source.Kind {
		case "imported_file":
			if s.DocumentParser == nil {
				return nil, nil, BadRequest("MINERU_UNAVAILABLE", "MinerU 解析能力不可用，无法解析需求文件。", "请确认 Engine 已启用 MinerU 文档解析器后重试。")
			}
			outputDir := filepath.Join(root, filepath.FromSlash(requirementsTempDir), runID, source.SourceID)
			content, err := s.DocumentParser.ParseDocument(ctx, source.AbsolutePath, outputDir)
			if err != nil {
				return nil, nil, BadRequest("MINERU_PARSE_FAILED", "MinerU 解析需求文件失败。", "请确认网络、MinerU Open API 或本地 CLI 可用，也可以更换文件后重试。")
			}
			content = limitRunes(strings.TrimSpace(content), maxSourceChars)
			if content == "" {
				return nil, nil, BadRequest("MINERU_PARSE_EMPTY", "MinerU 没有解析出可分析文本。", "请确认文件不是空白、加密或受损文档。")
			}
			source.CharCount = len([]rune(content))
			source.Summary = trimSummary(content)
			prepared = append(prepared, preparedRequirementSource{RequirementSource: source, Content: content})
		case "project_description":
			content := strings.TrimSpace(project.Description)
			source.CharCount = len([]rune(content))
			source.Summary = trimSummary(content)
			prepared = append(prepared, preparedRequirementSource{RequirementSource: source, Content: content})
		default:
			content, err := os.ReadFile(source.AbsolutePath)
			if err != nil {
				warnings = append(warnings, source.FileName+" 读取失败，已跳过。")
				continue
			}
			text := limitRunes(strings.TrimSpace(string(content)), maxSourceChars)
			if text == "" {
				warnings = append(warnings, source.FileName+" 内容为空，已跳过。")
				continue
			}
			source.CharCount = len([]rune(text))
			source.Summary = trimSummary(text)
			prepared = append(prepared, preparedRequirementSource{RequirementSource: source, Content: text})
		}
	}
	if len(prepared) == 0 {
		return nil, nil, BadRequest("REQUIREMENT_SOURCE_EMPTY", "没有可分析的需求文本。", "请导入有效文件或选择非空探索产物。")
	}
	return prepared, warnings, nil
}

func (s *Store) analyzeRequirements(ctx context.Context, project resources.Project, sources []preparedRequirementSource, prompt string) (RequirementAnalysisResult, []string, *AppError) {
	profile, provider, appErr := s.requirementModelBinding(project)
	if appErr != nil {
		return RequirementAnalysisResult{}, nil, appErr
	}
	if profile.Model == "model_generate_stub" || provider.ProviderID == "provider_local_stub" {
		return synthesizeRequirementAnalysis(project, sources, prompt), nil, nil
	}
	requestPrompt := buildRequirementPrompt(project, sources, prompt)
	messages := []resources.ChatGatewayMessage{
		{Role: "system", Content: "你是严谨的软件需求分析师。只输出合法 JSON，不要输出 Markdown。"},
		{Role: "user", Content: requestPrompt},
	}
	var builder strings.Builder
	streamErrCode := ""
	for chunk := range s.ModelGateway.StreamChat(ctx, resources.ToChatModelProvider(provider), resources.ToChatModelProfile(profile), messages) {
		if chunk.Error != nil {
			streamErrCode = chunk.Error.Code
			break
		}
		builder.WriteString(chunk.Delta)
	}
	if streamErrCode != "" {
		if ctx.Err() != nil {
			return RequirementAnalysisResult{}, nil, BadRequest(streamErrCode, "需求分析模型调用失败。", "请检查模型服务商配置后重试。")
		}
		analysis := synthesizeRequirementAnalysis(project, sources, prompt)
		analysis.Risks = append(analysis.Risks, "模型调用失败，本次已使用本地确定性需求分析兜底。")
		return analysis, []string{"模型调用失败，已使用本地确定性需求分析兜底。"}, nil
	}
	analysis, err := parseRequirementAnalysisJSON(builder.String())
	if err != nil {
		analysis := synthesizeRequirementAnalysis(project, sources, prompt)
		analysis.Risks = append(analysis.Risks, "模型未返回合法 JSON，本次已使用本地确定性需求分析兜底。")
		return analysis, []string{"模型未返回合法 JSON，已使用本地确定性需求分析兜底。"}, nil
	}
	analysis = normalizeRequirementAnalysis(project, sources, analysis)
	return analysis, nil, nil
}

func (s *Store) requirementModelBinding(project resources.Project) (resources.ModelProfile, resources.ModelProviderRecord, *AppError) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	agent, ok := s.Agents["agent_product_designer"]
	if !ok {
		agent = resources.AgentConfig{AgentID: "agent_product_designer", ModelProfileID: project.DefaultModelProfileID}
	}
	session := resources.ChatSession{
		SessionID:      "requirements_runtime",
		ProjectID:      &project.ProjectID,
		AgentID:        agent.AgentID,
		ModelProfileID: project.DefaultModelProfileID,
	}
	profile, provider, _, appErr := agentruntime.ResolveChatModelBinding(s.Store, session, agent)
	return profile, provider, appErr
}

func (s *Store) writeRequirementOutputs(root string, analysis RequirementAnalysisResult) ([]RequirementOutputFile, error) {
	outputDir := filepath.Join(root, filepath.FromSlash(requirementsOutputDir))
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}
	jsonPath := filepath.Join(outputDir, "requirements_analysis.json")
	payload, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(jsonPath, payload, 0o644); err != nil {
		return nil, err
	}
	xlsxPath := filepath.Join(outputDir, "feature_list.xlsx")
	if err := writeFeatureListXLSX(xlsxPath, analysis); err != nil {
		return nil, err
	}
	docxPath := filepath.Join(outputDir, "requirements_spec.docx")
	if err := writeRequirementSpecDOCX(docxPath, analysis); err != nil {
		return nil, err
	}
	paths := []struct {
		kind string
		path string
	}{
		{"analysis_json", jsonPath},
		{"feature_excel", xlsxPath},
		{"requirements_word", docxPath},
	}
	files := make([]RequirementOutputFile, 0, len(paths))
	for _, item := range paths {
		relative, err := relativeProjectPath(root, item.path)
		if err != nil {
			return nil, err
		}
		files = append(files, RequirementOutputFile{
			Kind:         item.kind,
			FileName:     filepath.Base(item.path),
			RelativePath: filepath.ToSlash(relative),
			AbsolutePath: item.path,
		})
	}
	return files, nil
}

func (s *Store) persistRequirementRun(projectID string, result RequirementAnalysisRun) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	projectModules, ok := s.Modules[projectID]
	if !ok {
		return
	}
	module, ok := projectModules["product"]
	if !ok {
		return
	}
	config := cloneAnyMap(module.Config)
	config["requirementAnalysisRunId"] = result.RunID
	config["requirementAnalysisStatus"] = result.Status
	config["requirementAnalysisFeatureCount"] = result.FeatureCount
	config["requirementAnalysisUpdatedAt"] = result.CreatedAt
	if len(result.OutputFiles) > 0 {
		config["requirementAnalysisOutputDir"] = requirementsOutputDir
	}
	module.Config = config
	module.Status = "completed"
	module.OutputArtifacts = []string{"feature_list.xlsx", "requirements_spec.docx", "requirements_analysis.json"}
	projectModules["product"] = module
	_ = s.PersistWorkspaceSnapshotLocked()
}

func buildRequirementPrompt(project resources.Project, sources []preparedRequirementSource, extraPrompt string) string {
	var builder strings.Builder
	builder.WriteString("请根据以下项目背景和需求来源，抽取功能清单并输出 JSON。\n")
	builder.WriteString("JSON 字段必须为：projectTitle, summary, sources, roles, features, nonFunctionalRequirements, risks, openQuestions。\n")
	builder.WriteString("features 每项字段必须为：featureId, module, name, role, scenario, description, priority, type, inputs, outputs, acceptanceCriteria, dependencies, sourceRefs, notes。\n")
	builder.WriteString("priority 只能用 P0/P1/P2，type 用 functional/data/integration/security/operation 中之一。\n")
	builder.WriteString("项目名称：" + project.Title + "\n项目描述：" + project.Description + "\n")
	if strings.TrimSpace(extraPrompt) != "" {
		builder.WriteString("用户补充要求：" + strings.TrimSpace(extraPrompt) + "\n")
	}
	for _, source := range sources {
		builder.WriteString("\n--- SOURCE " + source.SourceID + " / " + source.FileName + " ---\n")
		builder.WriteString(source.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func parseRequirementAnalysisJSON(raw string) (RequirementAnalysisResult, error) {
	cleaned := strings.TrimSpace(raw)
	if strings.Contains(cleaned, "```") {
		cleaned = stripCodeFence(cleaned)
	}
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start >= 0 && end > start {
		cleaned = cleaned[start : end+1]
	}
	var result RequirementAnalysisResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return RequirementAnalysisResult{}, err
	}
	if len(result.Features) == 0 {
		return RequirementAnalysisResult{}, errors.New("empty features")
	}
	return result, nil
}

func normalizeRequirementAnalysis(project resources.Project, sources []preparedRequirementSource, analysis RequirementAnalysisResult) RequirementAnalysisResult {
	if strings.TrimSpace(analysis.ProjectTitle) == "" {
		analysis.ProjectTitle = project.Title
	}
	if strings.TrimSpace(analysis.Summary) == "" {
		analysis.Summary = "根据探索结果和需求文件生成的需求分析。"
	}
	if len(analysis.Sources) == 0 {
		for _, source := range sources {
			analysis.Sources = append(analysis.Sources, source.FileName)
		}
	}
	for index := range analysis.Features {
		feature := &analysis.Features[index]
		if strings.TrimSpace(feature.FeatureID) == "" {
			feature.FeatureID = fmt.Sprintf("FR-%03d", index+1)
		}
		if strings.TrimSpace(feature.Priority) == "" {
			feature.Priority = "P1"
		}
		if strings.TrimSpace(feature.Type) == "" {
			feature.Type = "functional"
		}
		if len(feature.AcceptanceCriteria) == 0 {
			feature.AcceptanceCriteria = []string{"用户可以完成该功能对应的核心任务。"}
		}
		if len(feature.SourceRefs) == 0 && len(sources) > 0 {
			feature.SourceRefs = []string{sources[0].SourceID}
		}
	}
	if len(analysis.NonFunctionalRequirements) == 0 {
		analysis.NonFunctionalRequirements = []string{"响应速度、权限边界、数据安全和可审计性需在详细设计中确认。"}
	}
	if len(analysis.OpenQuestions) == 0 {
		analysis.OpenQuestions = []string{"哪些功能必须进入首版，哪些可以进入后续迭代？"}
	}
	return analysis
}

func synthesizeRequirementAnalysis(project resources.Project, sources []preparedRequirementSource, extraPrompt string) RequirementAnalysisResult {
	features := extractStructuredFeatureItems(sources)
	if len(features) > 0 {
		summary := "根据项目描述、探索结果和需求文件整理出首批功能需求。"
		if strings.TrimSpace(extraPrompt) != "" {
			summary += " 已考虑用户补充要求。"
		}
		sourceNames := make([]string, 0, len(sources))
		for _, source := range sources {
			sourceNames = append(sourceNames, source.FileName)
		}
		return normalizeRequirementAnalysis(project, sources, RequirementAnalysisResult{
			ProjectTitle:              project.Title,
			Summary:                   summary,
			Sources:                   sourceNames,
			Roles:                     []string{"项目负责人", "产品经理", "项目成员"},
			Features:                  features,
			NonFunctionalRequirements: extractStructuredNFRs(sources),
			Risks:                     []string{"需求文件质量会影响分析结果。", "扫描件解析依赖 MinerU OCR 环境。"},
			OpenQuestions:             []string{"哪些 P1 功能应进入首版？", "是否存在必须遵守的行业规范或审批流程？"},
		})
	}
	features = []RequirementFeatureItem{}
	candidates := extractFeatureCandidates(sources)
	if len(candidates) == 0 {
		candidates = []string{"项目初始化与配置", "需求文件导入", "需求结构化分析", "功能清单导出", "需求规格说明生成"}
	}
	for index, name := range candidates {
		if index >= 8 {
			break
		}
		priority := "P1"
		if index < 2 {
			priority = "P0"
		} else if index > 5 {
			priority = "P2"
		}
		sourceRef := "project_description"
		if len(sources) > 0 {
			sourceRef = sources[min(index, len(sources)-1)].SourceID
		}
		features = append(features, RequirementFeatureItem{
			FeatureID:          fmt.Sprintf("FR-%03d", index+1),
			Module:             inferModule(name),
			Name:               name,
			Role:               "项目成员",
			Scenario:           "在项目推进过程中需要完成 " + name + "。",
			Description:        "系统应支持" + name + "，并保留过程状态和产物结果。",
			Priority:           priority,
			Type:               "functional",
			Inputs:             []string{"项目上下文", "需求来源"},
			Outputs:            []string{"结构化结果", "可追踪产物"},
			AcceptanceCriteria: []string{"用户可以在项目空间中完成该功能。", "功能结果可被后续模块读取。"},
			Dependencies:       []string{"项目本地目录已初始化"},
			SourceRefs:         []string{sourceRef},
			Notes:              "本地 Stub 模式生成，用于离线演示和测试。",
		})
	}
	summary := "根据项目描述、探索结果和需求文件整理出首批功能需求。"
	if strings.TrimSpace(extraPrompt) != "" {
		summary += " 已考虑用户补充要求。"
	}
	sourceNames := make([]string, 0, len(sources))
	for _, source := range sources {
		sourceNames = append(sourceNames, source.FileName)
	}
	return normalizeRequirementAnalysis(project, sources, RequirementAnalysisResult{
		ProjectTitle:              project.Title,
		Summary:                   summary,
		Sources:                   sourceNames,
		Roles:                     []string{"项目负责人", "产品经理", "项目成员"},
		Features:                  features,
		NonFunctionalRequirements: []string{"需求来源需可追踪。", "导出的 Excel 和 Word 文档应能在 Office/WPS 中打开。", "导入文件只允许写入项目目录。"},
		Risks:                     []string{"需求文件质量会影响分析准确性。", "扫描件解析依赖 MinerU OCR 环境。"},
		OpenQuestions:             []string{"哪些 P1 功能应进入首版？", "是否存在必须遵守的行业规范或审批流程？"},
	})
}

var featureIDPattern = regexp.MustCompile(`(?i)^FR-\d+`)
var nfrIDPattern = regexp.MustCompile(`(?i)^NFR-\d+`)
var htmlTableRowPattern = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
var htmlTableCellPattern = regexp.MustCompile(`(?is)<t[dh][^>]*>(.*?)</t[dh]>`)
var htmlTagPattern = regexp.MustCompile(`(?is)<[^>]+>`)

func extractStructuredFeatureItems(sources []preparedRequirementSource) []RequirementFeatureItem {
	features := []RequirementFeatureItem{}
	seen := map[string]bool{}
	for _, source := range sources {
		for _, row := range extractTableRows(source.Content) {
			index := indexMatching(row, featureIDPattern)
			if index < 0 || len(row) < index+5 {
				continue
			}
			featureID := strings.ToUpper(strings.TrimSpace(row[index]))
			if seen[featureID] {
				continue
			}
			seen[featureID] = true
			module := valueAt(row, index+1)
			name := valueAt(row, index+2)
			priority := valueAt(row, index+3)
			description := valueAt(row, index+4)
			if name == "" && description == "" {
				continue
			}
			if module == "" {
				module = inferModule(name)
			}
			if priority == "" {
				priority = "P1"
			}
			if description == "" {
				description = "系统应支持" + name + "。"
			}
			features = append(features, RequirementFeatureItem{
				FeatureID:          featureID,
				Module:             module,
				Name:               name,
				Role:               "项目成员",
				Scenario:           description,
				Description:        description,
				Priority:           priority,
				Type:               "functional",
				Inputs:             []string{"项目上下文", "需求来源"},
				Outputs:            []string{"结构化结果", "可追踪产物"},
				AcceptanceCriteria: []string{description},
				Dependencies:       []string{"项目本地目录已初始化"},
				SourceRefs:         []string{source.SourceID},
				Notes:              "从需求文件结构化表格抽取。",
			})
			if len(features) >= 32 {
				return features
			}
		}
	}
	return features
}

func extractStructuredNFRs(sources []preparedRequirementSource) []string {
	items := []string{}
	seen := map[string]bool{}
	for _, source := range sources {
		for _, row := range extractTableRows(source.Content) {
			index := indexMatching(row, nfrIDPattern)
			if index < 0 || len(row) < index+3 {
				continue
			}
			item := strings.TrimSpace(valueAt(row, index+1) + "：" + valueAt(row, index+2))
			if item == "：" || seen[item] {
				continue
			}
			seen[item] = true
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return []string{"需求来源需可追踪。", "导出的 Excel 和 Word 文档应能在 Office/WPS 中打开。", "导入文件只允许写入项目目录。"}
	}
	return items
}

func extractTableRows(content string) [][]string {
	rows := [][]string{}
	for _, match := range htmlTableRowPattern.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		cells := []string{}
		for _, cellMatch := range htmlTableCellPattern.FindAllStringSubmatch(match[1], -1) {
			if len(cellMatch) < 2 {
				continue
			}
			text := html.UnescapeString(htmlTagPattern.ReplaceAllString(cellMatch[1], " "))
			text = strings.Join(strings.Fields(text), " ")
			if text != "" {
				cells = append(cells, text)
			}
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		cells := []string{}
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" && !strings.Contains(part, "---") {
				cells = append(cells, part)
			}
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	return rows
}

func indexMatching(values []string, pattern *regexp.Regexp) int {
	for index, value := range values {
		if pattern.MatchString(strings.TrimSpace(value)) {
			return index
		}
	}
	return -1
}

func valueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return strings.TrimSpace(values[index])
}

func extractFeatureCandidates(sources []preparedRequirementSource) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, source := range sources {
		for _, line := range strings.Split(source.Content, "\n") {
			line = strings.TrimSpace(strings.Trim(line, "#-* 　\t"))
			if line == "" || len([]rune(line)) > 42 {
				continue
			}
			if strings.Contains(line, "功能") || strings.Contains(line, "需求") || strings.Contains(line, "管理") || strings.Contains(line, "生成") || strings.Contains(line, "导入") || strings.Contains(line, "分析") {
				if !seen[line] {
					seen[line] = true
					result = append(result, line)
				}
			}
		}
	}
	return result
}

func inferModule(name string) string {
	switch {
	case strings.Contains(name, "导入") || strings.Contains(name, "文件"):
		return "文件导入"
	case strings.Contains(name, "分析") || strings.Contains(name, "需求"):
		return "需求分析"
	case strings.Contains(name, "生成") || strings.Contains(name, "导出"):
		return "产物生成"
	default:
		return "核心功能"
	}
}

func sourceIDForPath(path string) string {
	sum := sha1.Sum([]byte(filepath.ToSlash(path)))
	return "src_" + hex.EncodeToString(sum[:])[:12]
}

func safeFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	if name == "" || name == "." {
		return "requirement.docx"
	}
	return name
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s_%d%s", base, index, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func copyFile(source string, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func relativeProjectPath(root string, target string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if relative == "." || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", errors.New("path escapes project directory")
	}
	return relative, nil
}

func trimSummary(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return limitRunes(value, 120)
}

func limitRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func fileModTime(path string, fallback string) string {
	info, err := os.Stat(path)
	if err != nil {
		return fallback
	}
	return info.ModTime().Format(time.RFC3339)
}

func mimeTypeForTextPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".md":
		return "text/markdown"
	default:
		return http.DetectContentType([]byte(filepath.Base(path)))
	}
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func stripCodeFence(value string) string {
	lines := strings.Split(value, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func cloneAnyMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
