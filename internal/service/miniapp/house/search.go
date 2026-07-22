package house

import (
	"context"
	"log/slog"

	"house-manager/pkg/errcode"
)

func (s *HouseService) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	slog.InfoContext(ctx, "miniapp.house.search.start", "city", input.City,
		"district", input.District,
		"rent_mode", input.RentMode,
		"asset_mode", input.AssetMode,
		"keyword", input.Keyword,
		"page", input.Page,
		"page_size", input.PageSize,
	)
	if s == nil || s.miniappListings == nil {
		err := errcode.DatabaseError.WithErrorf("house search repository is nil")
		slog.ErrorContext(ctx, "miniapp.house.search.failed", "error", err)
		return nil, err
	}

	filter, page, pageSize, err := searchFilter(input)
	if err != nil {
		slog.WarnContext(ctx, "miniapp.house.search.failed", "error", err)
		return nil, err
	}

	listings, err := s.miniappListings.SearchMiniapp(ctx, filter)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("search house listings: %w", err)
		slog.ErrorContext(ctx, "miniapp.house.search.failed", "step", "search_listings",
			"page", page,
			"page_size", pageSize,
			"error", wrapped,
		)
		return nil, wrapped
	}
	total, err := s.miniappListings.CountMiniapp(ctx, filter)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("count house listings: %w", err)
		slog.ErrorContext(ctx, "miniapp.house.search.failed", "step", "count_listings",
			"page", page,
			"page_size", pageSize,
			"partial_count", len(listings),
			"error", wrapped,
		)
		return nil, wrapped
	}
	result := &SearchResult{
		List:     ListingItems(listings),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	slog.InfoContext(ctx, "miniapp.house.search.success", "page", page,
		"page_size", pageSize,
		"returned_count", len(result.List),
		"total", total,
	)
	return result, nil
}
