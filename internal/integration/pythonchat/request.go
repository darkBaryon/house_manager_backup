package pythonchat

import chatservice "house-manager/internal/service/chat"

type respondRequest struct {
	SessionID      string           `json:"session_id"`
	Message        string           `json:"message"`
	RecentMessages []respondMessage `json:"recent_messages"`
	RuntimeContext map[string]any   `json:"runtime_context"`
}

type respondMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func newRespondRequest(input chatservice.AIRespondInput) respondRequest {
	return respondRequest{
		SessionID:      input.SessionID,
		Message:        input.Message,
		RecentMessages: respondMessages(input.RecentMessages),
		RuntimeContext: input.RuntimeContext,
	}
}

func respondMessages(items []chatservice.RecentMessage) []respondMessage {
	if len(items) == 0 {
		return []respondMessage{}
	}
	out := make([]respondMessage, 0, len(items))
	for _, item := range items {
		out = append(out, respondMessage{Role: item.Role, Content: item.Content})
	}
	return out
}
