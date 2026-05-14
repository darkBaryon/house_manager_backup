package house

import (
	"context"

	"house-manager/pkg/errcode"
)

func (s *HouseService) GetPublicDetail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	if s == nil || s.miniappListings == nil {
		return nil, errcode.DatabaseError.WithErrorf("house detail repository is nil")
	}
	if input.ListingID.IsZero() {
		return nil, errcode.InvalidParam.WithErrorf("listing_id is required")
	}

	listing, err := s.miniappListings.FindOnlineDetail(ctx, input.ListingID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("find house public detail: %w", err)
	}
	if listing == nil {
		return nil, errcode.NotFound.WithErrorf("house public detail not found")
	}

	isFavorited := false
	if !input.UserID.IsZero() && s.favorites != nil {
		ok, err := s.favorites.IsFavorited(ctx, input.UserID, input.ListingID)
		if err != nil {
			return nil, errcode.DatabaseError.WithErrorf("find house favorite status: %w", err)
		}
		isFavorited = ok
	}

	return &DetailResult{
		House:       listingDetail(*listing),
		IsFavorited: isFavorited,
	}, nil
}
