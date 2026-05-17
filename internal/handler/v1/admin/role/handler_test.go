package role

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/internal/middleware"
	rolesvc "house-manager/internal/service/admin/role"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type roleEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestCreateReturnsRole(t *testing.T) {
	svc := &fakeRoleService{
		createResult: &rolesvc.CreateResult{
			Role: rolesvc.RoleDetail{RoleSummary: rolesvc.RoleSummary{
				RoleID:          "role-id",
				RoleName:        "运营管理员",
				RoleCode:        "ops_admin",
				PermissionCodes: []string{"staff.view"},
				PermissionNames: []string{"查看员工"},
				Status:          1,
			}},
		},
	}
	resp := performRoleRequest(t, svc, "/api/v1/role/create", `{"role_name":"运营管理员","role_code":"ops_admin","permission_codes":["staff.view"]}`, rolePrincipal([]string{"role.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.createInput.OperatorStaffID != "staff-id" || svc.createInput.RoleCode != "ops_admin" {
		t.Fatalf("unexpected input: %#v", svc.createInput)
	}
	var envelope roleEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data createResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Role.RoleID != "role-id" || data.Role.RoleCode != "ops_admin" {
		t.Fatalf("unexpected role response: %#v", data.Role)
	}
}

func TestCreateRequiresPermission(t *testing.T) {
	resp := performRoleRequest(t, &fakeRoleService{}, "/api/v1/role/create", `{}`, rolePrincipal([]string{"role.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope roleEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权创建角色" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListReturnsRoles(t *testing.T) {
	svc := &fakeRoleService{listResult: &rolesvc.ListResult{
		List: []rolesvc.ListItem{{RoleSummary: rolesvc.RoleSummary{RoleID: "role-id", RoleName: "运营管理员", RoleCode: "ops_admin"}}},
		Page: 2, PageSize: 10, Total: 1,
	}}
	resp := performRoleRequest(t, svc, "/api/v1/role/list", `{"keyword":"运营","page":2,"page_size":10}`, rolePrincipal([]string{"role.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.listInput.Keyword != "运营" || svc.listInput.Page != 2 {
		t.Fatalf("unexpected list input: %#v", svc.listInput)
	}
}

func TestDetailRejectsInvalidJSON(t *testing.T) {
	resp := performRoleRequest(t, &fakeRoleService{}, "/api/v1/role/detail", `{bad`, rolePrincipal([]string{"role.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestUpdatePropagatesServiceError(t *testing.T) {
	svc := &fakeRoleService{updateErr: errcode.NotFound.WithError(fmt.Errorf("角色不存在或已停用"))}
	resp := performRoleRequest(t, svc, "/api/v1/role/update", `{"role_id":"role-id","role_name":"运营"}`, rolePrincipal([]string{"role.edit"}))
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func performRoleRequest(t *testing.T, svc Service, path string, body string, principal session.Principal) *httptest.ResponseRecorder {
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

func rolePrincipal(permissionCodes []string) session.Principal {
	return session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        session.TerminalAdmin,
		Phone:           "13800000000",
		PermissionCodes: permissionCodes,
	}
}

type fakeRoleService struct {
	listInput    rolesvc.ListInput
	listResult   *rolesvc.ListResult
	listErr      error
	detailInput  rolesvc.DetailInput
	detailResult *rolesvc.DetailResult
	detailErr    error
	createInput  rolesvc.CreateInput
	createResult *rolesvc.CreateResult
	createErr    error
	updateInput  rolesvc.UpdateInput
	updateResult *rolesvc.UpdateResult
	updateErr    error
}

func (f *fakeRoleService) List(ctx context.Context, input rolesvc.ListInput) (*rolesvc.ListResult, error) {
	f.listInput = input
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeRoleService) Detail(ctx context.Context, input rolesvc.DetailInput) (*rolesvc.DetailResult, error) {
	f.detailInput = input
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detailResult, nil
}

func (f *fakeRoleService) Create(ctx context.Context, input rolesvc.CreateInput) (*rolesvc.CreateResult, error) {
	f.createInput = input
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeRoleService) Update(ctx context.Context, input rolesvc.UpdateInput) (*rolesvc.UpdateResult, error) {
	f.updateInput = input
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateResult, nil
}
