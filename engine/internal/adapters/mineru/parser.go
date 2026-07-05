package mineru

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	minerusdk "github.com/opendatalab/MinerU-Ecosystem/sdk/go"
)

type Parser struct{}

const flashMaxBytes = 10 * 1024 * 1024

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseDocument(ctx context.Context, inputPath string, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	if command, ok, err := resolveCommand(); err != nil {
		return "", err
	} else if ok {
		return parseWithCLI(ctx, command, inputPath, outputDir)
	}
	return parseWithOpenAPI(ctx, inputPath, outputDir)
}

func resolveCommand() (string, bool, error) {
	command := strings.TrimSpace(os.Getenv("MINERU_COMMAND"))
	if command == "" {
		found, err := exec.LookPath("mineru")
		if err != nil {
			return "", false, nil
		}
		return found, true, nil
	}
	if _, err := exec.LookPath(command); err != nil && !filepath.IsAbs(command) {
		return "", false, err
	}
	return command, true, nil
}

func parseWithCLI(ctx context.Context, command string, inputPath string, outputDir string) (string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(runCtx, command, minerUArgs(inputPath, outputDir)...)
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		if len(output) > 0 {
			return "", errors.New(limitRunes(string(output), 800))
		}
		return "", err
	}
	return collectMinerUOutput(outputDir)
}

func parseWithOpenAPI(ctx context.Context, inputPath string, outputDir string) (string, error) {
	client, mode, err := newMinerUAPIClient()
	if err != nil {
		return "", err
	}
	if mode == "flash" {
		if info, statErr := os.Stat(inputPath); statErr == nil && info.Size() > flashMaxBytes {
			return "", fmt.Errorf("mineru flash mode supports files up to 10 MB; configure MINERU_TOKEN or a local MinerU CLI for larger files")
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	result, err := extractWithOpenAPI(runCtx, client, mode, inputPath)
	if err != nil {
		return "", fmt.Errorf("mineru open api %s extract failed: %w", mode, err)
	}
	if result.Err() != nil {
		return "", result.Err()
	}
	markdown := strings.TrimSpace(result.Markdown)
	if markdown == "" {
		return "", errors.New("mineru open api returned empty markdown")
	}
	outputPath := filepath.Join(outputDir, safeOutputStem(inputPath)+".md")
	if err := os.WriteFile(outputPath, []byte(markdown), 0o644); err != nil {
		return "", err
	}
	return markdown, nil
}

func newMinerUAPIClient() (*minerusdk.Client, string, error) {
	options := []minerusdk.ClientOption{
		minerusdk.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
	}
	if baseURL := strings.TrimSpace(os.Getenv("MINERU_BASE_URL")); baseURL != "" {
		options = append(options, minerusdk.WithBaseURL(baseURL))
	}
	if baseURL := strings.TrimSpace(os.Getenv("MINERU_FLASH_BASE_URL")); baseURL != "" {
		options = append(options, minerusdk.WithFlashBaseURL(baseURL))
	}
	if strings.TrimSpace(os.Getenv("MINERU_TOKEN")) != "" {
		client, err := minerusdk.New("", options...)
		if err != nil {
			return nil, "", err
		}
		client.SetSource("dreamworker-requirements")
		return client, "precision", nil
	}
	client := minerusdk.NewFlash(options...)
	client.SetSource("dreamworker-requirements")
	return client, "flash", nil
}

func extractWithOpenAPI(ctx context.Context, client *minerusdk.Client, mode string, inputPath string) (*minerusdk.ExtractResult, error) {
	if mode == "precision" {
		return client.Extract(
			ctx,
			inputPath,
			minerusdk.WithLanguage(minerULanguage()),
			minerusdk.WithOCR(minerUFlashOCR()),
			minerusdk.WithFormula(true),
			minerusdk.WithTable(true),
			minerusdk.WithPollTimeout(10*time.Minute),
		)
	}
	return client.FlashExtract(
		ctx,
		inputPath,
		minerusdk.WithFlashLanguage(minerULanguage()),
		minerusdk.WithFlashOCR(minerUFlashOCR()),
		minerusdk.WithFlashFormula(true),
		minerusdk.WithFlashTable(true),
		minerusdk.WithFlashTimeout(10*time.Minute),
	)
}

func minerUArgs(inputPath string, outputDir string) []string {
	template := strings.TrimSpace(os.Getenv("MINERU_ARGS_TEMPLATE"))
	if template == "" {
		return []string{"--path", inputPath, "--output", outputDir}
	}
	parts := strings.Fields(template)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ReplaceAll(part, "{input}", inputPath)
		part = strings.ReplaceAll(part, "{output}", outputDir)
		result = append(result, part)
	}
	return result
}

func minerULanguage() string {
	language := strings.TrimSpace(os.Getenv("MINERU_LANGUAGE"))
	if language == "" {
		return "ch"
	}
	return language
}

func minerUFlashOCR() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("MINERU_FLASH_OCR")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func collectMinerUOutput(outputDir string) (string, error) {
	var builder strings.Builder
	err := filepath.WalkDir(outputDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".json" && ext != ".txt" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.TrimSpace(string(content)) == "" {
			return nil
		}
		builder.WriteString("\n\n--- MinerU Output: ")
		builder.WriteString(filepath.Base(path))
		builder.WriteString(" ---\n")
		builder.Write(content)
		return nil
	})
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(builder.String())
	if text == "" {
		return "", errors.New("mineru output is empty")
	}
	return text, nil
}

func safeOutputStem(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name = strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_").Replace(name)
	if name == "" || name == "." {
		return "mineru_output"
	}
	return name
}

func limitRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
