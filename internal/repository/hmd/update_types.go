package hmd

import (
	"fmt"
	"strings"

	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BuildingBaseInfoUpdate struct {
	BuildingName      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []hmdmodel.ListingFacility
}

type CentralizedBaseInfoUpdate struct {
	ProjectName string
	City        string
	District    string
	AddressText string
	Geo         *hmdmodel.GeoPoint
	BrandName   string
}

type DecentralizedBaseInfoUpdate struct {
	CommunityName string
	City          string
	District      string
	BizArea       string
	AddressText   string
	Geo           *hmdmodel.GeoPoint
	SubwayStation string
}

type RoomTypeCentralizedBaseInfoUpdate struct {
	RoomTypeName    string
	RoomCount       int
	HallCount       int
	BathroomCount   int
	KitchenCount    int
	AreaSize        int
	Orientation     hmdmodel.Orientation
	DecorationLevel hmdmodel.DecorationLevel
	PaymentCycle    hmdmodel.PaymentCycle
	Rent            int
	Deposit         int
	ServiceFee      int
	AgencyFeeMode   hmdmodel.AgencyFeeMode
	AgencyFeeValue  int
	Images          []hmdmodel.TaggedImage
	RoomFacilities  []hmdmodel.RoomFacility
}

type RoomCentralizedBaseInfoUpdate struct {
	RoomTypeID        bson.ObjectID
	RoomNo            string
	FloorNo           int
	RentMode          hmdmodel.RentMode
	LayoutText        string
	RoomCount         int
	HallCount         int
	BathroomCount     int
	KitchenCount      int
	AreaSize          int
	Orientation       hmdmodel.Orientation
	DecorationLevel   hmdmodel.DecorationLevel
	PaymentCycle      hmdmodel.PaymentCycle
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     hmdmodel.AgencyFeeMode
	AgencyFeeValue    int
	ViewingTimeRule   hmdmodel.ViewingTimeRule
	StartRentRule     hmdmodel.StartRentRule
	Images            []hmdmodel.TaggedImage
	RoomFacilities    []hmdmodel.RoomFacility
	ListingFacilities []hmdmodel.ListingFacility
}

type RoomDecentralizedBaseInfoUpdate struct {
	RoomNo            string
	FloorNo           int
	RentMode          hmdmodel.RentMode
	LayoutText        string
	RoomCount         int
	HallCount         int
	BathroomCount     int
	KitchenCount      int
	AreaSize          int
	Orientation       hmdmodel.Orientation
	DecorationLevel   hmdmodel.DecorationLevel
	PaymentCycle      hmdmodel.PaymentCycle
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     hmdmodel.AgencyFeeMode
	AgencyFeeValue    int
	ViewingTimeRule   hmdmodel.ViewingTimeRule
	StartRentRule     hmdmodel.StartRentRule
	Images            []hmdmodel.TaggedImage
	RoomFacilities    []hmdmodel.RoomFacility
	ListingFacilities []hmdmodel.ListingFacility
}

func centralizedBaseInfoUpdateFields(update CentralizedBaseInfoUpdate) (bson.M, error) {
	if err := validateRequiredText("project_name", update.ProjectName); err != nil {
		return nil, err
	}
	if err := validateRequiredText("city", update.City); err != nil {
		return nil, err
	}
	return bson.M{
		"project_name": update.ProjectName,
		"city":         update.City,
		"district":     update.District,
		"address_text": update.AddressText,
		"geo":          update.Geo,
		"brand_name":   update.BrandName,
	}, nil
}

func buildingBaseInfoUpdateFields(update BuildingBaseInfoUpdate) (bson.M, error) {
	if err := validateRequiredText("building_name", update.BuildingName); err != nil {
		return nil, err
	}
	if err := validateNonNegativeInt("floor_total", update.FloorTotal); err != nil {
		return nil, err
	}
	if err := validateListingFacilities(update.ListingFacilities); err != nil {
		return nil, err
	}
	return bson.M{
		"building_name":      update.BuildingName,
		"floor_total":        update.FloorTotal,
		"manager_name":       update.ManagerName,
		"manager_phone":      update.ManagerPhone,
		"photos":             update.Photos,
		"listing_facilities": update.ListingFacilities,
	}, nil
}

func decentralizedBaseInfoUpdateFields(update DecentralizedBaseInfoUpdate) (bson.M, error) {
	if err := validateRequiredText("community_name", update.CommunityName); err != nil {
		return nil, err
	}
	if err := validateRequiredText("city", update.City); err != nil {
		return nil, err
	}
	return bson.M{
		"community_name": update.CommunityName,
		"city":           update.City,
		"district":       update.District,
		"biz_area":       update.BizArea,
		"address_text":   update.AddressText,
		"geo":            update.Geo,
		"subway_station": update.SubwayStation,
	}, nil
}

func roomTypeCentralizedBaseInfoUpdateFields(update RoomTypeCentralizedBaseInfoUpdate) (bson.M, error) {
	if err := validateRequiredText("room_type_name", update.RoomTypeName); err != nil {
		return nil, err
	}
	if err := validateRoomShapeAndPrice(
		update.RoomCount,
		update.HallCount,
		update.BathroomCount,
		update.KitchenCount,
		update.AreaSize,
		update.Rent,
		update.Deposit,
		update.ServiceFee,
		update.AgencyFeeValue,
		update.Orientation,
		update.DecorationLevel,
		update.PaymentCycle,
		update.AgencyFeeMode,
		"",
		"",
		update.Images,
		update.RoomFacilities,
		nil,
	); err != nil {
		return nil, err
	}
	return bson.M{
		"room_type_name":   update.RoomTypeName,
		"room_count":       update.RoomCount,
		"hall_count":       update.HallCount,
		"bathroom_count":   update.BathroomCount,
		"kitchen_count":    update.KitchenCount,
		"area_size":        update.AreaSize,
		"orientation":      update.Orientation,
		"decoration_level": update.DecorationLevel,
		"payment_cycle":    update.PaymentCycle,
		"rent":             update.Rent,
		"deposit":          update.Deposit,
		"service_fee":      update.ServiceFee,
		"agency_fee_mode":  update.AgencyFeeMode,
		"agency_fee_value": update.AgencyFeeValue,
		"images":           update.Images,
		"room_facilities":  update.RoomFacilities,
	}, nil
}

func roomCentralizedBaseInfoUpdateFields(update RoomCentralizedBaseInfoUpdate) (bson.M, error) {
	if err := validateRoomTypeID(update.RoomTypeID); err != nil {
		return nil, err
	}
	if err := validateRoomBaseInfo(
		update.RoomNo,
		update.FloorNo,
		update.RentMode,
		update.RoomCount,
		update.HallCount,
		update.BathroomCount,
		update.KitchenCount,
		update.AreaSize,
		update.Rent,
		update.Deposit,
		update.ServiceFee,
		update.AgencyFeeValue,
		update.Orientation,
		update.DecorationLevel,
		update.PaymentCycle,
		update.AgencyFeeMode,
		update.ViewingTimeRule,
		update.StartRentRule,
		update.Images,
		update.RoomFacilities,
		update.ListingFacilities,
	); err != nil {
		return nil, err
	}
	fields := roomBaseInfoFieldsFromUpdate(roomBaseInfoUpdate{
		RoomNo:            update.RoomNo,
		FloorNo:           update.FloorNo,
		RentMode:          update.RentMode,
		LayoutText:        update.LayoutText,
		RoomCount:         update.RoomCount,
		HallCount:         update.HallCount,
		BathroomCount:     update.BathroomCount,
		KitchenCount:      update.KitchenCount,
		AreaSize:          update.AreaSize,
		Orientation:       update.Orientation,
		DecorationLevel:   update.DecorationLevel,
		PaymentCycle:      update.PaymentCycle,
		Rent:              update.Rent,
		Deposit:           update.Deposit,
		ServiceFee:        update.ServiceFee,
		AgencyFeeMode:     update.AgencyFeeMode,
		AgencyFeeValue:    update.AgencyFeeValue,
		ViewingTimeRule:   update.ViewingTimeRule,
		StartRentRule:     update.StartRentRule,
		Images:            update.Images,
		RoomFacilities:    update.RoomFacilities,
		ListingFacilities: update.ListingFacilities,
	})
	fields["room_type_id"] = update.RoomTypeID
	return fields, nil
}

func roomDecentralizedBaseInfoUpdateFields(update RoomDecentralizedBaseInfoUpdate) (bson.M, error) {
	if err := validateRoomBaseInfo(
		update.RoomNo,
		update.FloorNo,
		update.RentMode,
		update.RoomCount,
		update.HallCount,
		update.BathroomCount,
		update.KitchenCount,
		update.AreaSize,
		update.Rent,
		update.Deposit,
		update.ServiceFee,
		update.AgencyFeeValue,
		update.Orientation,
		update.DecorationLevel,
		update.PaymentCycle,
		update.AgencyFeeMode,
		update.ViewingTimeRule,
		update.StartRentRule,
		update.Images,
		update.RoomFacilities,
		update.ListingFacilities,
	); err != nil {
		return nil, err
	}
	return roomBaseInfoFieldsFromUpdate(roomBaseInfoUpdate{
		RoomNo:            update.RoomNo,
		FloorNo:           update.FloorNo,
		RentMode:          update.RentMode,
		LayoutText:        update.LayoutText,
		RoomCount:         update.RoomCount,
		HallCount:         update.HallCount,
		BathroomCount:     update.BathroomCount,
		KitchenCount:      update.KitchenCount,
		AreaSize:          update.AreaSize,
		Orientation:       update.Orientation,
		DecorationLevel:   update.DecorationLevel,
		PaymentCycle:      update.PaymentCycle,
		Rent:              update.Rent,
		Deposit:           update.Deposit,
		ServiceFee:        update.ServiceFee,
		AgencyFeeMode:     update.AgencyFeeMode,
		AgencyFeeValue:    update.AgencyFeeValue,
		ViewingTimeRule:   update.ViewingTimeRule,
		StartRentRule:     update.StartRentRule,
		Images:            update.Images,
		RoomFacilities:    update.RoomFacilities,
		ListingFacilities: update.ListingFacilities,
	}), nil
}

type roomBaseInfoUpdate struct {
	RoomNo            string
	FloorNo           int
	RentMode          hmdmodel.RentMode
	LayoutText        string
	RoomCount         int
	HallCount         int
	BathroomCount     int
	KitchenCount      int
	AreaSize          int
	Orientation       hmdmodel.Orientation
	DecorationLevel   hmdmodel.DecorationLevel
	PaymentCycle      hmdmodel.PaymentCycle
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     hmdmodel.AgencyFeeMode
	AgencyFeeValue    int
	ViewingTimeRule   hmdmodel.ViewingTimeRule
	StartRentRule     hmdmodel.StartRentRule
	Images            []hmdmodel.TaggedImage
	RoomFacilities    []hmdmodel.RoomFacility
	ListingFacilities []hmdmodel.ListingFacility
}

func roomBaseInfoFieldsFromUpdate(update roomBaseInfoUpdate) bson.M {
	return bson.M{
		"room_no":            update.RoomNo,
		"floor_no":           update.FloorNo,
		"rent_mode":          update.RentMode,
		"layout_text":        update.LayoutText,
		"room_count":         update.RoomCount,
		"hall_count":         update.HallCount,
		"bathroom_count":     update.BathroomCount,
		"kitchen_count":      update.KitchenCount,
		"area_size":          update.AreaSize,
		"orientation":        update.Orientation,
		"decoration_level":   update.DecorationLevel,
		"payment_cycle":      update.PaymentCycle,
		"rent":               update.Rent,
		"deposit":            update.Deposit,
		"service_fee":        update.ServiceFee,
		"agency_fee_mode":    update.AgencyFeeMode,
		"agency_fee_value":   update.AgencyFeeValue,
		"viewing_time_rule":  update.ViewingTimeRule,
		"start_rent_rule":    update.StartRentRule,
		"images":             update.Images,
		"room_facilities":    update.RoomFacilities,
		"listing_facilities": update.ListingFacilities,
	}
}

func validateRequiredText(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 不能为空", name)
	}
	return nil
}

func validateRoomTypeID(value bson.ObjectID) error {
	if value.IsZero() {
		return fmt.Errorf("room_type_id 不能为空")
	}
	return nil
}

func validateNonNegativeInt(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s 不能小于 0", name)
	}
	return nil
}

func validateLayoutCount(name string, value int) error {
	if value < hmdmodel.UnknownLayoutCount {
		return fmt.Errorf("%s 不能小于 %d", name, hmdmodel.UnknownLayoutCount)
	}
	return nil
}

func validateRoomBaseInfo(
	roomNo string,
	floorNo int,
	rentMode hmdmodel.RentMode,
	roomCount int,
	hallCount int,
	bathroomCount int,
	kitchenCount int,
	areaSize int,
	rent int,
	deposit int,
	serviceFee int,
	agencyFeeValue int,
	orientation hmdmodel.Orientation,
	decorationLevel hmdmodel.DecorationLevel,
	paymentCycle hmdmodel.PaymentCycle,
	agencyFeeMode hmdmodel.AgencyFeeMode,
	viewingTimeRule hmdmodel.ViewingTimeRule,
	startRentRule hmdmodel.StartRentRule,
	images []hmdmodel.TaggedImage,
	roomFacilities []hmdmodel.RoomFacility,
	listingFacilities []hmdmodel.ListingFacility,
) error {
	if err := validateRequiredText("room_no", roomNo); err != nil {
		return err
	}
	if err := validateNonNegativeInt("floor_no", floorNo); err != nil {
		return err
	}
	if !rentMode.Valid() {
		return fmt.Errorf("出租方式不合法")
	}
	return validateRoomShapeAndPrice(
		roomCount,
		hallCount,
		bathroomCount,
		kitchenCount,
		areaSize,
		rent,
		deposit,
		serviceFee,
		agencyFeeValue,
		orientation,
		decorationLevel,
		paymentCycle,
		agencyFeeMode,
		viewingTimeRule,
		startRentRule,
		images,
		roomFacilities,
		listingFacilities,
	)
}

func validateRoomShapeAndPrice(
	roomCount int,
	hallCount int,
	bathroomCount int,
	kitchenCount int,
	areaSize int,
	rent int,
	deposit int,
	serviceFee int,
	agencyFeeValue int,
	orientation hmdmodel.Orientation,
	decorationLevel hmdmodel.DecorationLevel,
	paymentCycle hmdmodel.PaymentCycle,
	agencyFeeMode hmdmodel.AgencyFeeMode,
	viewingTimeRule hmdmodel.ViewingTimeRule,
	startRentRule hmdmodel.StartRentRule,
	images []hmdmodel.TaggedImage,
	roomFacilities []hmdmodel.RoomFacility,
	listingFacilities []hmdmodel.ListingFacility,
) error {
	for name, value := range map[string]int{
		"room_count":       roomCount,
		"hall_count":       hallCount,
		"bathroom_count":   bathroomCount,
		"kitchen_count":    kitchenCount,
		"area_size":        areaSize,
		"rent":             rent,
		"deposit":          deposit,
		"service_fee":      serviceFee,
		"agency_fee_value": agencyFeeValue,
	} {
		if name == "room_count" || name == "hall_count" || name == "bathroom_count" || name == "kitchen_count" {
			if err := validateLayoutCount(name, value); err != nil {
				return err
			}
			continue
		}
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

func validateTaggedImages(images []hmdmodel.TaggedImage) error {
	for _, image := range images {
		if !image.Tag.ValidOptional() {
			return fmt.Errorf("图片标签 %q 不合法", image.Tag)
		}
	}
	return nil
}

func validateRoomFacilities(facilities []hmdmodel.RoomFacility) error {
	for _, facility := range facilities {
		if !facility.Valid() {
			return fmt.Errorf("房间设施 %q 不合法", facility)
		}
	}
	return nil
}

func validateListingFacilities(facilities []hmdmodel.ListingFacility) error {
	for _, facility := range facilities {
		if !facility.Valid() {
			return fmt.Errorf("房源设施 %q 不合法", facility)
		}
	}
	return nil
}
