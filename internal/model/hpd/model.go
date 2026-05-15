package hpd

import (
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionHpdListing          = "hs_hpd_listing"
	CollectionHpdMiniappListing   = "hs_hpd_miniapp_listing"
	CollectionHpdPublisherListing = "hs_hpd_publisher_listing"
	CollectionHpdRootScope        = "hs_hpd_root_scope_relation"
)

// HpdListing 对应 hs_hpd_listing，发布主实体和统一 listing identity。
type HpdListing struct {
	commonmodel.CommonFields `bson:",inline"`

	SourceType    HpdSourceType    `bson:"source_type" json:"sourceType"`
	SourceID      bson.ObjectID    `bson:"source_id" json:"sourceId"`
	AssetMode     HpdAssetMode     `bson:"asset_mode" json:"assetMode"`
	ListingStatus HpdListingStatus `bson:"listing_status" json:"listingStatus"`
	PublishedAt   int64            `bson:"published_at" json:"publishedAt"`
	OfflineAt     int64            `bson:"offline_at" json:"offlineAt"`
}

// HpdMiniappListing 对应 hs_hpd_miniapp_listing，小程序列表与详情 read model。
type HpdMiniappListing struct {
	commonmodel.CommonFields `bson:",inline"`

	ListingID               bson.ObjectID              `bson:"listing_id" json:"listingId"`
	SourceType              HpdSourceType              `bson:"source_type" json:"sourceType"`
	SourceID                bson.ObjectID              `bson:"source_id" json:"sourceId"`
	AssetMode               HpdAssetMode               `bson:"asset_mode" json:"assetMode"`
	RentMode                hmdmodel.RentMode          `bson:"rent_mode" json:"rentMode"`
	City                    string                     `bson:"city" json:"city"`
	District                string                     `bson:"district" json:"district"`
	BizArea                 string                     `bson:"biz_area" json:"bizArea"`
	CommunityName           string                     `bson:"community_name" json:"communityName"`
	BuildingOrCommunityName string                     `bson:"building_or_community_name" json:"buildingOrCommunityName"`
	SubwayStation           string                     `bson:"subway_station" json:"subwayStation"`
	SubwayDistanceM         int                        `bson:"subway_distance_m" json:"subwayDistanceM"`
	AddressText             string                     `bson:"address_text" json:"addressText"`
	Geo                     *hmdmodel.GeoPoint         `bson:"geo,omitempty" json:"geo,omitempty"`
	Title                   string                     `bson:"title" json:"title"`
	Subtitle                string                     `bson:"subtitle" json:"subtitle"`
	Price                   int                        `bson:"price" json:"price"`
	PriceText               string                     `bson:"price_text" json:"priceText"`
	LayoutText              string                     `bson:"layout_text" json:"layoutText"`
	AreaSize                int                        `bson:"area_size" json:"areaSize"`
	Orientation             hmdmodel.Orientation       `bson:"orientation" json:"orientation"`
	FloorText               string                     `bson:"floor_text" json:"floorText"`
	PaymentCycle            hmdmodel.PaymentCycle      `bson:"payment_cycle" json:"paymentCycle"`
	FeatureFlags            []string                   `bson:"feature_flags" json:"featureFlags"`
	ListingFacilities       []hmdmodel.ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
	PlatformTags            []string                   `bson:"platform_tags" json:"platformTags"`
	StartRentRule           hmdmodel.StartRentRule     `bson:"start_rent_rule" json:"startRentRule"`
	CostItems               []HpdCostItem              `bson:"cost_items" json:"costItems"`
	Description             string                     `bson:"description" json:"description"`
	RiskNotice              string                     `bson:"risk_notice" json:"riskNotice"`
	ContactPhone            string                     `bson:"contact_phone" json:"contactPhone"`
	Images                  []hmdmodel.TaggedImage     `bson:"images" json:"images"`
	WeightScore             int                        `bson:"weight_score" json:"weightScore"`
	IsOnline                HpdOnlineStatus            `bson:"is_online" json:"isOnline"`
}

// HpdPublisherListing 对应 hs_hpd_publisher_listing，房东端房源读模型。
type HpdPublisherListing struct {
	commonmodel.CommonFields `bson:",inline"`

	ListingID         bson.ObjectID              `bson:"listing_id" json:"listingId"`
	SourceType        HpdSourceType              `bson:"source_type" json:"sourceType"`
	SourceID          bson.ObjectID              `bson:"source_id" json:"sourceId"`
	AssetMode         HpdAssetMode               `bson:"asset_mode" json:"assetMode"`
	RootType          HpdRootScopeType           `bson:"root_type" json:"rootType"`
	RootID            bson.ObjectID              `bson:"root_id" json:"rootId"`
	ProjectID         bson.ObjectID              `bson:"project_id,omitempty" json:"projectId,omitempty"`
	ProjectName       string                     `bson:"project_name" json:"projectName"`
	BuildingID        bson.ObjectID              `bson:"building_id,omitempty" json:"buildingId,omitempty"`
	BuildingName      string                     `bson:"building_name" json:"buildingName"`
	RoomTypeID        bson.ObjectID              `bson:"room_type_id,omitempty" json:"roomTypeId,omitempty"`
	RoomTypeName      string                     `bson:"room_type_name" json:"roomTypeName"`
	DecentralizedID   bson.ObjectID              `bson:"decentralized_id,omitempty" json:"decentralizedId,omitempty"`
	CommunityName     string                     `bson:"community_name" json:"communityName"`
	RentMode          hmdmodel.RentMode          `bson:"rent_mode" json:"rentMode"`
	City              string                     `bson:"city" json:"city"`
	District          string                     `bson:"district" json:"district"`
	BizArea           string                     `bson:"biz_area" json:"bizArea"`
	SubwayStation     string                     `bson:"subway_station" json:"subwayStation"`
	AddressText       string                     `bson:"address_text" json:"addressText"`
	Geo               *hmdmodel.GeoPoint         `bson:"geo,omitempty" json:"geo,omitempty"`
	RoomNo            string                     `bson:"room_no" json:"roomNo"`
	FloorNo           int                        `bson:"floor_no" json:"floorNo"`
	Title             string                     `bson:"title" json:"title"`
	Subtitle          string                     `bson:"subtitle" json:"subtitle"`
	Price             int                        `bson:"price" json:"price"`
	PriceText         string                     `bson:"price_text" json:"priceText"`
	LayoutText        string                     `bson:"layout_text" json:"layoutText"`
	AreaSize          int                        `bson:"area_size" json:"areaSize"`
	Orientation       hmdmodel.Orientation       `bson:"orientation" json:"orientation"`
	DecorationLevel   hmdmodel.DecorationLevel   `bson:"decoration_level" json:"decorationLevel"`
	PaymentCycle      hmdmodel.PaymentCycle      `bson:"payment_cycle" json:"paymentCycle"`
	Deposit           int                        `bson:"deposit" json:"deposit"`
	ServiceFee        int                        `bson:"service_fee" json:"serviceFee"`
	AgencyFeeMode     hmdmodel.AgencyFeeMode     `bson:"agency_fee_mode" json:"agencyFeeMode"`
	AgencyFeeValue    int                        `bson:"agency_fee_value" json:"agencyFeeValue"`
	RoomStatus        hmdmodel.RoomStatus        `bson:"room_status" json:"roomStatus"`
	ListingStatus     HpdListingStatus           `bson:"listing_status" json:"listingStatus"`
	ViewingTimeRule   hmdmodel.ViewingTimeRule   `bson:"viewing_time_rule" json:"viewingTimeRule"`
	StartRentRule     hmdmodel.StartRentRule     `bson:"start_rent_rule" json:"startRentRule"`
	FeatureFlags      []string                   `bson:"feature_flags" json:"featureFlags"`
	ListingFacilities []hmdmodel.ListingFacility `bson:"listing_facilities" json:"listingFacilities"`
	RoomFacilities    []hmdmodel.RoomFacility    `bson:"room_facilities" json:"roomFacilities"`
	Images            []hmdmodel.TaggedImage     `bson:"images" json:"images"`
	IsOnline          HpdOnlineStatus            `bson:"is_online" json:"isOnline"`
}

// HpdCostItem 是小程序详情费用明细项。
type HpdCostItem struct {
	Name   string `bson:"name" json:"name"`
	Amount int    `bson:"amount" json:"amount"`
	Unit   string `bson:"unit" json:"unit"`
	Remark string `bson:"remark" json:"remark"`
}

// HpdRootScopeRelation 对应 hs_hpd_root_scope_relation，描述项目/小区级房东归属。
type HpdRootScopeRelation struct {
	commonmodel.CommonFields `bson:",inline"`

	RootType       HpdRootScopeType  `bson:"root_type" json:"rootType"`
	RootID         bson.ObjectID     `bson:"root_id" json:"rootId"`
	OwnerPhone     string            `bson:"owner_phone" json:"ownerPhone"`
	RelationStatus HpdRelationStatus `bson:"relation_status" json:"relationStatus"`
	EffectiveFrom  int64             `bson:"effective_from" json:"effectiveFrom"`
	EffectiveTo    int64             `bson:"effective_to" json:"effectiveTo"`
}
