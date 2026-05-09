package publish

type decentralizedCommunityRequest struct {
	ID            string           `json:"id"`
	CommunityName string           `json:"community_name" binding:"required"`
	City          string           `json:"city" binding:"required"`
	District      string           `json:"district"`
	BizArea       string           `json:"biz_area"`
	AddressText   string           `json:"address_text"`
	Geo           *geoPointRequest `json:"geo"`
	SubwayStation string           `json:"subway_station"`
}
