package house

import "house-manager/internal/service/miniapp/listingview"

type SearchInput struct {
	City          string
	District      string
	BizArea       string
	RentMode      string
	AssetMode     string
	MinPrice      int
	MaxPrice      int
	RoomCount     *int
	HallCount     *int
	BathroomCount *int
	KitchenCount  *int
	Keyword       string
	FeatureFlags  []string
	Page          int
	PageSize      int
}

type SearchResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}

type ListItem = listingview.Item
type TaggedImage = listingview.TaggedImage
