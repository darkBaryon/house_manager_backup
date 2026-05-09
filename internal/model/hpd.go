package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionHpdListing        = "hs_hpd_listing"
	CollectionHpdMiniappListing = "hs_hpd_miniapp_listing"
)

// HpdListing 对应 hs_hpd_listing，发布主实体和统一 listing identity。
type HpdListing struct {
	CommonFields `bson:",inline"`

	SourceType    HpdSourceType    `bson:"source_type" json:"sourceType"`
	SourceID      bson.ObjectID    `bson:"source_id" json:"sourceId"`
	AssetMode     HpdAssetMode     `bson:"asset_mode" json:"assetMode"`
	ListingStatus HpdListingStatus `bson:"listing_status" json:"listingStatus"`
	PublishedAt   int64            `bson:"published_at" json:"publishedAt"`
	OfflineAt     int64            `bson:"offline_at" json:"offlineAt"`
}

// HpdMiniappListing 对应 hs_hpd_miniapp_listing，小程序列表与详情 read model。
type HpdMiniappListing struct {
	CommonFields `bson:",inline"`

	ListingID               bson.ObjectID     `bson:"listing_id" json:"listingId"`
	SourceType              HpdSourceType     `bson:"source_type" json:"sourceType"`
	SourceID                bson.ObjectID     `bson:"source_id" json:"sourceId"`
	AssetMode               HpdAssetMode      `bson:"asset_mode" json:"assetMode"`
	RentMode                RentMode          `bson:"rent_mode" json:"rentMode"`
	City                    string            `bson:"city" json:"city"`
	District                string            `bson:"district" json:"district"`
	BizArea                 string            `bson:"biz_area" json:"bizArea"`
	CommunityName           string            `bson:"community_name" json:"communityName"`
	BuildingOrCommunityName string            `bson:"building_or_community_name" json:"buildingOrCommunityName"`
	SubwayStation           string            `bson:"subway_station" json:"subwayStation"`
	SubwayDistanceM         int               `bson:"subway_distance_m" json:"subwayDistanceM"`
	AddressText             string            `bson:"address_text" json:"addressText"`
	Geo                     *GeoPoint         `bson:"geo,omitempty" json:"geo,omitempty"`
	Title                   string            `bson:"title" json:"title"`
	Subtitle                string            `bson:"subtitle" json:"subtitle"`
	Price                   int               `bson:"price" json:"price"`
	PriceText               string            `bson:"price_text" json:"priceText"`
	LayoutText              string            `bson:"layout_text" json:"layoutText"`
	AreaSize                int               `bson:"area_size" json:"areaSize"`
	Orientation             Orientation       `bson:"orientation" json:"orientation"`
	FloorText               string            `bson:"floor_text" json:"floorText"`
	PaymentCycle            PaymentCycle      `bson:"payment_cycle" json:"paymentCycle"`
	FeatureFlags            []string          `bson:"feature_flags" json:"featureFlags"`
	ListingFacilities       []ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
	PlatformTags            []string          `bson:"platform_tags" json:"platformTags"`
	StartRentRule           StartRentRule     `bson:"start_rent_rule" json:"startRentRule"`
	CostItems               []HpdCostItem     `bson:"cost_items" json:"costItems"`
	Description             string            `bson:"description" json:"description"`
	RiskNotice              string            `bson:"risk_notice" json:"riskNotice"`
	ContactPhone            string            `bson:"contact_phone" json:"contactPhone"`
	Images                  []TaggedImage     `bson:"images" json:"images"`
	WeightScore             int               `bson:"weight_score" json:"weightScore"`
	IsOnline                HpdOnlineStatus   `bson:"is_online" json:"isOnline"`
}

// HpdCostItem 是小程序详情费用明细项。
type HpdCostItem struct {
	Name   string `bson:"name" json:"name"`
	Amount int    `bson:"amount" json:"amount"`
	Unit   string `bson:"unit" json:"unit"`
	Remark string `bson:"remark" json:"remark"`
}
