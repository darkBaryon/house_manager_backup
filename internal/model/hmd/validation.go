package hmd

import (
	"fmt"

	commonmodel "house-manager/internal/model/common"
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
		m.ViewingTimeRule,
		m.StartRentRule,
		m.Images,
		m.RoomFacilities,
		m.ListingFacilities,
	)
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
		if err := validateLayoutCount(name, value); err != nil {
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

func validateLayoutCount(name string, value int) error {
	if value < UnknownLayoutCount {
		return fmt.Errorf("%s 不能小于 %d", name, UnknownLayoutCount)
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
