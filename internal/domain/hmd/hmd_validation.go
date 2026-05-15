package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

func (s *Service) requireCentralizedProject(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: projectID is required", action)
	}
	project, err := s.centralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if project == nil {
		return nil, notFoundf("%s: centralized project not found", action)
	}
	return project, nil
}

func (s *Service) requireBuilding(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdBuilding, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: buildingID is required", action)
	}
	building, err := s.buildingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if building == nil {
		return nil, notFoundf("%s: building not found", action)
	}
	return building, nil
}

func (s *Service) requireRoomType(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomTypeCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: roomTypeID is required", action)
	}
	roomType, err := s.roomTypeCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if roomType == nil {
		return nil, notFoundf("%s: room type not found", action)
	}
	return roomType, nil
}

func (s *Service) requireCentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomCentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: roomID is required", action)
	}
	room, err := s.roomCentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if room == nil {
		return nil, notFoundf("%s: centralized room not found", action)
	}
	return room, nil
}

func (s *Service) requireDecentralized(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdDecentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: decentralizedID is required", action)
	}
	community, err := s.decentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if community == nil {
		return nil, notFoundf("%s: decentralized community not found", action)
	}
	return community, nil
}

func (s *Service) requireDecentralizedRoom(ctx context.Context, id bson.ObjectID, action string) (*hmdmodel.HmdRoomDecentralized, error) {
	if id.IsZero() {
		return nil, invalidParamf("%s: roomID is required", action)
	}
	room, err := s.roomDecentralizedRepo.FindByID(ctx, id)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	if room == nil {
		return nil, notFoundf("%s: decentralized room not found", action)
	}
	return room, nil
}

func (s *Service) validateRoomTypeOwnership(ctx context.Context, projectID, buildingID bson.ObjectID, action string) (bson.ObjectID, bson.ObjectID, error) {
	if projectID.IsZero() && buildingID.IsZero() {
		return bson.NilObjectID, bson.NilObjectID, invalidParamf("%s: projectID or buildingID is required", action)
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
			return bson.NilObjectID, bson.NilObjectID, invalidParamf("%s: building does not belong to project", action)
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
		return nil, nil, nil, invalidParamf("%s: building does not belong to project", action)
	}

	if roomTypeID.IsZero() {
		return project, building, nil, nil
	}

	roomType, err := s.requireRoomType(ctx, roomTypeID, action)
	if err != nil {
		return nil, nil, nil, err
	}
	if !roomType.ProjectID.IsZero() && roomType.ProjectID != project.ID {
		return nil, nil, nil, invalidParamf("%s: room type does not belong to project", action)
	}
	if !roomType.BuildingID.IsZero() && roomType.BuildingID != building.ID {
		return nil, nil, nil, invalidParamf("%s: room type does not belong to building", action)
	}

	return project, building, roomType, nil
}

func (s *Service) ensureRoomTypeNameAvailable(ctx context.Context, projectID, buildingID bson.ObjectID, roomTypeName string, currentID bson.ObjectID, action string) error {
	if roomTypeName == "" {
		return invalidParamf("%s: roomTypeName is required", action)
	}

	if !projectID.IsZero() {
		existing, err := s.roomTypeCentralizedRepo.FindByProjectAndName(ctx, projectID, roomTypeName)
		if err != nil {
			return databasef("%s: find existing room type by project: %w", action, err)
		}
		if existing != nil && existing.ID != currentID {
			return alreadyExistsf("%s: roomTypeName already exists under project", action)
		}
	}

	if !buildingID.IsZero() {
		roomTypes, err := s.roomTypeCentralizedRepo.ListByBuildingID(ctx, buildingID)
		if err != nil {
			return databasef("%s: list room types by building: %w", action, err)
		}
		for _, roomType := range roomTypes {
			if roomType.RoomTypeName == roomTypeName && roomType.ID != currentID {
				return alreadyExistsf("%s: roomTypeName already exists under building", action)
			}
		}
	}

	return nil
}
