package wire

import (
	"context"

	"house-manager/internal/config"
	publishhandler "house-manager/internal/handler/v1/publish"
	publishauthhandler "house-manager/internal/handler/v1/publish/auth"
	publishauthrepo "house-manager/internal/repository/publish_auth"
	publishsvc "house-manager/internal/service/publish"
	publishauthsvc "house-manager/internal/service/publish/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newAdmStaffRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.StaffRepository, error) {
	repo := publishauthrepo.NewStaffRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmRoleRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.RoleRepository, error) {
	repo := publishauthrepo.NewRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmPermissionRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.PermissionRepository, error) {
	repo := publishauthrepo.NewPermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmStaffRoleRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.StaffRoleRepository, error) {
	repo := publishauthrepo.NewStaffRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmRolePermissionRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.RolePermissionRepository, error) {
	repo := publishauthrepo.NewRolePermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmLoginLogRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.LoginLogRepository, error) {
	repo := publishauthrepo.NewLoginLogRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newPublishAuthEnv(cfg *config.Config) string {
	return cfg.Log.Env
}

var PublishSet = wire.NewSet(
	publishsvc.NewPublishService,
	publishhandler.NewPublishHandler,
	newAdmStaffRepository,
	newAdmRoleRepository,
	newAdmPermissionRepository,
	newAdmStaffRoleRepository,
	newAdmRolePermissionRepository,
	newAdmLoginLogRepository,
	newPublishAuthEnv,
	publishauthsvc.NewService,
	publishauthhandler.NewPublicHandler,
	publishauthhandler.NewHandler,
)
