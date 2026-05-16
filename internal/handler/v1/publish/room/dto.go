package room

import (
	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
	publishsvc "house-manager/internal/service/publish"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedRequest struct {
	ID                string                      `json:"id"`
	ProjectID         string                      `json:"project_id"`
	BuildingID        string                      `json:"building_id"`
	RoomTypeID        *string                     `json:"room_type_id"`
	RoomNo            string                      `json:"room_no" binding:"required"`
	FloorNo           int                         `json:"floor_no"`
	RentMode          string                      `json:"rent_mode" binding:"required"`
	LayoutText        string                      `json:"layout_text"`
	AreaSize          int                         `json:"area_size"`
	Orientation       string                      `json:"orientation"`
	DecorationLevel   string                      `json:"decoration_level"`
	PaymentCycle      string                      `json:"payment_cycle"`
	Rent              int                         `json:"rent"`
	Deposit           int                         `json:"deposit"`
	ServiceFee        int                         `json:"service_fee"`
	AgencyFeeMode     string                      `json:"agency_fee_mode"`
	AgencyFeeValue    int                         `json:"agency_fee_value"`
	ViewingTimeRule   string                      `json:"viewing_time_rule"`
	StartRentRule     string                      `json:"start_rent_rule"`
	Images            []common.TaggedImageRequest `json:"images"`
	RoomFacilities    []string                    `json:"room_facilities"`
	ListingFacilities []string                    `json:"listing_facilities"`
}

func (r centralizedRequest) toCreateInput(projectID, buildingID, roomTypeID bson.ObjectID) publishsvc.CreateCentralizedRoomInput {
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
		Images:            common.TaggedImageInputs(r.Images),
		RoomFacilities:    r.RoomFacilities,
		ListingFacilities: r.ListingFacilities,
	}
}

func (r centralizedRequest) toUpdateInput(id bson.ObjectID, roomTypeID *bson.ObjectID) publishsvc.UpdateCentralizedRoomInput {
	input := r.toCreateInput(bson.NilObjectID, bson.NilObjectID, bson.NilObjectID)
	return publishsvc.UpdateCentralizedRoomInput{
		ID:                id,
		RoomTypeID:        roomTypeID,
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

type decentralizedRequest struct {
	ID                string                      `json:"id"`
	DecentralizedID   string                      `json:"decentralized_id"`
	RoomNo            string                      `json:"room_no" binding:"required"`
	FloorNo           int                         `json:"floor_no"`
	RentMode          string                      `json:"rent_mode" binding:"required"`
	LayoutText        string                      `json:"layout_text"`
	AreaSize          int                         `json:"area_size"`
	Orientation       string                      `json:"orientation"`
	DecorationLevel   string                      `json:"decoration_level"`
	PaymentCycle      string                      `json:"payment_cycle"`
	Rent              int                         `json:"rent"`
	Deposit           int                         `json:"deposit"`
	ServiceFee        int                         `json:"service_fee"`
	AgencyFeeMode     string                      `json:"agency_fee_mode"`
	AgencyFeeValue    int                         `json:"agency_fee_value"`
	ViewingTimeRule   string                      `json:"viewing_time_rule"`
	StartRentRule     string                      `json:"start_rent_rule"`
	Images            []common.TaggedImageRequest `json:"images"`
	RoomFacilities    []string                    `json:"room_facilities"`
	ListingFacilities []string                    `json:"listing_facilities"`
}

func (r decentralizedRequest) toCreateInput(decentralizedID bson.ObjectID) publishsvc.CreateDecentralizedRoomInput {
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
		Images:            common.TaggedImageInputs(r.Images),
		RoomFacilities:    r.RoomFacilities,
		ListingFacilities: r.ListingFacilities,
	}
}

func (r decentralizedRequest) toUpdateInput(id bson.ObjectID) publishsvc.UpdateDecentralizedRoomInput {
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

type centralizedResponse struct {
	common.EntityMetaResponse

	ProjectID         string                       `json:"project_id"`
	BuildingID        string                       `json:"building_id"`
	RoomTypeID        string                       `json:"room_type_id"`
	RoomNo            string                       `json:"room_no"`
	FloorNo           int                          `json:"floor_no"`
	RentMode          string                       `json:"rent_mode"`
	LayoutText        string                       `json:"layout_text"`
	AreaSize          int                          `json:"area_size"`
	Orientation       string                       `json:"orientation"`
	DecorationLevel   string                       `json:"decoration_level"`
	PaymentCycle      string                       `json:"payment_cycle"`
	Rent              int                          `json:"rent"`
	Deposit           int                          `json:"deposit"`
	ServiceFee        int                          `json:"service_fee"`
	AgencyFeeMode     string                       `json:"agency_fee_mode"`
	AgencyFeeValue    int                          `json:"agency_fee_value"`
	RoomStatus        int                          `json:"room_status"`
	ViewingTimeRule   string                       `json:"viewing_time_rule"`
	StartRentRule     string                       `json:"start_rent_rule"`
	Images            []common.TaggedImageResponse `json:"images"`
	RoomFacilities    []string                     `json:"room_facilities"`
	ListingFacilities []string                     `json:"listing_facilities"`
}

type decentralizedResponse struct {
	common.EntityMetaResponse

	DecentralizedID   string                       `json:"decentralized_id"`
	RoomNo            string                       `json:"room_no"`
	FloorNo           int                          `json:"floor_no"`
	RentMode          string                       `json:"rent_mode"`
	LayoutText        string                       `json:"layout_text"`
	AreaSize          int                          `json:"area_size"`
	Orientation       string                       `json:"orientation"`
	DecorationLevel   string                       `json:"decoration_level"`
	PaymentCycle      string                       `json:"payment_cycle"`
	Rent              int                          `json:"rent"`
	Deposit           int                          `json:"deposit"`
	ServiceFee        int                          `json:"service_fee"`
	AgencyFeeMode     string                       `json:"agency_fee_mode"`
	AgencyFeeValue    int                          `json:"agency_fee_value"`
	RoomStatus        int                          `json:"room_status"`
	ViewingTimeRule   string                       `json:"viewing_time_rule"`
	StartRentRule     string                       `json:"start_rent_rule"`
	Images            []common.TaggedImageResponse `json:"images"`
	RoomFacilities    []string                     `json:"room_facilities"`
	ListingFacilities []string                     `json:"listing_facilities"`
}

func toCentralizedResponse(room *hmdmodel.HmdRoomCentralized) *centralizedResponse {
	if room == nil {
		return nil
	}
	return &centralizedResponse{
		EntityMetaResponse: common.EntityMeta(room.CommonFields),
		ProjectID:          common.ObjectIDHex(room.ProjectID),
		BuildingID:         common.ObjectIDHex(room.BuildingID),
		RoomTypeID:         common.ObjectIDHex(room.RoomTypeID),
		RoomNo:             room.RoomNo,
		FloorNo:            room.FloorNo,
		RentMode:           string(room.RentMode),
		LayoutText:         room.LayoutText,
		AreaSize:           room.AreaSize,
		Orientation:        string(room.Orientation),
		DecorationLevel:    string(room.DecorationLevel),
		PaymentCycle:       string(room.PaymentCycle),
		Rent:               room.Rent,
		Deposit:            room.Deposit,
		ServiceFee:         room.ServiceFee,
		AgencyFeeMode:      string(room.AgencyFeeMode),
		AgencyFeeValue:     room.AgencyFeeValue,
		RoomStatus:         int(room.RoomStatus),
		ViewingTimeRule:    string(room.ViewingTimeRule),
		StartRentRule:      string(room.StartRentRule),
		Images:             common.TaggedImages(room.Images),
		RoomFacilities:     common.RoomFacilities(room.RoomFacilities),
		ListingFacilities:  common.ListingFacilities(room.ListingFacilities),
	}
}

func toCentralizedListResponse(rooms []hmdmodel.HmdRoomCentralized) common.ListResponse[centralizedResponse] {
	list := make([]centralizedResponse, 0, len(rooms))
	for i := range rooms {
		item := toCentralizedResponse(&rooms[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[centralizedResponse]{List: list}
}

func toDecentralizedResponse(room *hmdmodel.HmdRoomDecentralized) *decentralizedResponse {
	if room == nil {
		return nil
	}
	return &decentralizedResponse{
		EntityMetaResponse: common.EntityMeta(room.CommonFields),
		DecentralizedID:    common.ObjectIDHex(room.DecentralizedID),
		RoomNo:             room.RoomNo,
		FloorNo:            room.FloorNo,
		RentMode:           string(room.RentMode),
		LayoutText:         room.LayoutText,
		AreaSize:           room.AreaSize,
		Orientation:        string(room.Orientation),
		DecorationLevel:    string(room.DecorationLevel),
		PaymentCycle:       string(room.PaymentCycle),
		Rent:               room.Rent,
		Deposit:            room.Deposit,
		ServiceFee:         room.ServiceFee,
		AgencyFeeMode:      string(room.AgencyFeeMode),
		AgencyFeeValue:     room.AgencyFeeValue,
		RoomStatus:         int(room.RoomStatus),
		ViewingTimeRule:    string(room.ViewingTimeRule),
		StartRentRule:      string(room.StartRentRule),
		Images:             common.TaggedImages(room.Images),
		RoomFacilities:     common.RoomFacilities(room.RoomFacilities),
		ListingFacilities:  common.ListingFacilities(room.ListingFacilities),
	}
}

func toDecentralizedListResponse(rooms []hmdmodel.HmdRoomDecentralized) common.ListResponse[decentralizedResponse] {
	list := make([]decentralizedResponse, 0, len(rooms))
	for i := range rooms {
		item := toDecentralizedResponse(&rooms[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[decentralizedResponse]{List: list}
}
