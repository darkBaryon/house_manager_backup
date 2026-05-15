package hmd

import (
	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionHmdCentralized         = "hs_hmd_centralized"
	CollectionHmdBuilding            = "hs_hmd_building"
	CollectionHmdDecentralized       = "hs_hmd_decentralized"
	CollectionHmdRoomTypeCentralized = "hs_hmd_room_type_centralized"
	CollectionHmdRoomCentralized     = "hs_hmd_room_centralized"
	CollectionHmdRoomDecentralized   = "hs_hmd_room_decentralized"
)

// GeoPoint 通用经纬度坐标。
type GeoPoint struct {
	Lng float64 `bson:"lng" json:"lng"`
	Lat float64 `bson:"lat" json:"lat"`
}

// TaggedImage 通用图片对象。
type TaggedImage struct {
	URL string   `bson:"url" json:"url"`
	Tag ImageTag `bson:"tag" json:"tag"`
}

// HmdCentralized 对应 hs_hmd_centralized，集中式项目主档。
type HmdCentralized struct {
	commonmodel.CommonFields `bson:",inline"`

	ProjectName string    `bson:"project_name" json:"projectName"`
	ProjectCode string    `bson:"project_code" json:"projectCode"`
	City        string    `bson:"city" json:"city"`
	District    string    `bson:"district" json:"district"`
	AddressText string    `bson:"address_text" json:"addressText"`
	Geo         *GeoPoint `bson:"geo,omitempty" json:"geo,omitempty"`
	BrandName   string    `bson:"brand_name" json:"brandName"`
}

// HmdBuilding 对应 hs_hmd_building，楼栋主档。
type HmdBuilding struct {
	commonmodel.CommonFields `bson:",inline"`

	ProjectID         bson.ObjectID     `bson:"project_id" json:"projectId"`
	BuildingName      string            `bson:"building_name" json:"buildingName"`
	BuildingCode      string            `bson:"building_code" json:"buildingCode"`
	FloorTotal        int               `bson:"floor_total" json:"floorTotal"`
	ManagerName       string            `bson:"manager_name" json:"managerName"`
	ManagerPhone      string            `bson:"manager_phone" json:"managerPhone"`
	Photos            []string          `bson:"photos" json:"photos"`
	ListingFacilities []ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
}

// HmdDecentralized 对应 hs_hmd_decentralized，分散式主档。
type HmdDecentralized struct {
	commonmodel.CommonFields `bson:",inline"`

	CommunityName string    `bson:"community_name" json:"communityName"`
	City          string    `bson:"city" json:"city"`
	District      string    `bson:"district" json:"district"`
	BizArea       string    `bson:"biz_area" json:"bizArea"`
	AddressText   string    `bson:"address_text" json:"addressText"`
	Geo           *GeoPoint `bson:"geo,omitempty" json:"geo,omitempty"`
	SubwayStation string    `bson:"subway_station" json:"subwayStation"`
}

// HmdRoomTypeCentralized 对应 hs_hmd_room_type_centralized，集中式房型模板。
type HmdRoomTypeCentralized struct {
	commonmodel.CommonFields `bson:",inline"`

	ProjectID       bson.ObjectID   `bson:"project_id,omitempty" json:"projectId,omitempty"`
	BuildingID      bson.ObjectID   `bson:"building_id,omitempty" json:"buildingId,omitempty"`
	RoomTypeName    string          `bson:"room_type_name" json:"roomTypeName"`
	RoomCount       int             `bson:"room_count" json:"roomCount"`
	HallCount       int             `bson:"hall_count" json:"hallCount"`
	BathroomCount   int             `bson:"bathroom_count" json:"bathroomCount"`
	KitchenCount    int             `bson:"kitchen_count" json:"kitchenCount"`
	AreaSize        int             `bson:"area_size" json:"areaSize"`
	Orientation     Orientation     `bson:"orientation" json:"orientation"`
	DecorationLevel DecorationLevel `bson:"decoration_level" json:"decorationLevel"`
	PaymentCycle    PaymentCycle    `bson:"payment_cycle" json:"paymentCycle"`
	Rent            int             `bson:"rent" json:"rent"`
	Deposit         int             `bson:"deposit" json:"deposit"`
	ServiceFee      int             `bson:"service_fee" json:"serviceFee"`
	AgencyFeeMode   AgencyFeeMode   `bson:"agency_fee_mode" json:"agencyFeeMode"`
	AgencyFeeValue  int             `bson:"agency_fee_value" json:"agencyFeeValue"`
	Images          []TaggedImage   `bson:"images" json:"images"`
	RoomFacilities  []RoomFacility  `bson:"room_facilities" json:"roomFacilities"`
}

// HmdRoomCentralized 对应 hs_hmd_room_centralized，集中式房间实例。
type HmdRoomCentralized struct {
	commonmodel.CommonFields `bson:",inline"`

	ProjectID         bson.ObjectID     `bson:"project_id,omitempty" json:"projectId,omitempty"`
	BuildingID        bson.ObjectID     `bson:"building_id,omitempty" json:"buildingId,omitempty"`
	RoomTypeID        bson.ObjectID     `bson:"room_type_id,omitempty" json:"roomTypeId,omitempty"`
	RoomNo            string            `bson:"room_no" json:"roomNo"`
	FloorNo           int               `bson:"floor_no" json:"floorNo"`
	RentMode          RentMode          `bson:"rent_mode" json:"rentMode"`
	LayoutText        string            `bson:"layout_text" json:"layoutText"`
	AreaSize          int               `bson:"area_size" json:"areaSize"`
	Orientation       Orientation       `bson:"orientation" json:"orientation"`
	DecorationLevel   DecorationLevel   `bson:"decoration_level" json:"decorationLevel"`
	PaymentCycle      PaymentCycle      `bson:"payment_cycle" json:"paymentCycle"`
	Rent              int               `bson:"rent" json:"rent"`
	Deposit           int               `bson:"deposit" json:"deposit"`
	ServiceFee        int               `bson:"service_fee" json:"serviceFee"`
	AgencyFeeMode     AgencyFeeMode     `bson:"agency_fee_mode" json:"agencyFeeMode"`
	AgencyFeeValue    int               `bson:"agency_fee_value" json:"agencyFeeValue"`
	RoomStatus        RoomStatus        `bson:"room_status" json:"roomStatus"`
	ViewingTimeRule   ViewingTimeRule   `bson:"viewing_time_rule" json:"viewingTimeRule"`
	StartRentRule     StartRentRule     `bson:"start_rent_rule" json:"startRentRule"`
	Images            []TaggedImage     `bson:"images" json:"images"`
	RoomFacilities    []RoomFacility    `bson:"room_facilities" json:"roomFacilities"`
	ListingFacilities []ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
}

// HmdRoomDecentralized 对应 hs_hmd_room_decentralized，分散式房间实例。
type HmdRoomDecentralized struct {
	commonmodel.CommonFields `bson:",inline"`

	DecentralizedID   bson.ObjectID     `bson:"decentralized_id" json:"decentralizedId"`
	RoomNo            string            `bson:"room_no" json:"roomNo"`
	FloorNo           int               `bson:"floor_no" json:"floorNo"`
	RentMode          RentMode          `bson:"rent_mode" json:"rentMode"`
	LayoutText        string            `bson:"layout_text" json:"layoutText"`
	AreaSize          int               `bson:"area_size" json:"areaSize"`
	Orientation       Orientation       `bson:"orientation" json:"orientation"`
	DecorationLevel   DecorationLevel   `bson:"decoration_level" json:"decorationLevel"`
	PaymentCycle      PaymentCycle      `bson:"payment_cycle" json:"paymentCycle"`
	Rent              int               `bson:"rent" json:"rent"`
	Deposit           int               `bson:"deposit" json:"deposit"`
	ServiceFee        int               `bson:"service_fee" json:"serviceFee"`
	AgencyFeeMode     AgencyFeeMode     `bson:"agency_fee_mode" json:"agencyFeeMode"`
	AgencyFeeValue    int               `bson:"agency_fee_value" json:"agencyFeeValue"`
	RoomStatus        RoomStatus        `bson:"room_status" json:"roomStatus"`
	ViewingTimeRule   ViewingTimeRule   `bson:"viewing_time_rule" json:"viewingTimeRule"`
	StartRentRule     StartRentRule     `bson:"start_rent_rule" json:"startRentRule"`
	Images            []TaggedImage     `bson:"images" json:"images"`
	RoomFacilities    []RoomFacility    `bson:"room_facilities" json:"roomFacilities"`
	ListingFacilities []ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
}
