package wire

import (
	"context"

	adminauthhandler "house-manager/internal/handler/v1/admin/auth"
	adminhousehandler "house-manager/internal/handler/v1/admin/house"
	adminproviderhandler "house-manager/internal/handler/v1/admin/provider"
	adminrolehandler "house-manager/internal/handler/v1/admin/role"
	adminstaffhandler "house-manager/internal/handler/v1/admin/staff"
	admrepo "house-manager/internal/repository/adm"
	adminauthsvc "house-manager/internal/service/admin/auth"
	adminhousesvc "house-manager/internal/service/admin/house"
	adminprovidersvc "house-manager/internal/service/admin/provider"
	adminrolesvc "house-manager/internal/service/admin/role"
	adminstaffsvc "house-manager/internal/service/admin/staff"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newAdminStaffRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.StaffRepository, error) {
	repo := admrepo.NewStaffRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminStaffAuthRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.StaffAuthRepository, error) {
	repo := admrepo.NewStaffAuthRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminRoleRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.RoleRepository, error) {
	repo := admrepo.NewRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminPermissionRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.PermissionRepository, error) {
	repo := admrepo.NewPermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminStaffRoleRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.StaffRoleRepository, error) {
	repo := admrepo.NewStaffRoleRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminRolePermissionRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.RolePermissionRepository, error) {
	repo := admrepo.NewRolePermissionRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newAdminLoginLogRepository(ctx context.Context, client *dbmongo.Client) (*admrepo.LoginLogRepository, error) {
	repo := admrepo.NewLoginLogRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

var AdminSet = wire.NewSet(
	newAdminStaffRepository,
	newAdminStaffAuthRepository,
	newAdminRoleRepository,
	newAdminPermissionRepository,
	newAdminStaffRoleRepository,
	newAdminRolePermissionRepository,
	newAdminLoginLogRepository,
	adminauthsvc.NewService,
	adminstaffsvc.NewService,
	adminrolesvc.NewService,
	adminprovidersvc.NewService,
	adminhousesvc.NewService,
	adminauthhandler.NewPublicHandler,
	adminauthhandler.NewHandler,
	adminstaffhandler.NewHandler,
	adminrolehandler.NewHandler,
	adminproviderhandler.NewHandler,
	adminhousehandler.NewHandler,
)
