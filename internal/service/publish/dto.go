package publish

import "go.mongodb.org/mongo-driver/v2/bson"

type ListCentralizedProjectsInput struct {
	City     string
	District string
}

type CreateCentralizedProjectInput struct {
	ProjectName string
	ProjectCode string
	City        string
	District    string
	AddressText string
	Geo         *GeoPointInput
	BrandName   string
}

type UpdateCentralizedProjectInput struct {
	ID          bson.ObjectID
	ProjectName string
	City        string
	District    string
	AddressText string
	Geo         *GeoPointInput
	BrandName   string
}

type CreateBuildingInput struct {
	ProjectID         bson.ObjectID
	BuildingName      string
	BuildingCode      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []string
}

type UpdateBuildingInput struct {
	ID                bson.ObjectID
	BuildingName      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []string
}

type CreateRoomTypeInput struct {
	ProjectID       bson.ObjectID
	BuildingID      bson.ObjectID
	RoomTypeName    string
	RoomCount       int
	HallCount       int
	BathroomCount   int
	KitchenCount    int
	AreaSize        int
	Orientation     string
	DecorationLevel string
	PaymentCycle    string
	Rent            int
	Deposit         int
	ServiceFee      int
	AgencyFeeMode   string
	AgencyFeeValue  int
	Images          []TaggedImageInput
	RoomFacilities  []string
}

type UpdateRoomTypeInput struct {
	ID              bson.ObjectID
	RoomTypeName    string
	RoomCount       int
	HallCount       int
	BathroomCount   int
	KitchenCount    int
	AreaSize        int
	Orientation     string
	DecorationLevel string
	PaymentCycle    string
	Rent            int
	Deposit         int
	ServiceFee      int
	AgencyFeeMode   string
	AgencyFeeValue  int
	Images          []TaggedImageInput
	RoomFacilities  []string
}

type CreateCentralizedRoomInput struct {
	ProjectID         bson.ObjectID
	BuildingID        bson.ObjectID
	RoomTypeID        bson.ObjectID
	RoomNo            string
	FloorNo           int
	RentMode          string
	LayoutText        string
	AreaSize          int
	Orientation       string
	DecorationLevel   string
	PaymentCycle      string
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     string
	AgencyFeeValue    int
	ViewingTimeRule   string
	StartRentRule     string
	Images            []TaggedImageInput
	RoomFacilities    []string
	ListingFacilities []string
}

type UpdateCentralizedRoomInput struct {
	ID                bson.ObjectID
	RoomNo            string
	FloorNo           int
	RentMode          string
	LayoutText        string
	AreaSize          int
	Orientation       string
	DecorationLevel   string
	PaymentCycle      string
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     string
	AgencyFeeValue    int
	ViewingTimeRule   string
	StartRentRule     string
	Images            []TaggedImageInput
	RoomFacilities    []string
	ListingFacilities []string
}

type UpdateCentralizedRoomStatusInput struct {
	ID         bson.ObjectID
	RoomStatus int
}

type ListDecentralizedCommunitiesInput struct {
	City     string
	District string
}

type CreateDecentralizedCommunityInput struct {
	CommunityName string
	City          string
	District      string
	BizArea       string
	AddressText   string
	Geo           *GeoPointInput
	SubwayStation string
}

type UpdateDecentralizedCommunityInput struct {
	ID            bson.ObjectID
	CommunityName string
	City          string
	District      string
	BizArea       string
	AddressText   string
	Geo           *GeoPointInput
	SubwayStation string
}

type CreateDecentralizedRoomInput struct {
	DecentralizedID   bson.ObjectID
	RoomTypeID        bson.ObjectID
	RoomNo            string
	FloorNo           int
	RentMode          string
	LayoutText        string
	AreaSize          int
	Orientation       string
	DecorationLevel   string
	PaymentCycle      string
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     string
	AgencyFeeValue    int
	ViewingTimeRule   string
	StartRentRule     string
	Images            []TaggedImageInput
	RoomFacilities    []string
	ListingFacilities []string
}

type UpdateDecentralizedRoomInput struct {
	ID                bson.ObjectID
	RoomNo            string
	FloorNo           int
	RentMode          string
	LayoutText        string
	AreaSize          int
	Orientation       string
	DecorationLevel   string
	PaymentCycle      string
	Rent              int
	Deposit           int
	ServiceFee        int
	AgencyFeeMode     string
	AgencyFeeValue    int
	ViewingTimeRule   string
	StartRentRule     string
	Images            []TaggedImageInput
	RoomFacilities    []string
	ListingFacilities []string
}

type UpdateDecentralizedRoomStatusInput struct {
	ID         bson.ObjectID
	RoomStatus int
}

type GeoPointInput struct {
	Lng float64
	Lat float64
}

type TaggedImageInput struct {
	URL string
	Tag string
}
