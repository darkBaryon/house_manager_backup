package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/internal/middleware"
	authsvc "house-manager/internal/service/admin/auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type adminAuthEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestLoginReturnsStaffProfileAndPermissions(t *testing.T) {
	svc := &fakeAdminAuthService{
		loginResult: &authsvc.LoginResult{
			Token: "opaque-token",
			AuthSession: authsvc.AuthSession{
				Principal: session.Principal{
					PrincipalType:   session.PrincipalTypeStaff,
					PrincipalID:     "staff-id",
					Terminal:        session.TerminalAdmin,
					Phone:           "13800000000",
					RoleCodes:       []string{"super_admin"},
					PermissionCodes: []string{"staff.view", "staff.edit"},
				},
				StaffProfile: authsvc.StaffProfile{
					StaffID:       "staff-id",
					Name:          "管理员",
					Phone:         "13800000000",
					Email:         "admin@example.com",
					Department:    "运营",
					JobTitle:      "主管",
					ContactQRCode: "https://cdn.example.com/qr.png",
				},
				RoleCodes:       []string{"super_admin"},
				PermissionCodes: []string{"staff.view", "staff.edit"},
			},
		},
	}

	resp := performAdminAuthRequest(t, svc, "/api/v1/admin_auth/login", `{"phone":"13800000000","password":"secret123"}`, nil, true)

	envelope := assertAdminAuthResponse(t, resp, http.StatusOK, 0)
	if svc.loginCalls != 1 || svc.loginInput.Phone != "13800000000" || svc.loginInput.Password != "secret123" {
		t.Fatalf("unexpected login call: %#v", svc)
	}
	var data struct {
		Token     string `json:"token"`
		Principal struct {
			PrincipalType string   `json:"principal_type"`
			PrincipalID   string   `json:"principal_id"`
			Terminal      string   `json:"terminal"`
			RoleCodes     []string `json:"role_codes"`
		} `json:"principal"`
		StaffProfile struct {
			StaffID string `json:"staff_id"`
			Name    string `json:"name"`
		} `json:"staff_profile"`
		RoleCodes       []string `json:"role_codes"`
		PermissionCodes []string `json:"permission_codes"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Token != "opaque-token" || data.Principal.PrincipalType != session.PrincipalTypeStaff || data.Principal.PrincipalID != "staff-id" {
		t.Fatalf("unexpected data: %#v", data)
	}
	if data.StaffProfile.StaffID != "staff-id" || data.StaffProfile.Name != "管理员" {
		t.Fatalf("unexpected staff profile: %#v", data.StaffProfile)
	}
	if len(data.RoleCodes) != 1 || len(data.PermissionCodes) != 2 {
		t.Fatalf("unexpected role/permission codes: %#v", data)
	}
}

func TestLoginInvalidJSONReturnsChineseInvalidParam(t *testing.T) {
	resp := performAdminAuthRequest(t, &fakeAdminAuthService{}, "/api/v1/admin_auth/login", `{`, nil, true)

	envelope := assertAdminAuthResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("expected chinese error, got %q", envelope.Error)
	}
}

func TestLoginServiceError(t *testing.T) {
	svc := &fakeAdminAuthService{loginErr: errcode.Unauthorized.WithError(fmt.Errorf("手机号或密码错误，请重新输入"))}

	resp := performAdminAuthRequest(t, svc, "/api/v1/admin_auth/login", `{"phone":"13800000000","password":"bad"}`, nil, true)

	envelope := assertAdminAuthResponse(t, resp, http.StatusUnauthorized, errcode.Unauthorized.Code)
	if envelope.Error != "手机号或密码错误，请重新输入" {
		t.Fatalf("expected detailed error, got %q body=%s", envelope.Error, resp.Body.String())
	}
}

func TestSessionRequiresPrincipal(t *testing.T) {
	resp := performAdminAuthRequest(t, &fakeAdminAuthService{}, "/api/v1/admin_auth/session", `{}`, nil, false)

	envelope := assertAdminAuthResponse(t, resp, http.StatusUnauthorized, errcode.Unauthorized.Code)
	if envelope.Error != "未登录或登录已过期，请重新登录" {
		t.Fatalf("expected chinese unauthorized, got %q", envelope.Error)
	}
}

func TestSessionReturnsPrincipal(t *testing.T) {
	svc := &fakeAdminAuthService{
		sessionResult: &authsvc.AuthSession{
			Principal: session.Principal{
				PrincipalType:   session.PrincipalTypeStaff,
				PrincipalID:     "staff-id",
				Terminal:        session.TerminalAdmin,
				Phone:           "13800000000",
				RoleCodes:       []string{"super_admin"},
				PermissionCodes: []string{"staff.view"},
			},
			StaffProfile: authsvc.StaffProfile{
				StaffID: "staff-id",
				Name:    "管理员",
				Phone:   "13800000000",
			},
			RoleCodes:       []string{"super_admin"},
			PermissionCodes: []string{"staff.view"},
		},
	}
	principal := session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   "staff-id",
		Terminal:      session.TerminalAdmin,
		Phone:         "13800000000",
	}

	resp := performAdminAuthRequest(t, svc, "/api/v1/admin_auth/session", `{}`, func(c *gin.Context) {
		c.Set(middleware.ContextPrincipal, principal)
	}, false)

	envelope := assertAdminAuthResponse(t, resp, http.StatusOK, 0)
	if svc.sessionCalls != 1 || svc.sessionPrincipal.PrincipalID != "staff-id" {
		t.Fatalf("unexpected session call: %#v", svc)
	}
	if !bytes.Contains(envelope.Data, []byte("staff_profile")) {
		t.Fatalf("expected session response body, got %s", string(envelope.Data))
	}
}

func TestLogoutDeletesCurrentToken(t *testing.T) {
	svc := &fakeAdminAuthService{}

	resp := performAdminAuthRequest(t, svc, "/api/v1/admin_auth/logout", `{}`, func(c *gin.Context) {
		c.Set(middleware.ContextToken, "opaque-token")
	}, false)

	envelope := assertAdminAuthResponse(t, resp, http.StatusOK, 0)
	if svc.logoutCalls != 1 || svc.logoutToken != "opaque-token" {
		t.Fatalf("unexpected logout call: %#v", svc)
	}
	if !bytes.Contains(envelope.Data, []byte(`"success":true`)) {
		t.Fatalf("expected logout success body, got %s", string(envelope.Data))
	}
}

func performAdminAuthRequest(t *testing.T, svc *fakeAdminAuthService, path string, body string, before gin.HandlerFunc, public bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	if before != nil {
		group.Use(before)
	}
	if public {
		(&PublicHandler{service: svc}).RegisterRoutes(group)
	} else {
		newHandler(svc).RegisterRoutes(group)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertAdminAuthResponse(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) adminAuthEnvelope {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var envelope adminAuthEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, envelope.Code, resp.Body.String())
	}
	return envelope
}

type fakeAdminAuthService struct {
	loginCalls  int
	loginInput  authsvc.LoginInput
	loginResult *authsvc.LoginResult
	loginErr    error

	sessionCalls     int
	sessionPrincipal session.Principal
	sessionResult    *authsvc.AuthSession
	sessionErr       error

	logoutCalls int
	logoutToken string
	logoutErr   error
}

func (f *fakeAdminAuthService) Login(ctx context.Context, input authsvc.LoginInput) (*authsvc.LoginResult, error) {
	f.loginCalls++
	f.loginInput = input
	return f.loginResult, f.loginErr
}

func (f *fakeAdminAuthService) Session(ctx context.Context, principal session.Principal) (*authsvc.AuthSession, error) {
	f.sessionCalls++
	f.sessionPrincipal = principal
	return f.sessionResult, f.sessionErr
}

func (f *fakeAdminAuthService) Logout(ctx context.Context, token string) error {
	f.logoutCalls++
	f.logoutToken = token
	return f.logoutErr
}
