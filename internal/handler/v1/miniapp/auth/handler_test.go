package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type authEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestWechatLoginBindsRequestAndReturnsToken(t *testing.T) {
	svc := &fakeAuthService{loginToken: "opaque-token"}

	resp := performAuthRequest(t, svc, "/api/v1/auth/wechat_login", `{"code":"wx-code"}`, nil)

	envelope := assertAuthResponse(t, resp, http.StatusOK, 0)
	if svc.loginCalls != 1 || svc.loginCode != "wx-code" {
		t.Fatalf("unexpected login call: calls=%d code=%q", svc.loginCalls, svc.loginCode)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Token != "opaque-token" {
		t.Fatalf("unexpected token: %q", data.Token)
	}
}

func TestWechatRegisterBindsRequestAndReturnsToken(t *testing.T) {
	svc := &fakeAuthService{registerToken: "opaque-register-token"}

	resp := performAuthRequest(t, svc, "/api/v1/auth/wechat_register", `{"code":"wx-code","phone_code":"phone-code"}`, nil)

	assertAuthResponse(t, resp, http.StatusOK, 0)
	if svc.registerCalls != 1 || svc.registerCode != "wx-code" || svc.registerPhoneCode != "phone-code" {
		t.Fatalf("unexpected register call: %#v", svc)
	}
}

func TestWechatLoginInvalidJSONReturnsInvalidParam(t *testing.T) {
	svc := &fakeAuthService{}

	resp := performAuthRequest(t, svc, "/api/v1/auth/wechat_login", `{`, nil)

	assertAuthResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.loginCalls != 0 {
		t.Fatalf("expected service not to be called, got %d", svc.loginCalls)
	}
}

func TestWechatLoginServiceErrcode(t *testing.T) {
	svc := &fakeAuthService{loginErr: errcode.Unauthorized.WithError(fmt.Errorf("not registered"))}

	resp := performAuthRequest(t, svc, "/api/v1/auth/wechat_login", `{"code":"wx-code"}`, nil)

	assertAuthResponse(t, resp, http.StatusUnauthorized, errcode.Unauthorized.Code)
}

func TestSessionRequiresUserID(t *testing.T) {
	resp := performAuthRequest(t, &fakeAuthService{}, "/api/v1/auth/session", `{}`, nil)

	assertAuthResponse(t, resp, http.StatusUnauthorized, errcode.Unauthorized.Code)
}

func TestSessionReturnsUserID(t *testing.T) {
	resp := performAuthRequest(t, &fakeAuthService{}, "/api/v1/auth/session", `{}`, func(c *gin.Context) {
		c.Set("userId", "user-id")
	})

	envelope := assertAuthResponse(t, resp, http.StatusOK, 0)
	var data struct {
		UserID string `json:"userId"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.UserID != "user-id" {
		t.Fatalf("unexpected userId: %q", data.UserID)
	}
}

func performAuthRequest(t *testing.T, svc *fakeAuthService, path string, body string, before gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	if before != nil {
		group.Use(before)
	}
	newAuthHandler(svc).RegisterRoutes(group)
	NewSessionHandler().RegisterRoutes(group)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertAuthResponse(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) authEnvelope {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var envelope authEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, envelope.Code, resp.Body.String())
	}
	return envelope
}

type fakeAuthService struct {
	loginCalls int
	loginCode  string
	loginToken string
	loginErr   error

	registerCalls     int
	registerCode      string
	registerPhoneCode string
	registerToken     string
	registerErr       error
}

func (f *fakeAuthService) WechatLogin(ctx context.Context, code, loginIP string) (string, error) {
	f.loginCalls++
	f.loginCode = code
	return f.loginToken, f.loginErr
}

func (f *fakeAuthService) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (string, error) {
	f.registerCalls++
	f.registerCode = code
	f.registerPhoneCode = phoneCode
	return f.registerToken, f.registerErr
}
