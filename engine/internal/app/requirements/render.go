package requirements

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func writeFeatureListXLSX(path string, analysis RequirementAnalysisResult) error {
	rows := [][]string{{
		"功能ID", "模块", "功能名称", "用户角色", "场景", "描述", "优先级", "类型", "输入", "输出", "验收标准", "依赖/约束", "来源", "备注",
	}}
	for _, feature := range analysis.Features {
		rows = append(rows, []string{
			feature.FeatureID,
			feature.Module,
			feature.Name,
			feature.Role,
			feature.Scenario,
			feature.Description,
			feature.Priority,
			feature.Type,
			strings.Join(feature.Inputs, "\n"),
			strings.Join(feature.Outputs, "\n"),
			strings.Join(feature.AcceptanceCriteria, "\n"),
			strings.Join(feature.Dependencies, "\n"),
			strings.Join(feature.SourceRefs, "\n"),
			feature.Notes,
		})
	}
	files := map[string]string{
		"[Content_Types].xml":        xlsxContentTypes(),
		"_rels/.rels":                packageRels("xl/workbook.xml"),
		"docProps/app.xml":           appProps("DreamWorker"),
		"docProps/core.xml":          coreProps(analysis.ProjectTitle),
		"xl/workbook.xml":            workbookXML(),
		"xl/_rels/workbook.xml.rels": workbookRels(),
		"xl/styles.xml":              xlsxStyles(),
		"xl/worksheets/sheet1.xml":   worksheetXML(rows),
		"xl/theme/theme1.xml":        xlsxTheme(),
	}
	return writeZip(path, files)
}

func writeRequirementSpecDOCX(path string, analysis RequirementAnalysisResult) error {
	var body strings.Builder
	body.WriteString(docxParagraph(analysis.ProjectTitle+" 需求规格说明", "Title"))
	body.WriteString(docxHeading("1. 文档信息", 1))
	body.WriteString(docxParagraph("生成工具：DreamWorker 需求分析模块", ""))
	body.WriteString(docxParagraph("生成时间："+time.Now().Format(time.RFC3339), ""))
	body.WriteString(docxHeading("2. 项目背景", 1))
	body.WriteString(docxParagraph(analysis.Summary, ""))
	body.WriteString(docxHeading("3. 需求来源", 1))
	for _, source := range analysis.Sources {
		body.WriteString(docxParagraph("• "+source, ""))
	}
	body.WriteString(docxHeading("4. 用户角色", 1))
	for _, role := range analysis.Roles {
		body.WriteString(docxParagraph("• "+role, ""))
	}
	body.WriteString(docxHeading("5. 功能需求", 1))
	body.WriteString(docxFeatureTable(analysis.Features))
	body.WriteString(docxHeading("6. 非功能需求", 1))
	for _, item := range analysis.NonFunctionalRequirements {
		body.WriteString(docxParagraph("• "+item, ""))
	}
	body.WriteString(docxHeading("7. 风险与约束", 1))
	for _, risk := range analysis.Risks {
		body.WriteString(docxParagraph("• "+risk, ""))
	}
	body.WriteString(docxHeading("8. 待确认问题", 1))
	for _, question := range analysis.OpenQuestions {
		body.WriteString(docxParagraph("• "+question, ""))
	}
	body.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1200" w:bottom="1440" w:left="1200" w:header="720" w:footer="720" w:gutter="0"/></w:sectPr>`)
	files := map[string]string{
		"[Content_Types].xml": docxContentTypes(),
		"_rels/.rels":         packageRels("word/document.xml"),
		"docProps/app.xml":    appProps("DreamWorker"),
		"docProps/core.xml":   coreProps(analysis.ProjectTitle),
		"word/document.xml":   docxDocument(body.String()),
		"word/styles.xml":     docxStyles(),
	}
	return writeZip(path, files)
}

func writeZip(path string, files map[string]string) error {
	buffer := bytes.Buffer{}
	writer := zip.NewWriter(&buffer)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		file, err := writer.Create(name)
		if err != nil {
			return err
		}
		if _, err := file.Write([]byte(files[name])); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buffer.Bytes(), 0o644)
}

func worksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews><cols>`)
	for col := 1; col <= 14; col++ {
		width := 18
		if col == 5 || col == 6 || col == 11 {
			width = 32
		}
		builder.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%d" customWidth="1"/>`, col, col, width))
	}
	builder.WriteString(`</cols><sheetData>`)
	for rowIndex, row := range rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex+1))
		for colIndex, value := range row {
			cell := cellName(colIndex+1, rowIndex+1)
			style := 0
			if rowIndex == 0 {
				style = 1
			}
			builder.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">%s</t></is></c>`, cell, style, xmlEscape(value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData><autoFilter ref="A1:N1"/></worksheet>`)
	return builder.String()
}

func cellName(col int, row int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
	}
	return fmt.Sprintf("%s%d", name, row)
}

func workbookXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="功能清单" sheetId="1" r:id="rId1"/></sheets></workbook>`
}

func workbookRels() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/></Relationships>`
}

func xlsxStyles() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts><fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills><borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="1"><alignment wrapText="1" vertical="top"/></xf><xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1" applyAlignment="1"><alignment wrapText="1" vertical="center"/></xf></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`
}

func xlsxTheme() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="DreamWorker"><a:themeElements><a:clrScheme name="DreamWorker"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="1F2937"/></a:dk2><a:lt2><a:srgbClr val="F8FAFC"/></a:lt2><a:accent1><a:srgbClr val="14B8A6"/></a:accent1><a:accent2><a:srgbClr val="2563EB"/></a:accent2><a:accent3><a:srgbClr val="F59E0B"/></a:accent3><a:accent4><a:srgbClr val="10B981"/></a:accent4><a:accent5><a:srgbClr val="EC4899"/></a:accent5><a:accent6><a:srgbClr val="64748B"/></a:accent6><a:hlink><a:srgbClr val="2563EB"/></a:hlink><a:folHlink><a:srgbClr val="7C3AED"/></a:folHlink></a:clrScheme><a:fontScheme name="DreamWorker"><a:majorFont><a:latin typeface="Calibri"/></a:majorFont><a:minorFont><a:latin typeface="Calibri"/></a:minorFont></a:fontScheme><a:fmtScheme name="DreamWorker"><a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst><a:lnStyleLst><a:ln w="9525"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln></a:lnStyleLst><a:effectStyleLst><a:effectStyle/></a:effectStyleLst><a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst></a:fmtScheme></a:themeElements></a:theme>`
}

func xlsxContentTypes() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/><Override PartName="/xl/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`
}

func docxDocument(body string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + body + `</w:body></w:document>`
}

func docxParagraph(text string, style string) string {
	styleXML := ""
	if style != "" {
		styleXML = `<w:pPr><w:pStyle w:val="` + style + `"/></w:pPr>`
	}
	return `<w:p>` + styleXML + `<w:r><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`
}

func docxHeading(text string, level int) string {
	return docxParagraph(text, fmt.Sprintf("Heading%d", level))
}

func docxFeatureTable(features []RequirementFeatureItem) string {
	rows := [][]string{{"功能ID", "模块", "功能名称", "优先级", "说明", "验收标准"}}
	for _, feature := range features {
		rows = append(rows, []string{
			feature.FeatureID,
			feature.Module,
			feature.Name,
			feature.Priority,
			feature.Description,
			strings.Join(feature.AcceptanceCriteria, "\n"),
		})
	}
	var builder strings.Builder
	builder.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/><w:left w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/><w:right w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="CBD5E1"/></w:tblBorders></w:tblPr>`)
	for _, row := range rows {
		builder.WriteString(`<w:tr>`)
		for _, cell := range row {
			builder.WriteString(`<w:tc><w:tcPr><w:tcW w:w="2400" w:type="dxa"/></w:tcPr>`)
			for _, line := range strings.Split(cell, "\n") {
				builder.WriteString(docxParagraph(line, ""))
			}
			builder.WriteString(`</w:tc>`)
		}
		builder.WriteString(`</w:tr>`)
	}
	builder.WriteString(`</w:tbl>`)
	return builder.String()
}

func docxStyles() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:rPr><w:sz w:val="22"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:rPr><w:b/><w:sz w:val="40"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:rPr><w:b/><w:sz w:val="30"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:rPr><w:b/><w:sz w:val="26"/></w:rPr></w:style></w:styles>`
}

func docxContentTypes() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`
}

func packageRels(target string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="` + target + `"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`
}

func coreProps(title string) string {
	now := time.Now().UTC().Format(time.RFC3339)
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>` + xmlEscape(title) + `</dc:title><dc:creator>DreamWorker</dc:creator><cp:lastModifiedBy>DreamWorker</cp:lastModifiedBy><dcterms:created xsi:type="dcterms:W3CDTF">` + now + `</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">` + now + `</dcterms:modified></cp:coreProperties>`
}

func appProps(app string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>` + xmlEscape(app) + `</Application></Properties>`
}

func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}
