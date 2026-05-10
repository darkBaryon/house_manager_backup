package favorite

import (
	"context"

	"house-manager/internal/model"
	minicommon "house-manager/internal/service/miniapp/common"
	housesvc "house-manager/internal/service/miniapp/house"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	favorites       favoriteRepository
	miniappListings miniappListingRepository
}

type favoriteRepository interface {
	Upsert(ctx context.Context, userID, listingID bson.ObjectID) error
	SoftRemove(ctx context.Context, userID, listingID bson.ObjectID) error
	Exists(ctx context.Context, userID, listingID bson.ObjectID) (bool, error)
	List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]model.Favorite, error)
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

type miniappListingRepository interface {
	FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*model.HpdMiniappListing, error)
	FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]model.HpdMiniappListing, error)
}

func NewService(favorites favoriteRepository, miniappListings miniappListingRepository) *Service {
	return &Service{favorites: favorites, miniappListings: miniappListings}
}

func (s *Service) Add(ctx context.Context, input AddInput) (*MutationResult, error) {
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		return nil, err
	}
	if err := s.requireOnlineListing(ctx, input.ListingID); err != nil {
		return nil, err
	}
	if err := s.favorites.Upsert(ctx, input.UserID, input.ListingID); err != nil {
		return nil, databasef("add favorite: %w", err)
	}
	return &MutationResult{ListingID: input.ListingID.Hex(), IsFavorited: true}, nil
}

func (s *Service) Remove(ctx context.Context, input RemoveInput) (*MutationResult, error) {
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		return nil, err
	}
	if err := s.favorites.SoftRemove(ctx, input.UserID, input.ListingID); err != nil {
		return nil, databasef("remove favorite: %w", err)
	}
	return &MutationResult{ListingID: input.ListingID.Hex(), IsFavorited: false}, nil
}

func (s *Service) Exists(ctx context.Context, input ExistsInput) (bool, error) {
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		return false, err
	}
	exists, err := s.favorites.Exists(ctx, input.UserID, input.ListingID)
	if err != nil {
		return false, databasef("favorite exists: %w", err)
	}
	return exists, nil
}

func (s *Service) IsFavorited(ctx context.Context, userID, listingID bson.ObjectID) (bool, error) {
	return s.Exists(ctx, ExistsInput{UserID: userID, ListingID: listingID})
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if input.UserID.IsZero() {
		return nil, invalidParamf("user_id is required")
	}
	page, pageSize := minicommon.NormalizePage(input.Page, input.PageSize)
	favorites, err := s.favorites.List(ctx, input.UserID, 0, 0)
	if err != nil {
		return nil, databasef("list favorites: %w", err)
	}
	listings, err := s.miniappListings.FindOnlineByListingIDs(ctx, favoriteListingIDs(favorites))
	if err != nil {
		return nil, databasef("list favorite listings: %w", err)
	}
	items := orderListingItems(favoriteListingIDs(favorites), listings)
	total := int64(len(items))
	return &ListResult{
		List:     pageListingItems(items, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func pageListingItems(items []housesvc.ListItem, page, pageSize int) []housesvc.ListItem {
	if len(items) == 0 {
		return []housesvc.ListItem{}
	}
	start := int(minicommon.Skip(page, pageSize))
	if start >= len(items) {
		return []housesvc.ListItem{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (s *Service) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	if userID.IsZero() {
		return 0, invalidParamf("user_id is required")
	}
	total, err := s.favorites.Count(ctx, userID)
	if err != nil {
		return 0, databasef("count favorites: %w", err)
	}
	return total, nil
}

func (s *Service) requireOnlineListing(ctx context.Context, listingID bson.ObjectID) error {
	if s == nil || s.favorites == nil || s.miniappListings == nil {
		return databasef("favorite service dependency is nil")
	}
	listing, err := s.miniappListings.FindOnlineDetail(ctx, listingID)
	if err != nil {
		return databasef("find favorite listing: %w", err)
	}
	if listing == nil {
		return notFoundf("listing not found")
	}
	return nil
}

func validateUserListing(userID, listingID bson.ObjectID) error {
	if userID.IsZero() {
		return invalidParamf("user_id is required")
	}
	if listingID.IsZero() {
		return invalidParamf("listing_id is required")
	}
	return nil
}

func favoriteListingIDs(favorites []model.Favorite) []bson.ObjectID {
	if len(favorites) == 0 {
		return nil
	}
	ids := make([]bson.ObjectID, 0, len(favorites))
	for _, favorite := range favorites {
		if !favorite.ListingID.IsZero() {
			ids = append(ids, favorite.ListingID)
		}
	}
	return ids
}

func orderListingItems(ids []bson.ObjectID, listings []model.HpdMiniappListing) []housesvc.ListItem {
	if len(ids) == 0 || len(listings) == 0 {
		return []housesvc.ListItem{}
	}
	byID := make(map[bson.ObjectID]model.HpdMiniappListing, len(listings))
	for _, listing := range listings {
		byID[listing.ListingID] = listing
	}
	items := make([]housesvc.ListItem, 0, len(listings))
	for _, id := range ids {
		if listing, ok := byID[id]; ok {
			items = append(items, housesvc.ListingItem(listing))
		}
	}
	return items
}
