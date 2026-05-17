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

func TestListRootsReturnsRootList(t *testing.T) {
	svc := &fakeHouseService{
		rootListResult: &housesvc.RootListResult{
			List: []housesvc.RootListItem{{
				RootSummary: housesvc.RootSummary{
					RootID:    "root-id",
					RootName:  "星海公寓",
					RootType:  "centralized_project",
					RoomCount: 20,
				},
			}},
			Page:     1,
			PageSize: 20,
			Total:    1,
		},
	}

	resp := performHouseRequest(t, svc, "/api/v1/house_root/list", "", housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data rootListResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(data.List) != 1 || data.List[0].RootID != "root-id" || data.List[0].RootName != "星海公寓" {
		t.Fatalf("unexpected root list response: %#v", data)
	}
}

func TestListBuildingsRequiresPermission(t *testing.T) {
	resp := performHouseRequest(t, &fakeHouseService{}, "/api/v1/house_building/list", `{"root_id":"root-id"}`, housePrincipal([]string{"provider.view"}))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error != "当前账号无权查看楼栋列表" {
		t.Fatalf("unexpected error: %q", envelope.Error)
	}
}

func TestListRoomsPropagatesServiceError(t *testing.T) {
	svc := &fakeHouseService{roomListErr: errcode.InvalidParam.WithError(fmt.Errorf("项目/小区参数不正确"))}
	resp := performHouseRequest(t, svc, "/api/v1/house_room/list", `{"root_id":"bad"}`, housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestDetailRoomReturnsRoom(t *testing.T) {
	svc := &fakeHouseService{
		roomDetailResult: &housesvc.DetailResult{
			House: housesvc.HouseSummary{
				ListingID:  "listing-id",
				SourceType: "centralized_room",
				Title:      "房间详情",
			},
		},
	}

	resp := performHouseRequest(t, svc, "/api/v1/house_room/detail", `{"listing_id":"listing-id"}`, housePrincipal([]string{"house.view"}))
	if resp.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d body=%s", resp.Code, resp.Body.String())
	}
	if svc.roomDetailInput.ListingID != "listing-id" {
		t.Fatalf("unexpected detail input: %#v", svc.roomDetailInput)
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var data detailResponse
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Room.SourceType != "centralized_room" || data.Room.Title != "房间详情" {
		t.Fatalf("unexpected detail response: %#v", data)
	}
}

func TestDetailRoomRejectsInvalidJSON(t *testing.T) {
	resp := performHouseRequest(t, &fakeHouseService{}, "/api/v1/house_room/detail", `{bad`, housePrincipal([]string{"house.view"}))
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
	rootListInput      housesvc.RootListInput
	rootListResult     *housesvc.RootListResult
	rootListErr        error
	buildingListInput  housesvc.BuildingListInput
	buildingListResult *housesvc.BuildingListResult
	buildingListErr    error
	roomListInput      housesvc.ListInput
	roomListResult     *housesvc.ListResult
	roomListErr        error
	roomDetailInput    housesvc.DetailInput
	roomDetailResult   *housesvc.DetailResult
	roomDetailErr      error
}

func (f *fakeHouseService) ListRoots(ctx context.Context, input housesvc.RootListInput) (*housesvc.RootListResult, error) {
	f.rootListInput = input
	if f.rootListErr != nil {
		return nil, f.rootListErr
	}
	return f.rootListResult, nil
}

func (f *fakeHouseService) ListBuildings(ctx context.Context, input housesvc.BuildingListInput) (*housesvc.BuildingListResult, error) {
	f.buildingListInput = input
	if f.buildingListErr != nil {
		return nil, f.buildingListErr
	}
	return f.buildingListResult, nil
}

func (f *fakeHouseService) ListRooms(ctx context.Context, input housesvc.ListInput) (*housesvc.ListResult, error) {
	f.roomListInput = input
	if f.roomListErr != nil {
		return nil, f.roomListErr
	}
	return f.roomListResult, nil
}

func (f *fakeHouseService) DetailRoom(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error) {
	f.roomDetailInput = input
	if f.roomDetailErr != nil {
		return nil, f.roomDetailErr
	}
	return f.roomDetailResult, nil
}
