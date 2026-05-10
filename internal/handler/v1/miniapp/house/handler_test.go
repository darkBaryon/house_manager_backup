package house

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/errcode"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type houseEnvelope struct {
	Code    int             `json:"code"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
	MaxSize *int64          `json:"maxSize,omitempty"`
	Size    *int            `json:"size,omitempty"`
}

func TestHouseSearchReturnsSnakeCasePagedData(t *testing.T) {
	listingID := bson.NewObjectID().Hex()
	svc := &fakeHouseService{
		searchResult: &housesvc.SearchResult{
			List: []housesvc.ListItem{{
				ListingID:               listingID,
				AssetMode:               "centralized",
				RentMode:                "whole",
				City:                    "深圳",
				District:                "南山区",
				BizArea:                 "科技园",
				BuildingOrCommunityName: "测试公寓",
				Price:                   5200,
				PriceText:               "5200元/月",
				FeatureFlags:            []string{"near_subway"},
				ListingFacilities:       []string{"elevator"},
				PlatformTags:            []string{"精选"},
				Images:                  []housesvc.TaggedImage{{URL: "https://example.com/a.jpg", Tag: "cover"}},
			}},
			Page:     2,
			PageSize: 10,
			Total:    23,
		},
	}
	body := mustHouseJSON(t, map[string]any{
		"city":          "深圳",
		"asset_mode":    "centralized",
		"feature_flags": []string{"near_subway"},
		"page":          2,
		"page_size":     10,
	})

	resp := performHouseRequest(t, svc, "/api/v1/house/search", body)
	envelope := assertHouseResponse(t, resp, http.StatusOK, 0)
	if envelope.MaxSize != nil || envelope.Size != nil {
		t.Fatalf("house search must not use SuccessPage shape, got body=%s", resp.Body.String())
	}
	if svc.searchCalls != 1 {
		t.Fatalf("expected Search to be called once, got %d", svc.searchCalls)
	}
	if svc.searchInput.City != "深圳" || svc.searchInput.AssetMode != "centralized" || len(svc.searchInput.FeatureFlags) != 1 {
		t.Fatalf("unexpected service input: %#v", svc.searchInput)
	}

	var data struct {
		List     []map[string]any `json:"list"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Total    int64            `json:"total"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data.Page != 2 || data.PageSize != 10 || data.Total != 23 || len(data.List) != 1 {
		t.Fatalf("unexpected paged data: %#v", data)
	}
	item := data.List[0]
	for _, key := range []string{"listing_id", "asset_mode", "rent_mode", "price_text", "feature_flags", "listing_facilities"} {
		if _, ok := item[key]; !ok {
			t.Fatalf("expected snake_case key %s in item: %#v", key, item)
		}
	}
	for _, key := range []string{"listingId", "assetMode", "rentMode", "priceText", "featureFlags", "listingFacilities"} {
		if _, ok := item[key]; ok {
			t.Fatalf("unexpected camelCase key %s in item: %#v", key, item)
		}
	}
}

func TestHouseSearchBindsFullRequest(t *testing.T) {
	svc := &fakeHouseService{searchResult: &housesvc.SearchResult{Page: 3, PageSize: 15}}
	body := mustHouseJSON(t, map[string]any{
		"city":          "深圳",
		"district":      "南山",
		"biz_area":      "科技园",
		"rent_mode":     "whole",
		"asset_mode":    "centralized",
		"min_price":     3000,
		"max_price":     6000,
		"keyword":       "高新园",
		"feature_flags": []string{"subway", "elevator"},
		"page":          3,
		"page_size":     15,
	})

	resp := performHouseRequest(t, svc, "/api/v1/house/search", body)

	assertHouseResponse(t, resp, http.StatusOK, 0)
	input := svc.searchInput
	if input.City != "深圳" ||
		input.District != "南山" ||
		input.BizArea != "科技园" ||
		input.RentMode != "whole" ||
		input.AssetMode != "centralized" ||
		input.MinPrice != 3000 ||
		input.MaxPrice != 6000 ||
		input.Keyword != "高新园" ||
		input.Page != 3 ||
		input.PageSize != 15 {
		t.Fatalf("unexpected search input: %#v", input)
	}
	if len(input.FeatureFlags) != 2 || input.FeatureFlags[0] != "subway" || input.FeatureFlags[1] != "elevator" {
		t.Fatalf("unexpected feature flags: %#v", input.FeatureFlags)
	}
}

func TestHousePublicDetailParsesListingIDAndReturnsSnakeCase(t *testing.T) {
	listingID := bson.NewObjectID()
	svc := &fakeHouseService{
		detailResult: &housesvc.DetailResult{
			House: housesvc.Detail{
				ListItem: housesvc.ListItem{
					ListingID: listingID.Hex(),
					AssetMode: "decentralized",
					RentMode:  "shared",
					City:      "深圳",
					Title:     "合租单间",
					Price:     2600,
				},
				AddressText:   "南山区测试路",
				Geo:           &housesvc.GeoPoint{Lng: 113.1, Lat: 22.2},
				StartRentRule: "long_one_year",
				CostItems:     []housesvc.CostItem{{Name: "押金", Amount: 2600, Unit: "元"}},
				ContactPhone:  "18800000000",
			},
			IsFavorited: false,
		},
	}

	resp := performHouseRequest(t, svc, "/api/v1/house/public_detail", mustHouseJSON(t, map[string]any{"listing_id": listingID.Hex()}))
	envelope := assertHouseResponse(t, resp, http.StatusOK, 0)
	if svc.detailCalls != 1 || svc.detailInput.ListingID != listingID {
		t.Fatalf("unexpected detail input: %#v", svc.detailInput)
	}

	var data map[string]any
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if _, ok := data["is_favorited"]; !ok {
		t.Fatalf("expected is_favorited, got %#v", data)
	}
	house, ok := data["house"].(map[string]any)
	if !ok {
		t.Fatalf("expected house object, got %#v", data["house"])
	}
	if house["listing_id"] != listingID.Hex() || house["address_text"] != "南山区测试路" {
		t.Fatalf("unexpected house data: %#v", house)
	}
	if _, ok := house["startRentRule"]; ok {
		t.Fatalf("unexpected camelCase field in detail: %#v", house)
	}
	if _, ok := house["start_rent_rule"]; !ok {
		t.Fatalf("expected start_rent_rule in detail: %#v", house)
	}
}

func TestHousePublicDetailMissingListingIDReturnsInvalidParam(t *testing.T) {
	svc := &fakeHouseService{}
	resp := performHouseRequest(t, svc, "/api/v1/house/public_detail", `{}`)

	assertHouseResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.detailCalls != 0 {
		t.Fatalf("expected service not to be called, got %d", svc.detailCalls)
	}
}

func TestHousePublicDetailInvalidListingIDReturnsInvalidParam(t *testing.T) {
	svc := &fakeHouseService{}
	resp := performHouseRequest(t, svc, "/api/v1/house/public_detail", `{"listing_id":"bad-id"}`)

	assertHouseResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.detailCalls != 0 {
		t.Fatalf("expected service not to be called, got %d", svc.detailCalls)
	}
}

func TestHouseSearchServiceErrcode(t *testing.T) {
	svc := &fakeHouseService{searchErr: errcode.DatabaseError.WithError(fmt.Errorf("mongo failed"))}
	resp := performHouseRequest(t, svc, "/api/v1/house/search", `{}`)

	assertHouseResponse(t, resp, http.StatusInternalServerError, errcode.DatabaseError.Code)
}

func TestHousePublicDetailServiceErrcodes(t *testing.T) {
	listingID := bson.NewObjectID()
	cases := []struct {
		name       string
		err        error
		httpStatus int
		code       int
	}{
		{
			name:       "not found",
			err:        errcode.NotFound.WithError(fmt.Errorf("not found")),
			httpStatus: http.StatusNotFound,
			code:       errcode.NotFound.Code,
		},
		{
			name:       "database",
			err:        errcode.DatabaseError.WithError(fmt.Errorf("mongo failed")),
			httpStatus: http.StatusInternalServerError,
			code:       errcode.DatabaseError.Code,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeHouseService{detailErr: tc.err}
			resp := performHouseRequest(t, svc, "/api/v1/house/public_detail", mustHouseJSON(t, map[string]any{"listing_id": listingID.Hex()}))

			assertHouseResponse(t, resp, tc.httpStatus, tc.code)
			if svc.detailCalls != 1 {
				t.Fatalf("expected detail service to be called once, got %d", svc.detailCalls)
			}
		})
	}
}

func performHouseRequest(t *testing.T, svc *fakeHouseService, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	newHouseHandler(svc).RegisterRoutes(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertHouseResponse(t *testing.T, resp *httptest.ResponseRecorder, httpStatus int, code int) houseEnvelope {
	t.Helper()
	if resp.Code != httpStatus {
		t.Fatalf("expected HTTP %d, got %d body=%s", httpStatus, resp.Code, resp.Body.String())
	}
	var envelope houseEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != code {
		t.Fatalf("expected code %d, got %d body=%s", code, envelope.Code, resp.Body.String())
	}
	return envelope
}

func mustHouseJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

type fakeHouseService struct {
	searchCalls  int
	searchInput  housesvc.SearchInput
	searchResult *housesvc.SearchResult
	searchErr    error

	detailCalls  int
	detailInput  housesvc.DetailInput
	detailResult *housesvc.DetailResult
	detailErr    error
}

func (f *fakeHouseService) Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error) {
	f.searchCalls++
	f.searchInput = input
	return f.searchResult, f.searchErr
}

func (f *fakeHouseService) GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error) {
	f.detailCalls++
	f.detailInput = input
	return f.detailResult, f.detailErr
}

func TestHouseSearchInvalidJSONReturnsInvalidParam(t *testing.T) {
	svc := &fakeHouseService{}
	resp := performHouseRequest(t, svc, "/api/v1/house/search", "{")

	assertHouseResponse(t, resp, http.StatusBadRequest, errcode.InvalidParam.Code)
	if svc.searchCalls != 0 {
		t.Fatalf("expected service not to be called, got %d", svc.searchCalls)
	}
}

func TestHouseSearchEmptyResultKeepsListArray(t *testing.T) {
	svc := &fakeHouseService{searchResult: &housesvc.SearchResult{Page: 1, PageSize: 20}}
	resp := performHouseRequest(t, svc, "/api/v1/house/search", `{}`)

	envelope := assertHouseResponse(t, resp, http.StatusOK, 0)
	if !strings.Contains(string(envelope.Data), `"list":[]`) {
		t.Fatalf("expected empty list array in data, got %s", string(envelope.Data))
	}
}
