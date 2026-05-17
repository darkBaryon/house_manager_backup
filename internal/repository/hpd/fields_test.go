package hpd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"strings"
	"testing"
)

func TestActiveFilterOverridesIncomingStatus(t *testing.T) {
	filter := activeFilter(bson.M{"city": "深圳", "status": commonmodel.StatusDeleted})

	if got := filter["status"]; got != commonmodel.StatusActive {
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
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), hpdmodel.HpdListingStatusUnspecified)
	if err == nil || !strings.Contains(err.Error(), "listingStatus is invalid") {
		t.Fatalf("expected invalid listingStatus error, got %v", err)
	}
}

func TestListingFieldsDoNotIncludeLifecycleFields(t *testing.T) {
	fields := listingFields(&hpdmodel.HpdListing{
		SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeCentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
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
		RentMode: hmdmodel.RentModeWhole,
		PriceMin: 3000,
		PriceMax: 6000,
	})
	if err != nil {
		t.Fatalf("expected search filter to build, got %v", err)
	}
	if got := filter["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status filter, got %v", got)
	}
	if got := filter["is_online"]; got != hpdmodel.HpdOnlineStatusYes {
		t.Fatalf("expected online filter, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city filter, got %v", got)
	}
	if got := filter["rent_mode"]; got != hmdmodel.RentModeWhole {
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
	_, err := miniappSearchFilter(MiniappListingSearchFilter{RentMode: hmdmodel.RentMode("daily")})
	if err == nil || !strings.Contains(err.Error(), "rentMode is invalid") {
		t.Fatalf("expected invalid rentMode error, got %v", err)
	}
}

func TestMiniappSearchFilterBuildsApiContractFields(t *testing.T) {
	filter, err := miniappSearchFilter(MiniappListingSearchFilter{
		AssetMode:    hpdmodel.HpdAssetModeCentralized,
		Keyword:      "南山.*",
		FeatureFlags: []string{"near_subway", "balcony"},
	})
	if err != nil {
		t.Fatalf("expected search filter to build, got %v", err)
	}
	if got := filter["asset_mode"]; got != hpdmodel.HpdAssetModeCentralized {
		t.Fatalf("expected asset mode filter, got %v", got)
	}

	flags, ok := filter["feature_flags"].(bson.M)
	if !ok {
		t.Fatalf("expected feature flags filter, got %T", filter["feature_flags"])
	}
	all, ok := flags["$all"].([]string)
	if !ok {
		t.Fatalf("expected $all string slice, got %T", flags["$all"])
	}
	if len(all) != 2 || all[0] != "near_subway" || all[1] != "balcony" {
		t.Fatalf("unexpected feature flags: %#v", all)
	}

	keyword, ok := filter["$or"].(bson.A)
	if !ok || len(keyword) != 5 {
		t.Fatalf("expected keyword OR filter over 5 fields, got %#v", filter["$or"])
	}
	titleFilter, ok := keyword[0].(bson.M)
	if !ok {
		t.Fatalf("expected first keyword clause to be bson.M, got %#v", keyword[0])
	}
	title, ok := titleFilter["title"].(bson.Regex)
	if !ok {
		t.Fatalf("expected title regex, got %#v", titleFilter["title"])
	}
	if title.Pattern != "南山\\.\\*" || title.Options != "i" {
		t.Fatalf("expected escaped case-insensitive regex, got %#v", title)
	}
}

func TestMiniappSearchFilterRejectsInvalidAssetMode(t *testing.T) {
	_, err := miniappSearchFilter(MiniappListingSearchFilter{AssetMode: hpdmodel.HpdAssetMode("villa")})
	if err == nil || !strings.Contains(err.Error(), "assetMode is invalid") {
		t.Fatalf("expected invalid assetMode error, got %v", err)
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

func TestActiveRootScopeAccessFilterBuildsOwnerScope(t *testing.T) {
	rootID := bson.NewObjectID()
	landlordID := bson.NewObjectID()
	filter, err := activeRootScopeAccessFilter(hpdmodel.HpdRootScopeTypeCentralizedProject, rootID, landlordID)
	if err != nil {
		t.Fatalf("expected access filter, got %v", err)
	}
	if got := filter["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status, got %v", got)
	}
	if got := filter["relation_status"]; got != hpdmodel.HpdRelationStatusActive {
		t.Fatalf("expected active relation status, got %v", got)
	}
	if got := filter["root_type"]; got != hpdmodel.HpdRootScopeTypeCentralizedProject {
		t.Fatalf("expected root type, got %v", got)
	}
	if got := filter["root_id"]; got != rootID {
		t.Fatalf("expected root id, got %v", got)
	}
	if got := filter["owner_landlord_id"]; got != landlordID {
		t.Fatalf("expected owner landlord filter, got %#v", got)
	}
}

func TestRootScopeRepositoryRejectsEmptyScopeInputs(t *testing.T) {
	repo := &RootScopeRepository{}
	if _, err := repo.FindActiveByRootAndOwner(context.Background(), hpdmodel.HpdRootScopeType(""), bson.NewObjectID(), bson.NewObjectID()); err == nil || !strings.Contains(err.Error(), "rootType is invalid") {
		t.Fatalf("expected root type error, got %v", err)
	}
	if _, err := repo.ListActiveRootIDsByOwnerLandlordID(context.Background(), hpdmodel.HpdRootScopeTypeCentralizedProject, bson.NilObjectID); err == nil || !strings.Contains(err.Error(), "ownerLandlordID is required") {
		t.Fatalf("expected owner landlord error, got %v", err)
	}
}

func TestPublisherListingFieldsIncludeRootAndRoomFields(t *testing.T) {
	listingID := bson.NewObjectID()
	rootID := bson.NewObjectID()
	fields := publisherListingFields(&hpdmodel.HpdPublisherListing{
		ListingID:            listingID,
		OwnerLandlordID:      bson.NewObjectID(),
		OwnerPhoneSnapshot:   "13800000000",
		LandlordNameSnapshot: "房东A",
		RootType:             hpdmodel.HpdRootScopeTypeCentralizedProject,
		RootID:               rootID,
		RoomNo:               "1201",
		ListingStatus:        hpdmodel.HpdListingStatusDraft,
	})

	if got := fields["listing_id"]; got != listingID {
		t.Fatalf("expected listing id, got %v", got)
	}
	if got := fields["root_type"]; got != hpdmodel.HpdRootScopeTypeCentralizedProject {
		t.Fatalf("expected root type, got %v", got)
	}
	if got := fields["root_id"]; got != rootID {
		t.Fatalf("expected root id, got %v", got)
	}
	if got := fields["room_no"]; got != "1201" {
		t.Fatalf("expected room no, got %v", got)
	}
	if got := fields["owner_phone_snapshot"]; got != "13800000000" {
		t.Fatalf("expected owner phone snapshot, got %v", got)
	}
}

func TestAdminListingFieldsIncludeAuditAndOwnerFields(t *testing.T) {
	listingID := bson.NewObjectID()
	auditTaskID := bson.NewObjectID()
	fields := adminListingFields(&hpdmodel.HpdAdminListing{
		ListingID:          listingID,
		OwnerLandlordID:    bson.NewObjectID(),
		OwnerPhoneSnapshot: "13800000000",
		RootType:           hpdmodel.HpdRootScopeTypeCentralizedProject,
		RootID:             bson.NewObjectID(),
		RoomNo:             "1201",
		ListingStatus:      hpdmodel.HpdListingStatusPending,
		AuditStatus:        hpdmodel.HpdAuditStatusPending,
		LatestAuditTaskID:  auditTaskID,
		LatestSubmittedAt:  123,
	})

	if got := fields["listing_id"]; got != listingID {
		t.Fatalf("expected listing id, got %v", got)
	}
	if got := fields["audit_status"]; got != hpdmodel.HpdAuditStatusPending {
		t.Fatalf("expected audit status, got %v", got)
	}
	if got := fields["latest_audit_task_id"]; got != auditTaskID {
		t.Fatalf("expected latest audit task id, got %v", got)
	}
	if got := fields["owner_phone_snapshot"]; got != "13800000000" {
		t.Fatalf("expected owner phone snapshot, got %v", got)
	}
}

func TestAdminUpdateProjectionRejectsIdentityField(t *testing.T) {
	repo := &AdminListingRepository{}
	err := repo.UpdateProjectionFields(context.Background(), bson.NewObjectID(), bson.M{"listing_id": bson.NewObjectID()})
	if err == nil || !strings.Contains(err.Error(), "listing_id") {
		t.Fatalf("expected listing_id to be rejected, got %v", err)
	}
}
