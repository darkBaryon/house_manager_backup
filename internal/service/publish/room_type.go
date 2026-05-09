package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.roomTypes.CreateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return s.roomTypes.GetRoomType(ctx, id)
}

func (s *PublishService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.roomTypes.ListRoomTypesByProject(ctx, projectID)
}

func (s *PublishService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.roomTypes.ListRoomTypesByBuilding(ctx, buildingID)
}

func (s *PublishService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.roomTypes.UpdateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
