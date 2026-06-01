package house

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeHouseService struct {
	searchInput housesvc.SearchInput
	detailInput housesvc.DetailInput
}

func (s *fakeHouseService) Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error) {
	s.searchInput = input
	return &housesvc.SearchResult{
		List: []housesvc.ListItem{{
			ListingID:       "661a0c6af8d56b3a1a4b2024",
			Title:           "科技园公寓 A栋 1208",
			Price:           4200,
			PriceText:       "4200元/月",
			District:        "南山",
			BizArea:         "科技园",
			LayoutText:      "单间",
			SubwayDistanceM: 650,
			Images:          []housesvc.TaggedImage{{URL: "https://example.com/room.jpg", Tag: "bedroom"}},
		}},
		Page:     1,
		PageSize: input.PageSize,
		Total:    1,
	}, nil
}

func (s *fakeHouseService) GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error) {
	s.detailInput = input
	return &housesvc.DetailResult{
		House: housesvc.Detail{
			ListItem: housesvc.ListItem{
				ListingID:       input.ListingID.Hex(),
				Title:           "科技园公寓 A栋 1208",
				Price:           4200,
				PriceText:       "4200元/月",
				District:        "南山",
				BizArea:         "科技园",
				LayoutText:      "单间",
				SubwayDistanceM: 650,
				Images:          []housesvc.TaggedImage{{URL: "https://example.com/room.jpg", Tag: "bedroom"}},
			},
			AddressText: "南山区科技园科发路",
			Description: "采光好，近地铁，适合通勤。",
		},
		IsFavorited: true,
	}, nil
}

func TestSearchMapsInternalToolResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeHouseService{}
	router := gin.New()
	newHandler(service).RegisterRoutes(router.Group(""))

	body := []byte(`{
		"session_id":"sess-1",
		"payload":{"district":"南山","keyword":"单间","page":1,"page_size":99}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tools/house/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-ID", "req-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	if service.searchInput.PageSize != internalToolMaxPageSize {
		t.Fatalf("page size = %d", service.searchInput.PageSize)
	}
	var got response.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != 0 {
		t.Fatalf("code = %d error = %s", got.Code, got.Error)
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T", got.Data)
	}
	list, ok := data["house_list"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("house_list = %#v", data["house_list"])
	}
	item := list[0].(map[string]any)
	if item["listing_id"] != "661a0c6af8d56b3a1a4b2024" || item["title"] != "科技园公寓 A栋 1208" {
		t.Fatalf("item = %#v", item)
	}
	if _, ok := item["house_id"]; ok {
		t.Fatalf("unexpected old field house_id in response: %#v", item)
	}
}

func TestPublicDetailOmitsFavoriteStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeHouseService{}
	router := gin.New()
	newHandler(service).RegisterRoutes(router.Group(""))

	listingID := bson.NewObjectID()
	body := []byte(`{
		"session_id":"sess-1",
		"listing_id":"` + listingID.Hex() + `"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tools/house/public_detail", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-ID", "req-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	if service.detailInput.UserID != bson.NilObjectID {
		t.Fatalf("user id should not be passed to public detail, got %s", service.detailInput.UserID.Hex())
	}
	var got response.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T", got.Data)
	}
	if data["listing_id"] != listingID.Hex() {
		t.Fatalf("listing_id = %#v", data["listing_id"])
	}
	if _, ok := data["is_favorited"]; ok {
		t.Fatalf("unexpected is_favorited in response: %#v", data)
	}
}
