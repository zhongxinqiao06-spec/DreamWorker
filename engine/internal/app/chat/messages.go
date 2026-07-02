package chat

func (s *Store) ListChatMessages(sessionID string) ([]ChatMessage, *AppError) {
	if sessionID == "" {
		return nil, BadRequest("BAD_REQUEST", "缺少 sessionId。", "请选择一个聊天会话。")
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, ok := s.Sessions[sessionID]; !ok {
		return nil, NotFound("SESSION_NOT_FOUND", "会话不存在。", "请创建新的聊天会话。")
	}
	return append([]ChatMessage{}, s.Messages[sessionID]...), nil
}
