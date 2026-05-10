package house

import "context"

func (s *HouseService) GetPublicDetail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	if s == nil || s.miniappListings == nil {
		return nil, databasef("house detail repository is nil")
	}
	if input.ListingID.IsZero() {
		return nil, invalidParamf("listing_id is required")
	}

	listing, err := s.miniappListings.FindOnlineDetail(ctx, input.ListingID)
	if err != nil {
		return nil, databasef("find house public detail: %w", err)
	}
	if listing == nil {
		return nil, notFoundf("house public detail not found")
	}

	isFavorited := false
	if !input.UserID.IsZero() && s.favorites != nil {
		ok, err := s.favorites.IsFavorited(ctx, input.UserID, input.ListingID)
		if err != nil {
			return nil, databasef("find house favorite status: %w", err)
		}
		isFavorited = ok
	}

	return &DetailResult{
		House:       listingDetail(*listing),
		IsFavorited: isFavorited,
	}, nil
}
