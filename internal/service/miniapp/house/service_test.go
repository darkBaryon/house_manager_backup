package house

import (
	"context"
	"errors"
	"testing"

	"house-manager/internal/model"
	repohpd "house-manager/internal/repository/hpd"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHouseSearchBuildsFilterAndMapsResult(t *testing.T) {
	listingID := bson.NewObjectID()
	repo := &fakeMiniappListingRepository{
		searchResult: []model.HpdMiniappListing{{
			ListingID:               listingID,
			AssetMode:               model.HpdAssetModeCentralized,
			RentMode:                model.RentModeWhole,
			City:                    "深圳",
			District:                "南山区",
			BizArea:                 "科技园",
			BuildingOrCommunityName: "测试公寓",
			Title:                   "整租一房",
			Price:                   5200,
			PriceText:               "5200元/月",
			PaymentCycle:            model.PaymentCycleMonthly,
			FeatureFlags:            []string{"near_subway"},
			ListingFacilities:       []model.ListingFacility{model.ListingFacilityElevator},
			PlatformTags:            []string{"精选"},
			Images:                  []model.TaggedImage{{URL: "https://example.com/a.jpg"}},
		}},
		countResult: 12,
	}
	svc := &HouseService{miniappListings: repo}

	result, err := svc.Search(context.Background(), SearchInput{
		City:         " 深圳 ",
		District:     "南山区",
		BizArea:      "科技园",
		RentMode:     string(model.RentModeWhole),
		AssetMode:    string(model.HpdAssetModeCentralized),
		MinPrice:     3000,
		MaxPrice:     8000,
		Keyword:      " 公寓 ",
		FeatureFlags: []string{"near_subway", " "},
		Page:         2,
		PageSize:     10,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if repo.searchCalls != 1 || repo.countCalls != 1 {
		t.Fatalf("expected search and count once, got %d/%d", repo.searchCalls, repo.countCalls)
	}
	if repo.searchFilter.City != "深圳" || repo.searchFilter.Keyword != "公寓" {
		t.Fatalf("unexpected trimmed filter: %#v", repo.searchFilter)
	}
	if repo.searchFilter.Skip != 10 || repo.searchFilter.Limit != 10 {
		t.Fatalf("unexpected pagination filter: %#v", repo.searchFilter)
	}
	if len(repo.searchFilter.FeatureFlags) != 1 || repo.searchFilter.FeatureFlags[0] != "near_subway" {
		t.Fatalf("unexpected feature flags: %#v", repo.searchFilter.FeatureFlags)
	}
	if result.Page != 2 || result.PageSize != 10 || result.Total != 12 {
		t.Fatalf("unexpected paging result: %#v", result)
	}
	if len(result.List) != 1 || result.List[0].ListingID != listingID.Hex() {
		t.Fatalf("unexpected list result: %#v", result.List)
	}
	if len(result.List[0].ListingFacilities) != 1 || result.List[0].ListingFacilities[0] != string(model.ListingFacilityElevator) {
		t.Fatalf("unexpected mapped facilities: %#v", result.List[0].ListingFacilities)
	}
}

func TestHouseSearchNormalizesPageSizeLimit(t *testing.T) {
	repo := &fakeMiniappListingRepository{}
	svc := &HouseService{miniappListings: repo}

	result, err := svc.Search(context.Background(), SearchInput{Page: -1, PageSize: 500})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if repo.searchFilter.Skip != 0 || repo.searchFilter.Limit != maxPageSize {
		t.Fatalf("unexpected normalized filter: %#v", repo.searchFilter)
	}
	if result.Page != defaultPage || result.PageSize != maxPageSize {
		t.Fatalf("unexpected normalized result: %#v", result)
	}
}

func TestHouseSearchReturnsInvalidParam(t *testing.T) {
	repo := &fakeMiniappListingRepository{}
	svc := &HouseService{miniappListings: repo}

	_, err := svc.Search(context.Background(), SearchInput{AssetMode: "villa"})
	assertErrCode(t, err, errcode.InvalidParam.Code)
	if repo.searchCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d", repo.searchCalls)
	}
}

func TestHouseSearchWrapsRepositoryError(t *testing.T) {
	repo := &fakeMiniappListingRepository{searchErr: errors.New("mongo down")}
	svc := &HouseService{miniappListings: repo}

	_, err := svc.Search(context.Background(), SearchInput{})
	assertErrCode(t, err, errcode.DatabaseError.Code)
	if repo.countCalls != 0 {
		t.Fatalf("expected count not to be called after search error, got %d", repo.countCalls)
	}
}

func TestHousePublicDetailMapsResult(t *testing.T) {
	listingID := bson.NewObjectID()
	repo := &fakeMiniappListingRepository{detailResult: &model.HpdMiniappListing{
		ListingID:     listingID,
		AssetMode:     model.HpdAssetModeDecentralized,
		RentMode:      model.RentModeShared,
		City:          "深圳",
		Title:         "合租单间",
		Price:         2600,
		AddressText:   "南山区测试路",
		Geo:           &model.GeoPoint{Lng: 113.1, Lat: 22.2},
		StartRentRule: model.StartRentRuleLongOneYear,
		CostItems:     []model.HpdCostItem{{Name: "押金", Amount: 2600, Unit: "元"}},
		ContactPhone:  "18800000000",
	}}
	svc := &HouseService{miniappListings: repo}

	result, err := svc.GetPublicDetail(context.Background(), DetailInput{ListingID: listingID})
	if err != nil {
		t.Fatalf("GetPublicDetail returned error: %v", err)
	}
	if repo.detailCalls != 1 || repo.detailID != listingID {
		t.Fatalf("unexpected detail repository call: %d %s", repo.detailCalls, repo.detailID.Hex())
	}
	if result.IsFavorited {
		t.Fatalf("expected anonymous detail not to be favorited")
	}
	if result.House.ListingID != listingID.Hex() || result.House.AddressText != "南山区测试路" {
		t.Fatalf("unexpected detail result: %#v", result.House)
	}
	if result.House.Geo == nil || result.House.Geo.Lng != 113.1 {
		t.Fatalf("unexpected geo: %#v", result.House.Geo)
	}
}

func TestHousePublicDetailReturnsNotFound(t *testing.T) {
	repo := &fakeMiniappListingRepository{}
	svc := &HouseService{miniappListings: repo}

	_, err := svc.GetPublicDetail(context.Background(), DetailInput{ListingID: bson.NewObjectID()})
	assertErrCode(t, err, errcode.NotFound.Code)
}

func assertErrCode(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %d, got nil", code)
	}
	e := errcode.FromError(err)
	if e == nil || e.Code != code {
		t.Fatalf("expected error code %d, got %v from %v", code, e, err)
	}
}

type fakeMiniappListingRepository struct {
	searchCalls  int
	searchFilter repohpd.MiniappListingSearchFilter
	searchResult []model.HpdMiniappListing
	searchErr    error

	countCalls  int
	countFilter repohpd.MiniappListingSearchFilter
	countResult int64
	countErr    error

	detailCalls  int
	detailID     bson.ObjectID
	detailResult *model.HpdMiniappListing
	detailErr    error
}

func (f *fakeMiniappListingRepository) SearchMiniapp(ctx context.Context, search repohpd.MiniappListingSearchFilter) ([]model.HpdMiniappListing, error) {
	f.searchCalls++
	f.searchFilter = search
	return f.searchResult, f.searchErr
}

func (f *fakeMiniappListingRepository) CountMiniapp(ctx context.Context, search repohpd.MiniappListingSearchFilter) (int64, error) {
	f.countCalls++
	f.countFilter = search
	return f.countResult, f.countErr
}

func (f *fakeMiniappListingRepository) FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*model.HpdMiniappListing, error) {
	f.detailCalls++
	f.detailID = listingID
	return f.detailResult, f.detailErr
}
