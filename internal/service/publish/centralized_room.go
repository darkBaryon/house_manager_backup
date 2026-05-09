package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.centralizedRooms.CreateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return s.centralizedRooms.GetCentralizedRoom(ctx, id)
}

func (s *PublishService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.centralizedRooms.ListCentralizedRoomsByProject(ctx, projectID)
}

func (s *PublishService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.centralizedRooms.ListCentralizedRoomsByBuilding(ctx, buildingID)
}

func (s *PublishService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.centralizedRooms.UpdateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	result, err := s.centralizedRooms.UpdateCentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
