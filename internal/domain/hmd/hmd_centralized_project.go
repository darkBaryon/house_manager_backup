package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func (s *Service) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*HmdMutationResult[hmdmodel.HmdCentralized], error) {
	entity := &hmdmodel.HmdCentralized{
		ProjectName: strings.TrimSpace(input.ProjectName),
		ProjectCode: strings.TrimSpace(input.ProjectCode),
		City:        strings.TrimSpace(input.City),
		District:    strings.TrimSpace(input.District),
		AddressText: strings.TrimSpace(input.AddressText),
		Geo:         toGeoPoint(input.Geo),
		BrandName:   strings.TrimSpace(input.BrandName),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create centralized project", err)
	}

	existing, err := s.centralizedRepo.FindByProjectCode(ctx, entity.ProjectCode)
	if err != nil {
		return nil, databasef("create centralized project: find existing project code: %w", err)
	}
	if existing != nil {
		return nil, alreadyExistsf("create centralized project: projectCode already exists")
	}

	if err := s.centralizedRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create centralized project", err)
	}
	return hmdMutationResult(entity, HmdChange{
		Action:     HmdChangeCreated,
		EntityType: HmdEntityCentralizedProject,
		EntityID:   entity.ID,
		Scope:      HmdScopeCentralizedProject,
		ProjectID:  entity.ID,
	}), nil
}

func (s *Service) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	return s.requireCentralizedProject(ctx, id, "get centralized project")
}

func (s *Service) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	projects, err := s.centralizedRepo.List(ctx, city, district)
	if err != nil {
		return nil, databasef("list centralized projects: %w", err)
	}
	return projects, nil
}

func (s *Service) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*HmdMutationResult[hmdmodel.HmdCentralized], error) {
	project, err := s.requireCentralizedProject(ctx, input.ID, "update centralized project")
	if err != nil {
		return nil, err
	}

	fields := bsonFields(
		"project_name", strings.TrimSpace(input.ProjectName),
		"city", strings.TrimSpace(input.City),
		"district", strings.TrimSpace(input.District),
		"address_text", strings.TrimSpace(input.AddressText),
		"geo", toGeoPoint(input.Geo),
		"brand_name", strings.TrimSpace(input.BrandName),
	)
	if err := s.centralizedRepo.UpdateBaseInfo(ctx, project.ID, fields); err != nil {
		return nil, mutationError("update centralized project", err)
	}
	updated, err := s.requireCentralizedProject(ctx, project.ID, "update centralized project")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, HmdChange{
		Action:     HmdChangeUpdated,
		EntityType: HmdEntityCentralizedProject,
		EntityID:   updated.ID,
		Scope:      HmdScopeCentralizedProject,
		ProjectID:  updated.ID,
	}), nil
}
