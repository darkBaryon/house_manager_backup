package hmd

import (
	"fmt"
	"strings"

	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m *HmdCentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("项目信息不能为空")
	}
	if commonmodel.IsBlank(m.ProjectName) || commonmodel.IsBlank(m.City) {
		return fmt.Errorf("项目名称和城市不能为空")
	}
	return nil
}

func (m *HmdBuilding) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("楼栋信息不能为空")
	}
	if m.ProjectID.IsZero() || commonmodel.IsBlank(m.BuildingName) {
		return fmt.Errorf("项目 ID 和楼栋名称不能为空")
	}
	if err := validateNonNegativeInt("floorTotal", m.FloorTotal); err != nil {
		return err
	}
	return validateListingFacilities(m.ListingFacilities)
}

func (m *HmdDecentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("小区信息不能为空")
	}
	if commonmodel.IsBlank(m.CommunityName) || commonmodel.IsBlank(m.City) {
		return fmt.Errorf("小区名称和城市不能为空")
	}
	return nil
}

func (m *HmdRoomTypeCentralized) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("房型信息不能为空")
	}
	if commonmodel.IsBlank(m.RoomTypeName) {
		return fmt.Errorf("房型名称不能为空")
	}
	if m.ProjectID.IsZero() && m.BuildingID.IsZero() {
		return fmt.Errorf("项目 ID 和楼栋 ID 至少填写一个")
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
		return fmt.Errorf("集中式房间信息不能为空")
	}
	if m.ProjectID.IsZero() || m.BuildingID.IsZero() || commonmodel.IsBlank(m.RoomNo) || m.RentMode == "" {
		return fmt.Errorf("项目 ID、楼栋 ID、房间号和出租方式不能为空")
	}
	if !m.RentMode.Valid() {
		return fmt.Errorf("出租方式不合法")
	}
	if !m.RoomStatus.Valid() {
		return fmt.Errorf("房间状态不合法")
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
		return fmt.Errorf("分散式房间信息不能为空")
	}
	if m.DecentralizedID.IsZero() || commonmodel.IsBlank(m.RoomNo) || m.RentMode == "" {
		return fmt.Errorf("小区 ID、房间号和出租方式不能为空")
	}
	if !m.RentMode.Valid() {
		return fmt.Errorf("出租方式不合法")
	}
	if !m.RoomStatus.Valid() {
		return fmt.Errorf("房间状态不合法")
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
			return fmt.Errorf("%s 不能为空", key)
		}
	case "room_type_id":
		objectID, ok := value.(bson.ObjectID)
		if !ok {
			return fmt.Errorf("room_type_id 必须是对象 ID")
		}
		if objectID.IsZero() {
			return fmt.Errorf("room_type_id 不能为空")
		}
	case "floor_total", "room_count", "hall_count", "bathroom_count", "kitchen_count",
		"floor_no", "area_size", "rent", "deposit", "service_fee", "agency_fee_value":
		intValue, ok := anyInt(value)
		if !ok {
			return fmt.Errorf("%s 必须是整数", key)
		}
		return validateNonNegativeInt(key, intValue)
	case "rent_mode":
		if !RentMode(fmt.Sprint(value)).Valid() {
			return fmt.Errorf("出租方式不合法")
		}
	case "room_status":
		intValue, ok := anyInt(value)
		if !ok || !RoomStatus(intValue).Valid() {
			return fmt.Errorf("房间状态不合法")
		}
	case "orientation":
		if !Orientation(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("朝向不合法")
		}
	case "decoration_level":
		if !DecorationLevel(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("装修等级不合法")
		}
	case "payment_cycle":
		if !PaymentCycle(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("付款周期不合法")
		}
	case "agency_fee_mode":
		if !AgencyFeeMode(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("中介费模式不合法")
		}
	case "viewing_time_rule":
		if !ViewingTimeRule(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("看房时间规则不合法")
		}
	case "start_rent_rule":
		if !StartRentRule(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("起租规则不合法")
		}
	case "images":
		images, ok := value.([]TaggedImage)
		if !ok {
			return fmt.Errorf("images 必须是图片数组")
		}
		return validateTaggedImages(images)
	case "room_facilities":
		facilities, ok := value.([]RoomFacility)
		if !ok {
			return fmt.Errorf("roomFacilities 必须是房间设施数组")
		}
		return validateRoomFacilities(facilities)
	case "listing_facilities":
		facilities, ok := value.([]ListingFacility)
		if !ok {
			return fmt.Errorf("listingFacilities 必须是房源设施数组")
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
		return fmt.Errorf("朝向不合法")
	}
	if !decorationLevel.ValidOptional() {
		return fmt.Errorf("装修等级不合法")
	}
	if !paymentCycle.ValidOptional() {
		return fmt.Errorf("付款周期不合法")
	}
	if !agencyFeeMode.ValidOptional() {
		return fmt.Errorf("中介费模式不合法")
	}
	if !viewingTimeRule.ValidOptional() {
		return fmt.Errorf("看房时间规则不合法")
	}
	if !startRentRule.ValidOptional() {
		return fmt.Errorf("起租规则不合法")
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
		return fmt.Errorf("%s 不能小于 0", name)
	}
	return nil
}

func validateTaggedImages(images []TaggedImage) error {
	for _, image := range images {
		if image.Tag == "" {
			continue
		}
		if _, ok := validImageTags[image.Tag]; !ok {
			return fmt.Errorf("图片标签 %q 不合法", image.Tag)
		}
	}
	return nil
}

func validateRoomFacilities(facilities []RoomFacility) error {
	for _, facility := range facilities {
		if _, ok := validRoomFacilities[facility]; !ok {
			return fmt.Errorf("房间设施 %q 不合法", facility)
		}
	}
	return nil
}

func validateListingFacilities(facilities []ListingFacility) error {
	for _, facility := range facilities {
		if _, ok := validListingFacilities[facility]; !ok {
			return fmt.Errorf("房源设施 %q 不合法", facility)
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
