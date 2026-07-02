package projects

func (s *Store) SeedDefaults(timestamp string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if len(s.Projects) > 0 {
		return
	}
	s.createProjectLocked(CreateProjectInput{
		Title:       "独立开发者 AI 项目孵化器",
		Description: "从机会探索、产品定义、工程开发到销售发布的默认项目空间。",
	}, timestamp)
}
