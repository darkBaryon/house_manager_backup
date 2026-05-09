package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	resp := performPublishRequest(t, svc, "/api/v1/publish/create_centralized_project", "{")

	assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.createCentralizedProjectCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", svc.createCentralizedProjectCalls)
	}
}

func TestPublishHandlerMissingRequiredFieldReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishService{}
	body := `{"project_code":"p001","city":"深圳"}`
	resp := performPublishRequest(t, svc, "/api/v1/publish/create_centralized_project", body)

	assertPublishResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.createCentralizedProjectCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", svc.createCentralizedProjectCalls)
	}
}

func TestPublishHandlerInvalidObjectIDReturnsInvalidParam(t *testing.T) {
	svc := &fakePublishService{}
	resp := performPublishRequest(t, svc, "/api/v1/publish/building_detail", `{"id":"bad-id"}`)

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
	resp := performPublishRequest(t, svc, "/api/v1/publish/create_building", body)

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

	var data model.HmdBuilding
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	if data.ID != buildingID || data.BuildingName != "A栋" {
		t.Fatalf("unexpected response data: %#v", data)
	}
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
			path: "/api/v1/publish/building_detail",
			body: mustJSON(t, map[string]any{"id": id.Hex()}),
			setup: func(s *fakePublishService) {
				s.getBuildingErr = errcode.NotFound.WithError(fmt.Errorf("missing building"))
			},
			httpStatus: http.StatusNotFound,
			code:       errcode.NotFound.Code,
		},
		{
			name: "already exists",
			path: "/api/v1/publish/create_building",
			body: buildingBody,
			setup: func(s *fakePublishService) {
				s.createBuildingErr = errcode.AlreadyExists.WithError(fmt.Errorf("duplicate building"))
			},
			httpStatus: http.StatusConflict,
			code:       errcode.AlreadyExists.Code,
		},
		{
			name: "database error",
			path: "/api/v1/publish/create_building",
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

type fakePublishService struct {
	unimplementedPublishService

	createCentralizedProjectCalls int
	createBuildingCalls           int
	createBuildingInput           publishsvc.CreateBuildingInput
	createBuildingResult          *model.HmdBuilding
	createBuildingErr             error
	getBuildingCalls              int
	getBuildingResult             *model.HmdBuilding
	getBuildingErr                error
}

func (f *fakePublishService) CreateCentralizedProject(ctx context.Context, input publishsvc.CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	f.createCentralizedProjectCalls++
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
