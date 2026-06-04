package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func (s *Service) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	project, building, roomType, err := s.validateCentralizedRoomDependencies(ctx, input.ProjectID, input.BuildingID, input.RoomTypeID, "create centralized room")
	if err != nil {
		return nil, err
	}

	roomCount, hallCount, bathroomCount, kitchenCount := centralizedRoomShape(input.RoomCount, input.HallCount, input.BathroomCount, input.KitchenCount, roomType)
	entity := &hmdmodel.HmdRoomCentralized{
		ProjectID:         project.ID,
		BuildingID:        building.ID,
		RoomTypeID:        roomTypeID(roomType),
		RoomNo:            strings.TrimSpace(input.RoomNo),
		FloorNo:           input.FloorNo,
		RentMode:          hmdmodel.RentMode(strings.TrimSpace(input.RentMode)),
		LayoutText:        strings.TrimSpace(input.LayoutText),
		RoomCount:         roomCount,
		HallCount:         hallCount,
		BathroomCount:     bathroomCount,
		KitchenCount:      kitchenCount,
		AreaSize:          input.AreaSize,
		Orientation:       hmdmodel.Orientation(strings.TrimSpace(input.Orientation)),
		DecorationLevel:   hmdmodel.DecorationLevel(strings.TrimSpace(input.DecorationLevel)),
		PaymentCycle:      hmdmodel.PaymentCycle(strings.TrimSpace(input.PaymentCycle)),
		Rent:              input.Rent,
		Deposit:           input.Deposit,
		ServiceFee:        input.ServiceFee,
		AgencyFeeMode:     hmdmodel.AgencyFeeMode(strings.TrimSpace(input.AgencyFeeMode)),
		AgencyFeeValue:    input.AgencyFeeValue,
		ViewingTimeRule:   hmdmodel.ViewingTimeRule(strings.TrimSpace(input.ViewingTimeRule)),
		StartRentRule:     hmdmodel.StartRentRule(strings.TrimSpace(input.StartRentRule)),
		Images:            toTaggedImages(input.Images),
		RoomFacilities:    toRoomFacilities(input.RoomFacilities),
		ListingFacilities: toListingFacilities(input.ListingFacilities),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create centralized room", err)
	}

	existing, err := s.roomCentralizedRepo.FindByBuildingAndRoomNo(ctx, building.ID, entity.RoomNo)
	if err != nil {
		return nil, databasef("create centralized room: find existing room: %w", err)
	}
	if existing != nil {
		return nil, alreadyExistsf("当前楼栋下已存在相同房间号")
	}

	if err := s.roomCentralizedRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create centralized room", err)
	}
	return hmdMutationResult(entity, centralizedRoomChange(HmdChangeCreated, entity)), nil
}

func (s *Service) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	return s.requireCentralizedRoom(ctx, id, "get centralized room")
}

func (s *Service) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list centralized rooms by project"); err != nil {
		return nil, err
	}
	rooms, err := s.roomCentralizedRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, databasef("list centralized rooms by project: %w", err)
	}
	return rooms, nil
}

func (s *Service) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	if _, err := s.requireBuilding(ctx, buildingID, "list centralized rooms by building"); err != nil {
		return nil, err
	}
	rooms, err := s.roomCentralizedRepo.ListByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, databasef("list centralized rooms by building: %w", err)
	}
	return rooms, nil
}

func (s *Service) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	room, err := s.requireCentralizedRoom(ctx, input.ID, "update centralized room")
	if err != nil {
		return nil, err
	}

	roomTypeID := room.RoomTypeID
	if input.RoomTypeID != nil {
		_, _, roomType, err := s.validateCentralizedRoomDependencies(ctx, room.ProjectID, room.BuildingID, *input.RoomTypeID, "update centralized room")
		if err != nil {
			return nil, err
		}
		if roomType != nil {
			roomTypeID = roomType.ID
		}
	}
	var roomType *hmdmodel.HmdRoomTypeCentralized
	if !roomTypeID.IsZero() {
		roomType, err = s.roomTypeCentralizedRepo.FindByID(ctx, roomTypeID)
		if err != nil {
			return nil, databasef("update centralized room: find room type: %w", err)
		}
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomCentralizedRepo.FindByBuildingAndRoomNo(ctx, room.BuildingID, roomNo)
	if err != nil {
		return nil, databasef("update centralized room: find existing room: %w", err)
	}
	if existing != nil && existing.ID != room.ID {
		return nil, alreadyExistsf("当前楼栋下已存在相同房间号")
	}

	roomCount, hallCount, bathroomCount, kitchenCount := centralizedRoomShape(input.RoomCount, input.HallCount, input.BathroomCount, input.KitchenCount, roomType)
	fields := bsonFields(
		"room_type_id", roomTypeID,
		"room_no", roomNo,
		"floor_no", input.FloorNo,
		"rent_mode", hmdmodel.RentMode(strings.TrimSpace(input.RentMode)),
		"layout_text", strings.TrimSpace(input.LayoutText),
		"room_count", roomCount,
		"hall_count", hallCount,
		"bathroom_count", bathroomCount,
		"kitchen_count", kitchenCount,
		"area_size", input.AreaSize,
		"orientation", hmdmodel.Orientation(strings.TrimSpace(input.Orientation)),
		"decoration_level", hmdmodel.DecorationLevel(strings.TrimSpace(input.DecorationLevel)),
		"payment_cycle", hmdmodel.PaymentCycle(strings.TrimSpace(input.PaymentCycle)),
		"rent", input.Rent,
		"deposit", input.Deposit,
		"service_fee", input.ServiceFee,
		"agency_fee_mode", hmdmodel.AgencyFeeMode(strings.TrimSpace(input.AgencyFeeMode)),
		"agency_fee_value", input.AgencyFeeValue,
		"viewing_time_rule", hmdmodel.ViewingTimeRule(strings.TrimSpace(input.ViewingTimeRule)),
		"start_rent_rule", hmdmodel.StartRentRule(strings.TrimSpace(input.StartRentRule)),
		"images", toTaggedImages(input.Images),
		"room_facilities", toRoomFacilities(input.RoomFacilities),
		"listing_facilities", toListingFacilities(input.ListingFacilities),
	)
	if err := s.roomCentralizedRepo.UpdateBaseInfo(ctx, room.ID, fields); err != nil {
		return nil, mutationError("update centralized room", err)
	}
	updated, err := s.requireCentralizedRoom(ctx, room.ID, "update centralized room")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, centralizedRoomChange(HmdChangeUpdated, updated)), nil
}

func centralizedRoomShape(roomCount, hallCount, bathroomCount, kitchenCount *int, roomType *hmdmodel.HmdRoomTypeCentralized) (int, int, int, int) {
	if roomType == nil {
		return layoutCountValue(roomCount), layoutCountValue(hallCount), layoutCountValue(bathroomCount), layoutCountValue(kitchenCount)
	}
	return layoutCountValueOrFallback(roomCount, roomType.RoomCount),
		layoutCountValueOrFallback(hallCount, roomType.HallCount),
		layoutCountValueOrFallback(bathroomCount, roomType.BathroomCount),
		layoutCountValueOrFallback(kitchenCount, roomType.KitchenCount)
}

func (s *Service) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	room, err := s.requireCentralizedRoom(ctx, input.ID, "update centralized room status")
	if err != nil {
		return nil, err
	}
	if err := s.roomCentralizedRepo.UpdateStatus(ctx, room.ID, input.RoomStatus); err != nil {
		return nil, mutationError("update centralized room status", err)
	}
	updated, err := s.requireCentralizedRoom(ctx, room.ID, "update centralized room status")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, centralizedRoomChange(HmdChangeStatusUpdated, updated)), nil
}

func centralizedRoomChange(action HmdChangeAction, room *hmdmodel.HmdRoomCentralized) HmdChange {
	return HmdChange{
		Action:     action,
		EntityType: HmdEntityRoomCentralized,
		EntityID:   room.ID,
		Scope:      HmdScopeCentralizedRoom,
		ProjectID:  room.ProjectID,
		BuildingID: room.BuildingID,
		RoomTypeID: room.RoomTypeID,
	}
}
