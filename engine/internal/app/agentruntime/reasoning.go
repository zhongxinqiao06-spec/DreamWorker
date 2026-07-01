package agentruntime

import "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"

func MergeUsage(current *resources.ChatModelUsage, next *resources.ChatModelUsage) *resources.ChatModelUsage {
	if next == nil {
		return current
	}
	if current == nil {
		copy := *next
		if copy.TotalTokens == 0 {
			copy.TotalTokens = copy.InputTokens + copy.OutputTokens
		}
		return &copy
	}
	if next.InputTokens > 0 {
		current.InputTokens = next.InputTokens
	}
	if next.OutputTokens > 0 {
		current.OutputTokens = next.OutputTokens
	}
	if next.TotalTokens > 0 {
		current.TotalTokens = next.TotalTokens
	} else {
		current.TotalTokens = current.InputTokens + current.OutputTokens
	}
	if next.CostUSD > 0 {
		current.CostUSD = next.CostUSD
	}
	return current
}
