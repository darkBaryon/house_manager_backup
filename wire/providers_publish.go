package wire

import (
	"context"

	publishhandler "house-manager/internal/handler/v1/publish"
	publishauthhandler "house-manager/internal/handler/v1/publish/auth"
	landlordrepo "house-manager/internal/repository/landlord"
	publishsvc "house-manager/internal/service/publish"
	publishauthsvc "house-manager/internal/service/publish/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newPublishLandlordRepository(ctx context.Context, client *dbmongo.Client) (*landlordrepo.LandlordRepository, error) {
	repo := landlordrepo.NewLandlordRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newPublishLandlordAuthRepository(ctx context.Context, client *dbmongo.Client) (*landlordrepo.LandlordAuthRepository, error) {
	repo := landlordrepo.NewLandlordAuthRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

var PublishSet = wire.NewSet(
	publishsvc.NewPublishService,
	publishhandler.NewPublishHandler,
	newPublishLandlordRepository,
	newPublishLandlordAuthRepository,
	publishauthsvc.NewService,
	publishauthhandler.NewPublicHandler,
	publishauthhandler.NewHandler,
)
