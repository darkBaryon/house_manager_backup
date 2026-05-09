package model

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m *HmdCentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd centralized is nil")
	}
	if isBlank(m.ProjectName) || isBlank(m.ProjectCode) || isBlank(m.City) {
		return fmt.Errorf("projectName, projectCode and city are required")
	}
	return nil
}

func (m *HmdBuilding) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd building is nil")
	}
	if m.ProjectID.IsZero() || isBlank(m.BuildingName) {
		return fmt.Errorf("projectID and buildingName are required")
	}
	if err := validateNonNegativeInt("floorTotal", m.FloorTotal); err != nil {
		return err
	}
	return validateListingFacilities(m.ListingFacilities)
}

func (m *HmdDecentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd decentralized is nil")
	}
	if isBlank(m.CommunityName) || isBlank(m.City) {
		return fmt.Errorf("communityName and city are required")
	}
	return nil
}

func (m *HmdRoomTypeCentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd room type centralized is nil")
	}
	if isBlank(m.RoomTypeName) {
		return fmt.Errorf("roomTypeName is required")
	}
	if m.ProjectID.IsZero() && m.BuildingID.IsZero() {
		return fmt.Errorf("projectID or buildingID is required")
	}
	return validateHmdRoomShapeAndPrice(
		m.RoomCount,
		m.HallCount,
		m.BathroomCount,
		m.KitchenCount,
		m.AreaSize,
		m.Rent,
		m.Deposit,
		m.ServiceFee,
		m.AgencyFeeValue,
		m.Orientation,
		m.DecorationLevel,
		m.PaymentCycle,
		m.AgencyFeeMode,
		"",
		"",
		m.Images,
		m.RoomFacilities,
		nil,
	)
}

func (m *HmdRoomCentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd room centralized is nil")
	}
	if m.ProjectID.IsZero() || m.BuildingID.IsZero() || isBlank(m.RoomNo) || m.RentMode == "" {
		return fmt.Errorf("projectID, buildingID, roomNo and rentMode are required")
	}
	if !m.RentMode.Valid() {
		return fmt.Errorf("rentMode is invalid")
	}
	if !m.RoomStatus.Valid() {
		return fmt.Errorf("roomStatus is invalid")
	}
	return validateHmdRoomShapeAndPrice(
		0,
		0,
		0,
		0,
		m.AreaSize,
		m.Rent,
		m.Deposit,
		m.ServiceFee,
		m.AgencyFeeValue,
		m.Orientation,
		m.DecorationLevel,
		m.PaymentCycle,
		m.AgencyFeeMode,
		m.ViewingTimeRule,
		m.StartRentRule,
		m.Images,
		m.RoomFacilities,
		m.ListingFacilities,
	)
}

func (m *HmdRoomDecentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hmd room decentralized is nil")
	}
	if m.DecentralizedID.IsZero() || isBlank(m.RoomNo) || m.RentMode == "" {
		return fmt.Errorf("decentralizedID, roomNo and rentMode are required")
	}
	if !m.RentMode.Valid() {
		return fmt.Errorf("rentMode is invalid")
	}
	if !m.RoomStatus.Valid() {
		return fmt.Errorf("roomStatus is invalid")
	}
	return validateHmdRoomShapeAndPrice(
		0,
		0,
		0,
		0,
		m.AreaSize,
		m.Rent,
		m.Deposit,
		m.ServiceFee,
		m.AgencyFeeValue,
		m.Orientation,
		m.DecorationLevel,
		m.PaymentCycle,
		m.AgencyFeeMode,
		m.ViewingTimeRule,
		m.StartRentRule,
		m.Images,
		m.RoomFacilities,
		m.ListingFacilities,
	)
}

func ValidateHmdUpdateFields(fields bson.M) error {
	for key, value := range fields {
		if err := validateHmdUpdateField(key, value); err != nil {
			return err
		}
	}
	return nil
}

func validateHmdUpdateField(key string, value any) error {
	switch key {
	case "project_name", "city", "building_name", "community_name", "room_type_name", "room_no":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return fmt.Errorf("%s is required", key)
		}
	case "floor_total", "room_count", "hall_count", "bathroom_count", "kitchen_count",
		"floor_no", "area_size", "rent", "deposit", "service_fee", "agency_fee_value":
		intValue, ok := anyInt(value)
		if !ok {
			return fmt.Errorf("%s must be int", key)
		}
		return validateNonNegativeInt(key, intValue)
	case "rent_mode":
		if !RentMode(fmt.Sprint(value)).Valid() {
			return fmt.Errorf("rentMode is invalid")
		}
	case "room_status":
		intValue, ok := anyInt(value)
		if !ok || !RoomStatus(intValue).Valid() {
			return fmt.Errorf("roomStatus is invalid")
		}
	case "orientation":
		if !Orientation(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("orientation is invalid")
		}
	case "decoration_level":
		if !DecorationLevel(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("decorationLevel is invalid")
		}
	case "payment_cycle":
		if !PaymentCycle(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("paymentCycle is invalid")
		}
	case "agency_fee_mode":
		if !AgencyFeeMode(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("agencyFeeMode is invalid")
		}
	case "viewing_time_rule":
		if !ViewingTimeRule(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("viewingTimeRule is invalid")
		}
	case "start_rent_rule":
		if !StartRentRule(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("startRentRule is invalid")
		}
	case "images":
		images, ok := value.([]TaggedImage)
		if !ok {
			return fmt.Errorf("images must be []TaggedImage")
		}
		return validateTaggedImages(images)
	case "room_facilities":
		facilities, ok := value.([]RoomFacility)
		if !ok {
			return fmt.Errorf("roomFacilities must be []RoomFacility")
		}
		return validateRoomFacilities(facilities)
	case "listing_facilities":
		facilities, ok := value.([]ListingFacility)
		if !ok {
			return fmt.Errorf("listingFacilities must be []ListingFacility")
		}
		return validateListingFacilities(facilities)
	}
	return nil
}

func validateHmdRoomShapeAndPrice(
	roomCount int,
	hallCount int,
	bathroomCount int,
	kitchenCount int,
	areaSize int,
	rent int,
	deposit int,
	serviceFee int,
	agencyFeeValue int,
	orientation Orientation,
	decorationLevel DecorationLevel,
	paymentCycle PaymentCycle,
	agencyFeeMode AgencyFeeMode,
	viewingTimeRule ViewingTimeRule,
	startRentRule StartRentRule,
	images []TaggedImage,
	roomFacilities []RoomFacility,
	listingFacilities []ListingFacility,
) error {
	for name, value := range map[string]int{
		"roomCount":      roomCount,
		"hallCount":      hallCount,
		"bathroomCount":  bathroomCount,
		"kitchenCount":   kitchenCount,
		"areaSize":       areaSize,
		"rent":           rent,
		"deposit":        deposit,
		"serviceFee":     serviceFee,
		"agencyFeeValue": agencyFeeValue,
	} {
		if err := validateNonNegativeInt(name, value); err != nil {
			return err
		}
	}
	if !orientation.ValidOptional() {
		return fmt.Errorf("orientation is invalid")
	}
	if !decorationLevel.ValidOptional() {
		return fmt.Errorf("decorationLevel is invalid")
	}
	if !paymentCycle.ValidOptional() {
		return fmt.Errorf("paymentCycle is invalid")
	}
	if !agencyFeeMode.ValidOptional() {
		return fmt.Errorf("agencyFeeMode is invalid")
	}
	if !viewingTimeRule.ValidOptional() {
		return fmt.Errorf("viewingTimeRule is invalid")
	}
	if !startRentRule.ValidOptional() {
		return fmt.Errorf("startRentRule is invalid")
	}
	if err := validateTaggedImages(images); err != nil {
		return err
	}
	if err := validateRoomFacilities(roomFacilities); err != nil {
		return err
	}
	return validateListingFacilities(listingFacilities)
}

func validateNonNegativeInt(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

func validateTaggedImages(images []TaggedImage) error {
	for _, image := range images {
		if image.Tag == "" {
			continue
		}
		if _, ok := validImageTags[image.Tag]; !ok {
			return fmt.Errorf("image tag %q is invalid", image.Tag)
		}
	}
	return nil
}

func validateRoomFacilities(facilities []RoomFacility) error {
	for _, facility := range facilities {
		if _, ok := validRoomFacilities[facility]; !ok {
			return fmt.Errorf("room facility %q is invalid", facility)
		}
	}
	return nil
}

func validateListingFacilities(facilities []ListingFacility) error {
	for _, facility := range facilities {
		if _, ok := validListingFacilities[facility]; !ok {
			return fmt.Errorf("listing facility %q is invalid", facility)
		}
	}
	return nil
}

func anyInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case RoomStatus:
		return int(v), true
	default:
		return 0, false
	}
}
