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
	logPublishInfo(ctx, "publish.project.create.start", "project_code", input.ProjectCode, "city", input.City, "district", input.District)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_code", input.ProjectCode)
		return nil, err
	}
	result, err := s.hmd.CreateCentralizedProject(ctx, input)
	createdID := mutationEntityID(result)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		err = rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, createdID, "publish.project.create", err)
		logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_id", createdID.Hex(), "project_code", input.ProjectCode)
		return nil, err
	}
	if entity != nil {
		logPublishInfo(ctx, "publish.project.create.hmd_created", "project_id", entity.ID.Hex(), "project_code", entity.ProjectCode)
		if err := registerRootScope(ctx, s.access, hpdmodel.HpdRootScopeTypeCentralizedProject, entity.ID, scope.principal); err != nil {
			err = rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, entity.ID, "publish.project.create", err)
			logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_id", entity.ID.Hex())
			return nil, err
		}
		logPublishInfo(ctx, "publish.project.create.root_scope_created", "project_id", entity.ID.Hex())
	}
	logPublishInfo(ctx, "publish.project.create.success", "project_id", entity.ID.Hex(), "project_code", entity.ProjectCode)
	return entity, nil
}

func (s *centralizedProjectService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	logPublishInfo(ctx, "publish.project.detail.start", "project_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.project.detail.success", "publish.project.detail.failed", err, "project_id", id.Hex())
		return nil, err
	}
	project, err := s.hmd.GetCentralizedProject(ctx, id)
	if err != nil {
		logPublishResult(ctx, "publish.project.detail.success", "publish.project.detail.failed", err, "project_id", id.Hex())
		return nil, err
	}
	if project != nil {
		if err := requireCentralizedProjectAccess(ctx, scope, project.ID, "get centralized project"); err != nil {
			logPublishWarn(ctx, "publish.project.detail.denied", "project_id", id.Hex(), "error", err)
			return nil, err
		}
	}
	logPublishInfo(ctx, "publish.project.detail.success", "project_id", id.Hex(), "found", project != nil)
	return project, nil
}

func (s *centralizedProjectService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	logPublishInfo(ctx, "publish.project.list.start", "city", input.City, "district", input.District)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.project.list.success", "publish.project.list.failed", err, "city", input.City, "district", input.District)
		return nil, err
	}
	projects, err := s.hmd.ListCentralizedProjects(ctx, input)
	if err != nil {
		logPublishResult(ctx, "publish.project.list.success", "publish.project.list.failed", err, "city", input.City, "district", input.District)
		return nil, err
	}
	filtered, err := filterCentralizedProjectsByScope(ctx, scope, projects)
	if err != nil {
		logPublishResult(ctx, "publish.project.list.success", "publish.project.list.failed", err, "input_count", len(projects))
		return nil, err
	}
	logPublishInfo(ctx, "publish.project.list.success", "input_count", len(projects), "result_count", len(filtered))
	return filtered, nil
}

func (s *centralizedProjectService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmdmodel.HmdCentralized, error) {
	logPublishInfo(ctx, "publish.project.update.start", "project_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.project.update.success", "publish.project.update.failed", err, "project_id", input.ID.Hex())
		return nil, err
	}
	if err := requireCentralizedProjectAccess(ctx, scope, input.ID, "update centralized project"); err != nil {
		logPublishWarn(ctx, "publish.project.update.denied", "project_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateCentralizedProject(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.project.update.success", "publish.project.update.failed", err, "project_id", input.ID.Hex())
	return entity, err
}
