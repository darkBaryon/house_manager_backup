package hpd

import (
	"strings"
	"testing"

	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHpdListingValidateForCreateRejectsUnspecifiedListingStatus(t *testing.T) {
	listing := &HpdListing{
		SourceType: HpdSourceTypeCentralizedRoom,
		SourceID:   bson.NewObjectID(),
		AssetMode:  HpdAssetModeCentralized,
	}

	err := listing.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "listingStatus is invalid") {
		t.Fatalf("expected invalid listingStatus error, got %v", err)
	}
}

func TestHpdMiniappListingValidateForCreateRejectsInvalidOnlineStatus(t *testing.T) {
	listing := validHpdMiniappListing()
	listing.IsOnline = HpdOnlineStatus(2)

	err := listing.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "isOnline is invalid") {
		t.Fatalf("expected invalid isOnline error, got %v", err)
	}
}

func TestHpdPublisherListingValidateForCreateRequiresRootAndRoomFields(t *testing.T) {
	listing := &HpdPublisherListing{
		ListingID:     bson.NewObjectID(),
		SourceType:    HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     HpdAssetModeCentralized,
		ListingStatus: HpdListingStatusDraft,
		RoomStatus:    hmdmodel.RoomStatusAvailable,
		RentMode:      hmdmodel.RentModeWhole,
		City:          "杭州",
		Title:         "测试房源",
	}

	err := listing.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "rootID") {
		t.Fatalf("expected rootID required error, got %v", err)
	}
}

func TestHpdPublisherListingValidateForCreateAcceptsMinimalValidListing(t *testing.T) {
	listing := &HpdPublisherListing{
		ListingID:     bson.NewObjectID(),
		SourceType:    HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     HpdAssetModeCentralized,
		RootType:      HpdRootScopeTypeCentralizedProject,
		RootID:        bson.NewObjectID(),
		RentMode:      hmdmodel.RentModeWhole,
		City:          "杭州",
		RoomNo:        "1201",
		Title:         "测试房源",
		ListingStatus: HpdListingStatusDraft,
		RoomStatus:    hmdmodel.RoomStatusAvailable,
		IsOnline:      HpdOnlineStatusNo,
	}

	if err := listing.ValidateForCreate(); err != nil {
		t.Fatalf("expected valid publisher listing, got %v", err)
	}
}

func TestHpdRootScopeRelationValidateForCreateRequiresOwnerPhone(t *testing.T) {
	relation := &HpdRootScopeRelation{
		RootType:       HpdRootScopeTypeCentralizedProject,
		RootID:         bson.NewObjectID(),
		RelationStatus: HpdRelationStatusActive,
	}

	err := relation.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "ownerPhone is required") {
		t.Fatalf("expected owner phone requirement error, got %v", err)
	}
}

func TestHpdRootScopeRelationValidateForCreateAcceptsValidOwnerRelation(t *testing.T) {
	relation := &HpdRootScopeRelation{
		RootType:       HpdRootScopeTypeDecentralizedCommunity,
		RootID:         bson.NewObjectID(),
		OwnerPhone:     "13800000000",
		RelationStatus: HpdRelationStatusActive,
	}

	if err := relation.ValidateForCreate(); err != nil {
		t.Fatalf("expected valid root scope relation, got %v", err)
	}
}

func TestHpdListingValidateForCreateRejectsSourceAssetModeMismatch(t *testing.T) {
	listing := &HpdListing{
		SourceType:    HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     HpdAssetModeDecentralized,
		ListingStatus: HpdListingStatusDraft,
	}

	err := listing.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "sourceType and assetMode mismatch") {
		t.Fatalf("expected sourceType and assetMode mismatch error, got %v", err)
	}
}

func TestHpdMiniappListingValidateForCreateRejectsSourceAssetModeMismatch(t *testing.T) {
	listing := validHpdMiniappListing()
	listing.AssetMode = HpdAssetModeDecentralized

	err := listing.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "sourceType and assetMode mismatch") {
		t.Fatalf("expected sourceType and assetMode mismatch error, got %v", err)
	}
}

func TestHpdMiniappOnlineStatusRequiresPublishedAndAvailable(t *testing.T) {
	tests := []struct {
		name          string
		listingStatus HpdListingStatus
		roomStatus    hmdmodel.RoomStatus
		want          HpdOnlineStatus
	}{
		{
			name:          "published available",
			listingStatus: HpdListingStatusPublished,
			roomStatus:    hmdmodel.RoomStatusAvailable,
			want:          HpdOnlineStatusYes,
		},
		{
			name:          "published rented",
			listingStatus: HpdListingStatusPublished,
			roomStatus:    hmdmodel.RoomStatusRented,
			want:          HpdOnlineStatusNo,
		},
		{
			name:          "draft available",
			listingStatus: HpdListingStatusDraft,
			roomStatus:    hmdmodel.RoomStatusAvailable,
			want:          HpdOnlineStatusNo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HpdMiniappOnlineStatus(tt.listingStatus, tt.roomStatus)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestValidateHpdUpdateFieldsRejectsInvalidListingStatus(t *testing.T) {
	err := ValidateHpdUpdateFields(bson.M{"listing_status": HpdListingStatusUnspecified})
	if err == nil || !strings.Contains(err.Error(), "listingStatus is invalid") {
		t.Fatalf("expected invalid listingStatus error, got %v", err)
	}
}

func TestValidateHpdUpdateFieldsAllowsValidMiniappFields(t *testing.T) {
	err := ValidateHpdUpdateFields(bson.M{
		"source_type":        HpdSourceTypeCentralizedRoom,
		"asset_mode":         HpdAssetModeCentralized,
		"listing_status":     HpdListingStatusPublished,
		"rent_mode":          hmdmodel.RentModeWhole,
		"city":               "深圳",
		"title":              "星河湾 1001",
		"price":              5000,
		"subway_distance_m":  800,
		"area_size":          35,
		"weight_score":       10,
		"orientation":        hmdmodel.OrientationSouth,
		"payment_cycle":      hmdmodel.PaymentCycleMonthly,
		"start_rent_rule":    hmdmodel.StartRentRuleLongOneYear,
		"listing_facilities": []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator},
		"images":             []hmdmodel.TaggedImage{{URL: "https://example.com/room.jpg", Tag: hmdmodel.ImageTagBedroom}},
		"cost_items":         []HpdCostItem{{Name: "押金", Amount: 5000, Unit: "元"}},
		"is_online":          HpdOnlineStatusYes,
	})
	if err != nil {
		t.Fatalf("expected valid hpd update fields, got %v", err)
	}
}

func TestValidateHpdUpdateFieldsRejectsSourceAssetModeMismatch(t *testing.T) {
	err := ValidateHpdUpdateFields(bson.M{
		"source_type": HpdSourceTypeDecentralizedRoom,
		"asset_mode":  HpdAssetModeCentralized,
	})
	if err == nil || !strings.Contains(err.Error(), "sourceType and assetMode mismatch") {
		t.Fatalf("expected sourceType and assetMode mismatch error, got %v", err)
	}
}

func TestHpdAssetModeForSourceType(t *testing.T) {
	assetMode, ok := HpdAssetModeForSourceType(HpdSourceTypeCentralizedRoom)
	if !ok || assetMode != HpdAssetModeCentralized {
		t.Fatalf("expected centralized source to map to centralized asset mode, got %v %v", assetMode, ok)
	}

	assetMode, ok = HpdAssetModeForSourceType(HpdSourceTypeDecentralizedRoom)
	if !ok || assetMode != HpdAssetModeDecentralized {
		t.Fatalf("expected decentralized source to map to decentralized asset mode, got %v %v", assetMode, ok)
	}
}

func validHpdMiniappListing() *HpdMiniappListing {
	return &HpdMiniappListing{
		ListingID:  bson.NewObjectID(),
		SourceType: HpdSourceTypeCentralizedRoom,
		SourceID:   bson.NewObjectID(),
		AssetMode:  HpdAssetModeCentralized,
		RentMode:   hmdmodel.RentModeWhole,
		City:       "深圳",
		Title:      "星河湾 1001",
		Price:      5000,
		IsOnline:   HpdOnlineStatusYes,
	}
}
