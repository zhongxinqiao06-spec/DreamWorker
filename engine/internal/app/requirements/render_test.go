package requirements

import (
	"archive/zip"
	"path/filepath"
	"testing"
)

func TestWriteRequirementArtifactsCreateOfficePackages(t *testing.T) {
	analysis := RequirementAnalysisResult{
		ProjectTitle: "测试项目",
		Summary:      "根据需求来源生成结构化功能清单。",
		Sources:      []string{"project_description.txt"},
		Roles:        []string{"项目成员"},
		Features: []RequirementFeatureItem{
			{
				FeatureID:          "FR-001",
				Module:             "需求分析",
				Name:               "导入需求文件",
				Role:               "项目成员",
				Scenario:           "上传 Word 或 PDF 项目要求文件。",
				Description:        "系统应解析需求文件并抽取功能项。",
				Priority:           "P0",
				Type:               "functional",
				Inputs:             []string{"需求文件"},
				Outputs:            []string{"功能清单"},
				AcceptanceCriteria: []string{"Excel 和 Word 均可打开。"},
				Dependencies:       []string{"项目目录已初始化"},
				SourceRefs:         []string{"project_description"},
			},
		},
		NonFunctionalRequirements: []string{"产物可追溯到来源。"},
		Risks:                     []string{"需求文件质量会影响分析结果。"},
		OpenQuestions:             []string{"首版范围如何排序？"},
	}
	dir := t.TempDir()
	xlsxPath := filepath.Join(dir, "feature_list.xlsx")
	docxPath := filepath.Join(dir, "requirements_spec.docx")
	if err := writeFeatureListXLSX(xlsxPath, analysis); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	if err := writeRequirementSpecDOCX(docxPath, analysis); err != nil {
		t.Fatalf("write docx: %v", err)
	}
	assertZipEntries(t, xlsxPath, []string{
		"[Content_Types].xml",
		"xl/workbook.xml",
		"xl/worksheets/sheet1.xml",
	})
	assertZipEntries(t, docxPath, []string{
		"[Content_Types].xml",
		"word/document.xml",
		"word/styles.xml",
	})
}

func assertZipEntries(t *testing.T, path string, names []string) {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open zip %s: %v", filepath.Base(path), err)
	}
	defer reader.Close()
	found := map[string]bool{}
	for _, file := range reader.File {
		found[file.Name] = true
	}
	for _, name := range names {
		if !found[name] {
			t.Fatalf("%s missing zip entry %s", filepath.Base(path), name)
		}
	}
}
