package house

import (
	"context"

	"house-manager/internal/model"
	repohpd "house-manager/internal/repository/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type HouseService struct {
	miniappListings miniappListingRepository
	favorites       favoriteChecker
}

type miniappListingRepository interface {
	SearchMiniapp(ctx context.Context, search repohpd.MiniappListingSearchFilter) ([]model.HpdMiniappListing, error)
	CountMiniapp(ctx context.Context, search repohpd.MiniappListingSearchFilter) (int64, error)
	FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*model.HpdMiniappListing, error)
}

type favoriteChecker interface {
	IsFavorited(ctx context.Context, userID, listingID bson.ObjectID) (bool, error)
}

func NewHouseService(miniappListings *repohpd.MiniappListingRepository) *HouseService {
	return &HouseService{miniappListings: miniappListings}
}

func NewHouseServiceWithFavorite(miniappListings *repohpd.MiniappListingRepository, favorites favoriteChecker) *HouseService {
	return &HouseService{miniappListings: miniappListings, favorites: favorites}
}
