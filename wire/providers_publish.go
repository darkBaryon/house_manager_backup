package wire

import (
	"context"

	"house-manager/internal/config"
	publishhandler "house-manager/internal/handler/v1/publish"
	publishauthhandler "house-manager/internal/handler/v1/publish/auth"
	repoaccount "house-manager/internal/repository/account"
	publishsvc "house-manager/internal/service/publish"
	publishauthsvc "house-manager/internal/service/publish/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newAdmStaffRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.StaffRepository, error) {
	repo := repoaccount.NewStaffRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmRoleRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.RoleRepository, error) {
	repo := repoaccount.NewRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmPermissionRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.PermissionRepository, error) {
	repo := repoaccount.NewPermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmStaffRoleRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.StaffRoleRepository, error) {
	repo := repoaccount.NewStaffRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmRolePermissionRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.RolePermissionRepository, error) {
	repo := repoaccount.NewRolePermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdmLoginLogRepository(ctx context.Context, client *dbmongo.Client) (*repoaccount.LoginLogRepository, error) {
	repo := repoaccount.NewLoginLogRepository(client)
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
