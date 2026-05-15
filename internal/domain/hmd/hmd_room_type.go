package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func (s *Service) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error) {
	projectID, buildingID, err := s.validateRoomTypeOwnership(ctx, input.ProjectID, input.BuildingID, "create room type")
	if err != nil {
		return nil, err
	}

	roomTypeName := strings.TrimSpace(input.RoomTypeName)
	entity := &hmdmodel.HmdRoomTypeCentralized{
		ProjectID:       projectID,
		BuildingID:      buildingID,
		RoomTypeName:    roomTypeName,
		RoomCount:       input.RoomCount,
		HallCount:       input.HallCount,
		BathroomCount:   input.BathroomCount,
		KitchenCount:    input.KitchenCount,
		AreaSize:        input.AreaSize,
		Orientation:     hmdmodel.Orientation(strings.TrimSpace(input.Orientation)),
		DecorationLevel: hmdmodel.DecorationLevel(strings.TrimSpace(input.DecorationLevel)),
		PaymentCycle:    hmdmodel.PaymentCycle(strings.TrimSpace(input.PaymentCycle)),
		Rent:            input.Rent,
		Deposit:         input.Deposit,
		ServiceFee:      input.ServiceFee,
		AgencyFeeMode:   hmdmodel.AgencyFeeMode(strings.TrimSpace(input.AgencyFeeMode)),
		AgencyFeeValue:  input.AgencyFeeValue,
		Images:          toTaggedImages(input.Images),
		RoomFacilities:  toRoomFacilities(input.RoomFacilities),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create room type", err)
	}

	if err := s.ensureRoomTypeNameAvailable(ctx, projectID, buildingID, roomTypeName, bson.NilObjectID, "create room type"); err != nil {
		return nil, err
	}

	if err := s.roomTypeCentralizedRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create room type", err)
	}
	return hmdMutationResult(entity, HmdChange{
		Action:     HmdChangeCreated,
		EntityType: HmdEntityRoomTypeCentralized,
		EntityID:   entity.ID,
		Scope:      HmdScopeRoomTypeCentralized,
		ProjectID:  entity.ProjectID,
		BuildingID: entity.BuildingID,
		RoomTypeID: entity.ID,
	}), nil
}

func (s *Service) GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	return s.requireRoomType(ctx, id, "get room type")
}

func (s *Service) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list room types by project"); err != nil {
		return nil, err
	}
	roomTypes, err := s.roomTypeCentralizedRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, databasef("list room types by project: %w", err)
	}
	return roomTypes, nil
}

func (s *Service) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	if _, err := s.requireBuilding(ctx, buildingID, "list room types by building"); err != nil {
		return nil, err
	}
	roomTypes, err := s.roomTypeCentralizedRepo.ListByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, databasef("list room types by building: %w", err)
	}
	return roomTypes, nil
}

func (s *Service) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error) {
	roomType, err := s.requireRoomType(ctx, input.ID, "update room type")
	if err != nil {
		return nil, err
	}

	roomTypeName := strings.TrimSpace(input.RoomTypeName)
	if err := s.ensureRoomTypeNameAvailable(ctx, roomType.ProjectID, roomType.BuildingID, roomTypeName, roomType.ID, "update room type"); err != nil {
		return nil, err
	}

	fields := bsonFields(
		"room_type_name", roomTypeName,
		"room_count", input.RoomCount,
		"hall_count", input.HallCount,
		"bathroom_count", input.BathroomCount,
		"kitchen_count", input.KitchenCount,
		"area_size", input.AreaSize,
		"orientation", hmdmodel.Orientation(strings.TrimSpace(input.Orientation)),
		"decoration_level", hmdmodel.DecorationLevel(strings.TrimSpace(input.DecorationLevel)),
		"payment_cycle", hmdmodel.PaymentCycle(strings.TrimSpace(input.PaymentCycle)),
		"rent", input.Rent,
		"deposit", input.Deposit,
		"service_fee", input.ServiceFee,
		"agency_fee_mode", hmdmodel.AgencyFeeMode(strings.TrimSpace(input.AgencyFeeMode)),
		"agency_fee_value", input.AgencyFeeValue,
		"images", toTaggedImages(input.Images),
		"room_facilities", toRoomFacilities(input.RoomFacilities),
	)
	if err := s.roomTypeCentralizedRepo.UpdateBaseInfo(ctx, roomType.ID, fields); err != nil {
		return nil, mutationError("update room type", err)
	}
	updated, err := s.requireRoomType(ctx, roomType.ID, "update room type")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, HmdChange{
		Action:     HmdChangeUpdated,
		EntityType: HmdEntityRoomTypeCentralized,
		EntityID:   updated.ID,
		Scope:      HmdScopeRoomTypeCentralized,
		ProjectID:  updated.ProjectID,
		BuildingID: updated.BuildingID,
		RoomTypeID: updated.ID,
	}), nil
}
