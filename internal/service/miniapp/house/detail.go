package house

import (
	"context"
	"log/slog"

	"house-manager/pkg/errcode"
)

func (s *HouseService) GetPublicDetail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	attrs := []any{"listing_id", input.ListingID.Hex()}
	if !input.UserID.IsZero() {
		attrs = append(attrs, "user_id", input.UserID.Hex())
	}
	slog.InfoContext(ctx, "miniapp.house.detail.start", attrs...)
	if s == nil || s.miniappListings == nil {
		err := errcode.DatabaseError.WithErrorf("house detail repository is nil")
		slog.ErrorContext(ctx, "miniapp.house.detail.failed", "listing_id", input.ListingID.Hex(),
			"error", err,
		)
		return nil, err
	}
	if input.ListingID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("listing_id is required")
		slog.WarnContext(ctx, "miniapp.house.detail.failed", "error", err)
		return nil, err
	}

	listing, err := s.miniappListings.FindOnlineDetail(ctx, input.ListingID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("find house public detail: %w", err)
		slog.ErrorContext(ctx, "miniapp.house.detail.failed", "listing_id", input.ListingID.Hex(),
			"step", "find_online_detail",
			"error", wrapped,
		)
		return nil, wrapped
	}
	if listing == nil {
		err := errcode.NotFound.WithErrorf("house public detail not found")
		slog.WarnContext(ctx, "miniapp.house.detail.failed", "listing_id", input.ListingID.Hex(),
			"step", "find_online_detail",
			"error", err,
		)
		return nil, err
	}

	isFavorited := false
	if !input.UserID.IsZero() && s.favorites != nil {
		ok, err := s.favorites.IsFavorited(ctx, input.UserID, input.ListingID)
		if err != nil {
			wrapped := errcode.DatabaseError.WithErrorf("find house favorite status: %w", err)
			slog.ErrorContext(ctx, "miniapp.house.detail.failed", "listing_id", input.ListingID.Hex(),
				"user_id", input.UserID.Hex(),
				"step", "favorite_status",
				"error", wrapped,
			)
			return nil, wrapped
		}
		isFavorited = ok
	}
	result := &DetailResult{
		House:       listingDetail(*listing),
		IsFavorited: isFavorited,
	}
	slog.InfoContext(ctx, "miniapp.house.detail.success", "listing_id", input.ListingID.Hex(),
		"user_id_present", !input.UserID.IsZero(),
		"is_favorited", isFavorited,
	)
	return result, nil
}
