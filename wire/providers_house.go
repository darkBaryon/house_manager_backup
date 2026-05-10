package wire

import (
	househandler "house-manager/internal/handler/v1/miniapp/house"
	repohpd "house-manager/internal/repository/hpd"
	favoritesvc "house-manager/internal/service/miniapp/favorite"
	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/session"

	"github.com/google/wire"
)

func newMiniappHouseService(miniappListings *repohpd.MiniappListingRepository, favorites *favoritesvc.Service) *housesvc.HouseService {
	return housesvc.NewHouseServiceWithFavorite(miniappListings, favorites)
}

func newMiniappHouseHandler(service *housesvc.HouseService, store *session.Store) *househandler.HouseHandler {
	return househandler.NewHouseHandler(service, store)
}

var MiniappHouseSet = wire.NewSet(
	newMiniappHouseService,
	newMiniappHouseHandler,
)
