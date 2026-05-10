package history

import (
	"context"
	"strings"
	"time"

	"house-manager/internal/model"
	minicommon "house-manager/internal/service/miniapp/common"
	"house-manager/internal/service/miniapp/listingview"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	history         historyRepository
	miniappListings miniappListingRepository
}

type historyRepository interface {
	Upsert(ctx context.Context, entity *model.History) error
	List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]model.History, error)
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

type miniappListingRepository interface {
	FindOnlineDetail(ctx context.Context, listingID bson.ObjectID) (*model.HpdMiniappListing, error)
	FindOnlineByListingIDs(ctx context.Context, listingIDs []bson.ObjectID) ([]model.HpdMiniappListing, error)
}

func NewService(history historyRepository, miniappListings miniappListingRepository) *Service {
	return &Service{history: history, miniappListings: miniappListings}
}

func (s *Service) Add(ctx context.Context, input AddInput) (*AddResult, error) {
	if input.UserID.IsZero() {
		return nil, invalidParamf("user_id is required")
	}
	if input.ListingID.IsZero() {
		return nil, invalidParamf("listing_id is required")
	}
	source := strings.TrimSpace(input.Source)
	if !model.ValidHistorySource(source) {
		return nil, invalidParamf("source is invalid")
	}
	if err := s.requireOnlineListing(ctx, input.ListingID); err != nil {
		return nil, err
	}
	viewedAt := time.Now().Unix()
	if err := s.history.Upsert(ctx, &model.History{
		UserID:    input.UserID,
		ListingID: input.ListingID,
		Source:    source,
		ViewedAt:  viewedAt,
	}); err != nil {
		return nil, databasef("add history: %w", err)
	}
	return &AddResult{ListingID: input.ListingID.Hex(), ViewedAt: viewedAt}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if input.UserID.IsZero() {
		return nil, invalidParamf("user_id is required")
	}
	page, pageSize := minicommon.NormalizePage(input.Page, input.PageSize)
	// TODO: replace full-load plus in-memory online filtering/paging with repository methods
	// that can return online-visible count and page data in one bounded query path.
	historyItems, err := s.history.List(ctx, input.UserID, 0, 0)
	if err != nil {
		return nil, databasef("list history: %w", err)
	}
	listings, err := s.miniappListings.FindOnlineByListingIDs(ctx, historyListingIDs(historyItems))
	if err != nil {
		return nil, databasef("list history listings: %w", err)
	}
	items := orderHistoryItems(historyItems, listings)
	total := int64(len(items))
	return &ListResult{
		List:     pageHistoryItems(items, page, pageSize),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func pageHistoryItems(items []ListItem, page, pageSize int) []ListItem {
	if len(items) == 0 {
		return []ListItem{}
	}
	start := int(minicommon.Skip(page, pageSize))
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
		return 0, invalidParamf("user_id is required")
	}
	// Keep dashboard and list total semantics aligned: only online-visible listings count.
	result, err := s.List(ctx, ListInput{UserID: userID, Page: 1, PageSize: 1})
	if err != nil {
		return 0, databasef("count history: %w", err)
	}
	return result.Total, nil
}

func (s *Service) requireOnlineListing(ctx context.Context, listingID bson.ObjectID) error {
	if s == nil || s.history == nil || s.miniappListings == nil {
		return databasef("history service dependency is nil")
	}
	listing, err := s.miniappListings.FindOnlineDetail(ctx, listingID)
	if err != nil {
		return databasef("find history listing: %w", err)
	}
	if listing == nil {
		return notFoundf("listing not found")
	}
	return nil
}

func historyListingIDs(items []model.History) []bson.ObjectID {
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

func orderHistoryItems(historyItems []model.History, listings []model.HpdMiniappListing) []ListItem {
	if len(historyItems) == 0 || len(listings) == 0 {
		return []ListItem{}
	}
	byID := make(map[bson.ObjectID]model.HpdMiniappListing, len(listings))
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
