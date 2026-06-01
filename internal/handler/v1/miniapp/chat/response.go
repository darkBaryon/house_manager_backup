package chat

import chatsvc "house-manager/internal/service/chat"

type sendResponse struct {
	SessionID        string              `json:"session_id"`
	AssistantMessage string              `json:"assistant_message"`
	Sentences        []string            `json:"sentences"`
	HouseList        []houseItemResponse `json:"house_list"`
}

type houseItemResponse struct {
	ListingID       string                `json:"listing_id"`
	Title           string                `json:"title"`
	Price           int                   `json:"price"`
	PriceText       string                `json:"price_text"`
	District        string                `json:"district"`
	BizArea         string                `json:"biz_area"`
	LayoutText      string                `json:"layout_text"`
	SubwayDistanceM int                   `json:"subway_distance_m"`
	Images          []taggedImageResponse `json:"images"`
}

type taggedImageResponse struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func toSendResponse(result *chatsvc.SendResult) sendResponse {
	if result == nil {
		return sendResponse{Sentences: []string{}, HouseList: []houseItemResponse{}}
	}
	return sendResponse{
		SessionID:        result.SessionID,
		AssistantMessage: result.AssistantMessage,
		Sentences:        stringSlice(result.Sentences),
		HouseList:        houseItemResponses(result.HouseList),
	}
}

func houseItemResponses(items []chatsvc.HouseItem) []houseItemResponse {
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
			SubwayDistanceM: item.SubwayDistanceM,
			Images:          taggedImageResponses(item.Images),
		})
	}
	return out
}

func taggedImageResponses(images []chatsvc.TaggedImage) []taggedImageResponse {
	if len(images) == 0 {
		return []taggedImageResponse{}
	}
	out := make([]taggedImageResponse, 0, len(images))
	for _, image := range images {
		out = append(out, taggedImageResponse{URL: image.URL, Tag: image.Tag})
	}
	return out
}

func stringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}
