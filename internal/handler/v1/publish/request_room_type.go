package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type roomTypeRequest struct {
	ID              string               `json:"id"`
	ProjectID       string               `json:"project_id"`
	BuildingID      string               `json:"building_id"`
	RoomTypeName    string               `json:"room_type_name" binding:"required"`
	RoomCount       int                  `json:"room_count"`
	HallCount       int                  `json:"hall_count"`
	BathroomCount   int                  `json:"bathroom_count"`
	KitchenCount    int                  `json:"kitchen_count"`
	AreaSize        int                  `json:"area_size"`
	Orientation     string               `json:"orientation"`
	DecorationLevel string               `json:"decoration_level"`
	PaymentCycle    string               `json:"payment_cycle"`
	Rent            int                  `json:"rent"`
	Deposit         int                  `json:"deposit"`
	ServiceFee      int                  `json:"service_fee"`
	AgencyFeeMode   string               `json:"agency_fee_mode"`
	AgencyFeeValue  int                  `json:"agency_fee_value"`
	Images          []taggedImageRequest `json:"images"`
	RoomFacilities  []string             `json:"room_facilities"`
}

func (r roomTypeRequest) toCreateInput(projectID, buildingID bson.ObjectID) publishsvc.CreateRoomTypeInput {
	return publishsvc.CreateRoomTypeInput{
		ProjectID:       projectID,
		BuildingID:      buildingID,
		RoomTypeName:    r.RoomTypeName,
		RoomCount:       r.RoomCount,
		HallCount:       r.HallCount,
		BathroomCount:   r.BathroomCount,
		KitchenCount:    r.KitchenCount,
		AreaSize:        r.AreaSize,
		Orientation:     r.Orientation,
		DecorationLevel: r.DecorationLevel,
		PaymentCycle:    r.PaymentCycle,
		Rent:            r.Rent,
		Deposit:         r.Deposit,
		ServiceFee:      r.ServiceFee,
		AgencyFeeMode:   r.AgencyFeeMode,
		AgencyFeeValue:  r.AgencyFeeValue,
		Images:          taggedImageInputs(r.Images),
		RoomFacilities:  r.RoomFacilities,
	}
}

func (r roomTypeRequest) toUpdateInput(id bson.ObjectID) publishsvc.UpdateRoomTypeInput {
	input := r.toCreateInput(bson.NilObjectID, bson.NilObjectID)
	return publishsvc.UpdateRoomTypeInput{
		ID:              id,
		RoomTypeName:    input.RoomTypeName,
		RoomCount:       input.RoomCount,
		HallCount:       input.HallCount,
		BathroomCount:   input.BathroomCount,
		KitchenCount:    input.KitchenCount,
		AreaSize:        input.AreaSize,
		Orientation:     input.Orientation,
		DecorationLevel: input.DecorationLevel,
		PaymentCycle:    input.PaymentCycle,
		Rent:            input.Rent,
		Deposit:         input.Deposit,
		ServiceFee:      input.ServiceFee,
		AgencyFeeMode:   input.AgencyFeeMode,
		AgencyFeeValue:  input.AgencyFeeValue,
		Images:          input.Images,
		RoomFacilities:  input.RoomFacilities,
	}
}
