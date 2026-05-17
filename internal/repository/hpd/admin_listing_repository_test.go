package hpd

import (
	"testing"

	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAdminListingListFilterBuildsActiveFilter(t *testing.T) {
	providerID := bson.NewObjectID()
	roomStatus := hmdmodel.RoomStatusAvailable
	listingStatus := hpdmodel.HpdListingStatusPublished
	auditStatus := hpdmodel.HpdAuditStatusPending

	filter, err := adminListingListFilter(AdminListingListFilter{
		OwnerLandlordID: providerID,
		AssetMode:       hpdmodel.HpdAssetModeCentralized,
		City:            " 上海 ",
		District:        " 徐汇 ",
		RoomStatus:      &roomStatus,
		ListingStatus:   &listingStatus,
		AuditStatus:     &auditStatus,
	})
	if err != nil {
		t.Fatalf("build admin listing list filter: %v", err)
	}
	if filter["status"] != commonmodel.StatusActive ||
		filter["owner_landlord_id"] != providerID ||
		filter["asset_mode"] != hpdmodel.HpdAssetModeCentralized ||
		filter["city"] != "上海" ||
		filter["district"] != "徐汇" ||
		filter["room_status"] != roomStatus ||
		filter["listing_status"] != listingStatus ||
		filter["audit_status"] != auditStatus {
		t.Fatalf("unexpected filter: %#v", filter)
	}
}

func TestAdminListingListFilterRejectsInvalidEnum(t *testing.T) {
	_, err := adminListingListFilter(AdminListingListFilter{AssetMode: hpdmodel.HpdAssetMode("bad")})
	if err == nil {
		t.Fatalf("expected invalid enum error")
	}
}
