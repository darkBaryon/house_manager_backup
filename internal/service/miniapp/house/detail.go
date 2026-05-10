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

	return &DetailResult{
		House:       listingDetail(*listing),
		IsFavorited: false,
	}, nil
}
