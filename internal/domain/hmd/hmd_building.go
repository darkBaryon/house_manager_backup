package hmd

import (
	"context"
	"strings"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *Service) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*HmdMutationResult[model.HmdBuilding], error) {
	project, err := s.requireCentralizedProject(ctx, input.ProjectID, "create building")
	if err != nil {
		return nil, err
	}

	entity := &model.HmdBuilding{
		ProjectID:         project.ID,
		BuildingName:      strings.TrimSpace(input.BuildingName),
		BuildingCode:      strings.TrimSpace(input.BuildingCode),
		FloorTotal:        input.FloorTotal,
		ManagerName:       strings.TrimSpace(input.ManagerName),
		ManagerPhone:      strings.TrimSpace(input.ManagerPhone),
		Photos:            cloneStringSlice(input.Photos),
		ListingFacilities: toListingFacilities(input.ListingFacilities),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create building", err)
	}

	if entity.BuildingCode != "" {
		existing, err := s.buildingRepo.FindByBuildingCode(ctx, entity.BuildingCode)
		if err != nil {
			return nil, databasef("create building: find existing building code: %w", err)
		}
		if existing != nil {
			return nil, alreadyExistsf("create building: buildingCode already exists")
		}
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

func (s *Service) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return s.requireBuilding(ctx, id, "get building")
}

func (s *Service) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	if _, err := s.requireCentralizedProject(ctx, projectID, "list buildings by project"); err != nil {
		return nil, err
	}
	buildings, err := s.buildingRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, databasef("list buildings by project: %w", err)
	}
	return buildings, nil
}

func (s *Service) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*HmdMutationResult[model.HmdBuilding], error) {
	building, err := s.requireBuilding(ctx, input.ID, "update building")
	if err != nil {
		return nil, err
	}

	fields := bsonFields(
		"building_name", strings.TrimSpace(input.BuildingName),
		"floor_total", input.FloorTotal,
		"manager_name", strings.TrimSpace(input.ManagerName),
		"manager_phone", strings.TrimSpace(input.ManagerPhone),
		"photos", cloneStringSlice(input.Photos),
		"listing_facilities", toListingFacilities(input.ListingFacilities),
	)
	if err := s.buildingRepo.UpdateBaseInfo(ctx, building.ID, fields); err != nil {
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
