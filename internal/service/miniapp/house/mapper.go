package house

import (
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/service/miniapp/listingview"
)

func ListingItems(listings []hpdmodel.HpdMiniappListing) []ListItem {
	return listingview.Items(listings)
}

func ListingItem(listing hpdmodel.HpdMiniappListing) ListItem {
	return listingview.FromListing(listing)
}

func listingDetail(listing hpdmodel.HpdMiniappListing) Detail {
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

func costItems(items []hpdmodel.HpdCostItem) []CostItem {
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
