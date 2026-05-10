package house

import housesvc "house-manager/internal/service/miniapp/house"

type detailResponse struct {
	House       detailHouseResponse `json:"house"`
	IsFavorited bool                `json:"is_favorited"`
}

type detailHouseResponse struct {
	listingItemResponse
	AddressText   string             `json:"address_text"`
	Geo           *geoPointResponse  `json:"geo,omitempty"`
	StartRentRule string             `json:"start_rent_rule"`
	CostItems     []costItemResponse `json:"cost_items"`
	Description   string             `json:"description"`
	RiskNotice    string             `json:"risk_notice"`
	ContactPhone  string             `json:"contact_phone"`
}

type geoPointResponse struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type costItemResponse struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
	Remark string `json:"remark"`
}

func toDetailResponse(result *housesvc.DetailResult) detailResponse {
	if result == nil {
		return detailResponse{House: detailHouseResponse{
			listingItemResponse: listingItemResponse{
				FeatureFlags:      []string{},
				ListingFacilities: []string{},
				PlatformTags:      []string{},
				Images:            []taggedImageResponse{},
			},
			CostItems: []costItemResponse{},
		}}
	}

	house := result.House
	resp := detailHouseResponse{
		listingItemResponse: listingItemResponses([]housesvc.ListItem{house.ListItem})[0],
		AddressText:         house.AddressText,
		StartRentRule:       house.StartRentRule,
		CostItems:           costItemResponses(house.CostItems),
		Description:         house.Description,
		RiskNotice:          house.RiskNotice,
		ContactPhone:        house.ContactPhone,
	}
	if house.Geo != nil {
		resp.Geo = &geoPointResponse{Lng: house.Geo.Lng, Lat: house.Geo.Lat}
	}
	return detailResponse{House: resp, IsFavorited: result.IsFavorited}
}

func costItemResponses(items []housesvc.CostItem) []costItemResponse {
	if len(items) == 0 {
		return []costItemResponse{}
	}
	out := make([]costItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, costItemResponse{
			Name:   item.Name,
			Amount: item.Amount,
			Unit:   item.Unit,
			Remark: item.Remark,
		})
	}
	return out
}
