package auth

import "house-manager/pkg/session"

type LoginInput struct {
	Phone     string
	LoginIP   string
	UserAgent string
}

type Subject struct {
	ID    string
	Name  string
	Phone string
}

type AuthSession struct {
	Principal session.Principal
	Subject   Subject
}

type LoginResult struct {
	Token string
	AuthSession
}
