package agentruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type Runtime struct {
	State *resources.Store
}

func NewRuntime(state *resources.Store) Runtime {
	return Runtime{State: state}
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func sortedStrings(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}

func ContentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func LatencyMS(startedAt time.Time) int {
	return resources.LatencyMS(startedAt)
}

func ChatStep(stepID string, phase string, title string, summary string, status string, timestamp string) resources.ChatExecutionStep {
	completedAt := ""
	if status == "completed" || status == "ready" {
		completedAt = timestamp
	}
	return resources.ChatExecutionStep{
		StepID:      stepID,
		Phase:       phase,
		Title:       title,
		Summary:     summary,
		Status:      status,
		StartedAt:   timestamp,
		CompletedAt: completedAt,
	}
}

func BuildCompletedChatExecutionSteps(timestamp string, agent resources.AgentConfig) []resources.ChatExecutionStep {
	strategy := fallback(agent.Planner.Strategy, "plan-execute")
	return []resources.ChatExecutionStep{
		ChatStep("step_plan", "PLAN", "规划", "已绑定 Agent、模型配置与项目上下文。", "completed", timestamp),
		ChatStep("step_graph", "GRAPH", "任务图", fmt.Sprintf("已按 %s 策略创建可观测执行图。", strategy), "completed", timestamp),
		ChatStep("step_execute", "EXECUTE", "模型流", "已通过 Engine 流式调用模型供应商。", "completed", timestamp),
		ChatStep("step_observe", "OBSERVE", "持久化结果", "已记录响应、token 用量与 trace 信息。", "completed", timestamp),
		ChatStep("step_replan", "REPLAN", "等待下一步", "可继续追问、重试或切换项目上下文。", "ready", timestamp),
	}
}
