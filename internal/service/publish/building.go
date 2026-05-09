package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.buildings.CreateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return s.buildings.GetBuilding(ctx, id)
}

func (s *PublishService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	return s.buildings.ListBuildingsByProject(ctx, projectID)
}

func (s *PublishService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.buildings.UpdateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
