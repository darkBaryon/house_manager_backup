package hmd

type RentMode string

const (
	RentModeWhole  RentMode = "whole"
	RentModeShared RentMode = "shared"
)

type RoomStatus int

const (
	RoomStatusUnspecified RoomStatus = 0
	RoomStatusAvailable   RoomStatus = 1
	RoomStatusRented      RoomStatus = 2
	RoomStatusOccupied    RoomStatus = 3
	RoomStatusOffline     RoomStatus = -1
)

type DecorationLevel string

const (
	DecorationLevelRough  DecorationLevel = "rough"
	DecorationLevelSimple DecorationLevel = "simple"
	DecorationLevelFine   DecorationLevel = "fine"
	DecorationLevelLuxury DecorationLevel = "luxury"
)

type PaymentCycle string

const (
	PaymentCycleMonthly    PaymentCycle = "monthly"
	PaymentCycleQuarterly  PaymentCycle = "quarterly"
	PaymentCycleHalfYearly PaymentCycle = "half_yearly"
	PaymentCycleYearly     PaymentCycle = "yearly"
	PaymentCycleCustom     PaymentCycle = "custom"
)

type AgencyFeeMode string

const (
	AgencyFeeModeNone                  AgencyFeeMode = "none"
	AgencyFeeModeFixed                 AgencyFeeMode = "fixed"
	AgencyFeeModeMonthlyRentMultiplier AgencyFeeMode = "monthly_rent_multiple"
)

type ViewingTimeRule string

const (
	ViewingTimeRuleWeekendOnly         ViewingTimeRule = "weekend_only"
	ViewingTimeRuleWorkdayOnly         ViewingTimeRule = "workday_only"
	ViewingTimeRuleAnytime             ViewingTimeRule = "anytime"
	ViewingTimeRuleWorkdayNightWeekend ViewingTimeRule = "workday_night_weekend"
)

type StartRentRule string

const (
	StartRentRuleLongHalfYear    StartRentRule = "long_half_year"
	StartRentRuleLongOneYear     StartRentRule = "long_one_year"
	StartRentRuleShortOneMonth   StartRentRule = "short_one_month"
	StartRentRuleShortThreeMonth StartRentRule = "short_three_month"
	StartRentRuleDaily           StartRentRule = "daily"
)

type Orientation string

const (
	OrientationEast       Orientation = "east"
	OrientationSouth      Orientation = "south"
	OrientationWest       Orientation = "west"
	OrientationNorth      Orientation = "north"
	OrientationSoutheast  Orientation = "southeast"
	OrientationNortheast  Orientation = "northeast"
	OrientationSouthwest  Orientation = "southwest"
	OrientationNorthwest  Orientation = "northwest"
	OrientationSouthNorth Orientation = "south_north"
	OrientationUnknown    Orientation = "unknown"
)

type ListingFacility string

const (
	ListingFacilityElevator           ListingFacility = "elevator"
	ListingFacilitySubway             ListingFacility = "subway"
	ListingFacilityConvenienceStore   ListingFacility = "convenience_store"
	ListingFacilityParking            ListingFacility = "parking"
	ListingFacilityGym                ListingFacility = "gym"
	ListingFacilityActivityArea       ListingFacility = "activity_area"
	ListingFacilitySecurityMonitoring ListingFacility = "security_monitoring"
	ListingFacilityBookBar            ListingFacility = "book_bar"
	ListingFacilityBarCounter         ListingFacility = "bar_counter"
	ListingFacilityLounge             ListingFacility = "lounge"
	ListingFacilityBilliards          ListingFacility = "billiards"
	ListingFacilityDIYDiningBar       ListingFacility = "diy_dining_bar"
	ListingFacilityLaundryRoom        ListingFacility = "laundry_room"
	ListingFacilityTableSoccer        ListingFacility = "table_soccer"
	ListingFacilitySkyGarden          ListingFacility = "sky_garden"
	ListingFacilityCinemaArea         ListingFacility = "cinema_area"
	ListingFacilityFrontDesk          ListingFacility = "front_desk"
	ListingFacilityReceptionArea      ListingFacility = "reception_area"
	ListingFacilityDanceRoom          ListingFacility = "dance_room"
	ListingFacilityLocker             ListingFacility = "locker"
	ListingFacilityPetFriendly        ListingFacility = "pet_friendly"
)

type RoomFacility string

const (
	RoomFacilitySmartLock       RoomFacility = "smart_lock"
	RoomFacilityBed             RoomFacility = "bed"
	RoomFacilityWardrobe        RoomFacility = "wardrobe"
	RoomFacilityDeskChair       RoomFacility = "desk_chair"
	RoomFacilityHeating         RoomFacility = "heating"
	RoomFacilityGas             RoomFacility = "gas"
	RoomFacilityBroadband       RoomFacility = "broadband"
	RoomFacilityTV              RoomFacility = "tv"
	RoomFacilityFridge          RoomFacility = "fridge"
	RoomFacilityWashingMachine  RoomFacility = "washing_machine"
	RoomFacilityAirConditioner  RoomFacility = "air_conditioner"
	RoomFacilityWaterHeater     RoomFacility = "water_heater"
	RoomFacilityMicrowave       RoomFacility = "microwave"
	RoomFacilityRangeHood       RoomFacility = "range_hood"
	RoomFacilityInductionCooker RoomFacility = "induction_cooker"
	RoomFacilityBalcony         RoomFacility = "balcony"
	RoomFacilityCookingAllowed  RoomFacility = "cooking_allowed"
	RoomFacilityPrivateBathroom RoomFacility = "private_bathroom"
	RoomFacilitySofa            RoomFacility = "sofa"
	RoomFacilityWaterPurifier   RoomFacility = "water_purifier"
)

type ImageTag string

const (
	ImageTagUnknown              ImageTag = "unknown"
	ImageTagBedroom              ImageTag = "bedroom"
	ImageTagMasterBedroom        ImageTag = "master_bedroom"
	ImageTagSecondaryBedroom     ImageTag = "secondary_bedroom"
	ImageTagLivingRoom           ImageTag = "living_room"
	ImageTagDiningRoom           ImageTag = "dining_room"
	ImageTagHallway              ImageTag = "hallway"
	ImageTagKitchen              ImageTag = "kitchen"
	ImageTagBathroom             ImageTag = "bathroom"
	ImageTagFloorPlanStandard    ImageTag = "floor_plan_standard"
	ImageTagFloorPlanNonStandard ImageTag = "floor_plan_non_standard"
	ImageTagExterior             ImageTag = "exterior"
)

var validRentModes = map[RentMode]struct{}{
	RentModeWhole:  {},
	RentModeShared: {},
}

var validRoomStatuses = map[RoomStatus]struct{}{
	RoomStatusUnspecified: {},
	RoomStatusAvailable:   {},
	RoomStatusRented:      {},
	RoomStatusOccupied:    {},
	RoomStatusOffline:     {},
}

var validDecorationLevels = map[DecorationLevel]struct{}{
	DecorationLevelRough:  {},
	DecorationLevelSimple: {},
	DecorationLevelFine:   {},
	DecorationLevelLuxury: {},
}

var validPaymentCycles = map[PaymentCycle]struct{}{
	PaymentCycleMonthly:    {},
	PaymentCycleQuarterly:  {},
	PaymentCycleHalfYearly: {},
	PaymentCycleYearly:     {},
	PaymentCycleCustom:     {},
}

var validAgencyFeeModes = map[AgencyFeeMode]struct{}{
	AgencyFeeModeNone:                  {},
	AgencyFeeModeFixed:                 {},
	AgencyFeeModeMonthlyRentMultiplier: {},
}

var validViewingTimeRules = map[ViewingTimeRule]struct{}{
	ViewingTimeRuleWeekendOnly:         {},
	ViewingTimeRuleWorkdayOnly:         {},
	ViewingTimeRuleAnytime:             {},
	ViewingTimeRuleWorkdayNightWeekend: {},
}

var validStartRentRules = map[StartRentRule]struct{}{
	StartRentRuleLongHalfYear:    {},
	StartRentRuleLongOneYear:     {},
	StartRentRuleShortOneMonth:   {},
	StartRentRuleShortThreeMonth: {},
	StartRentRuleDaily:           {},
}

var validOrientations = map[Orientation]struct{}{
	OrientationEast:       {},
	OrientationSouth:      {},
	OrientationWest:       {},
	OrientationNorth:      {},
	OrientationSoutheast:  {},
	OrientationNortheast:  {},
	OrientationSouthwest:  {},
	OrientationNorthwest:  {},
	OrientationSouthNorth: {},
	OrientationUnknown:    {},
}

var validListingFacilities = map[ListingFacility]struct{}{
	ListingFacilityElevator:           {},
	ListingFacilitySubway:             {},
	ListingFacilityConvenienceStore:   {},
	ListingFacilityParking:            {},
	ListingFacilityGym:                {},
	ListingFacilityActivityArea:       {},
	ListingFacilitySecurityMonitoring: {},
	ListingFacilityBookBar:            {},
	ListingFacilityBarCounter:         {},
	ListingFacilityLounge:             {},
	ListingFacilityBilliards:          {},
	ListingFacilityDIYDiningBar:       {},
	ListingFacilityLaundryRoom:        {},
	ListingFacilityTableSoccer:        {},
	ListingFacilitySkyGarden:          {},
	ListingFacilityCinemaArea:         {},
	ListingFacilityFrontDesk:          {},
	ListingFacilityReceptionArea:      {},
	ListingFacilityDanceRoom:          {},
	ListingFacilityLocker:             {},
	ListingFacilityPetFriendly:        {},
}

var validRoomFacilities = map[RoomFacility]struct{}{
	RoomFacilitySmartLock:       {},
	RoomFacilityBed:             {},
	RoomFacilityWardrobe:        {},
	RoomFacilityDeskChair:       {},
	RoomFacilityHeating:         {},
	RoomFacilityGas:             {},
	RoomFacilityBroadband:       {},
	RoomFacilityTV:              {},
	RoomFacilityFridge:          {},
	RoomFacilityWashingMachine:  {},
	RoomFacilityAirConditioner:  {},
	RoomFacilityWaterHeater:     {},
	RoomFacilityMicrowave:       {},
	RoomFacilityRangeHood:       {},
	RoomFacilityInductionCooker: {},
	RoomFacilityBalcony:         {},
	RoomFacilityCookingAllowed:  {},
	RoomFacilityPrivateBathroom: {},
	RoomFacilitySofa:            {},
	RoomFacilityWaterPurifier:   {},
}

var validImageTags = map[ImageTag]struct{}{
	ImageTagUnknown:              {},
	ImageTagBedroom:              {},
	ImageTagMasterBedroom:        {},
	ImageTagSecondaryBedroom:     {},
	ImageTagLivingRoom:           {},
	ImageTagDiningRoom:           {},
	ImageTagHallway:              {},
	ImageTagKitchen:              {},
	ImageTagBathroom:             {},
	ImageTagFloorPlanStandard:    {},
	ImageTagFloorPlanNonStandard: {},
	ImageTagExterior:             {},
}

func (v RentMode) Valid() bool {
	_, ok := validRentModes[v]
	return ok
}

func (v RoomStatus) Valid() bool {
	_, ok := validRoomStatuses[v]
	return ok
}

func (v DecorationLevel) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validDecorationLevels[v]
	return ok
}

func (v PaymentCycle) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validPaymentCycles[v]
	return ok
}

func (v AgencyFeeMode) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validAgencyFeeModes[v]
	return ok
}

func (v ViewingTimeRule) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validViewingTimeRules[v]
	return ok
}

func (v StartRentRule) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validStartRentRules[v]
	return ok
}

func (v Orientation) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validOrientations[v]
	return ok
}

func (v ListingFacility) Valid() bool {
	_, ok := validListingFacilities[v]
	return ok
}

func (v RoomFacility) Valid() bool {
	_, ok := validRoomFacilities[v]
	return ok
}

func (v ImageTag) ValidOptional() bool {
	if v == "" {
		return true
	}
	_, ok := validImageTags[v]
	return ok
}

func IsValidRoomStatus(roomStatus int) bool {
	return RoomStatus(roomStatus).Valid()
}

func IsValidRoomStatusUpdateTarget(roomStatus int) bool {
	status := RoomStatus(roomStatus)
	return status != RoomStatusUnspecified && status.Valid()
}
