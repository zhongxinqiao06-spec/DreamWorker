package resources

import "strings"

func (s *Store) ListProfiles() []ModelProfile {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return sortedValues(s.Profiles, profileSortKey)
}

func EffectiveProviderModel(profile ModelProfile, provider ModelProviderRecord) (string, string) {
	providerID := strings.TrimSpace(profile.ProviderID)
	model := strings.TrimSpace(profile.Model)
	if providerID == "" {
		providerID = strings.TrimSpace(provider.ProviderID)
	}
	if model == "" {
		model = strings.TrimSpace(provider.DefaultModel)
	}
	return providerID, model
}

func ProfileIDForProviderModel(providerID string, model string) string {
	value := strings.TrimSpace(providerID) + "_" + strings.TrimSpace(model)
	return "profile_" + sanitizeID(value)
}

func ProfileFromProviderModel(provider ModelProviderRecord, model string, timestamp string) ModelProfile {
	selectedModel := strings.TrimSpace(model)
	if selectedModel == "" {
		selectedModel = provider.DefaultModel
	}
	return ensureModelProfileDefaults(ModelProfile{
		ProfileID:      ProfileIDForProviderModel(provider.ProviderID, selectedModel),
		DisplayName:    provider.DisplayName + " / " + selectedModel,
		ProviderID:     provider.ProviderID,
		Model:          selectedModel,
		Temperature:    0.4,
		MaxTokens:      4096,
		ContextWindow:  128000,
		ResponseFormat: "text",
		ToolMode:       "auto",
		TimeoutMS:      120000,
		Purpose:        "服务商模型默认配置",
		Enabled:        true,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
	})
}

func profileSortKey(item ModelProfile) string {
	switch item.ProfileID {
	case "profile_fast":
		return "00-" + item.DisplayName
	case "profile_pro":
		return "01-" + item.DisplayName
	case "profile_siliconflow":
		return "02-" + item.DisplayName
	case "profile_stub":
		return "99-" + item.DisplayName
	default:
		return "50-" + item.DisplayName
	}
}

func (s *Store) SaveProfile(input ModelProfile) (ModelProfile, *AppError) {
	if input.ProfileID == "" {
		return ModelProfile{}, BadRequest("BAD_REQUEST", "模型配置格式无效。", "请检查 profileId、供应商和模型。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	now := s.Now()
	existing, exists := s.Profiles[input.ProfileID]
	if !exists {
		input.CreatedAt = now
	} else {
		input.CreatedAt = existing.CreatedAt
	}
	input = ensureModelProfileDefaults(input)
	input.UpdatedAt = now
	s.Profiles[input.ProfileID] = input
	return input, nil
}

func (s *Store) DeleteProfile(profileID string) (DeleteResult, *AppError) {
	if profileID == "" {
		return DeleteResult{}, BadRequest("BAD_REQUEST", "缺少 profileId。", "请选择要删除的模型配置。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.Profiles, profileID)
	return DeleteResult{OK: true, DeletedID: profileID}, nil
}
