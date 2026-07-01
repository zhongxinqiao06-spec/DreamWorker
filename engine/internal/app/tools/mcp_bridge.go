package tools

import (
	"context"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

func (r Registry) lookupMCPBinding(toolID string) (resources.MCPToolBinding, bool) {
	r.State.Mu.Lock()
	defer r.State.Mu.Unlock()
	binding, ok := r.State.MCPTools[toolID]
	return binding, ok
}

func (r Registry) callMCPTool(ctx context.Context, binding resources.MCPToolBinding, arguments string) (string, error) {
	return r.State.CallMCPTool(ctx, binding, arguments)
}
