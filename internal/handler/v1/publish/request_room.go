package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedRoomRequest struct {
	ID                string               `json:"id"`
	ProjectID         string               `json:"project_id"`
	BuildingID        string               `json:"building_id"`
	RoomTypeID        string               `json:"room_type_id"`
	RoomNo            string               `json:"room_no" binding:"required"`
	FloorNo           int                  `json:"floor_no"`
	RentMode          string               `json:"rent_mode" binding:"required"`
	LayoutText        string               `json:"layout_text"`
	AreaSize          int                  `json:"area_size"`
	Orientation       string               `json:"orientation"`
	DecorationLevel   string               `json:"decoration_level"`
	PaymentCycle      string               `json:"payment_cycle"`
	Rent              int                  `json:"rent"`
	Deposit           int                  `json:"deposit"`
	ServiceFee        int                  `json:"service_fee"`
	AgencyFeeMode     string               `json:"agency_fee_mode"`
	AgencyFeeValue    int                  `json:"agency_fee_value"`
	ViewingTimeRule   string               `json:"viewing_time_rule"`
	StartRentRule     string               `json:"start_rent_rule"`
	Images            []taggedImageRequest `json:"images"`
	RoomFacilities    []string             `json:"room_facilities"`
	ListingFacilities []string             `json:"listing_facilities"`
}

func (r centralizedRoomRequest) toCreateInput(projectID, buildingID, roomTypeID bson.ObjectID) publishsvc.CreateCentralizedRoomInput {
	return publishsvc.CreateCentralizedRoomInput{
		ProjectID:         projectID,
		BuildingID:        buildingID,
		RoomTypeID:        roomTypeID,
		RoomNo:            r.RoomNo,
		FloorNo:           r.FloorNo,
		RentMode:          r.RentMode,
		LayoutText:        r.LayoutText,
		AreaSize:          r.AreaSize,
		Orientation:       r.Orientation,
		DecorationLevel:   r.DecorationLevel,
		PaymentCycle:      r.PaymentCycle,
		Rent:              r.Rent,
		Deposit:           r.Deposit,
		ServiceFee:        r.ServiceFee,
		AgencyFeeMode:     r.AgencyFeeMode,
		AgencyFeeValue:    r.AgencyFeeValue,
		ViewingTimeRule:   r.ViewingTimeRule,
		StartRentRule:     r.StartRentRule,
		Images:            taggedImageInputs(r.Images),
		RoomFacilities:    r.RoomFacilities,
		ListingFacilities: r.ListingFacilities,
	}
}

func (r centralizedRoomRequest) toUpdateInput(id bson.ObjectID) publishsvc.UpdateCentralizedRoomInput {
	input := r.toCreateInput(bson.NilObjectID, bson.NilObjectID, bson.NilObjectID)
	return publishsvc.UpdateCentralizedRoomInput{
		ID:                id,
		RoomNo:            input.RoomNo,
		FloorNo:           input.FloorNo,
		RentMode:          input.RentMode,
		LayoutText:        input.LayoutText,
		AreaSize:          input.AreaSize,
		Orientation:       input.Orientation,
		DecorationLevel:   input.DecorationLevel,
		PaymentCycle:      input.PaymentCycle,
		Rent:              input.Rent,
		Deposit:           input.Deposit,
		ServiceFee:        input.ServiceFee,
		AgencyFeeMode:     input.AgencyFeeMode,
		AgencyFeeValue:    input.AgencyFeeValue,
		ViewingTimeRule:   input.ViewingTimeRule,
		StartRentRule:     input.StartRentRule,
		Images:            input.Images,
		RoomFacilities:    input.RoomFacilities,
		ListingFacilities: input.ListingFacilities,
	}
}

type decentralizedRoomRequest struct {
	ID                string               `json:"id"`
	DecentralizedID   string               `json:"decentralized_id"`
	RoomNo            string               `json:"room_no" binding:"required"`
	FloorNo           int                  `json:"floor_no"`
	RentMode          string               `json:"rent_mode" binding:"required"`
	LayoutText        string               `json:"layout_text"`
	AreaSize          int                  `json:"area_size"`
	Orientation       string               `json:"orientation"`
	DecorationLevel   string               `json:"decoration_level"`
	PaymentCycle      string               `json:"payment_cycle"`
	Rent              int                  `json:"rent"`
	Deposit           int                  `json:"deposit"`
	ServiceFee        int                  `json:"service_fee"`
	AgencyFeeMode     string               `json:"agency_fee_mode"`
	AgencyFeeValue    int                  `json:"agency_fee_value"`
	ViewingTimeRule   string               `json:"viewing_time_rule"`
	StartRentRule     string               `json:"start_rent_rule"`
	Images            []taggedImageRequest `json:"images"`
	RoomFacilities    []string             `json:"room_facilities"`
	ListingFacilities []string             `json:"listing_facilities"`
}

func (r decentralizedRoomRequest) toCreateInput(decentralizedID bson.ObjectID) publishsvc.CreateDecentralizedRoomInput {
	return publishsvc.CreateDecentralizedRoomInput{
		DecentralizedID:   decentralizedID,
		RoomNo:            r.RoomNo,
		FloorNo:           r.FloorNo,
		RentMode:          r.RentMode,
		LayoutText:        r.LayoutText,
		AreaSize:          r.AreaSize,
		Orientation:       r.Orientation,
		DecorationLevel:   r.DecorationLevel,
		PaymentCycle:      r.PaymentCycle,
		Rent:              r.Rent,
		Deposit:           r.Deposit,
		ServiceFee:        r.ServiceFee,
		AgencyFeeMode:     r.AgencyFeeMode,
		AgencyFeeValue:    r.AgencyFeeValue,
		ViewingTimeRule:   r.ViewingTimeRule,
		StartRentRule:     r.StartRentRule,
		Images:            taggedImageInputs(r.Images),
		RoomFacilities:    r.RoomFacilities,
		ListingFacilities: r.ListingFacilities,
	}
}

func (r decentralizedRoomRequest) toUpdateInput(id bson.ObjectID) publishsvc.UpdateDecentralizedRoomInput {
	input := r.toCreateInput(bson.NilObjectID)
	return publishsvc.UpdateDecentralizedRoomInput{
		ID:                id,
		RoomNo:            input.RoomNo,
		FloorNo:           input.FloorNo,
		RentMode:          input.RentMode,
		LayoutText:        input.LayoutText,
		AreaSize:          input.AreaSize,
		Orientation:       input.Orientation,
		DecorationLevel:   input.DecorationLevel,
		PaymentCycle:      input.PaymentCycle,
		Rent:              input.Rent,
		Deposit:           input.Deposit,
		ServiceFee:        input.ServiceFee,
		AgencyFeeMode:     input.AgencyFeeMode,
		AgencyFeeValue:    input.AgencyFeeValue,
		ViewingTimeRule:   input.ViewingTimeRule,
		StartRentRule:     input.StartRentRule,
		Images:            input.Images,
		RoomFacilities:    input.RoomFacilities,
		ListingFacilities: input.ListingFacilities,
	}
}
