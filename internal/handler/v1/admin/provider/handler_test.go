package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/internal/middleware"
	providersvc "house-manager/internal/service/admin/provider"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type providerEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestCreateReturnsProvider(t *testing.T) {
	svc := &fakeProviderService{
		createResult: &providersvc.CreateResult{
			Provider: providersvc.ProviderSummary{
				ProviderID:        "provider-id",
				Phone:             "13800000000",
				Status:            1,
				CreatedByStaffID:  "staff-id",
				CreatedAt:         111,
				UpdatedAt:         222,
				PasswordUpdatedAt: 333,
			},
		},
	}

	resp := performProviderRequest(t, svc, "/api/v1/provider/create", `{"phone":"13800000000","password":"secret123"}`, providerPrincipal([]string{"provider.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.createInput.OperatorStaffID != "staff-id" || svc.createInput.Phone != "13800000000" {
		t.Fatalf("unexpected input: %#v", svc.createInput)
	}
	var envelope providerEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data createResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Provider.ProviderID != "provider-id" || data.Provider.PasswordUpdatedAt != 333 {
		t.Fatalf("unexpected provider response: %#v", data.Provider)
	}
}

func TestCreateRequiresPermission(t *testing.T) {
	resp := performProviderRequest(t, &fakeProviderService{}, "/api/v1/provider/create", `{}`, providerPrincipal([]string{"provider.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope providerEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权创建发房方" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListReturnsProvidersAndAllowsEmptyBody(t *testing.T) {
	svc := &fakeProviderService{
		listResult: &providersvc.ListResult{
			List: []providersvc.ListItem{{
				ProviderSummary: providersvc.ProviderSummary{
					ProviderID:        "provider-id",
					Phone:             "13800000000",
					Status:            1,
					PasswordUpdatedAt: 333,
					LastLoginAt:       444,
					LastLoginIP:       "127.0.0.1",
				},
			}},
			Page:     1,
			PageSize: 20,
			Total:    1,
		},
	}

	resp := performProviderRequest(t, svc, "/api/v1/provider/list", "", providerPrincipal([]string{"provider.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope providerEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data listResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(data.List) != 1 || data.List[0].LastLoginAt != 444 || data.Total != 1 {
		t.Fatalf("unexpected list response: %#v", data)
	}
}

func TestDetailRequiresPermission(t *testing.T) {
	resp := performProviderRequest(t, &fakeProviderService{}, "/api/v1/provider/detail", `{"provider_id":"provider-id"}`, providerPrincipal([]string{"provider.edit"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope providerEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看发房方详情" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestUpdatePropagatesServiceError(t *testing.T) {
	svc := &fakeProviderService{updateErr: errcode.NotFound.WithError(fmt.Errorf("发房方不存在或已删除"))}
	resp := performProviderRequest(t, svc, "/api/v1/provider/update", `{"provider_id":"provider-id","phone":"13800000001"}`, providerPrincipal([]string{"provider.edit"}))
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestDisableCallsService(t *testing.T) {
	svc := &fakeProviderService{disableResult: &providersvc.DisableResult{Success: true}}
	resp := performProviderRequest(t, svc, "/api/v1/provider/disable", `{"provider_id":"provider-id"}`, providerPrincipal([]string{"provider.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.disableInput.OperatorStaffID != "staff-id" || svc.disableInput.ProviderID != "provider-id" {
		t.Fatalf("unexpected input: %#v", svc.disableInput)
	}
}

func TestDetailRejectsInvalidJSON(t *testing.T) {
	resp := performProviderRequest(t, &fakeProviderService{}, "/api/v1/provider/detail", `{bad`, providerPrincipal([]string{"provider.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope providerEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func performProviderRequest(t *testing.T, svc Service, path string, body string, principal session.Principal) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextPrincipal, principal)
		c.Request = c.Request.WithContext(session.ContextWithPrincipal(c.Request.Context(), principal))
		c.Next()
	})
	handler := newHandler(svc)
	group := router.Group("/api/v1")
	handler.RegisterRoutes(group)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func providerPrincipal(permissionCodes []string) session.Principal {
	return session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        session.TerminalAdmin,
		Phone:           "13800000000",
		PermissionCodes: permissionCodes,
	}
}

type fakeProviderService struct {
	createInput   providersvc.CreateInput
	createResult  *providersvc.CreateResult
	createErr     error
	listInput     providersvc.ListInput
	listResult    *providersvc.ListResult
	listErr       error
	detailInput   providersvc.DetailInput
	detailResult  *providersvc.DetailResult
	detailErr     error
	updateInput   providersvc.UpdateInput
	updateResult  *providersvc.UpdateResult
	updateErr     error
	disableInput  providersvc.DisableInput
	disableResult *providersvc.DisableResult
	disableErr    error
}

func (f *fakeProviderService) Create(ctx context.Context, input providersvc.CreateInput) (*providersvc.CreateResult, error) {
	f.createInput = input
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeProviderService) List(ctx context.Context, input providersvc.ListInput) (*providersvc.ListResult, error) {
	f.listInput = input
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeProviderService) Detail(ctx context.Context, input providersvc.DetailInput) (*providersvc.DetailResult, error) {
	f.detailInput = input
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detailResult, nil
}

func (f *fakeProviderService) Update(ctx context.Context, input providersvc.UpdateInput) (*providersvc.UpdateResult, error) {
	f.updateInput = input
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeProviderService) Disable(ctx context.Context, input providersvc.DisableInput) (*providersvc.DisableResult, error) {
	f.disableInput = input
	if f.disableErr != nil {
		return nil, f.disableErr
	}
	return f.disableResult, nil
}
