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

func TestListNormalizesFiltersAndMapsAdminListings(t *testing.T) {
	providerID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	roomStatus := int(hmdmodel.RoomStatusAvailable)
	listingStatus := int(hpdmodel.HpdListingStatusPublished)
	auditStatus := int(hpdmodel.HpdAuditStatusPending)
	repo := &fakeAdminListingRepository{
		listItems: []hpdmodel.HpdAdminListing{{
			CommonFields:       commonmodel.CommonFields{UpdatedAt: 123},
			ListingID:          listingID,
			AssetMode:          hpdmodel.HpdAssetModeCentralized,
			OwnerLandlordID:    providerID,
			OwnerPhoneSnapshot: "13800000000",
			City:               "上海",
			District:           "徐汇",
			Title:              "整租一居室",
			Price:              5200,
			RoomStatus:         hmdmodel.RoomStatusAvailable,
			ListingStatus:      hpdmodel.HpdListingStatusPublished,
			AuditStatus:        hpdmodel.HpdAuditStatusPending,
			IsOnline:           hpdmodel.HpdOnlineStatusYes,
		}},
		listTotal: 1,
	}
	svc := newService(repo)

	result, err := svc.List(context.Background(), ListInput{
		ProviderID:    " " + providerID.Hex() + " ",
		AssetMode:     "centralized",
		City:          " 上海 ",
		District:      " 徐汇 ",
		RoomStatus:    &roomStatus,
		ListingStatus: &listingStatus,
		AuditStatus:   &auditStatus,
		Page:          -1,
		PageSize:      200,
	})
	if err != nil {
		t.Fatalf("list house: %v", err)
	}
	if result.Page != 1 || result.PageSize != 100 || result.Total != 1 {
		t.Fatalf("unexpected page result: %#v", result)
	}
	if repo.listInput.OwnerLandlordID != providerID || repo.listInput.AssetMode != hpdmodel.HpdAssetModeCentralized {
		t.Fatalf("unexpected list filter: %#v", repo.listInput)
	}
	if repo.listInput.City != "上海" || repo.listInput.District != "徐汇" || repo.listInput.Skip != 0 || repo.listInput.Limit != 100 {
		t.Fatalf("unexpected normalized filter: %#v", repo.listInput)
	}
	if len(result.List) != 1 || result.List[0].ListingID != listingID.Hex() || result.List[0].ProviderPhone != "13800000000" {
		t.Fatalf("unexpected list item: %#v", result.List)
	}
}

func TestListRejectsInvalidFilter(t *testing.T) {
	status := 99
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.List(context.Background(), ListInput{RoomStatus: &status})
	assertHouseErr(t, err, errcode.InvalidParam.Code, "房源筛选参数不正确")
}

func TestDetailReturnsHouse(t *testing.T) {
	listingID := bson.NewObjectID()
	sourceID := bson.NewObjectID()
	repo := &fakeAdminListingRepository{
		detail: &hpdmodel.HpdAdminListing{
			ListingID:     listingID,
			SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
			SourceID:      sourceID,
			AssetMode:     hpdmodel.HpdAssetModeCentralized,
			Title:         "房源详情",
			ListingStatus: hpdmodel.HpdListingStatusPublished,
			AuditStatus:   hpdmodel.HpdAuditStatusApproved,
		},
	}
	svc := newService(repo)

	result, err := svc.Detail(context.Background(), DetailInput{ListingID: listingID.Hex()})
	if err != nil {
		t.Fatalf("detail house: %v", err)
	}
	if repo.detailListingID != listingID || result.House.SourceID != sourceID.Hex() || result.House.AuditStatus != int(hpdmodel.HpdAuditStatusApproved) {
		t.Fatalf("unexpected detail result: input=%s result=%#v", repo.detailListingID.Hex(), result.House)
	}
}

func TestDetailReturnsNotFound(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.Detail(context.Background(), DetailInput{ListingID: bson.NewObjectID().Hex()})
	assertHouseErr(t, err, errcode.NotFound.Code, "房源不存在或已删除")
}

func TestDetailRejectsInvalidListingID(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{})
	_, err := svc.Detail(context.Background(), DetailInput{ListingID: "bad"})
	assertHouseErr(t, err, errcode.InvalidParam.Code, "房源参数不正确")
}

func TestListWrapsRepositoryErrorWithChineseMessage(t *testing.T) {
	svc := newService(&fakeAdminListingRepository{err: errors.New("mongo down")})
	_, err := svc.List(context.Background(), ListInput{})
	assertHouseErr(t, err, errcode.DatabaseError.Code, "获取房源列表失败，请稍后重试")
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
	listInput       hpdrepo.AdminListingListFilter
	listItems       []hpdmodel.HpdAdminListing
	listTotal       int64
	detailListingID bson.ObjectID
	detail          *hpdmodel.HpdAdminListing
	err             error
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
