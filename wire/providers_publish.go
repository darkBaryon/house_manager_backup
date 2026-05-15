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

func newPublishOwnerUserRepository(ctx context.Context, client *dbmongo.Client) (*publishauthrepo.OwnerUserRepository, error) {
	repo := publishauthrepo.NewOwnerUserRepository(client)
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
	newPublishOwnerUserRepository,
	newPublishAuthEnv,
	publishauthsvc.NewService,
	publishauthhandler.NewPublicHandler,
	publishauthhandler.NewHandler,
)
