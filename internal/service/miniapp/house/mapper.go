package house

import "house-manager/internal/model"

func ListingItems(listings []model.HpdMiniappListing) []ListItem {
	if len(listings) == 0 {
		return []ListItem{}
	}
	items := make([]ListItem, 0, len(listings))
	for _, listing := range listings {
		items = append(items, ListingItem(listing))
	}
	return items
}

func ListingItem(listing model.HpdMiniappListing) ListItem {
	return ListItem{
		ListingID:               listing.ListingID.Hex(),
		AssetMode:               string(listing.AssetMode),
		RentMode:                string(listing.RentMode),
		City:                    listing.City,
		District:                listing.District,
		BizArea:                 listing.BizArea,
		CommunityName:           listing.CommunityName,
		BuildingOrCommunityName: listing.BuildingOrCommunityName,
		SubwayStation:           listing.SubwayStation,
		SubwayDistanceM:         listing.SubwayDistanceM,
		Title:                   listing.Title,
		Subtitle:                listing.Subtitle,
		Price:                   listing.Price,
		PriceText:               listing.PriceText,
		LayoutText:              listing.LayoutText,
		AreaSize:                listing.AreaSize,
		Orientation:             string(listing.Orientation),
		FloorText:               listing.FloorText,
		PaymentCycle:            string(listing.PaymentCycle),
		FeatureFlags:            cloneStrings(listing.FeatureFlags),
		ListingFacilities:       listingFacilityStrings(listing.ListingFacilities),
		PlatformTags:            cloneStrings(listing.PlatformTags),
		Images:                  taggedImages(listing.Images),
	}
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

func taggedImages(images []model.TaggedImage) []TaggedImage {
	if len(images) == 0 {
		return []TaggedImage{}
	}
	out := make([]TaggedImage, 0, len(images))
	for _, image := range images {
		out = append(out, TaggedImage{URL: image.URL, Tag: string(image.Tag)})
	}
	return out
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

func listingFacilityStrings(items []model.ListingFacility) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, string(item))
	}
	return out
}

func cloneStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}
