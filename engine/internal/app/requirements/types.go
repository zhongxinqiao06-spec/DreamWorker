package requirements

import "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"

type Store struct {
	*resources.Store
}

func NewStore(state *resources.Store) *Store {
	return &Store{Store: state}
}

type AppError = resources.AppError
type ModelProfile = resources.ModelProfile
type ModelProviderRecord = resources.ModelProviderRecord

var BadRequest = resources.BadRequest
var NotFound = resources.NotFound

type ImportRequirementFilesInput struct {
	ProjectID string   `json:"projectId"`
	FilePaths []string `json:"filePaths"`
}

type RequirementImportResult struct {
	ProjectID string              `json:"projectId"`
	RunID     string              `json:"runId"`
	Sources   []RequirementSource `json:"sources"`
	Message   string              `json:"message"`
}

type RequirementSourcesResult struct {
	ProjectID string              `json:"projectId"`
	Sources   []RequirementSource `json:"sources"`
}

type PreviewRequirementSourceInput struct {
	ProjectID string `json:"projectId"`
	SourceID  string `json:"sourceId"`
}

type RequirementSourcePreviewResult struct {
	ProjectID string            `json:"projectId"`
	Source    RequirementSource `json:"source"`
	Parser    string            `json:"parser"`
	Content   string            `json:"content"`
	CharCount int               `json:"charCount"`
	Truncated bool              `json:"truncated"`
	TraceID   string            `json:"traceId"`
	CreatedAt string            `json:"createdAt"`
}

type RunRequirementAnalysisInput struct {
	ProjectID string   `json:"projectId"`
	SourceIDs []string `json:"sourceIds"`
	Prompt    string   `json:"prompt"`
}

type RequirementAnalysisRun struct {
	RunID        string                    `json:"runId"`
	ProjectID    string                    `json:"projectId"`
	Status       string                    `json:"status"`
	Sources      []RequirementSource       `json:"sources"`
	FeatureCount int                       `json:"featureCount"`
	OutputFiles  []RequirementOutputFile   `json:"outputFiles"`
	Warnings     []string                  `json:"warnings"`
	TraceID      string                    `json:"traceId"`
	CreatedAt    string                    `json:"createdAt"`
	Analysis     RequirementAnalysisResult `json:"analysis"`
}

type RequirementSource struct {
	SourceID     string `json:"sourceId"`
	Kind         string `json:"kind"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	AbsolutePath string `json:"absolutePath,omitempty"`
	MimeType     string `json:"mimeType"`
	CharCount    int    `json:"charCount"`
	ImportedAt   string `json:"importedAt"`
	Summary      string `json:"summary"`
}

type RequirementOutputFile struct {
	Kind         string `json:"kind"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	AbsolutePath string `json:"absolutePath"`
}

type RequirementAnalysisResult struct {
	ProjectTitle              string                   `json:"projectTitle"`
	Summary                   string                   `json:"summary"`
	Sources                   []string                 `json:"sources"`
	Roles                     []string                 `json:"roles"`
	Features                  []RequirementFeatureItem `json:"features"`
	NonFunctionalRequirements []string                 `json:"nonFunctionalRequirements"`
	Risks                     []string                 `json:"risks"`
	OpenQuestions             []string                 `json:"openQuestions"`
}

type RequirementFeatureItem struct {
	FeatureID          string   `json:"featureId"`
	Module             string   `json:"module"`
	Name               string   `json:"name"`
	Role               string   `json:"role"`
	Scenario           string   `json:"scenario"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	Type               string   `json:"type"`
	Inputs             []string `json:"inputs"`
	Outputs            []string `json:"outputs"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Dependencies       []string `json:"dependencies"`
	SourceRefs         []string `json:"sourceRefs"`
	Notes              string   `json:"notes"`
}

type preparedRequirementSource struct {
	RequirementSource
	Content string
}
