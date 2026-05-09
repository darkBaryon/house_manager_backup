package model

import "fmt"

func (m *UserAuth) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user auth is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if !m.AuthProvider.Valid() {
		return fmt.Errorf("authProvider is invalid")
	}
	if isBlank(m.OpenID) {
		return fmt.Errorf("openID is required")
	}
	return nil
}
