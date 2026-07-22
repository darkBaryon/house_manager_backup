package chat

import (
	"strings"

	chatsvc "house-manager/internal/service/chat"
)

type sendRequest struct {
	Message    string `json:"message" binding:"required"`
	NewSession bool   `json:"new_session"`
}

func (r sendRequest) toServiceInput(requestID string, userID string) chatsvc.SendInput {
	return chatsvc.SendInput{
		RequestID:  requestID,
		UserID:     userID,
		Message:    strings.TrimSpace(r.Message),
		NewSession: r.NewSession,
	}
}
