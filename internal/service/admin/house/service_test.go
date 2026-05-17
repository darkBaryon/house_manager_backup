package house

import (
	"context"
	"errors"
	"testing"

	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	hpdrepo "house-manager/internal/repository/hpd"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestListRootsNormalizesFiltersAndMapsItems(t *testing.T) {
	providerID := bson.NewObjectID()
	repo := &fakeAdminListingRepository{
		rootItems: []hpdrepo.AdminRootListItem{{
			RootID:        bson.NewObjectID(),
			RootType:      hpdmodel.HpdRootScopeTypeCentralizedProject,
			AssetMode:     hpdmodel.HpdAssetModeCentralized,
			ProviderID:    providerID,
			ProviderPhone: "13800000000",
			ProviderName:  "张三",
			ProjectName:   "星海公寓",
			City:          "上海",
			District:      "徐汇",
			BuildingCount: 2,
			RoomCount:     24,
			UpdatedAt:     123,
		}},
		rootTotal: 1,
	}
	svc := newService(repo)
	roomStatus := int(hmdmodel.RoomStatusAvailable)

	result, err := svc.ListRoots(context.Background(), RootListInput{
		ProviderID: " " + providerID.Hex() + " ",
		AssetMode:  "centralized",
		City:       " 上海 ",
		District:   " 徐汇 ",
		RoomStatus: &roomStatus,
		Page:       -1,
		PageSize:   999,
	})
	if err != nil {
		t.Fatalf("list roots: %v", err)
	}
	if result.Page != 1 || result.PageSize != 100 || result.Total != 1 {
		t.Fatalf("unexpected page result: %#v", result)
	}
	if repo.rootInput.OwnerLandlordID != providerID || repo.rootInput.City != "上海" || repo.rootInput.District != "徐汇" {
		t.Fatalf("unexpected root filter: %#v", repo.rootInput)
	}
	if len(result.List) != 1 || result.List[0].RootName != "星海公寓" || result.List[0].BuildingCount != 2 {
		t.Fatalf("unexpected root list: %#v", result.List)
	}
}

func TestListBuildingsRequiresRootID(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.ListBuildings(context.Background(), BuildingListInput{})
	assertHouseErr(t, err, errcode.InvalidParam.Code, "项目/小区参数不正确")
}

func TestListRoomsRequiresRootID(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.ListRooms(context.Background(), ListInput{})
	assertHouseErr(t, err, errcode.InvalidParam.Code, "项目/小区参数不正确")
}

func TestListRoomsNormalizesRootAndBuildingFilters(t *testing.T) {
	rootID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	repo := &fakeAdminListingRepository{
		listItems: []hpdmodel.HpdAdminListing{{
			CommonFields: commonmodel.CommonFields{UpdatedAt: 123},
			ListingID:    listingID,
			RootID:       rootID,
			BuildingID:   buildingID,
			AssetMode:    hpdmodel.HpdAssetModeCentralized,
			Title:        "A座 1001",
		}},
		listTotal: 1,
	}
	svc := newService(repo)

	result, err := svc.ListRooms(context.Background(), ListInput{
		RootID:     rootID.Hex(),
		BuildingID: buildingID.Hex(),
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("list rooms: %v", err)
	}
	if repo.listInput.RootID != rootID || repo.listInput.BuildingID != buildingID {
		t.Fatalf("unexpected room filter: %#v", repo.listInput)
	}
	if len(result.List) != 1 || result.List[0].ListingID != listingID.Hex() {
		t.Fatalf("unexpected room list: %#v", result.List)
	}
}

func TestDetailRoomReturnsNotFound(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.DetailRoom(context.Background(), DetailInput{ListingID: bson.NewObjectID().Hex()})
	assertHouseErr(t, err, errcode.NotFound.Code, "房源不存在或已删除")
}

func TestListRootsWrapsRepositoryErrorWithChineseMessage(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{err: errors.New("mongo down")})
	_, err := svc.ListRoots(context.Background(), RootListInput{})
	assertHouseErr(t, err, errcode.DatabaseError.Code, "获取项目/小区列表失败，请稍后重试")
}

func assertHouseErr(t *testing.T, err error, code int, message string) {
	t.Helper()
	got := errcode.FromError(err)
	if got == nil {
		t.Fatalf("expected errcode, got %v", err)
	}
	if got.Code != code || got.PublicMessage() != message {
		t.Fatalf("expected code=%d message=%q, got code=%d message=%q err=%v", code, message, got.Code, got.PublicMessage(), err)
	}
}

type fakeAdminListingRepository struct {
	rootInput       hpdrepo.AdminRootListFilter
	rootItems       []hpdrepo.AdminRootListItem
	rootTotal       int64
	buildingInput   hpdrepo.AdminBuildingListFilter
	buildingItems   []hpdrepo.AdminBuildingListItem
	buildingTotal   int64
	listInput       hpdrepo.AdminListingListFilter
	listItems       []hpdmodel.HpdAdminListing
	listTotal       int64
	detailListingID bson.ObjectID
	detail          *hpdmodel.HpdAdminListing
	err             error
}

func (f *fakeAdminListingRepository) ListAdminRoots(ctx context.Context, input hpdrepo.AdminRootListFilter) ([]hpdrepo.AdminRootListItem, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.rootInput = input
	return f.rootItems, f.rootTotal, nil
}

func (f *fakeAdminListingRepository) ListAdminBuildings(ctx context.Context, input hpdrepo.AdminBuildingListFilter) ([]hpdrepo.AdminBuildingListItem, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.buildingInput = input
	return f.buildingItems, f.buildingTotal, nil
}

func (f *fakeAdminListingRepository) ListAdmin(ctx context.Context, input hpdrepo.AdminListingListFilter) ([]hpdmodel.HpdAdminListing, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.listInput = input
	return f.listItems, f.listTotal, nil
}

func (f *fakeAdminListingRepository) FindByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdAdminListing, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.detailListingID = listingID
	return f.detail, nil
}
