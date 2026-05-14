package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type roomTypeService struct {
	hmd       roomTypeDomain
	publisher mutationPublisher
}

func newRoomTypeService(hmd roomTypeDomain, publisher mutationPublisher) *roomTypeService {
	return &roomTypeService{hmd: hmd, publisher: publisher}
}

func (s *roomTypeService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.hmd.CreateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *roomTypeService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return s.hmd.GetRoomType(ctx, id)
}

func (s *roomTypeService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.hmd.ListRoomTypesByProject(ctx, projectID)
}

func (s *roomTypeService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.hmd.ListRoomTypesByBuilding(ctx, buildingID)
}

func (s *roomTypeService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.hmd.UpdateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
