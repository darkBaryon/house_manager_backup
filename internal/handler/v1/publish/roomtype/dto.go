package roomtype

import (
	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
	publishsvc "house-manager/internal/service/publish"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type request struct {
	ID              string                      `json:"id"`
	ProjectID       string                      `json:"project_id"`
	BuildingID      string                      `json:"building_id"`
	RoomTypeName    string                      `json:"room_type_name" binding:"required"`
	RoomCount       *int                        `json:"room_count"`
	HallCount       *int                        `json:"hall_count"`
	BathroomCount   *int                        `json:"bathroom_count"`
	KitchenCount    *int                        `json:"kitchen_count"`
	AreaSize        int                         `json:"area_size"`
	Orientation     string                      `json:"orientation"`
	DecorationLevel string                      `json:"decoration_level"`
	PaymentCycle    string                      `json:"payment_cycle"`
	Rent            int                         `json:"rent"`
	Deposit         int                         `json:"deposit"`
	ServiceFee      int                         `json:"service_fee"`
	AgencyFeeMode   string                      `json:"agency_fee_mode"`
	AgencyFeeValue  int                         `json:"agency_fee_value"`
	Images          []common.TaggedImageRequest `json:"images"`
	RoomFacilities  []string                    `json:"room_facilities"`
}

func (r request) toCreateInput(projectID, buildingID bson.ObjectID) publishsvc.CreateRoomTypeInput {
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
		Images:          common.TaggedImageInputs(r.Images),
		RoomFacilities:  r.RoomFacilities,
	}
}

func (r request) toUpdateInput(id bson.ObjectID) publishsvc.UpdateRoomTypeInput {
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

type response struct {
	common.EntityMetaResponse

	ProjectID       string                       `json:"project_id"`
	BuildingID      string                       `json:"building_id"`
	RoomTypeName    string                       `json:"room_type_name"`
	RoomCount       int                          `json:"room_count"`
	HallCount       int                          `json:"hall_count"`
	BathroomCount   int                          `json:"bathroom_count"`
	KitchenCount    int                          `json:"kitchen_count"`
	AreaSize        int                          `json:"area_size"`
	Orientation     string                       `json:"orientation"`
	DecorationLevel string                       `json:"decoration_level"`
	PaymentCycle    string                       `json:"payment_cycle"`
	Rent            int                          `json:"rent"`
	Deposit         int                          `json:"deposit"`
	ServiceFee      int                          `json:"service_fee"`
	AgencyFeeMode   string                       `json:"agency_fee_mode"`
	AgencyFeeValue  int                          `json:"agency_fee_value"`
	Images          []common.TaggedImageResponse `json:"images"`
	RoomFacilities  []string                     `json:"room_facilities"`
}

func toResponse(roomType *hmdmodel.HmdRoomTypeCentralized) *response {
	if roomType == nil {
		return nil
	}
	return &response{
		EntityMetaResponse: common.EntityMeta(roomType.CommonFields),
		ProjectID:          common.ObjectIDHex(roomType.ProjectID),
		BuildingID:         common.ObjectIDHex(roomType.BuildingID),
		RoomTypeName:       roomType.RoomTypeName,
		RoomCount:          roomType.RoomCount,
		HallCount:          roomType.HallCount,
		BathroomCount:      roomType.BathroomCount,
		KitchenCount:       roomType.KitchenCount,
		AreaSize:           roomType.AreaSize,
		Orientation:        string(roomType.Orientation),
		DecorationLevel:    string(roomType.DecorationLevel),
		PaymentCycle:       string(roomType.PaymentCycle),
		Rent:               roomType.Rent,
		Deposit:            roomType.Deposit,
		ServiceFee:         roomType.ServiceFee,
		AgencyFeeMode:      string(roomType.AgencyFeeMode),
		AgencyFeeValue:     roomType.AgencyFeeValue,
		Images:             common.TaggedImages(roomType.Images),
		RoomFacilities:     common.RoomFacilities(roomType.RoomFacilities),
	}
}

func toListResponse(roomTypes []hmdmodel.HmdRoomTypeCentralized) common.ListResponse[response] {
	list := make([]response, 0, len(roomTypes))
	for i := range roomTypes {
		item := toResponse(&roomTypes[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[response]{List: list}
}
