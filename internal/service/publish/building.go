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
	logPublishInfo(ctx, "publish.building.create.start", "project_id", input.ProjectID.Hex(), "building_code", input.BuildingCode)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.building.create.success", "publish.building.create.failed", err, "project_id", input.ProjectID.Hex())
		return nil, err
	}
	if err := s.requireProjectAccess(ctx, scope, input.ProjectID, "create building"); err != nil {
		logPublishWarn(ctx, "publish.building.create.denied", "project_id", input.ProjectID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.CreateBuilding(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.building.create.success", "publish.building.create.failed", err, "project_id", input.ProjectID.Hex())
	return entity, err
}

func (s *buildingService) GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	logPublishInfo(ctx, "publish.building.detail.start", "building_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.building.detail.success", "publish.building.detail.failed", err, "building_id", id.Hex())
		return nil, err
	}
	building, err := s.hmd.GetBuilding(ctx, id)
	if err != nil {
		logPublishResult(ctx, "publish.building.detail.success", "publish.building.detail.failed", err, "building_id", id.Hex())
		return nil, err
	}
	if building != nil {
		if err := s.requireProjectAccess(ctx, scope, building.ProjectID, "get building"); err != nil {
			logPublishWarn(ctx, "publish.building.detail.denied", "building_id", id.Hex(), "project_id", building.ProjectID.Hex(), "error", err)
			return nil, err
		}
	}
	logPublishInfo(ctx, "publish.building.detail.success", "building_id", id.Hex(), "found", building != nil)
	return building, nil
}

func (s *buildingService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	logPublishInfo(ctx, "publish.building.list_by_project.start", "project_id", projectID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.building.list_by_project.success", "publish.building.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	allowed, err := s.canAccessProject(ctx, scope, projectID)
	if err != nil {
		logPublishResult(ctx, "publish.building.list_by_project.success", "publish.building.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	if !allowed {
		logPublishWarn(ctx, "publish.building.list_by_project.denied", "project_id", projectID.Hex())
		return []hmdmodel.HmdBuilding{}, nil
	}
	buildings, err := s.hmd.ListBuildingsByProject(ctx, projectID)
	logPublishResult(ctx, "publish.building.list_by_project.success", "publish.building.list_by_project.failed", err, "project_id", projectID.Hex(), "result_count", len(buildings))
	return buildings, err
}

func (s *buildingService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmdmodel.HmdBuilding, error) {
	logPublishInfo(ctx, "publish.building.update.start", "building_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.building.update.success", "publish.building.update.failed", err, "building_id", input.ID.Hex())
		return nil, err
	}
	building, err := s.hmd.GetBuilding(ctx, input.ID)
	if err != nil {
		logPublishResult(ctx, "publish.building.update.success", "publish.building.update.failed", err, "building_id", input.ID.Hex())
		return nil, err
	}
	if building == nil {
		err := scopeNotFound("update building")
		logPublishWarn(ctx, "publish.building.update.denied", "building_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := s.requireProjectAccess(ctx, scope, building.ProjectID, "update building"); err != nil {
		logPublishWarn(ctx, "publish.building.update.denied", "building_id", input.ID.Hex(), "project_id", building.ProjectID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateBuilding(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.building.update.success", "publish.building.update.failed", err, "building_id", input.ID.Hex())
	return entity, err
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
