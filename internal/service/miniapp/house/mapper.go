package house

import (
	"house-manager/internal/model"
	"house-manager/internal/service/miniapp/listingview"
)

func ListingItems(listings []model.HpdMiniappListing) []ListItem {
	return listingview.Items(listings)
}

func ListingItem(listing model.HpdMiniappListing) ListItem {
	return listingview.FromListing(listing)
}

func listingDetail(listing model.HpdMiniappListing) Detail {
	detail := Detail{
		ListItem:      ListingItem(listing),
		AddressText:   listing.AddressText,
		StartRentRule: string(listing.StartRentRule),
		CostItems:     costItems(listing.CostItems),
		Description:   listing.Description,
		RiskNotice:    listing.RiskNotice,
		ContactPhone:  listing.ContactPhone,
	}
	if listing.Geo != nil {
		detail.Geo = &GeoPoint{Lng: listing.Geo.Lng, Lat: listing.Geo.Lat}
	}
	return detail
}

func costItems(items []model.HpdCostItem) []CostItem {
	if len(items) == 0 {
		return []CostItem{}
	}
	out := make([]CostItem, 0, len(items))
	for _, item := range items {
		out = append(out, CostItem{
			Name:   item.Name,
			Amount: item.Amount,
			Unit:   item.Unit,
			Remark: item.Remark,
		})
	}
	return out
}
