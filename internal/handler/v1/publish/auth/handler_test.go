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
	authsvc "house-manager/internal/service/publish/auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type publishAuthEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestLoginBindsPhoneAndReturnsSnakeCasePrincipal(t *testing.T) {
	svc := &fakePublishAuthService{
		loginResult: &authsvc.LoginResult{
			Token: "opaque-token",
			AuthSession: authsvc.AuthSession{
				Principal: session.Principal{
					PrincipalType: session.PrincipalTypeLandlord,
					PrincipalID:   "landlord-id",
					Terminal:      session.TerminalPublish,
					Phone:         "13800000000",
				},
				Subject: authsvc.Subject{ID: "landlord-id", Name: "房东A", Phone: "13800000000"},
			},
		},
	}

	resp := performPublishAuthRequest(t, svc, "/api/v1/publish_auth/login", `{"phone":"13800000000","password":"secret123"}`, nil, true)

	envelope := assertPublishAuthResponse(t, resp, http.StatusOK, 0)
	if svc.loginCalls != 1 || svc.loginInput.Phone != "13800000000" || svc.loginInput.Password != "secret123" {
		t.Fatalf("unexpected login call: %#v", svc)
	}
	var data struct {
		Token     string `json:"token"`
		Principal struct {
			PrincipalType   string   `json:"principal_type"`
			PrincipalID     string   `json:"principal_id"`
			Terminal        string   `json:"terminal"`
			RoleCodes       []string `json:"role_codes"`
			PermissionCodes []string `json:"permission_codes"`
		} `json:"principal"`
		Subject struct {
			ID string `json:"id"`
		} `json:"subject"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Token != "opaque-token" || data.Principal.PrincipalType != session.PrincipalTypeLandlord || data.Principal.PrincipalID != "landlord-id" {
		t.Fatalf("unexpected data: %#v", data)
	}
}

func TestLoginInvalidJSONReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishAuthService{}

	resp := performPublishAuthRequest(t, svc, "/api/v1/publish_auth/login", `{`, nil, true)

	assertPublishAuthResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.loginCalls != 0 {
		t.Fatalf("expected service not called")
	}
}

func TestLoginServiceError(t *testing.T) {
	svc := &fakePublishAuthService{loginErr: errcode.Forbidden.WithError(fmt.Errorf("仅支持本地环境登录"))}

	resp := performPublishAuthRequest(t, svc, "/api/v1/publish_auth/login", `{"phone":"13800000000","password":"secret123"}`, nil, true)

	envelope := assertPublishAuthResponse(t, resp, http.StatusForbidden, errcode.Forbidden.Code)
	if envelope.Error != "仅支持本地环境登录" {
		t.Fatalf("expected detailed error, got %q body=%s", envelope.Error, resp.Body.String())
	}
}

func TestSessionRequiresPrincipal(t *testing.T) {
	resp := performPublishAuthRequest(t, &fakePublishAuthService{}, "/api/v1/publish_auth/session", `{}`, nil, false)

	assertPublishAuthResponse(t, resp, http.StatusUnauthorized, errcode.Unauthorized.Code)
}

func TestSessionReturnsPrincipal(t *testing.T) {
	svc := &fakePublishAuthService{
		sessionResult: &authsvc.AuthSession{
			Principal: session.Principal{
				PrincipalType: session.PrincipalTypeLandlord,
				PrincipalID:   "landlord-id",
				Terminal:      session.TerminalPublish,
				Phone:         "13800000000",
			},
			Subject: authsvc.Subject{ID: "landlord-id", Name: "房东A", Phone: "13800000000"},
		},
	}
	principal := session.Principal{PrincipalType: session.PrincipalTypeLandlord, PrincipalID: "landlord-id", Terminal: session.TerminalPublish, Phone: "13800000000"}

	resp := performPublishAuthRequest(t, svc, "/api/v1/publish_auth/session", `{}`, func(c *gin.Context) {
		c.Set(middleware.ContextPrincipal, principal)
	}, false)

	envelope := assertPublishAuthResponse(t, resp, http.StatusOK, 0)
	if svc.sessionCalls != 1 || svc.sessionPrincipal.PrincipalID != "landlord-id" {
		t.Fatalf("unexpected session call: %#v", svc)
	}
	if !bytes.Contains(envelope.Data, []byte("principal_type")) {
		t.Fatalf("expected snake_case principal, got %s", string(envelope.Data))
	}
}

func TestLogoutDeletesCurrentToken(t *testing.T) {
	svc := &fakePublishAuthService{}

	resp := performPublishAuthRequest(t, svc, "/api/v1/publish_auth/logout", `{}`, func(c *gin.Context) {
		c.Set(middleware.ContextToken, "opaque-token")
	}, false)

	assertPublishAuthResponse(t, resp, http.StatusOK, 0)
	if svc.logoutCalls != 1 || svc.logoutToken != "opaque-token" {
		t.Fatalf("unexpected logout call: %#v", svc)
	}
}

func performPublishAuthRequest(t *testing.T, svc *fakePublishAuthService, path string, body string, before gin.HandlerFunc, public bool) *httptest.ResponseRecorder {
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

func assertPublishAuthResponse(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) publishAuthEnvelope {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var envelope publishAuthEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, envelope.Code, resp.Body.String())
	}
	return envelope
}

type fakePublishAuthService struct {
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

func (f *fakePublishAuthService) Login(ctx context.Context, input authsvc.LoginInput) (*authsvc.LoginResult, error) {
	f.loginCalls++
	f.loginInput = input
	return f.loginResult, f.loginErr
}

func (f *fakePublishAuthService) Session(ctx context.Context, principal session.Principal) (*authsvc.AuthSession, error) {
	f.sessionCalls++
	f.sessionPrincipal = principal
	return f.sessionResult, f.sessionErr
}

func (f *fakePublishAuthService) Logout(ctx context.Context, token string) error {
	f.logoutCalls++
	f.logoutToken = token
	return f.logoutErr
}
