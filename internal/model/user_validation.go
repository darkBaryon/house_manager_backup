package model

import "fmt"

func (m *User) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user is nil")
	}
	if isBlank(m.Phone) {
		return fmt.Errorf("phone is required")
	}
	if !m.SourceChannel.ValidOptional() {
		return fmt.Errorf("sourceChannel is invalid")
	}
	return nil
}

func (m *UserProfileExt) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user profile ext is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if m.BudgetMin < 0 || m.BudgetMax < 0 {
		return fmt.Errorf("budget must be non-negative")
	}
	return nil
}
