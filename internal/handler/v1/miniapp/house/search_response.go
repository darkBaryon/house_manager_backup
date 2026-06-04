package house

import housesvc "house-manager/internal/service/miniapp/house"

type searchResponse struct {
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
	District                string                `json:"district"`
	BizArea                 string                `json:"biz_area"`
	CommunityName           string                `json:"community_name"`
	BuildingOrCommunityName string                `json:"building_or_community_name"`
	SubwayStation           string                `json:"subway_station"`
	SubwayDistanceM         int                   `json:"subway_distance_m"`
	Title                   string                `json:"title"`
	Subtitle                string                `json:"subtitle"`
	Price                   int                   `json:"price"`
	PriceText               string                `json:"price_text"`
	LayoutText              string                `json:"layout_text"`
	RoomCount               int                   `json:"room_count"`
	HallCount               int                   `json:"hall_count"`
	BathroomCount           int                   `json:"bathroom_count"`
	KitchenCount            int                   `json:"kitchen_count"`
	AreaSize                int                   `json:"area_size"`
	Orientation             string                `json:"orientation"`
	FloorText               string                `json:"floor_text"`
	PaymentCycle            string                `json:"payment_cycle"`
	FeatureFlags            []string              `json:"feature_flags"`
	ListingFacilities       []string              `json:"listing_facilities"`
	PlatformTags            []string              `json:"platform_tags"`
	Images                  []taggedImageResponse `json:"images"`
}

type taggedImageResponse struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func toSearchResponse(result *housesvc.SearchResult) searchResponse {
	if result == nil {
		return searchResponse{List: []listingItemResponse{}}
	}
	return searchResponse{
		List:     listingItemResponses(result.List),
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func listingItemResponses(items []housesvc.ListItem) []listingItemResponse {
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
			RoomCount:               item.RoomCount,
			HallCount:               item.HallCount,
			BathroomCount:           item.BathroomCount,
			KitchenCount:            item.KitchenCount,
			AreaSize:                item.AreaSize,
			Orientation:             item.Orientation,
			FloorText:               item.FloorText,
			PaymentCycle:            item.PaymentCycle,
			FeatureFlags:            item.FeatureFlags,
			ListingFacilities:       item.ListingFacilities,
			PlatformTags:            item.PlatformTags,
			Images:                  taggedImageResponses(item.Images),
		})
	}
	return out
}

func taggedImageResponses(images []housesvc.TaggedImage) []taggedImageResponse {
	if len(images) == 0 {
		return []taggedImageResponse{}
	}
	out := make([]taggedImageResponse, 0, len(images))
	for _, image := range images {
		out = append(out, taggedImageResponse{URL: image.URL, Tag: image.Tag})
	}
	return out
}
