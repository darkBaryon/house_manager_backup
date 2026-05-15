package history

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	useractivitymodel "house-manager/internal/model/useractivity"
	"house-manager/pkg/errcode"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHistoryAddValidatesSourceAndUpserts(t *testing.T) {
	userID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	repo := &fakeHistoryRepository{}
	listings := &fakeMiniappListingRepository{detail: &hpdmodel.HpdMiniappListing{ListingID: listingID, City: "深圳", Title: "测试房源"}}
	svc := NewService(repo, listings)

	result, err := svc.Add(context.Background(), AddInput{UserID: userID, ListingID: listingID, Source: useractivitymodel.HistorySourceDiscover})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if result.ListingID != listingID.Hex() || result.ViewedAt <= 0 {
		t.Fatalf("unexpected add result: %#v", result)
	}
	if repo.upsert == nil || repo.upsert.Source != useractivitymodel.HistorySourceDiscover {
		t.Fatalf("unexpected upsert entity: %#v", repo.upsert)
	}
}

func TestHistoryAddRejectsInvalidSource(t *testing.T) {
	svc := NewService(&fakeHistoryRepository{}, &fakeMiniappListingRepository{detail: &hpdmodel.HpdMiniappListing{ListingID: bson.NewObjectID()}})

	_, err := svc.Add(context.Background(), AddInput{UserID: bson.NewObjectID(), ListingID: bson.NewObjectID(), Source: "feed"})
	assertHistoryErrCode(t, err, errcode.InvalidParam.Code)
}

func TestHistoryListFiltersOfflineBeforePaging(t *testing.T) {
	userID := bson.NewObjectID()
	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	repo := &fakeHistoryRepository{
		list: []useractivitymodel.History{
			{UserID: userID, ListingID: id1, ViewedAt: 20},
			{UserID: userID, ListingID: id2, ViewedAt: 10},
		},
	}
	listings := &fakeMiniappListingRepository{
		online: []hpdmodel.HpdMiniappListing{{ListingID: id2, City: "深圳", Title: "二号", Price: 2000}},
	}
	svc := NewService(repo, listings)

	result, err := svc.List(context.Background(), ListInput{UserID: userID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Total != 1 || len(result.List) != 1 || result.List[0].ListingID != id2.Hex() || result.List[0].ViewedAt != 10 {
		t.Fatalf("unexpected filtered list: %#v", result)
	}
}

func TestHistoryCountMatchesOnlineFilteredListTotal(t *testing.T) {
	userID := bson.NewObjectID()
	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	repo := &fakeHistoryRepository{
		list: []useractivitymodel.History{
			{UserID: userID, ListingID: id1, ViewedAt: 20},
			{UserID: userID, ListingID: id2, ViewedAt: 10},
		},
		count: 99,
	}
	listings := &fakeMiniappListingRepository{
		online: []hpdmodel.HpdMiniappListing{{ListingID: id1, City: "深圳", Title: "一号"}},
	}
	svc := NewService(repo, listings)

	total, err := svc.Count(context.Background(), userID)
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected online filtered count 1, got %d", total)
	}
}

func assertHistoryErrCode(t *testing.T, err error, code int) {
	t.Helper()
	e := errcode.FromError(err)
	if e == nil || e.Code != code {
		t.Fatalf("expected error code %d, got %v from %v", code, e, err)
	}
}

type fakeHistoryRepository struct {
	upsert *useractivitymodel.History
	list   []useractivitymodel.History
	count  int64
}

func (f *fakeHistoryRepository) Upsert(ctx context.Context, entity *useractivitymodel.History) error {
	f.upsert = entity
	return nil
}

func (f *fakeHistoryRepository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.History, error) {
	return f.list, nil
}

func (f *fakeHistoryRepository) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	return f.count, nil
}

type fakeMiniappListingRepository struct {
	detail *hpdmodel.HpdMiniappListing
	online []hpdmodel.HpdMiniappListing
}

func (f *fakeMiniappListingRepository) FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error) {
	return f.detail, nil
}

func (f *fakeMiniappListingRepository) FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]hpdmodel.HpdMiniappListing, error) {
	return f.online, nil
}
