package staff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/internal/middleware"
	staffsvc "house-manager/internal/service/admin/staff"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type adminStaffEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestCreateReturnsCreatedStaff(t *testing.T) {
	svc := &fakeAdminStaffService{
		createResult: &staffsvc.CreateResult{
			Staff: staffsvc.StaffSummary{
				StaffID:       "staff-id",
				Name:          "运营一号",
				Phone:         "13800000000",
				Email:         "ops@example.com",
				Department:    "运营部",
				JobTitle:      "运营",
				ContactQRCode: "https://example.com/qr.png",
				Roles: []staffsvc.RoleSummary{{
					RoleID:   "role-id",
					RoleCode: "ops_admin",
					RoleName: "运营管理员",
				}},
				RoleNames: []string{"运营管理员"},
				Status:    1,
				CreatedAt: 111,
				UpdatedAt: 222,
			},
		},
	}

	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/create", `{"name":"运营一号","phone":"13800000000","password":"secret123","role_ids":["role-id"]}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.createInput.OperatorStaffID != "staff-id" || svc.createInput.Phone != "13800000000" {
		t.Fatalf("unexpected input: %#v", svc.createInput)
	}

	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data struct {
		Staff staffSummaryResponse `json:"staff"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Staff.StaffID != "staff-id" || len(data.Staff.Roles) != 1 || data.Staff.Roles[0].RoleCode != "ops_admin" {
		t.Fatalf("unexpected staff response: %#v", data.Staff)
	}
}

func TestCreateRejectsInvalidJSON(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/create", `{bad`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestCreateRequiresPermission(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/create", `{"name":"运营一号"}`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权创建员工" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestCreatePropagatesServiceError(t *testing.T) {
	svc := &fakeAdminStaffService{
		createErr: errcode.AlreadyExists.WithError(fmt.Errorf("员工手机号已存在，请更换后重试")),
	}
	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/create", `{"name":"运营一号","phone":"13800000000","password":"secret123"}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusConflict {
		t.Fatalf("expected conflict, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "员工手机号已存在，请更换后重试" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListReturnsStaffList(t *testing.T) {
	status := 1
	svc := &fakeAdminStaffService{
		listResult: &staffsvc.ListResult{
			List: []staffsvc.StaffListItem{{
				StaffSummary: staffsvc.StaffSummary{
					StaffID: "staff-id",
					Name:    "运营一号",
					Phone:   "13800000000",
					Roles: []staffsvc.RoleSummary{{
						RoleID:   "role-id",
						RoleCode: "ops_admin",
						RoleName: "运营管理员",
					}},
					RoleNames: []string{"运营管理员"},
					Status:    1,
					CreatedAt: 111,
					UpdatedAt: 222,
				},
				LastLoginAt: 333,
				LastLoginIP: "127.0.0.1",
			}},
			Page:     2,
			PageSize: 10,
			Total:    1,
		},
	}

	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/list", `{"keyword":"运营","status":1,"page":2,"page_size":10}`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.listInput.Keyword != "运营" || svc.listInput.Status == nil || *svc.listInput.Status != status || svc.listInput.Page != 2 {
		t.Fatalf("unexpected list input: %#v", svc.listInput)
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data listResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(data.List) != 1 || data.List[0].LastLoginAt != 333 || data.Total != 1 {
		t.Fatalf("unexpected list response: %#v", data)
	}
}

func TestListRequiresPermission(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/list", `{}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看员工列表" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListAllowsEmptyBody(t *testing.T) {
	svc := &fakeAdminStaffService{listResult: &staffsvc.ListResult{List: []staffsvc.StaffListItem{}, Page: 1, PageSize: 20}}
	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/list", "", adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestDetailReturnsStaffDetail(t *testing.T) {
	svc := &fakeAdminStaffService{
		detailResult: &staffsvc.DetailResult{
			Staff: staffsvc.StaffDetail{
				StaffSummary: staffsvc.StaffSummary{
					StaffID: "staff-id",
					Name:    "运营一号",
					Phone:   "13800000000",
					Roles: []staffsvc.RoleSummary{{
						RoleID:   "role-id",
						RoleCode: "ops_admin",
						RoleName: "运营管理员",
					}},
					RoleNames: []string{"运营管理员"},
					Status:    1,
					CreatedAt: 111,
					UpdatedAt: 222,
				},
				CreatedByStaffID:  "creator-id",
				PasswordUpdatedAt: 333,
				LastLoginAt:       444,
				LastLoginIP:       "127.0.0.1",
			},
		},
	}

	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/detail", `{"staff_id":"staff-id"}`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.detailInput.StaffID != "staff-id" {
		t.Fatalf("unexpected detail input: %#v", svc.detailInput)
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data detailResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Staff.StaffID != "staff-id" || data.Staff.PasswordUpdatedAt != 333 || data.Staff.LastLoginAt != 444 {
		t.Fatalf("unexpected detail response: %#v", data.Staff)
	}
}

func TestDetailRequiresPermission(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/detail", `{"staff_id":"staff-id"}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看员工详情" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestDetailRejectsInvalidJSON(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/detail", `{bad`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestUpdateReturnsStaffDetail(t *testing.T) {
	svc := &fakeAdminStaffService{
		updateResult: &staffsvc.UpdateResult{
			Staff: staffsvc.StaffDetail{
				StaffSummary: staffsvc.StaffSummary{
					StaffID: "staff-id",
					Name:    "运营二号",
					Phone:   "13800000000",
					Roles: []staffsvc.RoleSummary{{
						RoleID:   "role-id",
						RoleCode: "ops_admin",
						RoleName: "运营管理员",
					}},
					RoleNames: []string{"运营管理员"},
					Status:    -1,
					CreatedAt: 111,
					UpdatedAt: 222,
				},
				PasswordUpdatedAt: 333,
			},
		},
	}

	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/update", `{"staff_id":"staff-id","name":"运营二号","role_ids":["role-id"],"status":-1}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.updateInput.OperatorStaffID != "staff-id" || svc.updateInput.StaffID != "staff-id" || svc.updateInput.Name == nil || *svc.updateInput.Name != "运营二号" {
		t.Fatalf("unexpected update input: %#v", svc.updateInput)
	}
	if svc.updateInput.RoleIDs == nil || len(*svc.updateInput.RoleIDs) != 1 || svc.updateInput.Status == nil || *svc.updateInput.Status != -1 {
		t.Fatalf("unexpected update role/status input: %#v", svc.updateInput)
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data updateResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Staff.StaffID != "staff-id" || data.Staff.Status != -1 || data.Staff.PasswordUpdatedAt != 333 {
		t.Fatalf("unexpected update response: %#v", data.Staff)
	}
}

func TestUpdateRequiresPermission(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/update", `{"staff_id":"staff-id","name":"运营二号"}`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权编辑员工" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestUpdateRejectsInvalidJSON(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/update", `{bad`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestDisableReturnsSuccess(t *testing.T) {
	svc := &fakeAdminStaffService{
		disableResult: &staffsvc.DisableResult{Success: true},
	}

	resp := performAdminStaffRequest(t, svc, "/api/v1/staff/disable", `{"staff_id":"staff-id"}`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.disableInput.StaffID != "staff-id" {
		t.Fatalf("unexpected disable input: %#v", svc.disableInput)
	}
	if svc.disableInput.OperatorStaffID != "staff-id" {
		t.Fatalf("unexpected disable operator: %#v", svc.disableInput)
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data disableResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if !data.Success {
		t.Fatalf("expected success response")
	}
}

func TestDisableRequiresPermission(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/disable", `{"staff_id":"staff-id"}`, adminStaffPrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权禁用员工" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestDisableRejectsInvalidJSON(t *testing.T) {
	resp := performAdminStaffRequest(t, &fakeAdminStaffService{}, "/api/v1/staff/disable", `{bad`, adminStaffPrincipal([]string{"staff.edit"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope adminStaffEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func performAdminStaffRequest(t *testing.T, svc Service, path string, body string, principal session.Principal) *httptest.ResponseRecorder {
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

func adminStaffPrincipal(permissionCodes []string) session.Principal {
	return session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        session.TerminalAdmin,
		Phone:           "13800000000",
		RoleCodes:       []string{"super_admin"},
		PermissionCodes: permissionCodes,
	}
}

type fakeAdminStaffService struct {
	createInput   staffsvc.CreateInput
	createResult  *staffsvc.CreateResult
	createErr     error
	listInput     staffsvc.ListInput
	listResult    *staffsvc.ListResult
	listErr       error
	detailInput   staffsvc.DetailInput
	detailResult  *staffsvc.DetailResult
	detailErr     error
	updateInput   staffsvc.UpdateInput
	updateResult  *staffsvc.UpdateResult
	updateErr     error
	disableInput  staffsvc.DisableInput
	disableResult *staffsvc.DisableResult
	disableErr    error
}

func (f *fakeAdminStaffService) Create(ctx context.Context, input staffsvc.CreateInput) (*staffsvc.CreateResult, error) {
	f.createInput = input
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeAdminStaffService) List(ctx context.Context, input staffsvc.ListInput) (*staffsvc.ListResult, error) {
	f.listInput = input
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeAdminStaffService) Detail(ctx context.Context, input staffsvc.DetailInput) (*staffsvc.DetailResult, error) {
	f.detailInput = input
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detailResult, nil
}

func (f *fakeAdminStaffService) Update(ctx context.Context, input staffsvc.UpdateInput) (*staffsvc.UpdateResult, error) {
	f.updateInput = input
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeAdminStaffService) Disable(ctx context.Context, input staffsvc.DisableInput) (*staffsvc.DisableResult, error) {
	f.disableInput = input
	if f.disableErr != nil {
		return nil, f.disableErr
	}
	return f.disableResult, nil
}
