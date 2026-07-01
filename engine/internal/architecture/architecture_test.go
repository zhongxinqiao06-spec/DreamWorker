package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/zhongxinqiao06-spec/DreamWorker/engine"

func TestLayerBoundaries(t *testing.T) {
	internalRoot := filepath.Clean(filepath.Join(".."))

	err := filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		layer := layerForPath(path)
		if layer == "" {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imported := range file.Imports {
			assertLayerImportAllowed(t, path, layer, imported)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal packages: %v", err)
	}
}

func layerForPath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func assertLayerImportAllowed(t *testing.T, filePath string, layer string, imported *ast.ImportSpec) {
	t.Helper()

	importPath, err := strconv.Unquote(imported.Path.Value)
	if err != nil {
		t.Fatalf("%s: parse import path: %v", filePath, err)
	}

	switch layer {
	case "domain":
		assertDoesNotImportInternal(t, filePath, importPath, "adapters", "platform", "api")
	case "app":
		assertDoesNotImportInternal(t, filePath, importPath, "adapters", "platform", "api")
		assertNoExternalSDK(t, filePath, importPath)
	case "runtime":
		assertRuntimeImportAllowed(t, filePath, importPath)
	case "api":
		assertDoesNotImportInternal(t, filePath, importPath, "adapters", "platform")
	}
}

func assertDoesNotImportInternal(t *testing.T, filePath string, importPath string, denied ...string) {
	t.Helper()
	for _, layer := range denied {
		if strings.HasPrefix(importPath, modulePath+"/internal/"+layer) {
			t.Fatalf("%s: must not import %s", filePath, importPath)
		}
	}
}

func assertNoExternalSDK(t *testing.T, filePath string, importPath string) {
	t.Helper()
	if strings.HasPrefix(importPath, modulePath) {
		return
	}
	if strings.Contains(importPath, ".") {
		t.Fatalf("%s: app layer must not import external SDK %s", filePath, importPath)
	}
}

func assertRuntimeImportAllowed(t *testing.T, filePath string, importPath string) {
	t.Helper()
	if !strings.HasPrefix(importPath, modulePath) {
		assertNoExternalSDK(t, filePath, importPath)
		return
	}
	if strings.HasPrefix(importPath, modulePath+"/internal/domain") ||
		strings.HasPrefix(importPath, modulePath+"/internal/ports") {
		return
	}
	t.Fatalf("%s: runtime layer can import only domain and ports, got %s", filePath, importPath)
}
