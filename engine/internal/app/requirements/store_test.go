package requirements

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/projects"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type fakeRequirementParser struct {
	content   string
	inputPath string
	outputDir string
	calls     int
}

func (p *fakeRequirementParser) ParseDocument(_ context.Context, inputPath string, outputDir string) (string, error) {
	p.calls++
	p.inputPath = inputPath
	p.outputDir = outputDir
	return p.content, nil
}

func TestPreviewRequirementSourceParsesImportedFileWithDocumentParser(t *testing.T) {
	parser := &fakeRequirementParser{content: "FR-001 编码智能体应支持任务理解、代码修改和测试验收。"}
	state := resources.NewStore(
		resources.WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		resources.WithTraceID(func() string { return "tr_preview" }),
		resources.WithDocumentParser(parser),
	)
	root := t.TempDir()
	projectStore := projects.NewStore(state)
	project, appErr := projectStore.CreateProject(resources.CreateProjectInput{
		Title:         "编码智能体",
		Description:   "面向工程团队的编码智能体。",
		LocalRootPath: &root,
	})
	if appErr != nil {
		t.Fatalf("create project: %v", appErr.Message)
	}
	if _, appErr := projectStore.InitializeLocalDirectory(project.ProjectID); appErr != nil {
		t.Fatalf("initialize project directory: %v", appErr.Message)
	}
	uploadDir := t.TempDir()
	uploadPath := filepath.Join(uploadDir, "coding-agent-requirements.docx")
	if err := os.WriteFile(uploadPath, []byte("placeholder docx bytes"), 0o644); err != nil {
		t.Fatalf("write upload: %v", err)
	}

	store := NewStore(state)
	importResult, appErr := store.ImportRequirementFiles(ImportRequirementFilesInput{
		ProjectID: project.ProjectID,
		FilePaths: []string{uploadPath},
	})
	if appErr != nil {
		t.Fatalf("import requirement file: %v", appErr.Message)
	}
	preview, appErr := store.PreviewRequirementSource(context.Background(), PreviewRequirementSourceInput{
		ProjectID: project.ProjectID,
		SourceID:  importResult.Sources[0].SourceID,
	})
	if appErr != nil {
		t.Fatalf("preview requirement source: %v", appErr.Message)
	}

	if preview.Parser != "mineru" {
		t.Fatalf("expected mineru parser, got %s", preview.Parser)
	}
	if !strings.Contains(preview.Content, "编码智能体") {
		t.Fatalf("preview content did not include parsed text: %q", preview.Content)
	}
	if preview.TraceID != "tr_preview" || preview.CreatedAt != "2026-07-01T00:00:00Z" {
		t.Fatalf("unexpected preview metadata: %+v", preview)
	}
	if parser.calls != 1 {
		t.Fatalf("expected parser to be called once, got %d", parser.calls)
	}
	if !strings.Contains(filepath.ToSlash(parser.outputDir), "/workspace/temp/requirements/preview/") {
		t.Fatalf("preview output dir should stay under project temp directory, got %s", parser.outputDir)
	}
}

func TestPreviewRequirementSourceUsesDirectTextForProjectDescription(t *testing.T) {
	parser := &fakeRequirementParser{content: "should not be called"}
	state := resources.NewStore(
		resources.WithClock(func() string { return "2026-07-01T00:00:00Z" }),
		resources.WithDocumentParser(parser),
	)
	root := t.TempDir()
	projectStore := projects.NewStore(state)
	project, appErr := projectStore.CreateProject(resources.CreateProjectInput{
		Title:         "编码智能体",
		Description:   "项目描述中的需求来源。",
		LocalRootPath: &root,
	})
	if appErr != nil {
		t.Fatalf("create project: %v", appErr.Message)
	}
	if _, appErr := projectStore.InitializeLocalDirectory(project.ProjectID); appErr != nil {
		t.Fatalf("initialize project directory: %v", appErr.Message)
	}

	preview, appErr := NewStore(state).PreviewRequirementSource(context.Background(), PreviewRequirementSourceInput{
		ProjectID: project.ProjectID,
		SourceID:  "project_description",
	})
	if appErr != nil {
		t.Fatalf("preview project description: %v", appErr.Message)
	}
	if preview.Parser != "direct_text" {
		t.Fatalf("expected direct_text parser, got %s", preview.Parser)
	}
	if preview.Content != "项目描述中的需求来源。" {
		t.Fatalf("unexpected preview content: %q", preview.Content)
	}
	if parser.calls != 0 {
		t.Fatalf("parser should not be called for project description")
	}
}
