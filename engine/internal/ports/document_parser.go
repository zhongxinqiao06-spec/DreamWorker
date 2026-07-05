package ports

import "context"

type DocumentParser interface {
	ParseDocument(ctx context.Context, inputPath string, outputDir string) (string, error)
}
