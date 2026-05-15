package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
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
	return entity, nil
}

func (s *centralizedRoomService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	room, err := s.hmd.GetCentralizedRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if room != nil {
		if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "get centralized room"); err != nil {
			return nil, err
		}
	}
	return room, nil
}

func (s *centralizedRoomService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := canAccessCentralizedProject(ctx, scope, projectID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdRoomCentralized{}, nil
	}
	return s.hmd.ListCentralizedRoomsByProject(ctx, projectID)
}

func (s *centralizedRoomService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdRoomCentralized{}, nil
	}
	return s.hmd.ListCentralizedRoomsByBuilding(ctx, buildingID)
}

func (s *centralizedRoomService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdmodel.HmdRoomCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	room, err := s.hmd.GetCentralizedRoom(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, scopeNotFound("update centralized room")
	}
	if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "update centralized room"); err != nil {
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
	room, err := s.hmd.GetCentralizedRoom(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, scopeNotFound("update centralized room status")
	}
	if err := requireCentralizedProjectAccess(ctx, scope, room.ProjectID, "update centralized room status"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedRoomService) canAccessBuilding(ctx context.Context, scope PublishScope, buildingID bson.ObjectID) (bool, error) {
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
	return canAccessCentralizedProject(ctx, scope, building.ProjectID)
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
