package tools

import "github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"

type Registry struct {
	State *resources.Store
}

func NewRegistry(state *resources.Store) Registry {
	return Registry{State: state}
}
