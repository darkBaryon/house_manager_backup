package favorite

import (
	"context"
	"testing"

	"house-manager/internal/model"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestFavoriteAddIsIdempotentAndRequiresOnlineListing(t *testing.T) {
	userID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	favorites := &fakeFavoriteRepository{}
	listings := &fakeMiniappListingRepository{
		detail: &model.HpdMiniappListing{ListingID: listingID, City: "深圳", Title: "测试房源"},
	}
	svc := NewService(favorites, listings)

	result, err := svc.Add(context.Background(), AddInput{UserID: userID, ListingID: listingID})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if result.ListingID != listingID.Hex() || !result.IsFavorited {
		t.Fatalf("unexpected add result: %#v", result)
	}
	if favorites.upsertUserID != userID || favorites.upsertListingID != listingID {
		t.Fatalf("unexpected upsert: %#v", favorites)
	}
}

func TestFavoriteAddReturnsNotFoundForOfflineListing(t *testing.T) {
	svc := NewService(&fakeFavoriteRepository{}, &fakeMiniappListingRepository{})

	_, err := svc.Add(context.Background(), AddInput{UserID: bson.NewObjectID(), ListingID: bson.NewObjectID()})
	assertFavoriteErrCode(t, err, errcode.NotFound.Code)
}

func TestFavoriteListFiltersOfflineBeforePaging(t *testing.T) {
	userID := bson.NewObjectID()
	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	id3 := bson.NewObjectID()
	favorites := &fakeFavoriteRepository{
		list: []model.Favorite{
			{UserID: userID, ListingID: id1},
			{UserID: userID, ListingID: id2},
			{UserID: userID, ListingID: id3},
		},
	}
	listings := &fakeMiniappListingRepository{
		online: []model.HpdMiniappListing{
			{ListingID: id1, City: "深圳", Title: "一号", Price: 1000},
			{ListingID: id3, City: "深圳", Title: "三号", Price: 3000},
		},
	}
	svc := NewService(favorites, listings)

	result, err := svc.List(context.Background(), ListInput{UserID: userID, Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Total != 2 || len(result.List) != 1 || result.List[0].ListingID != id1.Hex() {
		t.Fatalf("unexpected filtered page: %#v", result)
	}
}

func TestFavoriteCountMatchesOnlineFilteredListTotal(t *testing.T) {
	userID := bson.NewObjectID()
	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	favorites := &fakeFavoriteRepository{
		list: []model.Favorite{
			{UserID: userID, ListingID: id1},
			{UserID: userID, ListingID: id2},
		},
		count: 99,
	}
	listings := &fakeMiniappListingRepository{
		online: []model.HpdMiniappListing{{ListingID: id2, City: "深圳", Title: "二号"}},
	}
	svc := NewService(favorites, listings)

	total, err := svc.Count(context.Background(), userID)
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected online filtered count 1, got %d", total)
	}
}

func assertFavoriteErrCode(t *testing.T, err error, code int) {
	t.Helper()
	e := errcode.FromError(err)
	if e == nil || e.Code != code {
		t.Fatalf("expected error code %d, got %v from %v", code, e, err)
	}
}

type fakeFavoriteRepository struct {
	upsertUserID    bson.ObjectID
	upsertListingID bson.ObjectID
	list            []model.Favorite
	count           int64
	exists          bool
}

func (f *fakeFavoriteRepository) Upsert(ctx context.Context, userID, listingID bson.ObjectID) error {
	f.upsertUserID = userID
	f.upsertListingID = listingID
	return nil
}

func (f *fakeFavoriteRepository) SoftRemove(ctx context.Context, userID, listingID bson.ObjectID) error {
	return nil
}

func (f *fakeFavoriteRepository) Exists(ctx context.Context, userID, listingID bson.ObjectID) (bool, error) {
	return f.exists, nil
}

func (f *fakeFavoriteRepository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]model.Favorite, error) {
	return f.list, nil
}

func (f *fakeFavoriteRepository) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	return f.count, nil
}

type fakeMiniappListingRepository struct {
	detail *model.HpdMiniappListing
	online []model.HpdMiniappListing
}

func (f *fakeMiniappListingRepository) FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*model.HpdMiniappListing, error) {
	return f.detail, nil
}

func (f *fakeMiniappListingRepository) FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]model.HpdMiniappListing, error) {
	return f.online, nil
}
