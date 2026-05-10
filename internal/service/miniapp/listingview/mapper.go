package listingview

import "house-manager/internal/model"

func Items(listings []model.HpdMiniappListing) []Item {
	if len(listings) == 0 {
		return []Item{}
	}
	items := make([]Item, 0, len(listings))
	for _, listing := range listings {
		items = append(items, FromListing(listing))
	}
	return items
}

func FromListing(listing model.HpdMiniappListing) Item {
	return Item{
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
