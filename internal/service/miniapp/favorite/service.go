package favorite

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	useractivitymodel "house-manager/internal/model/useractivity"
	"house-manager/internal/service/miniapp/listingview"
	miniapplog "house-manager/internal/service/miniapp/logging"
	"house-manager/internal/service/miniapp/paging"
	"house-manager/pkg/errcode"
	"log/slog"

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
	List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.Favorite, error)
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

type miniappListingRepository interface {
	FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error)
	FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]hpdmodel.HpdMiniappListing, error)
}

func NewService(favorites favoriteRepository, miniappListings miniappListingRepository) *Service {
	return &Service{favorites: favorites, miniappListings: miniappListings}
}

func (s *Service) Add(ctx context.Context, input AddInput) (*MutationResult, error) {
	slog.InfoContext(ctx, "miniapp.favorite.add.start", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
	)...)
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		slog.WarnContext(ctx, "miniapp.favorite.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"error", err,
		)...)
		return nil, err
	}
	if err := s.requireOnlineListing(ctx, input.ListingID); err != nil {
		slog.WarnContext(ctx, "miniapp.favorite.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"step", "require_online_listing",
			"error", err,
		)...)
		return nil, err
	}
	if err := s.favorites.Upsert(ctx, input.UserID, input.ListingID); err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("add favorite: %w", err)
		slog.ErrorContext(ctx, "miniapp.favorite.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"step", "upsert_favorite",
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	result := &MutationResult{ListingID: input.ListingID.Hex(), IsFavorited: true}
	slog.InfoContext(ctx, "miniapp.favorite.add.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
	)...)
	return result, nil
}

func (s *Service) Remove(ctx context.Context, input RemoveInput) (*MutationResult, error) {
	slog.InfoContext(ctx, "miniapp.favorite.remove.start", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
	)...)
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		slog.WarnContext(ctx, "miniapp.favorite.remove.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"error", err,
		)...)
		return nil, err
	}
	if err := s.favorites.SoftRemove(ctx, input.UserID, input.ListingID); err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("remove favorite: %w", err)
		slog.ErrorContext(ctx, "miniapp.favorite.remove.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	result := &MutationResult{ListingID: input.ListingID.Hex(), IsFavorited: false}
	slog.InfoContext(ctx, "miniapp.favorite.remove.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
	)...)
	return result, nil
}

func (s *Service) Exists(ctx context.Context, input ExistsInput) (bool, error) {
	if err := validateUserListing(input.UserID, input.ListingID); err != nil {
		slog.WarnContext(ctx, "miniapp.favorite.exists.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"error", err,
		)...)
		return false, err
	}
	exists, err := s.favorites.Exists(ctx, input.UserID, input.ListingID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("favorite exists: %w", err)
		slog.ErrorContext(ctx, "miniapp.favorite.exists.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"error", wrapped,
		)...)
		return false, wrapped
	}
	slog.InfoContext(ctx, "miniapp.favorite.exists.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
		"is_favorited", exists,
	)...)
	return exists, nil
}

func (s *Service) IsFavorited(ctx context.Context, userID, listingID bson.ObjectID) (bool, error) {
	return s.Exists(ctx, ExistsInput{UserID: userID, ListingID: listingID})
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	slog.InfoContext(ctx, "miniapp.favorite.list.start", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"page", input.Page,
		"page_size", input.PageSize,
	)...)
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.favorite.list.failed", miniapplog.Attrs(ctx,
			"error", err,
		)...)
		return nil, err
	}
	page, pageSize := paging.NormalizePage(input.Page, input.PageSize)
	// TODO: replace full-load plus in-memory online filtering/paging with repository methods
	// that can return online-visible count and page data in one bounded query path.
	favorites, err := s.favorites.List(ctx, input.UserID, 0, 0)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("list favorites: %w", err)
		slog.ErrorContext(ctx, "miniapp.favorite.list.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"step", "list_favorites",
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	listings, err := s.miniappListings.FindOnlineByListingIDs(ctx, favoriteListingIDs(favorites))
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("list favorite listings: %w", err)
		slog.ErrorContext(ctx, "miniapp.favorite.list.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"step", "list_online_listings",
			"favorite_count", len(favorites),
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	items := orderListingItems(favoriteListingIDs(favorites), listings)
	total := int64(len(items))
	result := &ListResult{
		List:     pageListingItems(items, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	slog.InfoContext(ctx, "miniapp.favorite.list.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"page", page,
		"page_size", pageSize,
		"returned_count", len(result.List),
		"total", total,
	)...)
	return result, nil
}

func pageListingItems(items []listingview.Item, page, pageSize int) []listingview.Item {
	if len(items) == 0 {
		return []listingview.Item{}
	}
	start := int(paging.Skip(page, pageSize))
	if start >= len(items) {
		return []listingview.Item{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (s *Service) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	if userID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.favorite.count.failed", miniapplog.Attrs(ctx,
			"error", err,
		)...)
		return 0, err
	}
	// Keep dashboard and list total semantics aligned: only online-visible listings count.
	result, err := s.List(ctx, ListInput{UserID: userID, Page: 1, PageSize: 1})
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("count favorites: %w", err)
		slog.WarnContext(ctx, "miniapp.favorite.count.failed", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"error", wrapped,
		)...)
		return 0, wrapped
	}
	slog.InfoContext(ctx, "miniapp.favorite.count.success", miniapplog.Attrs(ctx,
		"user_id", userID.Hex(),
		"total", result.Total,
	)...)
	return result.Total, nil
}

func (s *Service) requireOnlineListing(ctx context.Context, listingID bson.ObjectID) error {
	if s == nil || s.favorites == nil || s.miniappListings == nil {
		return errcode.DatabaseError.WithErrorf("favorite service dependency is nil")
	}
	listing, err := s.miniappListings.FindOnlineDetail(ctx, listingID)
	if err != nil {
		return errcode.DatabaseError.WithErrorf("find favorite listing: %w", err)
	}
	if listing == nil {
		return errcode.NotFound.WithErrorf("listing not found")
	}
	return nil
}

func validateUserListing(userID, listingID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithErrorf("user_id is required")
	}
	if listingID.IsZero() {
		return errcode.InvalidParam.WithErrorf("listing_id is required")
	}
	return nil
}

func favoriteListingIDs(favorites []useractivitymodel.Favorite) []bson.ObjectID {
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

func orderListingItems(ids []bson.ObjectID, listings []hpdmodel.HpdMiniappListing) []listingview.Item {
	if len(ids) == 0 || len(listings) == 0 {
		return []listingview.Item{}
	}
	byID := make(map[bson.ObjectID]hpdmodel.HpdMiniappListing, len(listings))
	for _, listing := range listings {
		byID[listing.ListingID] = listing
	}
	items := make([]listingview.Item, 0, len(listings))
	for _, id := range ids {
		if listing, ok := byID[id]; ok {
			items = append(items, listingview.FromListing(listing))
		}
	}
	return items
}
