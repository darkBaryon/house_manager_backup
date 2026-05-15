package history

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	useractivitymodel "house-manager/internal/model/useractivity"
	"house-manager/internal/service/miniapp/listingview"
	miniapplog "house-manager/internal/service/miniapp/logging"
	"house-manager/internal/service/miniapp/paging"
	"house-manager/pkg/errcode"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	history         historyRepository
	miniappListings miniappListingRepository
}

type historyRepository interface {
	Upsert(ctx context.Context, entity *useractivitymodel.History) error
	List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.History, error)
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

type miniappListingRepository interface {
	FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdMiniappListing, error)
	FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]hpdmodel.HpdMiniappListing, error)
}

func NewService(history historyRepository, miniappListings miniappListingRepository) *Service {
	return &Service{history: history, miniappListings: miniappListings}
}

func (s *Service) Add(ctx context.Context, input AddInput) (*AddResult, error) {
	slog.InfoContext(ctx, "miniapp.history.add.start", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
		"source", input.Source,
	)...)
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.history.add.failed", miniapplog.Attrs(ctx,
			"error", err,
		)...)
		return nil, err
	}
	if input.ListingID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("listing_id is required")
		slog.WarnContext(ctx, "miniapp.history.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"error", err,
		)...)
		return nil, err
	}
	source := strings.TrimSpace(input.Source)
	if !useractivitymodel.ValidHistorySource(source) {
		err := errcode.InvalidParam.WithErrorf("source is invalid")
		slog.WarnContext(ctx, "miniapp.history.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"source", source,
			"error", err,
		)...)
		return nil, err
	}
	if err := s.requireOnlineListing(ctx, input.ListingID); err != nil {
		slog.WarnContext(ctx, "miniapp.history.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"step", "require_online_listing",
			"error", err,
		)...)
		return nil, err
	}
	viewedAt := time.Now().Unix()
	if err := s.history.Upsert(ctx, &useractivitymodel.History{
		UserID:    input.UserID,
		ListingID: input.ListingID,
		Source:    source,
		ViewedAt:  viewedAt,
	}); err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("add history: %w", err)
		slog.ErrorContext(ctx, "miniapp.history.add.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"listing_id", input.ListingID.Hex(),
			"step", "upsert_history",
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	result := &AddResult{ListingID: input.ListingID.Hex(), ViewedAt: viewedAt}
	slog.InfoContext(ctx, "miniapp.history.add.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"listing_id", input.ListingID.Hex(),
		"viewed_at", viewedAt,
	)...)
	return result, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	slog.InfoContext(ctx, "miniapp.history.list.start", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"page", input.Page,
		"page_size", input.PageSize,
	)...)
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.history.list.failed", miniapplog.Attrs(ctx,
			"error", err,
		)...)
		return nil, err
	}
	page, pageSize := paging.NormalizePage(input.Page, input.PageSize)
	// TODO: replace full-load plus in-memory online filtering/paging with repository methods
	// that can return online-visible count and page data in one bounded query path.
	historyItems, err := s.history.List(ctx, input.UserID, 0, 0)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("list history: %w", err)
		slog.ErrorContext(ctx, "miniapp.history.list.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"step", "list_history",
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	listings, err := s.miniappListings.FindOnlineByListingIDs(ctx, historyListingIDs(historyItems))
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("list history listings: %w", err)
		slog.ErrorContext(ctx, "miniapp.history.list.failed", miniapplog.Attrs(ctx,
			"user_id", input.UserID.Hex(),
			"step", "list_online_listings",
			"history_count", len(historyItems),
			"error", wrapped,
		)...)
		return nil, wrapped
	}
	items := orderHistoryItems(historyItems, listings)
	total := int64(len(items))
	result := &ListResult{
		List:     pageHistoryItems(items, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	slog.InfoContext(ctx, "miniapp.history.list.success", miniapplog.Attrs(ctx,
		"user_id", input.UserID.Hex(),
		"page", page,
		"page_size", pageSize,
		"returned_count", len(result.List),
		"total", total,
	)...)
	return result, nil
}

func pageHistoryItems(items []ListItem, page, pageSize int) []ListItem {
	if len(items) == 0 {
		return []ListItem{}
	}
	start := int(paging.Skip(page, pageSize))
	if start >= len(items) {
		return []ListItem{}
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
		slog.WarnContext(ctx, "miniapp.history.count.failed", miniapplog.Attrs(ctx,
			"error", err,
		)...)
		return 0, err
	}
	// Keep dashboard and list total semantics aligned: only online-visible listings count.
	result, err := s.List(ctx, ListInput{UserID: userID, Page: 1, PageSize: 1})
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("count history: %w", err)
		slog.WarnContext(ctx, "miniapp.history.count.failed", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"error", wrapped,
		)...)
		return 0, wrapped
	}
	slog.InfoContext(ctx, "miniapp.history.count.success", miniapplog.Attrs(ctx,
		"user_id", userID.Hex(),
		"total", result.Total,
	)...)
	return result.Total, nil
}

func (s *Service) requireOnlineListing(ctx context.Context, listingID bson.ObjectID) error {
	if s == nil || s.history == nil || s.miniappListings == nil {
		return errcode.DatabaseError.WithErrorf("history service dependency is nil")
	}
	listing, err := s.miniappListings.FindOnlineDetail(ctx, listingID)
	if err != nil {
		return errcode.DatabaseError.WithErrorf("find history listing: %w", err)
	}
	if listing == nil {
		return errcode.NotFound.WithErrorf("listing not found")
	}
	return nil
}

func historyListingIDs(items []useractivitymodel.History) []bson.ObjectID {
	if len(items) == 0 {
		return nil
	}
	ids := make([]bson.ObjectID, 0, len(items))
	for _, item := range items {
		if !item.ListingID.IsZero() {
			ids = append(ids, item.ListingID)
		}
	}
	return ids
}

func orderHistoryItems(historyItems []useractivitymodel.History, listings []hpdmodel.HpdMiniappListing) []ListItem {
	if len(historyItems) == 0 || len(listings) == 0 {
		return []ListItem{}
	}
	byID := make(map[bson.ObjectID]hpdmodel.HpdMiniappListing, len(listings))
	for _, listing := range listings {
		byID[listing.ListingID] = listing
	}
	items := make([]ListItem, 0, len(listings))
	for _, historyItem := range historyItems {
		if listing, ok := byID[historyItem.ListingID]; ok {
			items = append(items, ListItem{
				Item:     listingview.FromListing(listing),
				ViewedAt: historyItem.ViewedAt,
			})
		}
	}
	return items
}
