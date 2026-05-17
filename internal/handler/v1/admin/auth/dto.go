package auth

import (
	authsvc "house-manager/internal/service/admin/auth"
	"house-manager/pkg/session"
)

type loginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type principalResponse struct {
	PrincipalType   string   `json:"principal_type"`
	PrincipalID     string   `json:"principal_id"`
	Terminal        string   `json:"terminal"`
	Phone           string   `json:"phone"`
	RoleCodes       []string `json:"role_codes"`
	PermissionCodes []string `json:"permission_codes"`
}

type staffProfileResponse struct {
	StaffID       string `json:"staff_id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Department    string `json:"department"`
	JobTitle      string `json:"job_title"`
	ContactQRCode string `json:"contact_qr_code"`
}

type loginResponse struct {
	Token           string               `json:"token"`
	Principal       principalResponse    `json:"principal"`
	StaffProfile    staffProfileResponse `json:"staff_profile"`
	RoleCodes       []string             `json:"role_codes"`
	PermissionCodes []string             `json:"permission_codes"`
}

type sessionResponse struct {
	Principal       principalResponse    `json:"principal"`
	StaffProfile    staffProfileResponse `json:"staff_profile"`
	RoleCodes       []string             `json:"role_codes"`
	PermissionCodes []string             `json:"permission_codes"`
}

type logoutResponse struct {
	Success bool `json:"success"`
}

func toLoginResponse(result *authsvc.LoginResult) loginResponse {
	return loginResponse{
		Token:           result.Token,
		Principal:       toPrincipalResponse(result.Principal),
		StaffProfile:    toStaffProfileResponse(result.StaffProfile),
		RoleCodes:       result.RoleCodes,
		PermissionCodes: result.PermissionCodes,
	}
}

func toSessionResponse(result *authsvc.AuthSession) sessionResponse {
	return sessionResponse{
		Principal:       toPrincipalResponse(result.Principal),
		StaffProfile:    toStaffProfileResponse(result.StaffProfile),
		RoleCodes:       result.RoleCodes,
		PermissionCodes: result.PermissionCodes,
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

func toStaffProfileResponse(profile authsvc.StaffProfile) staffProfileResponse {
	return staffProfileResponse{
		StaffID:       profile.StaffID,
		Name:          profile.Name,
		Phone:         profile.Phone,
		Email:         profile.Email,
		Department:    profile.Department,
		JobTitle:      profile.JobTitle,
		ContactQRCode: profile.ContactQRCode,
	}
}
