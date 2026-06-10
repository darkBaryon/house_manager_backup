package hpd

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
)

var (
	miniappProjectionFields = allowedFields(
		"rent_mode",
		"city",
		"district",
		"biz_area",
		"community_name",
		"building_or_community_name",
		"subway_station",
		"subway_distance_m",
		"address_text",
		"geo",
		"title",
		"subtitle",
		"price",
		"price_text",
		"layout_text",
		"room_count",
		"hall_count",
		"bathroom_count",
		"kitchen_count",
		"area_size",
		"orientation",
		"floor_text",
		"payment_cycle",
		"feature_flags",
		"listing_facilities",
		"platform_tags",
		"start_rent_rule",
		"cost_items",
		"description",
		"risk_notice",
		"contact_phone",
		"images",
		"weight_score",
		"is_online",
	)

	publisherProjectionFields = allowedFields(
		"owner_landlord_id",
		"owner_phone_snapshot",
		"landlord_name_snapshot",
		"root_type",
		"root_id",
		"project_id",
		"project_name",
		"building_id",
		"building_name",
		"room_type_id",
		"room_type_name",
		"decentralized_id",
		"community_name",
		"rent_mode",
		"city",
		"district",
		"biz_area",
		"subway_station",
		"address_text",
		"geo",
		"room_no",
		"floor_no",
		"title",
		"subtitle",
		"price",
		"price_text",
		"layout_text",
		"room_count",
		"hall_count",
		"bathroom_count",
		"kitchen_count",
		"area_size",
		"orientation",
		"decoration_level",
		"payment_cycle",
		"deposit",
		"service_fee",
		"agency_fee_mode",
		"agency_fee_value",
		"room_status",
		"listing_status",
		"viewing_time_rule",
		"start_rent_rule",
		"feature_flags",
		"listing_facilities",
		"room_facilities",
		"images",
		"is_online",
	)

	adminProjectionFields = allowedFields(
		"owner_landlord_id",
		"owner_phone_snapshot",
		"landlord_name_snapshot",
		"root_type",
		"root_id",
		"project_id",
		"project_name",
		"building_id",
		"building_name",
		"room_type_id",
		"room_type_name",
		"decentralized_id",
		"community_name",
		"rent_mode",
		"city",
		"district",
		"biz_area",
		"address_text",
		"room_no",
		"title",
		"price",
		"price_text",
		"layout_text",
		"room_count",
		"hall_count",
		"bathroom_count",
		"kitchen_count",
		"area_size",
		"room_status",
		"listing_status",
		"audit_status",
		"is_online",
		"latest_audit_task_id",
		"latest_submitted_at",
		"latest_reviewed_at",
		"reviewer_staff_id",
	)
)

func allowedFields(fields ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
	}
	return allowed
}

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

func pickAllowedFields(fields bson.M, allowed map[string]struct{}) (bson.M, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("fields are required")
	}

	picked := make(bson.M, len(fields))
	for k, v := range fields {
		if _, ok := allowed[k]; !ok {
			return nil, fmt.Errorf("field %q is not allowed to update", k)
		}
		picked[k] = v
	}
	return picked, nil
}

func cloneBsonM(src bson.M) bson.M {
	if src == nil {
		return nil
	}
	cloned := make(bson.M, len(src))
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
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
