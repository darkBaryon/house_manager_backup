package house

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"house-manager/internal/middleware"
	housesvc "house-manager/internal/service/admin/house"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type houseEnvelope struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestListReturnsHousesAndAllowsEmptyBody(t *testing.T) {
	svc := &fakeHouseService{
		listResult: &housesvc.ListResult{
			List: []housesvc.ListItem{{
				HouseSummary: housesvc.HouseSummary{
					ListingID:     "listing-id",
					AssetMode:     "centralized",
					ProviderID:    "provider-id",
					ProviderPhone: "13800000000",
					Title:         "整租一居室",
					ListingStatus: 3,
					AuditStatus:   1,
					IsOnline:      1,
					UpdatedAt:     123,
				},
			}},
			Page:     1,
			PageSize: 20,
			Total:    1,
		},
	}

	resp := performHouseRequest(t, svc, "/api/v1/house/list", "", housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data listResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(data.List) != 1 || data.List[0].ListingID != "listing-id" || data.Total != 1 {
		t.Fatalf("unexpected list response: %#v", data)
	}
}

func TestListRequiresPermission(t *testing.T) {
	resp := performHouseRequest(t, &fakeHouseService{}, "/api/v1/house/list", `{}`, housePrincipal([]string{"provider.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看房源列表" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListPropagatesServiceError(t *testing.T) {
	svc := &fakeHouseService{listErr: errcode.InvalidParam.WithError(fmt.Errorf("房源筛选参数不正确"))}
	resp := performHouseRequest(t, svc, "/api/v1/house/list", `{"room_status":99}`, housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestDetailReturnsHouse(t *testing.T) {
	svc := &fakeHouseService{
		detailResult: &housesvc.DetailResult{
			House: housesvc.HouseSummary{
				ListingID:     "listing-id",
				SourceType:    "centralized_room",
				SourceID:      "source-id",
				Title:         "房源详情",
				ListingStatus: 3,
			},
		},
	}

	resp := performHouseRequest(t, svc, "/api/v1/house/detail", `{"listing_id":"listing-id"}`, housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.detailInput.ListingID != "listing-id" {
		t.Fatalf("unexpected detail input: %#v", svc.detailInput)
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data detailResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.House.SourceType != "centralized_room" || data.House.Title != "房源详情" {
		t.Fatalf("unexpected detail response: %#v", data)
	}
}

func TestDetailRequiresPermission(t *testing.T) {
	resp := performHouseRequest(t, &fakeHouseService{}, "/api/v1/house/detail", `{"listing_id":"listing-id"}`, housePrincipal([]string{"staff.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看房源详情" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestDetailRejectsInvalidJSON(t *testing.T) {
	resp := performHouseRequest(t, &fakeHouseService{}, "/api/v1/house/detail", `{bad`, housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "请求参数格式不正确" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func performHouseRequest(t *testing.T, svc Service, path string, body string, principal session.Principal) *httptest.ResponseRecorder {
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

func housePrincipal(permissionCodes []string) session.Principal {
	return session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        session.TerminalAdmin,
		Phone:           "13800000000",
		PermissionCodes: permissionCodes,
	}
}

type fakeHouseService struct {
	listInput    housesvc.ListInput
	listResult   *housesvc.ListResult
	listErr      error
	detailInput  housesvc.DetailInput
	detailResult *housesvc.DetailResult
	detailErr    error
}

func (f *fakeHouseService) List(ctx context.Context, input housesvc.ListInput) (*housesvc.ListResult, error) {
	f.listInput = input
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeHouseService) Detail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error) {
	f.detailInput = input
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detailResult, nil
}
