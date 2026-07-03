package resources

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Store) ListSkills() []SkillConfig {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return sortedValues(s.Skills, func(item SkillConfig) string { return item.DisplayName })
}

func (s *Store) GetSkill(skillID string) (SkillConfig, *AppError) {
	if skillID == "" {
		return SkillConfig{}, BadRequest("BAD_REQUEST", "缺少 skillId。", "请选择要查看的 Skill。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	skill, ok := s.Skills[skillID]
	if !ok {
		return SkillConfig{}, NotFound("SKILL_NOT_FOUND", "Skill 不存在。", "请刷新 Skill 列表。")
	}
	return skill, nil
}

func (s *Store) SaveSkill(input SkillConfig) (SkillConfig, *AppError) {
	if input.SkillID == "" {
		return SkillConfig{}, BadRequest("BAD_REQUEST", "Skill 配置格式无效。", "请检查 skillId 和输出产物。")
	}
	input = ensureSkillDefaults(input)
	s.Mu.Lock()
	if existing, ok := s.Skills[input.SkillID]; ok {
		input.BuiltIn = existing.BuiltIn
		if input.SourcePath == "" {
			input.SourcePath = existing.SourcePath
		}
	}
	s.Mu.Unlock()
	written, err := s.writeAgentSkillFile(input)
	if err != nil {
		return SkillConfig{}, BadRequest("SKILL_WRITE_FAILED", "Skill 文件写入失败。", "请检查 .agent 目录权限。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	input = written
	s.Skills[input.SkillID] = input
	if appErr := s.persistWorkspaceLocked(); appErr != nil {
		return SkillConfig{}, appErr
	}
	return input, nil
}

func (s *Store) DeleteSkill(skillID string) (DeleteResult, *AppError) {
	if skillID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 skillId。", "请选择要删除的 Skill。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Skills, skillID)
	if appErr := s.persistWorkspaceLocked(); appErr != nil {
		return DeleteResult{}, appErr
	}
	return DeleteResult{OK: true, DeletedID: skillID}, nil
}

func (s *Store) loadAgentSkills() {
	for _, file := range discoverAgentSkillFiles(s.AgentDir) {
		skill, err := readAgentSkillFile(file)
		if err != nil {
			continue
		}
		s.Skills[skill.SkillID] = skill
	}
}

func discoverAgentSkillFiles(agentDir string) []string {
	if strings.TrimSpace(agentDir) == "" {
		return nil
	}
	var files []string
	roots := []string{filepath.Join(agentDir, "skills"), agentDir}
	seen := map[string]bool{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "skills" {
				continue
			}
			path := filepath.Join(root, entry.Name(), "SKILL.md")
			if seen[path] {
				continue
			}
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				files = append(files, path)
				seen[path] = true
			}
		}
	}
	sort.Strings(files)
	return files
}

func readAgentSkillFile(path string) (SkillConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SkillConfig{}, err
	}
	metadata, body := parseSkillMarkdown(string(data))
	command := filepath.Base(filepath.Dir(path))
	command = fallback(metadata["command"], command)
	name := fallback(metadata["name"], titleFromMarkdown(body, command))
	description := fallback(metadata["description"], firstParagraph(body))
	category := fallback(metadata["category"], "general")
	version := fallback(metadata["version"], "0.1.0")
	allowedTools := metadataList(metadata, "allowed-tools")
	outputArtifacts := metadataList(metadata, "output-artifacts")
	builtIn := metadata["dreamworker-built-in"] != "false"
	absolute, _ := filepath.Abs(path)
	return SkillConfig{
		SkillID:              skillIDFromCommand(command),
		CommandName:          command,
		DisplayName:          name,
		Description:          redactSecrets(description),
		WhenToUse:            redactSecrets(metadata["when_to_use"]),
		Instructions:         redactSecrets(strings.TrimSpace(body)),
		Category:             category,
		Version:              version,
		Enabled:              true,
		BuiltIn:              builtIn,
		SourcePath:           absolute,
		RequiredCapabilities: capabilitiesFromAllowedTools(allowedTools),
		OutputArtifacts:      outputArtifacts,
	}, nil
}

func parseSkillMarkdown(content string) (map[string]string, string) {
	metadata := map[string]string{}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return metadata, normalized
	}
	rest := strings.TrimPrefix(normalized, "---\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return metadata, normalized
	}
	frontmatter := rest[:end]
	body := strings.TrimPrefix(rest[end:], "\n---")
	body = strings.TrimPrefix(body, "\n")
	for _, line := range strings.Split(frontmatter, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		metadata[strings.ToLower(strings.TrimSpace(key))] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return metadata, body
}

func metadataList(metadata map[string]string, key string) []string {
	value := strings.TrimSpace(metadata[key])
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"'`)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func titleFromMarkdown(body string, fallbackValue string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	words := strings.Fields(strings.ReplaceAll(fallbackValue, "-", " "))
	for index, word := range words {
		if word == "" {
			continue
		}
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func firstParagraph(body string) string {
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block != "" && !strings.HasPrefix(block, "#") {
			return strings.Join(strings.Fields(block), " ")
		}
	}
	return ""
}

func skillIDFromCommand(command string) string {
	return "skill_" + sanitizeID(command)
}

func skillCommandFromID(skillID string) string {
	command := strings.TrimPrefix(skillID, "skill_")
	command = strings.ReplaceAll(command, "_", "-")
	return sanitizeCommand(command)
}

func sanitizeCommand(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z':
			builder.WriteRune(ch)
			lastDash = false
		case ch >= '0' && ch <= '9':
			builder.WriteRune(ch)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteRune('-')
				lastDash = true
			}
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "skill"
	}
	return result
}

func capabilitiesFromAllowedTools(tools []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, tool := range tools {
		capability := capabilityFromAllowedTool(tool)
		if capability == "" || seen[capability] {
			continue
		}
		seen[capability] = true
		result = append(result, capability)
	}
	return result
}

func capabilityFromAllowedTool(tool string) string {
	switch sanitizeID(tool) {
	case "artifact_read", "filesystem_project_read", "document_read", "code_reference_read":
		return "cap_artifact_read"
	case "artifact_write":
		return "cap_artifact_write"
	case "web_search":
		return "cap_web_search_stub"
	case "browser_readonly":
		return "cap_browser_readonly_stub"
	case "model_reasoning":
		return "cap_model_generate_stub"
	case "human_question":
		return "cap_human_input"
	default:
		return ""
	}
}

func allowedToolsFromCapabilities(capabilities []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, capability := range capabilities {
		tool := allowedToolFromCapability(capability)
		if tool == "" || seen[tool] {
			continue
		}
		seen[tool] = true
		result = append(result, tool)
	}
	return result
}

func allowedToolFromCapability(capability string) string {
	switch capability {
	case "cap_artifact_read":
		return "artifact_read"
	case "cap_artifact_write":
		return "artifact_write"
	case "cap_web_search_stub":
		return "web_search"
	case "cap_browser_readonly_stub":
		return "browser_readonly"
	case "cap_model_generate_stub":
		return "model_reasoning"
	case "cap_human_input":
		return "human_question"
	default:
		return ""
	}
}

func (s *Store) writeAgentSkillFile(skill SkillConfig) (SkillConfig, error) {
	skill = ensureSkillDefaults(skill)
	if strings.TrimSpace(skill.SourcePath) == "" {
		skill.SourcePath = filepath.Join(s.AgentDir, "skills", skill.CommandName, "SKILL.md")
	}
	if err := os.MkdirAll(filepath.Dir(skill.SourcePath), 0o755); err != nil {
		return skill, err
	}
	content := renderAgentSkillMarkdown(skill)
	if err := os.WriteFile(skill.SourcePath, []byte(content), 0o644); err != nil {
		return skill, err
	}
	absolute, _ := filepath.Abs(skill.SourcePath)
	skill.SourcePath = absolute
	return skill, nil
}

func ensureSkillDefaults(skill SkillConfig) SkillConfig {
	if strings.TrimSpace(skill.CommandName) == "" {
		skill.CommandName = skillCommandFromID(skill.SkillID)
	}
	if strings.TrimSpace(skill.SkillID) == "" {
		skill.SkillID = skillIDFromCommand(skill.CommandName)
	}
	if strings.TrimSpace(skill.DisplayName) == "" {
		skill.DisplayName = titleFromMarkdown(skill.Instructions, skill.CommandName)
	}
	if strings.TrimSpace(skill.Description) == "" {
		skill.Description = firstParagraph(skill.Instructions)
	}
	if strings.TrimSpace(skill.Category) == "" {
		skill.Category = "general"
	}
	if strings.TrimSpace(skill.Version) == "" {
		skill.Version = "0.1.0"
	}
	skill.Enabled = true
	return skill
}

func renderAgentSkillMarkdown(skill SkillConfig) string {
	allowedTools := allowedToolsFromCapabilities(skill.RequiredCapabilities)
	instructions := strings.TrimSpace(skill.Instructions)
	if instructions == "" {
		instructions = "## Instructions\n\n" + strings.TrimSpace(skill.Description)
	}
	return fmt.Sprintf(`---
name: %s
description: %s
when_to_use: %s
allowed-tools: %s
category: %s
version: %s
output-artifacts: %s
dreamworker-built-in: %t
---

%s
`,
		skill.DisplayName,
		skill.Description,
		skill.WhenToUse,
		strings.Join(allowedTools, ", "),
		skill.Category,
		skill.Version,
		strings.Join(skill.OutputArtifacts, ", "),
		skill.BuiltIn,
		instructions,
	)
}
