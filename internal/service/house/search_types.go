package house

type SearchInput struct {
	City         string
	District     string
	BizArea      string
	RentMode     string
	AssetMode    string
	MinPrice     int
	MaxPrice     int
	Keyword      string
	FeatureFlags []string
	Page         int
	PageSize     int
}

type SearchResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}

type ListItem struct {
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
