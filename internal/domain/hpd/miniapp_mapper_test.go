package hpd

import (
	"testing"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMapCentralizedMiniappListingBuildsOnlineSnapshot(t *testing.T) {
	listing := &model.HpdListing{
		CommonFields:  model.CommonFields{ID: bson.NewObjectID()},
		SourceType:    model.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     model.HpdAssetModeCentralized,
		ListingStatus: model.HpdListingStatusPublished,
	}
	room := &model.HmdRoomCentralized{
		CommonFields:      model.CommonFields{ID: listing.SourceID},
		RoomNo:            "1201",
		FloorNo:           12,
		RentMode:          model.RentModeWhole,
		AreaSize:          35,
		Orientation:       model.OrientationSouth,
		PaymentCycle:      model.PaymentCycleMonthly,
		Rent:              5800,
		RoomStatus:        model.RoomStatusAvailable,
		StartRentRule:     model.StartRentRuleLongOneYear,
		RoomFacilities:    []model.RoomFacility{model.RoomFacilityBed},
		ListingFacilities: []model.ListingFacility{model.ListingFacilityElevator},
	}
	project := &model.HmdCentralized{
		ProjectName: "泊寓南山科技园",
		City:        "深圳",
		District:    "南山",
		AddressText: "南山区科技园",
	}
	building := &model.HmdBuilding{
		BuildingName:      "A座",
		ListingFacilities: []model.ListingFacility{model.ListingFacilityElevator, model.ListingFacilityGym},
	}
	roomType := &model.HmdRoomTypeCentralized{
		RoomCount:     1,
		HallCount:     1,
		BathroomCount: 1,
		Images:        []model.TaggedImage{{URL: "https://example.com/type.jpg", Tag: model.ImageTagBedroom}},
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
	if got.IsOnline != model.HpdOnlineStatusYes {
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
	listing := &model.HpdListing{
		CommonFields:  model.CommonFields{ID: bson.NewObjectID()},
		SourceType:    model.HpdSourceTypeDecentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     model.HpdAssetModeDecentralized,
		ListingStatus: model.HpdListingStatusPublished,
	}
	room := &model.HmdRoomDecentralized{
		CommonFields:   model.CommonFields{ID: listing.SourceID},
		RoomNo:         "801",
		FloorNo:        8,
		RentMode:       model.RentModeWhole,
		LayoutText:     "两室一厅",
		AreaSize:       68,
		Orientation:    model.OrientationSouthNorth,
		PaymentCycle:   model.PaymentCycleQuarterly,
		Rent:           7200,
		RoomStatus:     model.RoomStatusRented,
		StartRentRule:  model.StartRentRuleLongOneYear,
		RoomFacilities: []model.RoomFacility{model.RoomFacilityFridge},
	}
	community := &model.HmdDecentralized{
		CommunityName: "南头城花园",
		City:          "深圳",
		District:      "南山",
		BizArea:       "科技园",
		AddressText:   "南山区南头",
		SubwayStation: "高新园",
	}

	got := mapDecentralizedMiniappListing(listing, room, community)

	if got.IsOnline != model.HpdOnlineStatusNo {
		t.Fatalf("expected unavailable room to stay offline, got %d", got.IsOnline)
	}
	if got.BuildingOrCommunityName != community.CommunityName {
		t.Fatalf("expected community name copied, got %s", got.BuildingOrCommunityName)
	}
	if len(got.FeatureFlags) != 1 || got.FeatureFlags[0] != string(model.RoomFacilityFridge) {
		t.Fatalf("expected room facility flags, got %#v", got.FeatureFlags)
	}
}
