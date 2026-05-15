package publish

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type centralizedProjectService struct {
	hmd       centralizedProjectScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newCentralizedProjectService(hmd centralizedProjectScopeDomain, publisher mutationPublisher, access publishAccessService) *centralizedProjectService {
	return &centralizedProjectService{hmd: hmd, publisher: publisher, access: access}
}

func (s *centralizedProjectService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmdmodel.HmdCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateCentralizedProject(ctx, input)
	createdID := mutationEntityID(result)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		return nil, rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, createdID, "create centralized project", err)
	}
	if entity != nil {
		if err := registerRootScope(ctx, s.access, hpdmodel.HpdRootScopeTypeCentralizedProject, entity.ID, scope.principal); err != nil {
			return nil, rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, entity.ID, "create centralized project", err)
		}
	}
	return entity, nil
}

func (s *centralizedProjectService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	project, err := s.hmd.GetCentralizedProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if project != nil {
		if err := requireCentralizedProjectAccess(ctx, scope, project.ID, "get centralized project"); err != nil {
			return nil, err
		}
	}
	return project, nil
}

func (s *centralizedProjectService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	projects, err := s.hmd.ListCentralizedProjects(ctx, input)
	if err != nil {
		return nil, err
	}
	return filterCentralizedProjectsByScope(ctx, scope, projects)
}

func (s *centralizedProjectService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmdmodel.HmdCentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := requireCentralizedProjectAccess(ctx, scope, input.ID, "update centralized project"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
