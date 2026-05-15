package hmd

import (
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHmdRoomCentralizedValidateRejectsInvalidRentMode(t *testing.T) {
	room := &HmdRoomCentralized{
		ProjectID:  bson.NewObjectID(),
		BuildingID: bson.NewObjectID(),
		RoomNo:     "A-1001",
		RentMode:   RentMode("daily"),
		RoomStatus: RoomStatusAvailable,
	}

	err := room.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "rentMode is invalid") {
		t.Fatalf("expected invalid rentMode error, got %v", err)
	}
}

func TestHmdBuildingValidateAllowsEmptyBuildingCode(t *testing.T) {
	building := &HmdBuilding{
		ProjectID:    bson.NewObjectID(),
		BuildingName: "1号楼",
	}

	if err := building.ValidateForCreate(); err != nil {
		t.Fatalf("expected empty buildingCode to be allowed, got %v", err)
	}
}

func TestHmdDecentralizedValidateAllowsEmptyDistrict(t *testing.T) {
	community := &HmdDecentralized{
		CommunityName: "星河湾",
		City:          "深圳",
	}

	if err := community.ValidateForCreate(); err != nil {
		t.Fatalf("expected empty district to be allowed, got %v", err)
	}
}

func TestHmdRoomCentralizedValidateRejectsInvalidEnumFields(t *testing.T) {
	room := &HmdRoomCentralized{
		ProjectID:       bson.NewObjectID(),
		BuildingID:      bson.NewObjectID(),
		RoomNo:          "A-1001",
		RentMode:        RentModeWhole,
		RoomStatus:      RoomStatusAvailable,
		DecorationLevel: DecorationLevel("too_good"),
	}

	err := room.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "decorationLevel is invalid") {
		t.Fatalf("expected invalid decorationLevel error, got %v", err)
	}
}

func TestHmdRoomCentralizedValidateRejectsInvalidFacilities(t *testing.T) {
	room := &HmdRoomCentralized{
		ProjectID:         bson.NewObjectID(),
		BuildingID:        bson.NewObjectID(),
		RoomNo:            "A-1001",
		RentMode:          RentModeWhole,
		RoomStatus:        RoomStatusAvailable,
		ListingFacilities: []ListingFacility{"moon_pool"},
	}

	err := room.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "listing facility") {
		t.Fatalf("expected invalid listing facility error, got %v", err)
	}
}

func TestValidateHmdUpdateFieldsRejectsEmptyRequiredString(t *testing.T) {
	err := ValidateHmdUpdateFields(bson.M{"building_name": ""})
	if err == nil || !strings.Contains(err.Error(), "building_name is required") {
		t.Fatalf("expected required building_name error, got %v", err)
	}
}

func TestValidateHmdUpdateFieldsRejectsInvalidPaymentCycle(t *testing.T) {
	err := ValidateHmdUpdateFields(bson.M{"payment_cycle": PaymentCycle("weekly")})
	if err == nil || !strings.Contains(err.Error(), "paymentCycle is invalid") {
		t.Fatalf("expected invalid paymentCycle error, got %v", err)
	}
}

func TestValidateHmdUpdateFieldsAllowsValidEnums(t *testing.T) {
	err := ValidateHmdUpdateFields(bson.M{
		"rent_mode":         RentModeShared,
		"orientation":       OrientationSouthNorth,
		"decoration_level":  DecorationLevelFine,
		"payment_cycle":     PaymentCycleMonthly,
		"agency_fee_mode":   AgencyFeeModeNone,
		"viewing_time_rule": ViewingTimeRuleAnytime,
		"start_rent_rule":   StartRentRuleLongOneYear,
		"room_facilities":   []RoomFacility{RoomFacilityBed},
		"listing_facilities": []ListingFacility{
			ListingFacilityElevator,
		},
		"images": []TaggedImage{{URL: "https://example.com/room.jpg", Tag: ImageTagBedroom}},
	})
	if err != nil {
		t.Fatalf("expected valid update fields, got %v", err)
	}
}
