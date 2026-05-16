package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

func (s *Service) requireCentralizedProject(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：项目 ID 不能为空", action)
	}
	project, err := s.centralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if project == nil {
		return nil, notFoundf("%s：项目不存在", action)
	}
	return project, nil
}

func (s *Service) requireBuilding(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdBuilding, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：楼栋 ID 不能为空", action)
	}
	building, err := s.buildingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if building == nil {
		return nil, notFoundf("%s：楼栋不存在", action)
	}
	return building, nil
}

func (s *Service) requireRoomType(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomTypeCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：房型 ID 不能为空", action)
	}
	roomType, err := s.roomTypeCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if roomType == nil {
		return nil, notFoundf("%s：房型不存在", action)
	}
	return roomType, nil
}

func (s *Service) requireCentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：房间 ID 不能为空", action)
	}
	room, err := s.roomCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if room == nil {
		return nil, notFoundf("%s：集中式房间不存在", action)
	}
	return room, nil
}

func (s *Service) requireDecentralized(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdDecentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：小区 ID 不能为空", action)
	}
	community, err := s.decentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if community == nil {
		return nil, notFoundf("%s：小区不存在", action)
	}
	return community, nil
}

func (s *Service) requireDecentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomDecentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s：房间 ID 不能为空", action)
	}
	room, err := s.roomDecentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if room == nil {
		return nil, notFoundf("%s：分散式房间不存在", action)
	}
	return room, nil
}

func (s *Service) validateRoomTypeOwnership(ctx context.Context, projectID, buildingID bson.ObjectID, action string) (bson.ObjectID, bson.ObjectID, error) {
	if projectID.IsZero() && buildingID.IsZero() {
		return bson.NilObjectID, bson.NilObjectID, invalidParamf("%s：项目 ID 和楼栋 ID 至少填写一个", action)
	}

	if !projectID.IsZero() {
		if _, err := s.requireCentralizedProject(ctx, projectID, action); err != nil {
			return bson.NilObjectID, bson.NilObjectID, err
		}
	}

	if !buildingID.IsZero() {
		building, err := s.requireBuilding(ctx, buildingID, action)
		if err != nil {
			return bson.NilObjectID, bson.NilObjectID, err
		}
		if !projectID.IsZero() && building.ProjectID != projectID {
			return bson.NilObjectID, bson.NilObjectID, invalidParamf("%s：楼栋不属于当前项目", action)
		}
		if projectID.IsZero() {
			projectID = building.ProjectID
		}
	}

	return projectID, buildingID, nil
}

func (s *Service) validateCentralizedRoomDependencies(ctx context.Context, projectID, buildingID, roomTypeID bson.ObjectID, action string) (*hmdmodel.HmdCentralized, *hmdmodel.HmdBuilding, *hmdmodel.HmdRoomTypeCentralized, error) {
	project, err := s.requireCentralizedProject(ctx, projectID, action)
	if err != nil {
		return nil, nil, nil, err
	}

	building, err := s.requireBuilding(ctx, buildingID, action)
	if err != nil {
		return nil, nil, nil, err
	}
	if building.ProjectID != project.ID {
		return nil, nil, nil, invalidParamf("%s：楼栋不属于当前项目", action)
	}

	if roomTypeID.IsZero() {
		return project, building, nil, nil
	}

	roomType, err := s.requireRoomType(ctx, roomTypeID, action)
	if err != nil {
		return nil, nil, nil, err
	}
	if !roomType.ProjectID.IsZero() && roomType.ProjectID != project.ID {
		return nil, nil, nil, invalidParamf("%s：房型不属于当前项目", action)
	}
	if !roomType.BuildingID.IsZero() && roomType.BuildingID != building.ID {
		return nil, nil, nil, invalidParamf("%s：房型不属于当前楼栋", action)
	}

	return project, building, roomType, nil
}

func (s *Service) ensureRoomTypeNameAvailable(ctx context.Context, projectID, buildingID bson.ObjectID, roomTypeName string, currentID bson.ObjectID, action string) error {
	if roomTypeName == "" {
		return invalidParamf("%s：房型名称不能为空", action)
	}

	if !projectID.IsZero() {
		existing, err := s.roomTypeCentralizedRepo.FindByProjectAndName(ctx, projectID, roomTypeName)
		if err != nil {
			return databasef("%s: find existing room type by project: %w", action, err)
		}
		if existing != nil && existing.ID != currentID {
			return alreadyExistsf("%s：当前项目下已存在同名房型", action)
		}
	}

	if !buildingID.IsZero() {
		roomTypes, err := s.roomTypeCentralizedRepo.ListByBuildingID(ctx, buildingID)
		if err != nil {
			return databasef("%s: list room types by building: %w", action, err)
		}
		for _, roomType := range roomTypes {
			if roomType.RoomTypeName == roomTypeName && roomType.ID != currentID {
				return alreadyExistsf("%s：当前楼栋下已存在同名房型", action)
			}
		}
	}

	return nil
}
