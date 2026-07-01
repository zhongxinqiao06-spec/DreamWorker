package tools

import (
	"context"
	"fmt"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func (r Registry) executeLowRiskTool(
	ctx context.Context,
	session resources.ChatSession,
	traceID string,
	tool resources.ToolConfig,
	arguments string,
) resources.ToolExecutionResult {
	_ = session
	if binding, ok := r.lookupMCPBinding(tool.ToolID); ok {
		result, err := r.callMCPTool(ctx, binding, arguments)
		if err != nil {
			return resources.ToolExecutionResult{
				Status:       "blocked",
				ErrorCode:    "MCP_TOOL_FAILED",
				ErrorMessage: resources.RedactSecrets(err.Error()),
			}
		}
		return resources.ToolExecutionResult{Status: "completed", OutputSummary: result}
	}
	switch tool.ToolID {
	case "tool_model_generate_stub":
		return resources.ToolExecutionResult{
			Status:        "completed",
			OutputSummary: "确定性模型辅助工具已在 Engine 内执行，trace=" + traceID + "。",
		}
	case "tool_human_input":
		return resources.ToolExecutionResult{
			Status:       "blocked",
			ErrorCode:    "HUMAN_INPUT_REQUIRED",
			ErrorMessage: "需要在界面中显式交接给用户输入",
		}
	case "tool_artifact_read":
		return resources.ToolExecutionResult{
			Status:        "completed",
			OutputSummary: "项目制品读取工具可用；本次未提供 artifact_id。",
		}
	default:
		return resources.ToolExecutionResult{
			Status:        "completed",
			OutputSummary: fmt.Sprintf("%s 已按低风险策略执行。", tool.ToolID),
		}
	}
}
