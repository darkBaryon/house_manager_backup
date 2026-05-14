package house

import (
	"strings"

	"house-manager/internal/model"
	repohpd "house-manager/internal/repository/hpd"
	"house-manager/pkg/errcode"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 50
)

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func searchFilter(input SearchInput) (repohpd.MiniappListingSearchFilter, int, int, error) {
	page, pageSize := normalizePage(input.Page, input.PageSize)
	if input.MinPrice < 0 || input.MaxPrice < 0 {
		return repohpd.MiniappListingSearchFilter{}, 0, 0, errcode.InvalidParam.WithErrorf("price range must be non-negative")
	}
	if input.MaxPrice > 0 && input.MinPrice > input.MaxPrice {
		return repohpd.MiniappListingSearchFilter{}, 0, 0, errcode.InvalidParam.WithErrorf("min_price must be less than or equal to max_price")
	}

	rentMode := model.RentMode(strings.TrimSpace(input.RentMode))
	if rentMode != "" && !rentMode.Valid() {
		return repohpd.MiniappListingSearchFilter{}, 0, 0, errcode.InvalidParam.WithErrorf("rent_mode is invalid")
	}

	assetMode := model.HpdAssetMode(strings.TrimSpace(input.AssetMode))
	if assetMode != "" && !assetMode.Valid() {
		return repohpd.MiniappListingSearchFilter{}, 0, 0, errcode.InvalidParam.WithErrorf("asset_mode is invalid")
	}

	filter := repohpd.MiniappListingSearchFilter{
		City:         strings.TrimSpace(input.City),
		District:     strings.TrimSpace(input.District),
		BizArea:      strings.TrimSpace(input.BizArea),
		RentMode:     rentMode,
		AssetMode:    assetMode,
		Keyword:      strings.TrimSpace(input.Keyword),
		FeatureFlags: compactStrings(input.FeatureFlags),
		PriceMin:     input.MinPrice,
		PriceMax:     input.MaxPrice,
		Skip:         int64((page - 1) * pageSize),
		Limit:        int64(pageSize),
	}
	return filter, page, pageSize, nil
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
