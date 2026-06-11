package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	repohmd "house-manager/internal/repository/hmd"
	"strings"
)

func (s *Service) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*HmdMutationResult[hmdmodel.HmdBuilding], error) {
	project, err := s.requireCentralizedProject(ctx, input.ProjectID, "create building")
	if err != nil {
		return nil, err
	}

	entity := &hmdmodel.HmdBuilding{
		ProjectID:         project.ID,
		BuildingName:      strings.TrimSpace(input.BuildingName),
		FloorTotal:        input.FloorTotal,
		ManagerName:       strings.TrimSpace(input.ManagerName),
		ManagerPhone:      strings.TrimSpace(input.ManagerPhone),
		Photos:            cloneStringSlice(input.Photos),
		ListingFacilities: toListingFacilities(input.ListingFacilities),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create building", err)
	}
	existing, err := s.buildingRepo.FindByProjectAndName(ctx, entity.ProjectID, entity.BuildingName)
	if err != nil {
		return nil, databasef("create building: find existing building name: %w", err)
	}
	if existing != nil {
		return nil, alreadyExistsf("当前项目下已存在同名楼栋")
	}

	if err := s.buildingRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create building", err)
	}
	return hmdMutationResult(entity, HmdChange{
		Action:     HmdChangeCreated,
		EntityType: HmdEntityBuilding,
		EntityID:   entity.ID,
		Scope:      HmdScopeBuilding,
		ProjectID:  entity.ProjectID,
		BuildingID: entity.ID,
	}), nil
}

func (s *Service) GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	return s.requireBuilding(ctx, id, "get building")
}

func (s *Service) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list buildings by project"); err != nil {
		return nil, err
	}
	buildings, err := s.buildingRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, databasef("list buildings by project: %w", err)
	}
	return buildings, nil
}

func (s *Service) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*HmdMutationResult[hmdmodel.HmdBuilding], error) {
	building, err := s.requireBuilding(ctx, input.ID, "update building")
	if err != nil {
		return nil, err
	}

	update := repohmd.BuildingBaseInfoUpdate{
		BuildingName:      strings.TrimSpace(input.BuildingName),
		FloorTotal:        input.FloorTotal,
		ManagerName:       strings.TrimSpace(input.ManagerName),
		ManagerPhone:      strings.TrimSpace(input.ManagerPhone),
		Photos:            cloneStringSlice(input.Photos),
		ListingFacilities: toListingFacilities(input.ListingFacilities),
	}
	if err := s.buildingRepo.UpdateBaseInfo(ctx, building.ID, update); err != nil {
		return nil, mutationError("update building", err)
	}
	updated, err := s.requireBuilding(ctx, building.ID, "update building")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, HmdChange{
		Action:     HmdChangeUpdated,
		EntityType: HmdEntityBuilding,
		EntityID:   updated.ID,
		Scope:      HmdScopeBuilding,
		ProjectID:  updated.ProjectID,
		BuildingID: updated.ID,
	}), nil
}
