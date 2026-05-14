package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type decentralizedRoomService struct {
	hmd       decentralizedRoomDomain
	publisher mutationPublisher
}

func newDecentralizedRoomService(hmd decentralizedRoomDomain, publisher mutationPublisher) *decentralizedRoomService {
	return &decentralizedRoomService{hmd: hmd, publisher: publisher}
}

func (s *decentralizedRoomService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.CreateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *decentralizedRoomService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	return s.hmd.GetDecentralizedRoom(ctx, id)
}

func (s *decentralizedRoomService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	return s.hmd.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
}

func (s *decentralizedRoomService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *decentralizedRoomService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
