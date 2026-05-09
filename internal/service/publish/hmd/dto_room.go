package hmd

import "go.mongodb.org/mongo-driver/v2/bson"

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

type CreateDecentralizedRoomInput struct {
	DecentralizedID   bson.ObjectID
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
