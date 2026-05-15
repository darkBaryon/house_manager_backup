package listingprojection

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"testing"
)

func TestMapCentralizedMiniappListingBuildsOnlineSnapshot(t *testing.T) {
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeCentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}
	room := &hmdmodel.HmdRoomCentralized{
		CommonFields:      commonmodel.CommonFields{ID: listing.SourceID},
		RoomNo:            "1201",
		FloorNo:           12,
		RentMode:          hmdmodel.RentModeWhole,
		AreaSize:          35,
		Orientation:       hmdmodel.OrientationSouth,
		PaymentCycle:      hmdmodel.PaymentCycleMonthly,
		Rent:              5800,
		RoomStatus:        hmdmodel.RoomStatusAvailable,
		StartRentRule:     hmdmodel.StartRentRuleLongOneYear,
		RoomFacilities:    []hmdmodel.RoomFacility{hmdmodel.RoomFacilityBed},
		ListingFacilities: []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator},
	}
	project := &hmdmodel.HmdCentralized{
		ProjectName: "泊寓南山科技园",
		City:        "深圳",
		District:    "南山",
		AddressText: "南山区科技园",
	}
	building := &hmdmodel.HmdBuilding{
		BuildingName:      "A座",
		ListingFacilities: []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator, hmdmodel.ListingFacilityGym},
	}
	roomType := &hmdmodel.HmdRoomTypeCentralized{
		RoomCount:     1,
		HallCount:     1,
		BathroomCount: 1,
		Images:        []hmdmodel.TaggedImage{{URL: "https://example.com/type.jpg", Tag: hmdmodel.ImageTagBedroom}},
	}

	got := mapCentralizedMiniappListing(listing, room, project, building, roomType)

	if got.ListingID != listing.ID {
		t.Fatalf("expected listing id copied, got %v", got.ListingID)
	}
	if got.Title != "泊寓南山科技园 A座 1201" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
	if got.LayoutText != "1室1厅1卫" {
		t.Fatalf("expected layout fallback from room type, got %s", got.LayoutText)
	}
	if got.IsOnline != hpdmodel.HpdOnlineStatusYes {
		t.Fatalf("expected online listing, got %d", got.IsOnline)
	}
	if len(got.ListingFacilities) != 2 {
		t.Fatalf("expected merged listing facilities, got %#v", got.ListingFacilities)
	}
	if len(got.Images) != 1 || got.Images[0].URL != "https://example.com/type.jpg" {
		t.Fatalf("expected room type image fallback, got %#v", got.Images)
	}
}

func TestMapDecentralizedMiniappListingKeepsOfflineWhenRoomUnavailable(t *testing.T) {
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: bson.NewObjectID()},
		SourceType:    hpdmodel.HpdSourceTypeDecentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeDecentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}
	room := &hmdmodel.HmdRoomDecentralized{
		CommonFields:   commonmodel.CommonFields{ID: listing.SourceID},
		RoomNo:         "801",
		FloorNo:        8,
		RentMode:       hmdmodel.RentModeWhole,
		LayoutText:     "两室一厅",
		AreaSize:       68,
		Orientation:    hmdmodel.OrientationSouthNorth,
		PaymentCycle:   hmdmodel.PaymentCycleQuarterly,
		Rent:           7200,
		RoomStatus:     hmdmodel.RoomStatusRented,
		StartRentRule:  hmdmodel.StartRentRuleLongOneYear,
		RoomFacilities: []hmdmodel.RoomFacility{hmdmodel.RoomFacilityFridge},
	}
	community := &hmdmodel.HmdDecentralized{
		CommunityName: "南头城花园",
		City:          "深圳",
		District:      "南山",
		BizArea:       "科技园",
		AddressText:   "南山区南头",
		SubwayStation: "高新园",
	}

	got := mapDecentralizedMiniappListing(listing, room, community)

	if got.IsOnline != hpdmodel.HpdOnlineStatusNo {
		t.Fatalf("expected unavailable room to stay offline, got %d", got.IsOnline)
	}
	if got.BuildingOrCommunityName != community.CommunityName {
		t.Fatalf("expected community name copied, got %s", got.BuildingOrCommunityName)
	}
	if len(got.FeatureFlags) != 1 || got.FeatureFlags[0] != string(hmdmodel.RoomFacilityFridge) {
		t.Fatalf("expected room facility flags, got %#v", got.FeatureFlags)
	}
}
