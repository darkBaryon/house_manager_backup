package hmd

import "go.mongodb.org/mongo-driver/v2/bson"

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
