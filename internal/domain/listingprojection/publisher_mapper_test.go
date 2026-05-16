package listingprojection

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

func TestMapCentralizedPublisherListingBuildsProjectScopedSnapshot(t *testing.T) {
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeCentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}
	room := &hmdmodel.HmdRoomCentralized{
		CommonFields:      commonmodel.CommonFields{ID: listing.SourceID},
		ProjectID:         bson.NewObjectID(),
		BuildingID:        bson.NewObjectID(),
		RoomTypeID:        bson.NewObjectID(),
		RoomNo:            "1201",
		FloorNo:           12,
		RentMode:          hmdmodel.RentModeWhole,
		Rent:              5800,
		RoomStatus:        hmdmodel.RoomStatusAvailable,
		RoomFacilities:    []hmdmodel.RoomFacility{hmdmodel.RoomFacilityBed},
		ListingFacilities: []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator},
	}
	project := &hmdmodel.HmdCentralized{
		CommonFields: commonmodel.CommonFields{ID: room.ProjectID},
		ProjectName:  "泊寓南山科技园",
		City:         "深圳",
		District:     "南山",
	}
	building := &hmdmodel.HmdBuilding{
		CommonFields: commonmodel.CommonFields{ID: room.BuildingID},
		BuildingName: "A座",
	}
	roomType := &hmdmodel.HmdRoomTypeCentralized{
		CommonFields: commonmodel.CommonFields{ID: room.RoomTypeID},
		RoomTypeName: "一居室",
	}
	owner := &hpdmodel.HpdRootScopeRelation{
		OwnerLandlordID: bson.NewObjectID(),
		OwnerPhone:      "13800000000",
	}

	got := mapCentralizedPublisherListing(listing, room, project, building, roomType, owner)
	if got.RootType != hpdmodel.HpdRootScopeTypeCentralizedProject || got.RootID != project.ID {
		t.Fatalf("expected project root scope, got %#v", got)
	}
	if got.OwnerLandlordID != owner.OwnerLandlordID || got.OwnerPhoneSnapshot != owner.OwnerPhone {
		t.Fatalf("unexpected owner mapping: %#v", got)
	}
	if got.ProjectName != project.ProjectName || got.BuildingName != building.BuildingName || got.RoomTypeName != roomType.RoomTypeName {
		t.Fatalf("unexpected name mapping: %#v", got)
	}
	if got.ListingStatus != listing.ListingStatus || got.RoomStatus != room.RoomStatus {
		t.Fatalf("unexpected status mapping: %#v", got)
	}
}

func TestMapDecentralizedPublisherListingBuildsCommunityScopedSnapshot(t *testing.T) {
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		SourceType:    hpdmodel.HpdSourceTypeDecentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeDecentralized,
		ListingStatus: hpdmodel.HpdListingStatusDraft,
	}
	room := &hmdmodel.HmdRoomDecentralized{
		CommonFields: commonmodel.CommonFields{ID: listing.SourceID},
		RoomNo:       "801",
		RentMode:     hmdmodel.RentModeShared,
		Rent:         3200,
		RoomStatus:   hmdmodel.RoomStatusOccupied,
	}
	community := &hmdmodel.HmdDecentralized{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		CommunityName: "南头城花园",
		City:          "深圳",
		District:      "南山",
	}
	owner := &hpdmodel.HpdRootScopeRelation{
		OwnerLandlordID: bson.NewObjectID(),
		OwnerPhone:      "13800000000",
	}

	got := mapDecentralizedPublisherListing(listing, room, community, owner)
	if got.RootType != hpdmodel.HpdRootScopeTypeDecentralizedCommunity || got.RootID != community.ID {
		t.Fatalf("expected community root scope, got %#v", got)
	}
	if got.OwnerLandlordID != owner.OwnerLandlordID || got.OwnerPhoneSnapshot != owner.OwnerPhone {
		t.Fatalf("unexpected owner mapping: %#v", got)
	}
	if got.CommunityName != community.CommunityName || got.RoomNo != room.RoomNo {
		t.Fatalf("unexpected field mapping: %#v", got)
	}
	if got.ListingStatus != listing.ListingStatus || got.RoomStatus != room.RoomStatus {
		t.Fatalf("unexpected status mapping: %#v", got)
	}
}
