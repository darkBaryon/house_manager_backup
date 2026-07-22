package hpd

import (
	"fmt"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	hpdFieldID              common.Field = "_id"
	hpdFieldStatus          common.Field = "status"
	hpdFieldCreatedAt       common.Field = "created_at"
	hpdFieldUpdatedAt       common.Field = "updated_at"
	hpdFieldVersion         common.Field = "version"
	hpdFieldListingID       common.Field = "listing_id"
	hpdFieldSourceType      common.Field = "source_type"
	hpdFieldSourceID        common.Field = "source_id"
	hpdFieldOwnerLandlordID common.Field = "owner_landlord_id"
	hpdFieldRootType        common.Field = "root_type"
	hpdFieldRootID          common.Field = "root_id"
	hpdFieldRelationStatus  common.Field = "relation_status"
	hpdFieldProjectID       common.Field = "project_id"
	hpdFieldBuildingID      common.Field = "building_id"
	hpdFieldDecentralizedID common.Field = "decentralized_id"
	hpdFieldRoomStatus      common.Field = "room_status"
	hpdFieldListingStatus   common.Field = "listing_status"
	hpdFieldAuditStatus     common.Field = "audit_status"
	hpdFieldAssetMode       common.Field = "asset_mode"
	hpdFieldCity            common.Field = "city"
	hpdFieldDistrict        common.Field = "district"
	hpdFieldIsOnline        common.Field = "is_online"
	hpdFieldWeightScore     common.Field = "weight_score"
)

func activeFilter(fields bson.M) bson.M {
	filter := make(bson.M, len(fields)+1)
	for k, v := range fields {
		if k == "status" {
			continue
		}
		filter[k] = v
	}
	filter["status"] = commonmodel.StatusActive
	return filter
}

func updateDocFromSetFields(fields bson.M) common.UpdateDoc {
	update := common.NewUpdateDoc().Inc(hpdFieldVersion, 1)
	for key, value := range fields {
		update = update.Set(common.Field(key), value)
	}
	return update
}

func listingFields(entity *hpdmodel.HpdListing) bson.M {
	return bson.M{
		"source_type": entity.SourceType,
		"source_id":   entity.SourceID,
		"asset_mode":  entity.AssetMode,
	}
}

func miniappListingFields(entity *hpdmodel.HpdMiniappListing) bson.M {
	return bson.M{
		"listing_id":                 entity.ListingID,
		"source_type":                entity.SourceType,
		"source_id":                  entity.SourceID,
		"asset_mode":                 entity.AssetMode,
		"rent_mode":                  entity.RentMode,
		"city":                       entity.City,
		"district":                   entity.District,
		"biz_area":                   entity.BizArea,
		"community_name":             entity.CommunityName,
		"building_or_community_name": entity.BuildingOrCommunityName,
		"subway_station":             entity.SubwayStation,
		"subway_distance_m":          entity.SubwayDistanceM,
		"address_text":               entity.AddressText,
		"geo":                        entity.Geo,
		"title":                      entity.Title,
		"subtitle":                   entity.Subtitle,
		"price":                      entity.Price,
		"price_text":                 entity.PriceText,
		"layout_text":                entity.LayoutText,
		"room_count":                 entity.RoomCount,
		"hall_count":                 entity.HallCount,
		"bathroom_count":             entity.BathroomCount,
		"kitchen_count":              entity.KitchenCount,
		"area_size":                  entity.AreaSize,
		"orientation":                entity.Orientation,
		"floor_text":                 entity.FloorText,
		"payment_cycle":              entity.PaymentCycle,
		"feature_flags":              entity.FeatureFlags,
		"listing_facilities":         entity.ListingFacilities,
		"platform_tags":              entity.PlatformTags,
		"start_rent_rule":            entity.StartRentRule,
		"cost_items":                 entity.CostItems,
		"description":                entity.Description,
		"risk_notice":                entity.RiskNotice,
		"contact_phone":              entity.ContactPhone,
		"images":                     entity.Images,
		"weight_score":               entity.WeightScore,
		"is_online":                  entity.IsOnline,
	}
}

func publisherListingFields(entity *hpdmodel.HpdPublisherListing) bson.M {
	return bson.M{
		"listing_id":             entity.ListingID,
		"source_type":            entity.SourceType,
		"source_id":              entity.SourceID,
		"asset_mode":             entity.AssetMode,
		"owner_landlord_id":      entity.OwnerLandlordID,
		"owner_phone_snapshot":   entity.OwnerPhoneSnapshot,
		"landlord_name_snapshot": entity.LandlordNameSnapshot,
		"root_type":              entity.RootType,
		"root_id":                entity.RootID,
		"project_id":             entity.ProjectID,
		"project_name":           entity.ProjectName,
		"building_id":            entity.BuildingID,
		"building_name":          entity.BuildingName,
		"room_type_id":           entity.RoomTypeID,
		"room_type_name":         entity.RoomTypeName,
		"decentralized_id":       entity.DecentralizedID,
		"community_name":         entity.CommunityName,
		"rent_mode":              entity.RentMode,
		"city":                   entity.City,
		"district":               entity.District,
		"biz_area":               entity.BizArea,
		"subway_station":         entity.SubwayStation,
		"address_text":           entity.AddressText,
		"geo":                    entity.Geo,
		"room_no":                entity.RoomNo,
		"floor_no":               entity.FloorNo,
		"title":                  entity.Title,
		"subtitle":               entity.Subtitle,
		"price":                  entity.Price,
		"price_text":             entity.PriceText,
		"layout_text":            entity.LayoutText,
		"room_count":             entity.RoomCount,
		"hall_count":             entity.HallCount,
		"bathroom_count":         entity.BathroomCount,
		"kitchen_count":          entity.KitchenCount,
		"area_size":              entity.AreaSize,
		"orientation":            entity.Orientation,
		"decoration_level":       entity.DecorationLevel,
		"payment_cycle":          entity.PaymentCycle,
		"deposit":                entity.Deposit,
		"service_fee":            entity.ServiceFee,
		"agency_fee_mode":        entity.AgencyFeeMode,
		"agency_fee_value":       entity.AgencyFeeValue,
		"room_status":            entity.RoomStatus,
		"listing_status":         entity.ListingStatus,
		"viewing_time_rule":      entity.ViewingTimeRule,
		"start_rent_rule":        entity.StartRentRule,
		"feature_flags":          entity.FeatureFlags,
		"listing_facilities":     entity.ListingFacilities,
		"room_facilities":        entity.RoomFacilities,
		"images":                 entity.Images,
		"is_online":              entity.IsOnline,
	}
}

func adminListingFields(entity *hpdmodel.HpdAdminListing) bson.M {
	return bson.M{
		"listing_id":             entity.ListingID,
		"source_type":            entity.SourceType,
		"source_id":              entity.SourceID,
		"asset_mode":             entity.AssetMode,
		"owner_landlord_id":      entity.OwnerLandlordID,
		"owner_phone_snapshot":   entity.OwnerPhoneSnapshot,
		"landlord_name_snapshot": entity.LandlordNameSnapshot,
		"root_type":              entity.RootType,
		"root_id":                entity.RootID,
		"project_id":             entity.ProjectID,
		"project_name":           entity.ProjectName,
		"building_id":            entity.BuildingID,
		"building_name":          entity.BuildingName,
		"room_type_id":           entity.RoomTypeID,
		"room_type_name":         entity.RoomTypeName,
		"decentralized_id":       entity.DecentralizedID,
		"community_name":         entity.CommunityName,
		"rent_mode":              entity.RentMode,
		"city":                   entity.City,
		"district":               entity.District,
		"biz_area":               entity.BizArea,
		"address_text":           entity.AddressText,
		"room_no":                entity.RoomNo,
		"title":                  entity.Title,
		"price":                  entity.Price,
		"price_text":             entity.PriceText,
		"layout_text":            entity.LayoutText,
		"room_count":             entity.RoomCount,
		"hall_count":             entity.HallCount,
		"bathroom_count":         entity.BathroomCount,
		"kitchen_count":          entity.KitchenCount,
		"area_size":              entity.AreaSize,
		"room_status":            entity.RoomStatus,
		"listing_status":         entity.ListingStatus,
		"audit_status":           entity.AuditStatus,
		"is_online":              entity.IsOnline,
		"latest_audit_task_id":   entity.LatestAuditTaskID,
		"latest_submitted_at":    entity.LatestSubmittedAt,
		"latest_reviewed_at":     entity.LatestReviewedAt,
		"reviewer_staff_id":      entity.ReviewerStaffID,
	}
}

func rootScopeRelationFields(entity *hpdmodel.HpdRootScopeRelation) bson.M {
	return bson.M{
		"root_type":         entity.RootType,
		"root_id":           entity.RootID,
		"owner_landlord_id": entity.OwnerLandlordID,
		"owner_phone":       entity.OwnerPhone,
		"relation_status":   entity.RelationStatus,
		"effective_from":    entity.EffectiveFrom,
		"effective_to":      entity.EffectiveTo,
	}
}

func activeRootScopeRelationFilter(fields bson.M) bson.M {
	filter := activeFilter(fields)
	filter["relation_status"] = hpdmodel.HpdRelationStatusActive
	return filter
}

func activeRootScopeAccessFilter(rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, ownerLandlordID bson.ObjectID) (bson.M, error) {
	if !rootType.Valid() {
		return nil, fmt.Errorf("rootType is invalid")
	}
	if rootID.IsZero() {
		return nil, fmt.Errorf("rootID is required")
	}
	if ownerLandlordID.IsZero() {
		return nil, fmt.Errorf("ownerLandlordID is required")
	}
	return activeRootScopeRelationFilter(bson.M{
		"root_type":         rootType,
		"root_id":           rootID,
		"owner_landlord_id": ownerLandlordID,
	}), nil
}
