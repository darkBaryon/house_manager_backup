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
	logPublishInfo(ctx, "publish.room_type.create.start", "project_id", input.ProjectID.Hex(), "building_id", input.BuildingID.Hex(), "room_type_name", input.RoomTypeName)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.create.success", "publish.room_type.create.failed", err)
		return nil, err
	}
	if err := s.requireRoomTypeParentAccess(ctx, scope, input.ProjectID, input.BuildingID, "create room type"); err != nil {
		logPublishWarn(ctx, "publish.room_type.create.denied", "project_id", input.ProjectID.Hex(), "building_id", input.BuildingID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.CreateRoomType(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.room_type.create.success", "publish.room_type.create.failed", err)
	return entity, err
}

func (s *roomTypeService) GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	logPublishInfo(ctx, "publish.room_type.detail.start", "room_type_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.detail.success", "publish.room_type.detail.failed", err, "room_type_id", id.Hex())
		return nil, err
	}
	roomType, err := s.hmd.GetRoomType(ctx, id)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.detail.success", "publish.room_type.detail.failed", err, "room_type_id", id.Hex())
		return nil, err
	}
	if roomType != nil {
		if err := s.requireRoomTypeAccess(ctx, scope, roomType, "get room type"); err != nil {
			logPublishWarn(ctx, "publish.room_type.detail.denied", "room_type_id", id.Hex(), "error", err)
			return nil, err
		}
	}
	logPublishInfo(ctx, "publish.room_type.detail.success", "room_type_id", id.Hex(), "found", roomType != nil)
	return roomType, nil
}

func (s *roomTypeService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	logPublishInfo(ctx, "publish.room_type.list_by_project.start", "project_id", projectID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.list_by_project.success", "publish.room_type.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	allowed, err := s.canAccessProject(ctx, scope, projectID)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.list_by_project.success", "publish.room_type.list_by_project.failed", err, "project_id", projectID.Hex())
		return nil, err
	}
	if !allowed {
		logPublishWarn(ctx, "publish.room_type.list_by_project.denied", "project_id", projectID.Hex())
		return []hmdmodel.HmdRoomTypeCentralized{}, nil
	}
	roomTypes, err := s.hmd.ListRoomTypesByProject(ctx, projectID)
	logPublishResult(ctx, "publish.room_type.list_by_project.success", "publish.room_type.list_by_project.failed", err, "project_id", projectID.Hex(), "result_count", len(roomTypes))
	return roomTypes, err
}

func (s *roomTypeService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	logPublishInfo(ctx, "publish.room_type.list_by_building.start", "building_id", buildingID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.list_by_building.success", "publish.room_type.list_by_building.failed", err, "building_id", buildingID.Hex())
		return nil, err
	}
	allowed, err := s.canAccessBuilding(ctx, scope, buildingID)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.list_by_building.success", "publish.room_type.list_by_building.failed", err, "building_id", buildingID.Hex())
		return nil, err
	}
	if !allowed {
		logPublishWarn(ctx, "publish.room_type.list_by_building.denied", "building_id", buildingID.Hex())
		return []hmdmodel.HmdRoomTypeCentralized{}, nil
	}
	roomTypes, err := s.hmd.ListRoomTypesByBuilding(ctx, buildingID)
	logPublishResult(ctx, "publish.room_type.list_by_building.success", "publish.room_type.list_by_building.failed", err, "building_id", buildingID.Hex(), "result_count", len(roomTypes))
	return roomTypes, err
}

func (s *roomTypeService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmdmodel.HmdRoomTypeCentralized, error) {
	logPublishInfo(ctx, "publish.room_type.update.start", "room_type_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.update.success", "publish.room_type.update.failed", err, "room_type_id", input.ID.Hex())
		return nil, err
	}
	roomType, err := s.hmd.GetRoomType(ctx, input.ID)
	if err != nil {
		logPublishResult(ctx, "publish.room_type.update.success", "publish.room_type.update.failed", err, "room_type_id", input.ID.Hex())
		return nil, err
	}
	if roomType == nil {
		err := scopeNotFound("update room type")
		logPublishWarn(ctx, "publish.room_type.update.denied", "room_type_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := s.requireRoomTypeAccess(ctx, scope, roomType, "update room type"); err != nil {
		logPublishWarn(ctx, "publish.room_type.update.denied", "room_type_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateRoomType(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.room_type.update.success", "publish.room_type.update.failed", err, "room_type_id", input.ID.Hex())
	return entity, err
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
