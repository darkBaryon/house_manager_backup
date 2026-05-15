package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

type buildingService struct {
	hmd       buildingScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newBuildingService(hmd buildingScopeDomain, publisher mutationPublisher, access publishAccessService) *buildingService {
	return &buildingService{hmd: hmd, publisher: publisher, access: access}
}

func (s *buildingService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmdmodel.HmdBuilding, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := s.requireProjectAccess(ctx, scope, input.ProjectID, "create building"); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *buildingService) GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	building, err := s.hmd.GetBuilding(ctx, id)
	if err != nil {
		return nil, err
	}
	if building != nil {
		if err := s.requireProjectAccess(ctx, scope, building.ProjectID, "get building"); err != nil {
			return nil, err
		}
	}
	return building, nil
}

func (s *buildingService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessProject(ctx, scope, projectID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdBuilding{}, nil
	}
	return s.hmd.ListBuildingsByProject(ctx, projectID)
}

func (s *buildingService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmdmodel.HmdBuilding, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	building, err := s.hmd.GetBuilding(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if building == nil {
		return nil, scopeNotFound("update building")
	}
	if err := s.requireProjectAccess(ctx, scope, building.ProjectID, "update building"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *buildingService) canAccessProject(ctx context.Context, scope PublishScope, projectID bson.ObjectID) (bool, error) {
	return canAccessCentralizedProject(ctx, scope, projectID)
}

func (s *buildingService) requireProjectAccess(ctx context.Context, scope PublishScope, projectID bson.ObjectID, action string) error {
	allowed, err := s.canAccessProject(ctx, scope, projectID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}
