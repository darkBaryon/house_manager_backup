package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

type roomTypeService struct {
	hmd       roomTypeDomain
	publisher mutationPublisher
	access    publishAccessService
	scopeHmd  roomTypeScopeDomain
}

func newRoomTypeService(hmd roomTypeScopeDomain, publisher mutationPublisher, access publishAccessService) *roomTypeService {
	return &roomTypeService{hmd: hmd, publisher: publisher, access: access, scopeHmd: hmd}
}

func (s *roomTypeService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmdmodel.HmdRoomTypeCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := s.requireRoomTypeParentAccess(ctx, scope, input.ProjectID, input.BuildingID, "create room type"); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *roomTypeService) GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	roomType, err := s.hmd.GetRoomType(ctx, id)
	if err != nil {
		return nil, err
	}
	if roomType != nil {
		if err := s.requireRoomTypeAccess(ctx, scope, roomType, "get room type"); err != nil {
			return nil, err
		}
	}
	return roomType, nil
}

func (s *roomTypeService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessProject(ctx, scope, projectID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdRoomTypeCentralized{}, nil
	}
	return s.hmd.ListRoomTypesByProject(ctx, projectID)
}

func (s *roomTypeService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdRoomTypeCentralized{}, nil
	}
	return s.hmd.ListRoomTypesByBuilding(ctx, buildingID)
}

func (s *roomTypeService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmdmodel.HmdRoomTypeCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	roomType, err := s.hmd.GetRoomType(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if roomType == nil {
		return nil, scopeNotFound("update room type")
	}
	if err := s.requireRoomTypeAccess(ctx, scope, roomType, "update room type"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *roomTypeService) canAccessProject(ctx context.Context, scope PublishScope, projectID bson.ObjectID) (bool, error) {
	return canAccessCentralizedProject(ctx, scope, projectID)
}

func (s *roomTypeService) canAccessBuilding(ctx context.Context, scope PublishScope, buildingID bson.ObjectID) (bool, error) {
	building, err := s.scopeHmd.GetBuilding(ctx, buildingID)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	if building == nil {
		return false, nil
	}
	return s.canAccessProject(ctx, scope, building.ProjectID)
}

func (s *roomTypeService) requireRoomTypeParentAccess(ctx context.Context, scope PublishScope, projectID, buildingID bson.ObjectID, action string) error {
	var (
		allowed bool
		err     error
	)
	if !buildingID.IsZero() {
		allowed, err = s.canAccessBuilding(ctx, scope, buildingID)
	} else {
		allowed, err = s.canAccessProject(ctx, scope, projectID)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

func (s *roomTypeService) requireRoomTypeAccess(ctx context.Context, scope PublishScope, roomType *hmdmodel.HmdRoomTypeCentralized, action string) error {
	var (
		allowed bool
		err     error
	)
	if !roomType.ProjectID.IsZero() {
		allowed, err = s.canAccessProject(ctx, scope, roomType.ProjectID)
	} else {
		allowed, err = s.canAccessBuilding(ctx, scope, roomType.BuildingID)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}
