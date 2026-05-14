package house

import (
	"context"

	"house-manager/pkg/errcode"
)

func (s *HouseService) Search(ctx context.Context, input SearchInput) (*SearchResult, error) {
	if s == nil || s.miniappListings == nil {
		return nil, errcode.DatabaseError.WithErrorf("house search repository is nil")
	}

	filter, page, pageSize, err := searchFilter(input)
	if err != nil {
		return nil, err
	}

	listings, err := s.miniappListings.SearchMiniapp(ctx, filter)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("search house listings: %w", err)
	}
	total, err := s.miniappListings.CountMiniapp(ctx, filter)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("count house listings: %w", err)
	}

	return &SearchResult{
		List:     ListingItems(listings),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
