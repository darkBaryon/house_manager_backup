package hpd

import (
	"context"
	"strings"
	"testing"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestActiveFilterOverridesIncomingStatus(t *testing.T) {
	filter := activeFilter(bson.M{"city": "深圳", "status": model.StatusDeleted})

	if got := filter["status"]; got != model.StatusActive {
		t.Fatalf("expected active status filter, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city preserved, got %v", got)
	}
}

func TestListingUpdateLifecycleRejectsIdentityField(t *testing.T) {
	repo := &ListingRepository{}
	err := repo.UpdateLifecycleFields(context.Background(), bson.NewObjectID(), bson.M{"source_id": bson.NewObjectID()})
	if err == nil || !strings.Contains(err.Error(), "source_id") {
		t.Fatalf("expected source_id to be rejected, got %v", err)
	}
}

func TestListingUpdateStatusRejectsUnspecified(t *testing.T) {
	repo := &ListingRepository{}
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), model.HpdListingStatusUnspecified)
	if err == nil || !strings.Contains(err.Error(), "listingStatus is invalid") {
		t.Fatalf("expected invalid listingStatus error, got %v", err)
	}
}

func TestListingFieldsDoNotIncludeLifecycleFields(t *testing.T) {
	fields := listingFields(&model.HpdListing{
		SourceType:    model.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     model.HpdAssetModeCentralized,
		ListingStatus: model.HpdListingStatusPublished,
		PublishedAt:   100,
		OfflineAt:     200,
	})

	for _, field := range []string{"listing_status", "published_at", "offline_at"} {
		if _, ok := fields[field]; ok {
			t.Fatalf("expected %s to be excluded from projector upsert fields", field)
		}
	}
}

func TestMiniappUpdateProjectionRejectsIdentityField(t *testing.T) {
	repo := &MiniappListingRepository{}
	err := repo.UpdateProjectionFields(context.Background(), bson.NewObjectID(), bson.M{"listing_id": bson.NewObjectID()})
	if err == nil || !strings.Contains(err.Error(), "listing_id") {
		t.Fatalf("expected listing_id to be rejected, got %v", err)
	}
}

func TestMiniappSearchFilterBuildsOnlinePriceRange(t *testing.T) {
	filter, err := miniappSearchFilter(MiniappListingSearchFilter{
		City:     "深圳",
		RentMode: model.RentModeWhole,
		PriceMin: 3000,
		PriceMax: 6000,
	})
	if err != nil {
		t.Fatalf("expected search filter to build, got %v", err)
	}
	if got := filter["status"]; got != model.StatusActive {
		t.Fatalf("expected active status filter, got %v", got)
	}
	if got := filter["is_online"]; got != model.HpdOnlineStatusYes {
		t.Fatalf("expected online filter, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city filter, got %v", got)
	}
	if got := filter["rent_mode"]; got != model.RentModeWhole {
		t.Fatalf("expected rent mode filter, got %v", got)
	}

	price, ok := filter["price"].(bson.M)
	if !ok {
		t.Fatalf("expected price range filter, got %T", filter["price"])
	}
	if got := price["$gte"]; got != 3000 {
		t.Fatalf("expected price min 3000, got %v", got)
	}
	if got := price["$lte"]; got != 6000 {
		t.Fatalf("expected price max 6000, got %v", got)
	}
}

func TestMiniappSearchFilterRejectsInvalidRentMode(t *testing.T) {
	_, err := miniappSearchFilter(MiniappListingSearchFilter{RentMode: model.RentMode("daily")})
	if err == nil || !strings.Contains(err.Error(), "rentMode is invalid") {
		t.Fatalf("expected invalid rentMode error, got %v", err)
	}
}

func TestMiniappSearchFilterRejectsInvalidPriceRange(t *testing.T) {
	_, err := miniappSearchFilter(MiniappListingSearchFilter{PriceMin: 6000, PriceMax: 3000})
	if err == nil || !strings.Contains(err.Error(), "priceMin") {
		t.Fatalf("expected invalid price range error, got %v", err)
	}
}

func TestMiniappCountUsesSearchValidation(t *testing.T) {
	repo := &MiniappListingRepository{}
	_, err := repo.CountMiniapp(context.Background(), MiniappListingSearchFilter{PriceMin: -1})
	if err == nil || !strings.Contains(err.Error(), "price range must be non-negative") {
		t.Fatalf("expected invalid price range error, got %v", err)
	}
}
