package publish

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/errcode"
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
	logPublishInfo(ctx, "publish.project.create.start", "project_name", input.ProjectName, "city", input.City, "district", input.District)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_name", input.ProjectName)
		return nil, err
	}
	if err := s.ensureNoDuplicateProjectForLandlord(ctx, scope, input); err != nil {
		logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_name", input.ProjectName, "city", input.City)
		return nil, err
	}
	result, err := s.hmd.CreateCentralizedProject(ctx, input)
	createdID := mutationEntityID(result)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		err = rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, createdID, "publish.project.create", err)
		logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_id", createdID.Hex(), "project_name", input.ProjectName)
		return nil, err
	}
	if entity != nil {
		logPublishInfo(ctx, "publish.project.create.hmd_created", "project_id", entity.ID.Hex(), "project_name", entity.ProjectName)
		if err := registerRootScope(ctx, s.access, hpdmodel.HpdRootScopeTypeCentralizedProject, entity.ID, scope.principal); err != nil {
			err = rollbackCreateFailure(ctx, s.hmd.RollbackCentralizedProjectCreate, entity.ID, "publish.project.create", err)
			logPublishResult(ctx, "publish.project.create.success", "publish.project.create.failed", err, "project_id", entity.ID.Hex())
			return nil, err
		}
		logPublishInfo(ctx, "publish.project.create.root_scope_created", "project_id", entity.ID.Hex())
	}
	logPublishInfo(ctx, "publish.project.create.success", "project_id", entity.ID.Hex(), "project_name", entity.ProjectName)
	return entity, nil
}

func (s *centralizedProjectService) ensureNoDuplicateProjectForLandlord(ctx context.Context, scope PublishScope, input CreateCentralizedProjectInput) error {
	projectIDs, err := scope.accessibleProjectIDs(ctx)
	if err != nil {
		return err
	}
	if len(projectIDs) == 0 {
		return nil
	}
	projects, err := s.hmd.ListCentralizedProjectsByIDs(ctx, projectIDs, ListCentralizedProjectsInput{City: input.City})
	if err != nil {
		return err
	}
	targetName := strings.TrimSpace(input.ProjectName)
	targetCity := strings.TrimSpace(input.City)
	for i := range projects {
		if strings.TrimSpace(projects[i].ProjectName) == targetName && strings.TrimSpace(projects[i].City) == targetCity {
			return errcode.AlreadyExists.WithError(fmt.Errorf("当前房东在该城市下已存在同名项目"))
		}
	}
	return nil
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
	projectIDs, err := scope.accessibleProjectIDs(ctx)
	if err != nil {
		logPublishResult(ctx, "publish.project.list.success", "publish.project.list.failed", err, "city", input.City, "district", input.District, "step", "resolve_scope")
		return nil, err
	}
	projects, err := s.hmd.ListCentralizedProjectsByIDs(ctx, projectIDs, input)
	if err != nil {
		logPublishResult(ctx, "publish.project.list.success", "publish.project.list.failed", err, "city", input.City, "district", input.District, "project_scope_count", len(projectIDs))
		return nil, err
	}
	logPublishInfo(ctx, "publish.project.list.success", "project_scope_count", len(projectIDs), "result_count", len(projects))
	return projects, nil
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
