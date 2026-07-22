package chat

import (
	"fmt"
	"strings"
)

func (m *Session) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("chat session is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if !ValidSessionStatus(m.SessionStatus) {
		return fmt.Errorf("sessionStatus is invalid")
	}
	if m.StartedAt <= 0 {
		return fmt.Errorf("startedAt is required")
	}
	if m.LastActiveAt <= 0 {
		return fmt.Errorf("lastActiveAt is required")
	}
	if m.LastSeq < 0 {
		return fmt.Errorf("lastSeq is invalid")
	}
	return nil
}

func (m *Message) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("chat message is nil")
	}
	if m.SessionID.IsZero() {
		return fmt.Errorf("sessionID is required")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if m.Seq <= 0 {
		return fmt.Errorf("seq is required")
	}
	if !ValidMessageRole(m.Role) {
		return fmt.Errorf("role is invalid")
	}
	if strings.TrimSpace(m.Content) == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

func ValidSessionStatus(status int) bool {
	switch status {
	case SessionStatusActive, SessionStatusEnded:
		return true
	default:
		return false
	}
}

func ValidMessageRole(role string) bool {
	switch strings.TrimSpace(role) {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem:
		return true
	default:
		return false
	}
}
