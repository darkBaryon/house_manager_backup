package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func (s *Service) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	community, err := s.requireDecentralized(ctx, input.DecentralizedID, "create decentralized room")
	if err != nil {
		return nil, err
	}

	entity := &hmdmodel.HmdRoomDecentralized{
		DecentralizedID:   community.ID,
		RoomNo:            strings.TrimSpace(input.RoomNo),
		FloorNo:           input.FloorNo,
		RentMode:          hmdmodel.RentMode(strings.TrimSpace(input.RentMode)),
		LayoutText:        strings.TrimSpace(input.LayoutText),
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
		return nil, mutationError("create decentralized room", err)
	}

	existing, err := s.roomDecentralizedRepo.FindByDecentralizedAndRoomNo(ctx, community.ID, entity.RoomNo)
	if err != nil {
		return nil, databasef("create decentralized room: find existing room: %w", err)
	}
	if existing != nil {
		return nil, alreadyExistsf("create decentralized room: roomNo already exists under community")
	}

	if err := s.roomDecentralizedRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create decentralized room", err)
	}
	return hmdMutationResult(entity, decentralizedRoomChange(HmdChangeCreated, entity)), nil
}

func (s *Service) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error) {
	return s.requireDecentralizedRoom(ctx, id, "get decentralized room")
}

func (s *Service) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error) {
	if _, err := s.requireDecentralized(ctx, decentralizedID, "list decentralized rooms by community"); err != nil {
		return nil, err
	}
	rooms, err := s.roomDecentralizedRepo.ListByDecentralizedID(ctx, decentralizedID)
	if err != nil {
		return nil, databasef("list decentralized rooms by community: %w", err)
	}
	return rooms, nil
}

func (s *Service) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	room, err := s.requireDecentralizedRoom(ctx, input.ID, "update decentralized room")
	if err != nil {
		return nil, err
	}

	roomNo := strings.TrimSpace(input.RoomNo)
	existing, err := s.roomDecentralizedRepo.FindByDecentralizedAndRoomNo(ctx, room.DecentralizedID, roomNo)
	if err != nil {
		return nil, databasef("update decentralized room: find existing room: %w", err)
	}
	if existing != nil && existing.ID != room.ID {
		return nil, alreadyExistsf("update decentralized room: roomNo already exists under community")
	}

	fields := bsonFields(
		"room_no", roomNo,
		"floor_no", input.FloorNo,
		"rent_mode", hmdmodel.RentMode(strings.TrimSpace(input.RentMode)),
		"layout_text", strings.TrimSpace(input.LayoutText),
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
	if err := s.roomDecentralizedRepo.UpdateBaseInfo(ctx, room.ID, fields); err != nil {
		return nil, mutationError("update decentralized room", err)
	}
	updated, err := s.requireDecentralizedRoom(ctx, room.ID, "update decentralized room")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, decentralizedRoomChange(HmdChangeUpdated, updated)), nil
}

func (s *Service) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	room, err := s.requireDecentralizedRoom(ctx, input.ID, "update decentralized room status")
	if err != nil {
		return nil, err
	}
	if err := s.roomDecentralizedRepo.UpdateStatus(ctx, room.ID, input.RoomStatus); err != nil {
		return nil, mutationError("update decentralized room status", err)
	}
	updated, err := s.requireDecentralizedRoom(ctx, room.ID, "update decentralized room status")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, decentralizedRoomChange(HmdChangeStatusUpdated, updated)), nil
}

func decentralizedRoomChange(action HmdChangeAction, room *hmdmodel.HmdRoomDecentralized) HmdChange {
	return HmdChange{
		Action:          action,
		EntityType:      HmdEntityRoomDecentralized,
		EntityID:        room.ID,
		Scope:           HmdScopeDecentralizedRoom,
		DecentralizedID: room.DecentralizedID,
	}
}
