package wire

import (
	"context"

	favoritehandler "house-manager/internal/handler/v1/miniapp/favorite"
	historyhandler "house-manager/internal/handler/v1/miniapp/history"
	userhandler "house-manager/internal/handler/v1/miniapp/user"
	repofavorite "house-manager/internal/repository/favorite"
	repohistory "house-manager/internal/repository/history"
	repohpd "house-manager/internal/repository/hpd"
	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	favoritesvc "house-manager/internal/service/miniapp/favorite"
	historysvc "house-manager/internal/service/miniapp/history"
	usersvc "house-manager/internal/service/miniapp/user"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newFavoriteRepository(ctx context.Context, client *dbmongo.Client) (*repofavorite.Repository, error) {
	repo := repofavorite.NewRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newHistoryRepository(ctx context.Context, client *dbmongo.Client) (*repohistory.Repository, error) {
	repo := repohistory.NewRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newMiniappFavoriteService(favorites *repofavorite.Repository, miniappListings *repohpd.MiniappListingRepository) *favoritesvc.Service {
	return favoritesvc.NewService(favorites, miniappListings)
}

func newMiniappHistoryService(history *repohistory.Repository, miniappListings *repohpd.MiniappListingRepository) *historysvc.Service {
	return historysvc.NewService(history, miniappListings)
}

func newMiniappUserService(
	users *miniappauthrepo.UserRepository,
	profiles *miniappauthrepo.UserProfileExtRepository,
	favorites *favoritesvc.Service,
	history *historysvc.Service,
) *usersvc.Service {
	return usersvc.NewService(users, profiles, favorites, history)
}

var MiniappUserActivitySet = wire.NewSet(
	newFavoriteRepository,
	newHistoryRepository,
	newMiniappFavoriteService,
	newMiniappHistoryService,
	newMiniappUserService,
	favoritehandler.NewHandler,
	historyhandler.NewHandler,
	userhandler.NewHandler,
)
