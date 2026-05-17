package auth

import "house-manager/pkg/session"

type LoginInput struct {
	Phone     string
	Password  string
	LoginIP   string
	UserAgent string
}

type StaffProfile struct {
	StaffID       string
	Name          string
	Phone         string
	Email         string
	Department    string
	JobTitle      string
	ContactQRCode string
}

type AuthSession struct {
	Principal       session.Principal
	StaffProfile    StaffProfile
	RoleCodes       []string
	PermissionCodes []string
}

type LoginResult struct {
	Token string
	AuthSession
}
