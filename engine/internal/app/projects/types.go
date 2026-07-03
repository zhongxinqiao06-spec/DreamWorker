package projects

import (
	"sort"
	"strings"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/resources"
)

type Store struct {
	*resources.Store
}

func NewStore(state *resources.Store) *Store {
	return &Store{Store: state}
}

type Project = resources.Project
type ProjectModuleConfig = resources.ProjectModuleConfig
type ProjectMemoryConfig = resources.ProjectMemoryConfig
type ProjectRunPolicy = resources.ProjectRunPolicy
type ProjectSecurityPolicy = resources.ProjectSecurityPolicy
type ProjectModule = resources.ProjectModule
type ProjectSubmodule = resources.ProjectSubmodule
type CreateProjectInput = resources.CreateProjectInput
type UpdateProjectInput = resources.UpdateProjectInput
type ProjectDirectoryEntryCheck = resources.ProjectDirectoryEntryCheck
type ProjectDirectoryCheck = resources.ProjectDirectoryCheck
type ProjectManifestExport = resources.ProjectManifestExport
type ModuleRequest = resources.ModuleRequest
type UpdateModuleConfigInput = resources.UpdateModuleConfigInput
type DeleteResult = resources.DeleteResult
type AppError = resources.AppError

var BadRequest = resources.BadRequest
var NotFound = resources.NotFound

func sortedValues[T any](items map[string]T, key func(T) string) []T {
	values := make([]T, 0, len(items))
	for _, value := range items {
		values = append(values, value)
	}
	return sortSlice(values, key)
}

func sortSlice[T any](values []T, key func(T) string) []T {
	sort.Slice(values, func(i, j int) bool {
		return key(values[i]) < key(values[j])
	})
	return values
}

func cloneAnyMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
