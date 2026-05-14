package auth

import (
	authsvc "house-manager/internal/service/publish/auth"
	"house-manager/pkg/session"
)

type loginRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type principalResponse struct {
	PrincipalType   string   `json:"principal_type"`
	PrincipalID     string   `json:"principal_id"`
	Terminal        string   `json:"terminal"`
	Phone           string   `json:"phone"`
	RoleCodes       []string `json:"role_codes"`
	PermissionCodes []string `json:"permission_codes"`
}

type subjectResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone"`
}

type loginResponse struct {
	Token     string            `json:"token"`
	Principal principalResponse `json:"principal"`
	Subject   subjectResponse   `json:"subject"`
}

type sessionResponse struct {
	Principal principalResponse `json:"principal"`
	Subject   subjectResponse   `json:"subject"`
}

type logoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}

func toLoginResponse(result *authsvc.LoginResult) loginResponse {
	return loginResponse{
		Token:     result.Token,
		Principal: toPrincipalResponse(result.Principal),
		Subject:   toSubjectResponse(result.Subject),
	}
}

func toSessionResponse(result *authsvc.AuthSession) sessionResponse {
	return sessionResponse{
		Principal: toPrincipalResponse(result.Principal),
		Subject:   toSubjectResponse(result.Subject),
	}
}

func toPrincipalResponse(principal session.Principal) principalResponse {
	return principalResponse{
		PrincipalType:   principal.PrincipalType,
		PrincipalID:     principal.PrincipalID,
		Terminal:        principal.Terminal,
		Phone:           principal.Phone,
		RoleCodes:       principal.RoleCodes,
		PermissionCodes: principal.PermissionCodes,
	}
}

func toSubjectResponse(subject authsvc.Subject) subjectResponse {
	return subjectResponse{
		ID:    subject.ID,
		Name:  subject.Name,
		Phone: subject.Phone,
	}
}
