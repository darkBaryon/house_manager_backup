package favorite

import (
	favoritesvc "house-manager/internal/service/miniapp/favorite"
	"house-manager/internal/service/miniapp/listingview"
)

type mutationResponse struct {
	ListingID   string `json:"listing_id"`
	IsFavorited bool   `json:"is_favorited"`
}

type listResponse struct {
	List     []listingItemResponse `json:"list"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

type listingItemResponse struct {
	ListingID               string                `json:"listing_id"`
	AssetMode               string                `json:"asset_mode"`
	RentMode                string                `json:"rent_mode"`
	City                    string                `json:"city"`
	District                string                `json:"district,omitempty"`
	BizArea                 string                `json:"biz_area,omitempty"`
	CommunityName           string                `json:"community_name,omitempty"`
	BuildingOrCommunityName string                `json:"building_or_community_name,omitempty"`
	SubwayStation           string                `json:"subway_station,omitempty"`
	SubwayDistanceM         int                   `json:"subway_distance_m,omitempty"`
	Title                   string                `json:"title"`
	Subtitle                string                `json:"subtitle,omitempty"`
	Price                   int                   `json:"price"`
	PriceText               string                `json:"price_text,omitempty"`
	LayoutText              string                `json:"layout_text,omitempty"`
	AreaSize                int                   `json:"area_size,omitempty"`
	Orientation             string                `json:"orientation,omitempty"`
	FloorText               string                `json:"floor_text,omitempty"`
	PaymentCycle            string                `json:"payment_cycle,omitempty"`
	FeatureFlags            []string              `json:"feature_flags"`
	ListingFacilities       []string              `json:"listing_facilities"`
	PlatformTags            []string              `json:"platform_tags"`
	Images                  []taggedImageResponse `json:"images"`
}

type taggedImageResponse struct {
	URL string `json:"url"`
	Tag string `json:"tag,omitempty"`
}

func toMutationResponse(result *favoritesvc.MutationResult) mutationResponse {
	return mutationResponse{ListingID: result.ListingID, IsFavorited: result.IsFavorited}
}

func toListResponse(result *favoritesvc.ListResult) listResponse {
	return listResponse{
		List:     toListingItemResponses(result.List),
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toListingItemResponses(items []listingview.Item) []listingItemResponse {
	if len(items) == 0 {
		return []listingItemResponse{}
	}
	out := make([]listingItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, listingItemResponse{
			ListingID:               item.ListingID,
			AssetMode:               item.AssetMode,
			RentMode:                item.RentMode,
			City:                    item.City,
			District:                item.District,
			BizArea:                 item.BizArea,
			CommunityName:           item.CommunityName,
			BuildingOrCommunityName: item.BuildingOrCommunityName,
			SubwayStation:           item.SubwayStation,
			SubwayDistanceM:         item.SubwayDistanceM,
			Title:                   item.Title,
			Subtitle:                item.Subtitle,
			Price:                   item.Price,
			PriceText:               item.PriceText,
			LayoutText:              item.LayoutText,
			AreaSize:                item.AreaSize,
			Orientation:             item.Orientation,
			FloorText:               item.FloorText,
			PaymentCycle:            item.PaymentCycle,
			FeatureFlags:            item.FeatureFlags,
			ListingFacilities:       item.ListingFacilities,
			PlatformTags:            item.PlatformTags,
			Images:                  toTaggedImageResponses(item.Images),
		})
	}
	return out
}

func toTaggedImageResponses(images []listingview.TaggedImage) []taggedImageResponse {
	if len(images) == 0 {
		return []taggedImageResponse{}
	}
	out := make([]taggedImageResponse, 0, len(images))
	for _, image := range images {
		out = append(out, taggedImageResponse{URL: image.URL, Tag: image.Tag})
	}
	return out
}
