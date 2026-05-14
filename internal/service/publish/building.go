package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type buildingService struct {
	hmd       buildingDomain
	publisher mutationPublisher
}

func newBuildingService(hmd buildingDomain, publisher mutationPublisher) *buildingService {
	return &buildingService{hmd: hmd, publisher: publisher}
}

func (s *buildingService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.hmd.CreateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *buildingService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return s.hmd.GetBuilding(ctx, id)
}

func (s *buildingService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	return s.hmd.ListBuildingsByProject(ctx, projectID)
}

func (s *buildingService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.hmd.UpdateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
