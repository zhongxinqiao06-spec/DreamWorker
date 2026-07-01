package capability

import (
	"context"
	"encoding/json"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/domain"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/ports"
)

const (
	BuiltinArtifactRead        = domain.CapabilityIDArtifactRead
	BuiltinArtifactWrite       = domain.CapabilityIDArtifactWrite
	BuiltinWebSearchStub       = domain.CapabilityIDWebSearchStub
	BuiltinBrowserReadonlyStub = domain.CapabilityIDBrowserReadonlyStub
	BuiltinModelGenerateStub   = domain.CapabilityIDModelGenerateStub
	BuiltinHumanInput          = domain.CapabilityIDHumanInput
)

func BuiltinManifests() []domain.CapabilityManifest {
	return []domain.CapabilityManifest{
		builtinManifest(BuiltinArtifactRead, "Artifact Read", domain.RiskLow, nil),
		builtinManifest(BuiltinArtifactWrite, "Artifact Write", domain.RiskLow, nil),
		builtinManifest(BuiltinWebSearchStub, "Web Search Stub", domain.RiskLow, nil),
		builtinManifest(BuiltinBrowserReadonlyStub, "Browser Readonly Stub", domain.RiskLow, nil),
		builtinManifest(BuiltinModelGenerateStub, "Model Generate Stub", domain.RiskLow, nil),
		builtinManifest(BuiltinHumanInput, "Human Input", domain.RiskLow, nil),
	}
}

func RegisterBuiltins(ctx context.Context, registry ports.CapabilityRegistry) error {
	for _, manifest := range BuiltinManifests() {
		record, err := registry.Discover(ctx, manifest, domain.TrustTrustedBuiltin)
		if err != nil && record.Manifest.Metadata.ID == "" {
			return err
		}
		for _, state := range []domain.LifecycleState{
			domain.CapabilityRegistered,
			domain.CapabilitySchemaValidated,
			domain.CapabilityRiskClassified,
			domain.CapabilityAuthorized,
			domain.CapabilityEnabled,
		} {
			if _, err := registry.Transition(ctx, manifest.Metadata.ID, state); err != nil {
				return err
			}
		}
	}
	return nil
}

func BuiltinHandlers(artifacts ports.ArtifactStore) map[string]domain.CapabilityHandler {
	return map[string]domain.CapabilityHandler{
		BuiltinArtifactRead:        artifactReadHandler(artifacts),
		BuiltinArtifactWrite:       artifactWriteHandler(artifacts),
		BuiltinWebSearchStub:       stubHandler("web_search_stub", "未访问外部网络，返回固定搜索摘要。"),
		BuiltinBrowserReadonlyStub: stubHandler("browser_readonly_stub", "未打开真实浏览器，返回固定只读页面摘要。"),
		BuiltinModelGenerateStub:   stubHandler("model_generate_stub", "未调用真实模型，返回固定结构化文本。"),
		BuiltinHumanInput:          stubHandler("human_input", "等待后续 UI 接入人工输入。"),
	}
}

func builtinManifest(
	id string,
	name string,
	risk domain.RiskLevel,
	riskActions []domain.RiskAction,
) domain.CapabilityManifest {
	reasons := make([]string, 0, len(riskActions))
	for _, action := range riskActions {
		reasons = append(reasons, string(action))
	}
	return domain.CapabilityManifest{
		APIVersion: domain.CapabilityAPIVersion,
		Kind:       domain.CapabilityKindBuiltin,
		Metadata: domain.CapabilityMetadata{
			ID:       id,
			Name:     name,
			Version:  "0.1.0",
			Provider: "builtin",
		},
		Protocol: domain.CapabilityProtocol{Type: domain.CapabilityProtocolBuiltin},
		InputSchema: map[string]any{
			"type": "object",
		},
		OutputSchema: map[string]any{
			"type": "object",
		},
		Permissions: map[string]any{},
		Risk: domain.CapabilityRisk{
			Level:   risk,
			Reasons: reasons,
		},
		Approval:      map[string]any{},
		Runtime:       map[string]any{"timeoutMs": 30000},
		Observability: map[string]any{"logInputs": "summary", "logOutputs": "summary"},
	}
}

func artifactReadHandler(artifacts ports.ArtifactStore) domain.CapabilityHandler {
	return func(request domain.CapabilityInvocationRequest) (domain.CapabilityInvocationResult, error) {
		var input struct {
			ArtifactID string `json:"artifact_id"`
			Version    int    `json:"version"`
		}
		if err := json.Unmarshal(request.Input, &input); err != nil {
			return domain.CapabilityInvocationResult{OK: false, ErrorCode: "INVALID_INPUT", ErrorMessage: err.Error()}, err
		}
		artifact, err := artifacts.Read(context.Background(), input.ArtifactID, input.Version)
		if err != nil {
			return domain.CapabilityInvocationResult{OK: false, ErrorCode: "ARTIFACT_READ_FAILED", ErrorMessage: err.Error()}, err
		}
		return domain.CapabilityInvocationResult{
			OK: true,
			Output: JSONOutput(map[string]any{
				"artifact_id": artifact.Meta.ArtifactID,
				"version":     artifact.Meta.Version,
				"content":     string(artifact.Content),
			}),
		}, nil
	}
}

func artifactWriteHandler(artifacts ports.ArtifactStore) domain.CapabilityHandler {
	return func(request domain.CapabilityInvocationRequest) (domain.CapabilityInvocationResult, error) {
		var input struct {
			ArtifactID  string `json:"artifact_id"`
			Kind        string `json:"kind"`
			Title       string `json:"title"`
			Version     int    `json:"version"`
			ContentType string `json:"content_type"`
			FileName    string `json:"file_name"`
			Content     string `json:"content"`
		}
		if err := json.Unmarshal(request.Input, &input); err != nil {
			return domain.CapabilityInvocationResult{OK: false, ErrorCode: "INVALID_INPUT", ErrorMessage: err.Error()}, err
		}
		meta, err := artifacts.Put(context.Background(), domain.ArtifactWrite{
			ArtifactID:  input.ArtifactID,
			MissionID:   request.MissionID,
			RunID:       request.RunID,
			Kind:        input.Kind,
			Title:       input.Title,
			Version:     input.Version,
			ContentType: input.ContentType,
			TraceID:     request.TraceID,
			FileName:    input.FileName,
			Content:     []byte(input.Content),
		})
		if err != nil {
			return domain.CapabilityInvocationResult{OK: false, ErrorCode: "ARTIFACT_WRITE_FAILED", ErrorMessage: err.Error()}, err
		}
		return domain.CapabilityInvocationResult{
			OK: true,
			Output: JSONOutput(map[string]any{
				"artifact_id": meta.ArtifactID,
				"version":     meta.Version,
				"uri":         meta.URI,
			}),
		}, nil
	}
}

func stubHandler(kind string, summary string) domain.CapabilityHandler {
	return func(request domain.CapabilityInvocationRequest) (domain.CapabilityInvocationResult, error) {
		return domain.CapabilityInvocationResult{
			OK: true,
			Output: JSONOutput(map[string]any{
				"kind":     kind,
				"summary":  summary,
				"trace_id": request.TraceID,
			}),
		}, nil
	}
}
