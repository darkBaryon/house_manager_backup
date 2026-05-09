package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.decentralizedRooms.CreateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	return s.decentralizedRooms.GetDecentralizedRoom(ctx, id)
}

func (s *PublishService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	return s.decentralizedRooms.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
}

func (s *PublishService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.decentralizedRooms.UpdateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.decentralizedRooms.UpdateDecentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
