package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedRoomService struct {
	hmd       centralizedRoomDomain
	publisher mutationPublisher
}

func newCentralizedRoomService(hmd centralizedRoomDomain, publisher mutationPublisher) *centralizedRoomService {
	return &centralizedRoomService{hmd: hmd, publisher: publisher}
}

func (s *centralizedRoomService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.CreateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedRoomService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return s.hmd.GetCentralizedRoom(ctx, id)
}

func (s *centralizedRoomService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.hmd.ListCentralizedRoomsByProject(ctx, projectID)
}

func (s *centralizedRoomService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.hmd.ListCentralizedRoomsByBuilding(ctx, buildingID)
}

func (s *centralizedRoomService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.UpdateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedRoomService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.UpdateCentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
