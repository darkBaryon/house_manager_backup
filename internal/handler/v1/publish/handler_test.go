package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"
	"house-manager/pkg/errcode"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type publishResponse struct {
	Code  int             `json:"code"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func TestPublishHandlerInvalidJSONReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishService{}
	resp := performPublishRequest(t, svc, "/api/v1/centralized_project/create", "{")

	assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.createCentralizedProjectCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", svc.createCentralizedProjectCalls)
	}
}

func TestPublishHandlerMissingRequiredFieldReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishService{}
	body := `{"project_code":"p001","city":"深圳"}`
	resp := performPublishRequest(t, svc, "/api/v1/centralized_project/create", body)

	assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.createCentralizedProjectCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", svc.createCentralizedProjectCalls)
	}
}

func TestPublishHandlerInvalidObjectIDReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishService{}
	resp := performPublishRequest(t, svc, "/api/v1/building/detail", `{"id":"bad-id"}`)

	assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.getBuildingCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", svc.getBuildingCalls)
	}
}

func TestPublishHandlerCreateBuildingBindsRequest(t *testing.T) {
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	svc := &fakePublishService{
		createBuildingResult: &model.HmdBuilding{
			CommonFields:      model.CommonFields{ID: buildingID},
			ProjectID:         projectID,
			BuildingName:      "A栋",
			BuildingCode:      "B001",
			FloorTotal:        18,
			ManagerName:       "测试管家",
			ManagerPhone:      "18800000000",
			Photos:            []string{"https://example.com/a.jpg"},
			ListingFacilities: []model.ListingFacility{model.ListingFacilityElevator},
		},
	}

	body := mustJSON(t, map[string]any{
		"project_id":         projectID.Hex(),
		"building_name":      "A栋",
		"building_code":      "B001",
		"floor_total":        18,
		"manager_name":       "测试管家",
		"manager_phone":      "18800000000",
		"photos":             []string{"https://example.com/a.jpg"},
		"listing_facilities": []string{string(model.ListingFacilityElevator)},
	})
	resp := performPublishRequest(t, svc, "/api/v1/building/create", body)

	envelope := assertPublishResponse(t, resp, http.StatusOK, 0)
	if svc.createBuildingCalls != 1 {
		t.Fatalf("expected CreateBuilding to be called once, got %d", svc.createBuildingCalls)
	}
	if svc.createBuildingInput.ProjectID != projectID {
		t.Fatalf("unexpected projectID: %s", svc.createBuildingInput.ProjectID.Hex())
	}
	if svc.createBuildingInput.BuildingName != "A栋" || svc.createBuildingInput.BuildingCode != "B001" {
		t.Fatalf("unexpected building input: %#v", svc.createBuildingInput)
	}
	if svc.createBuildingInput.FloorTotal != 18 || len(svc.createBuildingInput.ListingFacilities) != 1 {
		t.Fatalf("unexpected building details: %#v", svc.createBuildingInput)
	}

	var data map[string]any
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if data["id"] != buildingID.Hex() || data["building_name"] != "A栋" || data["project_id"] != projectID.Hex() {
		t.Fatalf("unexpected response data: %#v", data)
	}
	assertJSONKeys(t, envelope.Data, []string{"building_name", "building_code", "created_at", "updated_at", "listing_facilities"}, []string{"buildingName", "buildingCode", "createdAt", "updatedAt", "listingFacilities"})
}

func TestPublishHandlerListWrapsDataListAndKeepsEmptyArray(t *testing.T) {
	projectID := bson.NewObjectID()
	svc := &fakePublishService{}

	resp := performPublishRequest(t, svc, "/api/v1/building/list_by_project", mustJSON(t, map[string]any{"project_id": projectID.Hex()}))

	envelope := assertPublishResponse(t, resp, http.StatusOK, 0)
	var data struct {
		List []map[string]any `json:"list"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if data.List == nil || len(data.List) != 0 {
		t.Fatalf("expected empty list array, got %#v body=%s", data.List, string(envelope.Data))
	}
	if strings.HasPrefix(strings.TrimSpace(string(envelope.Data)), "[") {
		t.Fatalf("list response must be wrapped object, got %s", string(envelope.Data))
	}
}

func TestPublishHandlerCentralizedProjectListAllowsEmptyFilter(t *testing.T) {
	svc := &fakePublishService{}

	resp := performPublishRequest(t, svc, "/api/v1/centralized_project/list", `{}`)

	envelope := assertPublishResponse(t, resp, http.StatusOK, 0)
	if svc.listCentralizedProjectsCalls != 1 {
		t.Fatalf("expected ListCentralizedProjects to be called once, got %d", svc.listCentralizedProjectsCalls)
	}
	if svc.listCentralizedProjectsInput.City != "" || svc.listCentralizedProjectsInput.District != "" {
		t.Fatalf("expected empty list filter, got %#v", svc.listCentralizedProjectsInput)
	}
	var data struct {
		List []map[string]any `json:"list"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if data.List == nil {
		t.Fatalf("expected list array, got nil body=%s", string(envelope.Data))
	}
}

func TestPublishHandlerCentralizedRoomStatusReturnsFullSnakeCaseDTO(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	roomTypeID := bson.NewObjectID()
	svc := &fakePublishService{
		updateCentralizedRoomStatusResult: &model.HmdRoomCentralized{
			CommonFields: model.CommonFields{ID: roomID, CreatedAt: 11, UpdatedAt: 22, Status: model.StatusActive, Version: 3},
			ProjectID:    projectID,
			BuildingID:   buildingID,
			RoomTypeID:   roomTypeID,
			RoomNo:       "1208",
			RentMode:     model.RentModeWhole,
			RoomStatus:   model.RoomStatusRented,
		},
	}

	resp := performPublishRequest(t, svc, "/api/v1/centralized_room/update_status", mustJSON(t, map[string]any{
		"id":          roomID.Hex(),
		"room_status": int(model.RoomStatusRented),
	}))

	envelope := assertPublishResponse(t, resp, http.StatusOK, 0)
	if svc.updateCentralizedRoomStatusCalls != 1 {
		t.Fatalf("expected UpdateCentralizedRoomStatus to be called once, got %d", svc.updateCentralizedRoomStatusCalls)
	}
	if svc.updateCentralizedRoomStatusInput.ID != roomID || svc.updateCentralizedRoomStatusInput.RoomStatus != int(model.RoomStatusRented) {
		t.Fatalf("unexpected room status input: %#v", svc.updateCentralizedRoomStatusInput)
	}
	var data map[string]any
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if data["id"] != roomID.Hex() || data["project_id"] != projectID.Hex() || data["building_id"] != buildingID.Hex() || data["room_type_id"] != roomTypeID.Hex() {
		t.Fatalf("unexpected centralized room response: %#v", data)
	}
	if data["room_status"] != float64(model.RoomStatusRented) {
		t.Fatalf("unexpected room status: %#v", data["room_status"])
	}
	assertJSONKeys(t, envelope.Data, []string{"room_no", "rent_mode", "room_status", "room_type_id", "created_at"}, []string{"roomNo", "rentMode", "roomStatus", "roomTypeId", "createdAt"})
}

func TestPublishHandlerRoomStatusRequiresExplicitAllowedTarget(t *testing.T) {
	roomID := bson.NewObjectID()
	cases := []struct {
		name string
		body string
	}{
		{
			name: "missing room_status",
			body: mustJSON(t, map[string]any{"id": roomID.Hex()}),
		},
		{
			name: "zero room_status",
			body: mustJSON(t, map[string]any{"id": roomID.Hex(), "room_status": int(model.RoomStatusUnspecified)}),
		},
		{
			name: "unknown room_status",
			body: mustJSON(t, map[string]any{"id": roomID.Hex(), "room_status": 99}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakePublishService{}
			resp := performPublishRequest(t, svc, "/api/v1/centralized_room/update_status", tc.body)

			assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
			if svc.updateCentralizedRoomStatusCalls != 0 {
				t.Fatalf("expected service not to be called, got %d calls", svc.updateCentralizedRoomStatusCalls)
			}
		})
	}
}

func TestPublishHandlerDecentralizedRoomDTOHasNoRoomTypeID(t *testing.T) {
	roomID := bson.NewObjectID()
	decentralizedID := bson.NewObjectID()
	ignoredRoomTypeID := bson.NewObjectID()
	svc := &fakePublishService{
		createDecentralizedRoomResult: &model.HmdRoomDecentralized{
			CommonFields:      model.CommonFields{ID: roomID},
			DecentralizedID:   decentralizedID,
			RoomNo:            "3-201",
			RentMode:          model.RentModeShared,
			RoomStatus:        model.RoomStatusAvailable,
			RoomFacilities:    []model.RoomFacility{model.RoomFacilityBed},
			ListingFacilities: []model.ListingFacility{model.ListingFacilitySubway},
		},
	}

	resp := performPublishRequest(t, svc, "/api/v1/decentralized_room/create", mustJSON(t, map[string]any{
		"decentralized_id": decentralizedID.Hex(),
		"room_type_id":     ignoredRoomTypeID.Hex(),
		"room_no":          "3-201",
		"rent_mode":        string(model.RentModeShared),
	}))

	envelope := assertPublishResponse(t, resp, http.StatusOK, 0)
	if svc.createDecentralizedRoomCalls != 1 || svc.createDecentralizedRoomInput.DecentralizedID != decentralizedID {
		t.Fatalf("unexpected decentralized room input: %#v", svc.createDecentralizedRoomInput)
	}
	var data map[string]any
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if _, ok := data["room_type_id"]; ok {
		t.Fatalf("decentralized room response must not contain room_type_id: %#v", data)
	}
	assertJSONKeys(t, envelope.Data, []string{"decentralized_id", "room_no", "rent_mode", "room_status"}, []string{"decentralizedId", "roomNo", "rentMode", "roomStatus", "room_type_id"})
}

func TestPublishHandlerServiceErrcodes(t *testing.T) {
	id := bson.NewObjectID()
	buildingBody := mustJSON(t, map[string]any{
		"project_id":    bson.NewObjectID().Hex(),
		"building_name": "A栋",
	})

	cases := []struct {
		name       string
		path       string
		body       string
		setup      func(*fakePublishService)
		httpStatus int
		code       int
	}{
		{
			name: "not found",
			path: "/api/v1/building/detail",
			body: mustJSON(t, map[string]any{"id": id.Hex()}),
			setup: func(s *fakePublishService) {
				s.getBuildingErr = errcode.NotFound.WithError(fmt.Errorf("missing building"))
			},
			httpStatus: http.StatusNotFound,
			code:       errcode.NotFound.Code,
		},
		{
			name: "already exists",
			path: "/api/v1/building/create",
			body: buildingBody,
			setup: func(s *fakePublishService) {
				s.createBuildingErr = errcode.AlreadyExists.WithError(fmt.Errorf("duplicate building"))
			},
			httpStatus: http.StatusConflict,
			code:       errcode.AlreadyExists.Code,
		},
		{
			name: "database error",
			path: "/api/v1/building/create",
			body: buildingBody,
			setup: func(s *fakePublishService) {
				s.createBuildingErr = errcode.DatabaseError.WithError(fmt.Errorf("mongo failed"))
			},
			httpStatus: http.StatusInternalServerError,
			code:       errcode.DatabaseError.Code,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakePublishService{}
			tc.setup(svc)
			resp := performPublishRequest(t, svc, tc.path, tc.body)
			assertPublishResponse(t, resp, tc.httpStatus, tc.code)
		})
	}
}

func performPublishRequest(t *testing.T, svc *fakePublishService, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	newPublishHandler(svc).RegisterRoutes(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertPublishResponse(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) publishResponse {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var envelope publishResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, envelope.Code, resp.Body.String())
	}
	return envelope
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func assertJSONKeys(t *testing.T, data json.RawMessage, wantKeys []string, forbiddenKeys []string) {
	t.Helper()
	text := string(data)
	for _, key := range wantKeys {
		if !strings.Contains(text, `"`+key+`"`) {
			t.Fatalf("expected json key %q in %s", key, text)
		}
	}
	for _, key := range forbiddenKeys {
		if strings.Contains(text, `"`+key+`"`) {
			t.Fatalf("unexpected json key %q in %s", key, text)
		}
	}
}

type fakePublishService struct {
	unimplementedPublishService

	createCentralizedProjectCalls int
	listCentralizedProjectsCalls  int
	listCentralizedProjectsInput  publishsvc.ListCentralizedProjectsInput
	createBuildingCalls           int
	createBuildingInput           publishsvc.CreateBuildingInput
	createBuildingResult          *model.HmdBuilding
	createBuildingErr             error
	getBuildingCalls              int
	getBuildingResult             *model.HmdBuilding
	getBuildingErr                error
	listBuildingsByProjectResult  []model.HmdBuilding

	updateCentralizedRoomStatusCalls  int
	updateCentralizedRoomStatusResult *model.HmdRoomCentralized
	updateCentralizedRoomStatusInput  publishsvc.UpdateCentralizedRoomStatusInput

	createDecentralizedRoomCalls  int
	createDecentralizedRoomInput  publishsvc.CreateDecentralizedRoomInput
	createDecentralizedRoomResult *model.HmdRoomDecentralized
}

func (f *fakePublishService) CreateCentralizedProject(ctx context.Context, input publishsvc.CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	f.createCentralizedProjectCalls++
	return nil, nil
}

func (f *fakePublishService) ListCentralizedProjects(ctx context.Context, input publishsvc.ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	f.listCentralizedProjectsCalls++
	f.listCentralizedProjectsInput = input
	return nil, nil
}

func (f *fakePublishService) CreateBuilding(ctx context.Context, input publishsvc.CreateBuildingInput) (*model.HmdBuilding, error) {
	f.createBuildingCalls++
	f.createBuildingInput = input
	return f.createBuildingResult, f.createBuildingErr
}

func (f *fakePublishService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	f.getBuildingCalls++
	return f.getBuildingResult, f.getBuildingErr
}

func (f *fakePublishService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	return f.listBuildingsByProjectResult, nil
}

func (f *fakePublishService) UpdateCentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	f.updateCentralizedRoomStatusCalls++
	f.updateCentralizedRoomStatusInput = input
	return f.updateCentralizedRoomStatusResult, nil
}

func (f *fakePublishService) CreateDecentralizedRoom(ctx context.Context, input publishsvc.CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	f.createDecentralizedRoomCalls++
	f.createDecentralizedRoomInput = input
	return f.createDecentralizedRoomResult, nil
}

type unimplementedPublishService struct{}

func (unimplementedPublishService) CreateCentralizedProject(ctx context.Context, input publishsvc.CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	panic("unexpected CreateCentralizedProject call")
}

func (unimplementedPublishService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	panic("unexpected GetCentralizedProject call")
}

func (unimplementedPublishService) ListCentralizedProjects(ctx context.Context, input publishsvc.ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	panic("unexpected ListCentralizedProjects call")
}

func (unimplementedPublishService) UpdateCentralizedProject(ctx context.Context, input publishsvc.UpdateCentralizedProjectInput) (*model.HmdCentralized, error) {
	panic("unexpected UpdateCentralizedProject call")
}

func (unimplementedPublishService) CreateBuilding(ctx context.Context, input publishsvc.CreateBuildingInput) (*model.HmdBuilding, error) {
	panic("unexpected CreateBuilding call")
}

func (unimplementedPublishService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	panic("unexpected GetBuilding call")
}

func (unimplementedPublishService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	panic("unexpected ListBuildingsByProject call")
}

func (unimplementedPublishService) UpdateBuilding(ctx context.Context, input publishsvc.UpdateBuildingInput) (*model.HmdBuilding, error) {
	panic("unexpected UpdateBuilding call")
}

func (unimplementedPublishService) CreateRoomType(ctx context.Context, input publishsvc.CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	panic("unexpected CreateRoomType call")
}

func (unimplementedPublishService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	panic("unexpected GetRoomType call")
}

func (unimplementedPublishService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	panic("unexpected ListRoomTypesByProject call")
}

func (unimplementedPublishService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	panic("unexpected ListRoomTypesByBuilding call")
}

func (unimplementedPublishService) UpdateRoomType(ctx context.Context, input publishsvc.UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	panic("unexpected UpdateRoomType call")
}

func (unimplementedPublishService) CreateCentralizedRoom(ctx context.Context, input publishsvc.CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	panic("unexpected CreateCentralizedRoom call")
}

func (unimplementedPublishService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	panic("unexpected GetCentralizedRoom call")
}

func (unimplementedPublishService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	panic("unexpected ListCentralizedRoomsByProject call")
}

func (unimplementedPublishService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	panic("unexpected ListCentralizedRoomsByBuilding call")
}

func (unimplementedPublishService) UpdateCentralizedRoom(ctx context.Context, input publishsvc.UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	panic("unexpected UpdateCentralizedRoom call")
}

func (unimplementedPublishService) UpdateCentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	panic("unexpected UpdateCentralizedRoomStatus call")
}

func (unimplementedPublishService) CreateDecentralizedCommunity(ctx context.Context, input publishsvc.CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	panic("unexpected CreateDecentralizedCommunity call")
}

func (unimplementedPublishService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	panic("unexpected GetDecentralizedCommunity call")
}

func (unimplementedPublishService) ListDecentralizedCommunities(ctx context.Context, input publishsvc.ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	panic("unexpected ListDecentralizedCommunities call")
}

func (unimplementedPublishService) UpdateDecentralizedCommunity(ctx context.Context, input publishsvc.UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	panic("unexpected UpdateDecentralizedCommunity call")
}

func (unimplementedPublishService) CreateDecentralizedRoom(ctx context.Context, input publishsvc.CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	panic("unexpected CreateDecentralizedRoom call")
}

func (unimplementedPublishService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	panic("unexpected GetDecentralizedRoom call")
}

func (unimplementedPublishService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	panic("unexpected ListDecentralizedRoomsByCommunity call")
}

func (unimplementedPublishService) UpdateDecentralizedRoom(ctx context.Context, input publishsvc.UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	panic("unexpected UpdateDecentralizedRoom call")
}

func (unimplementedPublishService) UpdateDecentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error) {
	panic("unexpected UpdateDecentralizedRoomStatus call")
}
