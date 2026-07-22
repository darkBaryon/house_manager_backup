package pythonchat

import chatservice "house-manager/internal/service/chat"

type respondData struct {
	AssistantMessage string         `json:"assistant_message"`
	Sentences        []string       `json:"sentences"`
	HouseList        []houseItem    `json:"house_list"`
	UpdatedContext   map[string]any `json:"updated_context"`
	SafetyResult     map[string]any `json:"safety_result"`
	IntentResult     map[string]any `json:"intent_result"`
}

type houseItem struct {
	ListingID       string        `json:"listing_id"`
	Title           string        `json:"title"`
	Price           int           `json:"price"`
	PriceText       string        `json:"price_text"`
	District        string        `json:"district"`
	BizArea         string        `json:"biz_area"`
	LayoutText      string        `json:"layout_text"`
	SubwayDistanceM int           `json:"subway_distance_m"`
	Images          []taggedImage `json:"images"`
}

type taggedImage struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func toServiceOutput(data *respondData) chatservice.AIRespondOutput {
	if data == nil {
		return chatservice.AIRespondOutput{}
	}
	return chatservice.AIRespondOutput{
		AssistantMessage: data.AssistantMessage,
		Sentences:        data.Sentences,
		HouseList:        toServiceHouseItems(data.HouseList),
		UpdatedContext:   data.UpdatedContext,
		SafetyResult:     data.SafetyResult,
		IntentResult:     data.IntentResult,
	}
}

func toServiceHouseItems(items []houseItem) []chatservice.HouseItem {
	if len(items) == 0 {
		return []chatservice.HouseItem{}
	}
	out := make([]chatservice.HouseItem, 0, len(items))
	for _, item := range items {
		out = append(out, chatservice.HouseItem{
			ListingID:       item.ListingID,
			Title:           item.Title,
			Price:           item.Price,
			PriceText:       item.PriceText,
			District:        item.District,
			BizArea:         item.BizArea,
			LayoutText:      item.LayoutText,
			SubwayDistanceM: item.SubwayDistanceM,
			Images:          toServiceTaggedImages(item.Images),
		})
	}
	return out
}

func toServiceTaggedImages(images []taggedImage) []chatservice.TaggedImage {
	if len(images) == 0 {
		return []chatservice.TaggedImage{}
	}
	out := make([]chatservice.TaggedImage, 0, len(images))
	for _, image := range images {
		out = append(out, chatservice.TaggedImage{URL: image.URL, Tag: image.Tag})
	}
	return out
}
