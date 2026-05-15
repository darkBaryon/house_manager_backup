package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type centralizedRoomService struct {
	hmd       centralizedRoomScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newCentralizedRoomService(hmd centralizedRoomScopeDomain, publisher mutationPublisher, access publishAccessService) *centralizedRoomService {
	return &centralizedRoomService{hmd: hmd, publisher: publisher, access: access}
}

func (s *centralizedRoomService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := s.requireBuildingAccess(ctx, scope, input.BuildingID, "create centralized room"); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateCentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		return nil, err
	}
	if entity != nil {
		if err := registerRoomEntrust(ctx, s.access, hpdmodel.HpdSourceTypeCentralizedRoom, entity.ID, scope.principal); err != nil {
			return nil, err
		}
	}
	return entity, nil
}

func (s *centralizedRoomService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeCentralizedRoom, id, "get centralized room"); err != nil {
		return nil, err
	}
	return s.hmd.GetCentralizedRoom(ctx, id)
}

func (s *centralizedRoomService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	rooms, err := s.hmd.ListCentralizedRoomsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return filterCentralizedRoomsByScope(ctx, scope, rooms)
}

func (s *centralizedRoomService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	rooms, err := s.hmd.ListCentralizedRoomsByBuilding(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	return filterCentralizedRoomsByScope(ctx, scope, rooms)
}

func (s *centralizedRoomService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeCentralizedRoom, input.ID, "update centralized room"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedRoomService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeCentralizedRoom, input.ID, "update centralized room status"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedRoomService) canAccessBuilding(ctx context.Context, scope PublishScope, buildingID bson.ObjectID) (bool, error) {
	if scope.IsGlobal() {
		return true, nil
	}
	building, err := s.hmd.GetBuilding(ctx, buildingID)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	if building == nil {
		return false, nil
	}
	return canAccessCentralizedProject(ctx, scope, s.hmd, building.ProjectID)
}

func (s *centralizedRoomService) requireBuildingAccess(ctx context.Context, scope PublishScope, buildingID bson.ObjectID, action string) error {
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}
