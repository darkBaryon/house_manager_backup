package house

import housesvc "house-manager/internal/service/miniapp/house"

type searchResponse struct {
	HouseList []houseItemResponse `json:"house_list"`
}

type houseItemResponse struct {
	ListingID       string                `json:"listing_id"`
	Title           string                `json:"title"`
	Price           int                   `json:"price"`
	PriceText       string                `json:"price_text"`
	District        string                `json:"district"`
	BizArea         string                `json:"biz_area"`
	LayoutText      string                `json:"layout_text"`
	RoomCount       int                   `json:"room_count"`
	HallCount       int                   `json:"hall_count"`
	BathroomCount   int                   `json:"bathroom_count"`
	KitchenCount    int                   `json:"kitchen_count"`
	SubwayDistanceM int                   `json:"subway_distance_m"`
	Images          []taggedImageResponse `json:"images"`
}

type taggedImageResponse struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

type publicDetailResponse struct {
	ListingID       string                `json:"listing_id"`
	Title           string                `json:"title"`
	Price           int                   `json:"price"`
	PriceText       string                `json:"price_text"`
	District        string                `json:"district"`
	BizArea         string                `json:"biz_area"`
	LayoutText      string                `json:"layout_text"`
	RoomCount       int                   `json:"room_count"`
	HallCount       int                   `json:"hall_count"`
	BathroomCount   int                   `json:"bathroom_count"`
	KitchenCount    int                   `json:"kitchen_count"`
	SubwayDistanceM int                   `json:"subway_distance_m"`
	AddressText     string                `json:"address_text"`
	Description     string                `json:"description"`
	Images          []taggedImageResponse `json:"images"`
}

func toSearchResponse(result *housesvc.SearchResult) searchResponse {
	if result == nil || len(result.List) == 0 {
		return searchResponse{HouseList: []houseItemResponse{}}
	}
	return searchResponse{HouseList: houseItemResponses(result.List)}
}

func houseItemResponses(items []housesvc.ListItem) []houseItemResponse {
	if len(items) == 0 {
		return []houseItemResponse{}
	}
	out := make([]houseItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, houseItemResponse{
			ListingID:       item.ListingID,
			Title:           item.Title,
			Price:           item.Price,
			PriceText:       item.PriceText,
			District:        item.District,
			BizArea:         item.BizArea,
			LayoutText:      item.LayoutText,
			RoomCount:       item.RoomCount,
			HallCount:       item.HallCount,
			BathroomCount:   item.BathroomCount,
			KitchenCount:    item.KitchenCount,
			SubwayDistanceM: item.SubwayDistanceM,
			Images:          taggedImageResponses(item.Images),
		})
	}
	return out
}

func toPublicDetailResponse(result *housesvc.DetailResult) publicDetailResponse {
	if result == nil {
		return publicDetailResponse{Images: []taggedImageResponse{}}
	}
	house := result.House
	return publicDetailResponse{
		ListingID:       house.ListingID,
		Title:           house.Title,
		Price:           house.Price,
		PriceText:       house.PriceText,
		District:        house.District,
		BizArea:         house.BizArea,
		LayoutText:      house.LayoutText,
		RoomCount:       house.RoomCount,
		HallCount:       house.HallCount,
		BathroomCount:   house.BathroomCount,
		KitchenCount:    house.KitchenCount,
		SubwayDistanceM: house.SubwayDistanceM,
		AddressText:     house.AddressText,
		Description:     house.Description,
		Images:          taggedImageResponses(house.Images),
	}
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
