package hpd

import (
	"context"
	"fmt"

	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AdminRootListFilter struct {
	OwnerLandlordID bson.ObjectID
	AssetMode       hpdmodel.HpdAssetMode
	City            string
	District        string
	RoomStatus      *hmdmodel.RoomStatus
	ListingStatus   *hpdmodel.HpdListingStatus
	AuditStatus     *hpdmodel.HpdAuditStatus
	Skip            int64
	Limit           int64
}

type AdminRootListItem struct {
	RootID        bson.ObjectID             `bson:"root_id"`
	RootType      hpdmodel.HpdRootScopeType `bson:"root_type"`
	AssetMode     hpdmodel.HpdAssetMode     `bson:"asset_mode"`
	ProviderID    bson.ObjectID             `bson:"provider_id"`
	ProviderPhone string                    `bson:"provider_phone"`
	ProviderName  string                    `bson:"provider_name"`
	ProjectID     bson.ObjectID             `bson:"project_id,omitempty"`
	ProjectName   string                    `bson:"project_name"`
	CommunityID   bson.ObjectID             `bson:"community_id,omitempty"`
	CommunityName string                    `bson:"community_name"`
	City          string                    `bson:"city"`
	District      string                    `bson:"district"`
	BizArea       string                    `bson:"biz_area"`
	BuildingCount int                       `bson:"building_count"`
	RoomCount     int64                     `bson:"room_count"`
	UpdatedAt     int64                     `bson:"updated_at"`
}

type AdminBuildingListFilter struct {
	RootID        bson.ObjectID
	RoomStatus    *hmdmodel.RoomStatus
	ListingStatus *hpdmodel.HpdListingStatus
	AuditStatus   *hpdmodel.HpdAuditStatus
	Skip          int64
	Limit         int64
}

type AdminBuildingListItem struct {
	RootID       bson.ObjectID `bson:"root_id"`
	BuildingID   bson.ObjectID `bson:"building_id"`
	ProjectID    bson.ObjectID `bson:"project_id,omitempty"`
	ProjectName  string        `bson:"project_name"`
	BuildingName string        `bson:"building_name"`
	City         string        `bson:"city"`
	District     string        `bson:"district"`
	BizArea      string        `bson:"biz_area"`
	RoomCount    int64         `bson:"room_count"`
	UpdatedAt    int64         `bson:"updated_at"`
}

type aggregateListResult[T any] struct {
	Items []T `bson:"items"`
	Total []struct {
		Value int64 `bson:"value"`
	} `bson:"total"`
}

func (r *AdminListingRepository) ListAdminRoots(ctx context.Context, input AdminRootListFilter) ([]AdminRootListItem, int64, error) {
	match, err := adminRootListFilter(input)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin roots: %w", err)
	}
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "root_id", Value: "$root_id"},
				{Key: "root_type", Value: "$root_type"},
				{Key: "asset_mode", Value: "$asset_mode"},
				{Key: "provider_id", Value: "$owner_landlord_id"},
				{Key: "provider_phone", Value: "$owner_phone_snapshot"},
				{Key: "provider_name", Value: "$landlord_name_snapshot"},
				{Key: "project_id", Value: "$project_id"},
				{Key: "project_name", Value: "$project_name"},
				{Key: "community_id", Value: "$decentralized_id"},
				{Key: "community_name", Value: "$community_name"},
				{Key: "city", Value: "$city"},
				{Key: "district", Value: "$district"},
				{Key: "biz_area", Value: "$biz_area"},
			}},
			{Key: "building_ids", Value: bson.D{{Key: "$addToSet", Value: "$building_id"}}},
			{Key: "room_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "updated_at", Value: bson.D{{Key: "$max", Value: "$updated_at"}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "root_id", Value: "$_id.root_id"},
			{Key: "root_type", Value: "$_id.root_type"},
			{Key: "asset_mode", Value: "$_id.asset_mode"},
			{Key: "provider_id", Value: "$_id.provider_id"},
			{Key: "provider_phone", Value: "$_id.provider_phone"},
			{Key: "provider_name", Value: "$_id.provider_name"},
			{Key: "project_id", Value: "$_id.project_id"},
			{Key: "project_name", Value: "$_id.project_name"},
			{Key: "community_id", Value: "$_id.community_id"},
			{Key: "community_name", Value: "$_id.community_name"},
			{Key: "city", Value: "$_id.city"},
			{Key: "district", Value: "$_id.district"},
			{Key: "biz_area", Value: "$_id.biz_area"},
			{Key: "building_count", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$building_ids"},
				{Key: "as", Value: "building_id"},
				{Key: "cond", Value: bson.D{{Key: "$and", Value: bson.A{
					bson.D{{Key: "$ne", Value: bson.A{"$$building_id", nil}}},
					bson.D{{Key: "$ne", Value: bson.A{"$$building_id", bson.NilObjectID}}},
				}}}},
			}}}}}},
			{Key: "room_count", Value: 1},
			{Key: "updated_at", Value: 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "updated_at", Value: -1}, {Key: "root_id", Value: -1}}}},
		bson.D{{Key: "$facet", Value: facetStages(input.Skip, input.Limit)}},
	}
	result, err := aggregateAdminList[AdminRootListItem](ctx, r, pipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin roots: %w", err)
	}
	return result.Items, aggregateTotal(result.Total), nil
}

func (r *AdminListingRepository) ListAdminBuildings(ctx context.Context, input AdminBuildingListFilter) ([]AdminBuildingListItem, int64, error) {
	match, err := adminBuildingListFilter(input)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin buildings: %w", err)
	}
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "root_id", Value: "$root_id"},
				{Key: "building_id", Value: "$building_id"},
				{Key: "project_id", Value: "$project_id"},
				{Key: "project_name", Value: "$project_name"},
				{Key: "building_name", Value: "$building_name"},
				{Key: "city", Value: "$city"},
				{Key: "district", Value: "$district"},
				{Key: "biz_area", Value: "$biz_area"},
			}},
			{Key: "room_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "updated_at", Value: bson.D{{Key: "$max", Value: "$updated_at"}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "root_id", Value: "$_id.root_id"},
			{Key: "building_id", Value: "$_id.building_id"},
			{Key: "project_id", Value: "$_id.project_id"},
			{Key: "project_name", Value: "$_id.project_name"},
			{Key: "building_name", Value: "$_id.building_name"},
			{Key: "city", Value: "$_id.city"},
			{Key: "district", Value: "$_id.district"},
			{Key: "biz_area", Value: "$_id.biz_area"},
			{Key: "room_count", Value: 1},
			{Key: "updated_at", Value: 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "updated_at", Value: -1}, {Key: "building_id", Value: -1}}}},
		bson.D{{Key: "$facet", Value: facetStages(input.Skip, input.Limit)}},
	}
	result, err := aggregateAdminList[AdminBuildingListItem](ctx, r, pipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin buildings: %w", err)
	}
	return result.Items, aggregateTotal(result.Total), nil
}

func facetStages(skip, limit int64) bson.D {
	items := bson.A{}
	if skip > 0 {
		items = append(items, bson.D{{Key: "$skip", Value: skip}})
	}
	if limit > 0 {
		items = append(items, bson.D{{Key: "$limit", Value: limit}})
	}
	return bson.D{
		{Key: "items", Value: items},
		{Key: "total", Value: bson.A{bson.D{{Key: "$count", Value: "value"}}}},
	}
}

func aggregateAdminList[T any](ctx context.Context, r *AdminListingRepository, pipeline bson.A) (*aggregateListResult[T], error) {
	mongoPipeline := make(mongo.Pipeline, 0, len(pipeline))
	for _, stage := range pipeline {
		doc, ok := stage.(bson.D)
		if !ok {
			return nil, fmt.Errorf("aggregate stage must be bson.D, got %T", stage)
		}
		mongoPipeline = append(mongoPipeline, doc)
	}

	var results []aggregateListResult[T]
	if err := r.Aggregate(ctx, mongoPipeline, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &aggregateListResult[T]{Items: []T{}, Total: nil}, nil
	}
	if results[0].Items == nil {
		results[0].Items = []T{}
	}
	return &results[0], nil
}

func aggregateTotal(rows []struct {
	Value int64 `bson:"value"`
}) int64 {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].Value
}

func adminRootListFilter(input AdminRootListFilter) (bson.M, error) {
	return buildAdminHierarchyFilter(input.OwnerLandlordID, input.AssetMode, bson.NilObjectID, bson.NilObjectID, input.City, input.District, input.RoomStatus, input.ListingStatus, input.AuditStatus)
}

func adminBuildingListFilter(input AdminBuildingListFilter) (bson.M, error) {
	if input.RootID.IsZero() {
		return nil, fmt.Errorf("rootID is required")
	}
	filter, err := buildAdminHierarchyFilter(bson.NilObjectID, "", input.RootID, bson.NilObjectID, "", "", input.RoomStatus, input.ListingStatus, input.AuditStatus)
	if err != nil {
		return nil, err
	}
	filter["building_id"] = bson.M{"$exists": true, "$ne": bson.NilObjectID}
	return filter, nil
}

func buildAdminHierarchyFilter(
	ownerLandlordID bson.ObjectID,
	assetMode hpdmodel.HpdAssetMode,
	rootID bson.ObjectID,
	buildingID bson.ObjectID,
	city string,
	district string,
	roomStatus *hmdmodel.RoomStatus,
	listingStatus *hpdmodel.HpdListingStatus,
	auditStatus *hpdmodel.HpdAuditStatus,
) (bson.M, error) {
	if assetMode != "" && !assetMode.Valid() {
		return nil, fmt.Errorf("assetMode is invalid")
	}
	if roomStatus != nil && !roomStatus.Valid() {
		return nil, fmt.Errorf("roomStatus is invalid")
	}
	if listingStatus != nil && !listingStatus.Valid() {
		return nil, fmt.Errorf("listingStatus is invalid")
	}
	if auditStatus != nil && !auditStatus.Valid() {
		return nil, fmt.Errorf("auditStatus is invalid")
	}

	fields := bson.M{}
	if !ownerLandlordID.IsZero() {
		fields["owner_landlord_id"] = ownerLandlordID
	}
	if assetMode != "" {
		fields["asset_mode"] = assetMode
	}
	if !rootID.IsZero() {
		fields["root_id"] = rootID
	}
	if !buildingID.IsZero() {
		fields["building_id"] = buildingID
	}
	if value := strings.TrimSpace(city); value != "" {
		fields["city"] = value
	}
	if value := strings.TrimSpace(district); value != "" {
		fields["district"] = value
	}
	if roomStatus != nil {
		fields["room_status"] = *roomStatus
	}
	if listingStatus != nil {
		fields["listing_status"] = *listingStatus
	}
	if auditStatus != nil {
		fields["audit_status"] = *auditStatus
	}
	return activeFilter(fields), nil
}
