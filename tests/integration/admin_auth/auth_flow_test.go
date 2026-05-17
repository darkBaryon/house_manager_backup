package adminauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type envelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestIntegrationAdminAuthLoginSessionLogoutFlow(t *testing.T) {
	f := newIntegrationFixture(t)

	loginResp := performRequest(t, f, "/api/v1/admin_auth/login", `{"phone":"`+f.phone+`","password":"`+f.password+`"}`, "")
	loginBody := assertEnvelope(t, loginResp, http.StatusOK, 0)
	var loginData struct {
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
			Phone   string `json:"phone"`
		} `json:"staff_profile"`
		PermissionCodes []string `json:"permission_codes"`
	}
	if err := json.Unmarshal(loginBody.Data, &loginData); err != nil {
		t.Fatalf("decode login data: %v", err)
	}
	if loginData.Token == "" {
		t.Fatal("expected token")
	}
	if loginData.Principal.PrincipalType != "staff" || loginData.Principal.Terminal != "admin" {
		t.Fatalf("unexpected principal: %#v", loginData.Principal)
	}
	if loginData.StaffProfile.StaffID != f.staffID.Hex() || loginData.StaffProfile.Phone != f.phone {
		t.Fatalf("unexpected staff profile: %#v", loginData.StaffProfile)
	}

	sessionResp := performRequest(t, f, "/api/v1/admin_auth/session", `{}`, loginData.Token)
	sessionBody := assertEnvelope(t, sessionResp, http.StatusOK, 0)
	var sessionData struct {
		Principal struct {
			PrincipalID string `json:"principal_id"`
			Terminal    string `json:"terminal"`
		} `json:"principal"`
		RoleCodes       []string `json:"role_codes"`
		PermissionCodes []string `json:"permission_codes"`
	}
	if err := json.Unmarshal(sessionBody.Data, &sessionData); err != nil {
		t.Fatalf("decode session data: %v", err)
	}
	if sessionData.Principal.PrincipalID != f.staffID.Hex() || sessionData.Principal.Terminal != "admin" {
		t.Fatalf("unexpected session principal: %#v", sessionData.Principal)
	}
	if len(sessionData.RoleCodes) != 1 || sessionData.RoleCodes[0] != f.roleCode {
		t.Fatalf("unexpected role codes: %#v", sessionData.RoleCodes)
	}
	if len(sessionData.PermissionCodes) != 1 || sessionData.PermissionCodes[0] != f.permissionCode {
		t.Fatalf("unexpected permission codes: %#v", sessionData.PermissionCodes)
	}

	logoutResp := performRequest(t, f, "/api/v1/admin_auth/logout", `{}`, loginData.Token)
	logoutBody := assertEnvelope(t, logoutResp, http.StatusOK, 0)
	if !bytes.Contains(logoutBody.Data, []byte(`"success":true`)) {
		t.Fatalf("unexpected logout body: %s", string(logoutBody.Data))
	}

	sessionAfterLogoutResp := performRequest(t, f, "/api/v1/admin_auth/session", `{}`, loginData.Token)
	sessionAfterLogoutBody := assertEnvelope(t, sessionAfterLogoutResp, http.StatusUnauthorized, 10002)
	if sessionAfterLogoutBody.Error != "未登录或登录已过期，请重新登录" {
		t.Fatalf("expected chinese unauthorized message, got %q", sessionAfterLogoutBody.Error)
	}
}

func performRequest(t *testing.T, f *integrationFixture, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp := httptest.NewRecorder()
	f.router.ServeHTTP(resp, req)
	return resp
}

func assertEnvelope(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) envelope {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if body.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, body.Code, resp.Body.String())
	}
	return body
}
