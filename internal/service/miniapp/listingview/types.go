package listingview

type Item struct {
	ListingID               string
	AssetMode               string
	RentMode                string
	City                    string
	District                string
	BizArea                 string
	CommunityName           string
	BuildingOrCommunityName string
	SubwayStation           string
	SubwayDistanceM         int
	Title                   string
	Subtitle                string
	Price                   int
	PriceText               string
	LayoutText              string
	RoomCount               int
	HallCount               int
	BathroomCount           int
	KitchenCount            int
	AreaSize                int
	Orientation             string
	FloorText               string
	PaymentCycle            string
	FeatureFlags            []string
	ListingFacilities       []string
	PlatformTags            []string
	Images                  []TaggedImage
}

type TaggedImage struct {
	URL string
	Tag string
}
