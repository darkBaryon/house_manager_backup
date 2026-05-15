package hpd

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
	"strings"
)

var (
	listingLifecycleFields = allowedFields(
		"listing_status",
		"published_at",
		"offline_at",
	)

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

func listingStatusUpdateFields(listingStatus hpdmodel.HpdListingStatus) bson.M {
	return bson.M{"listing_status": listingStatus}
}

func entrustRelationFields(entity *hpdmodel.HpdEntrustRelation) bson.M {
	return bson.M{
		"listing_id":          entity.ListingID,
		"owner_name":          entity.OwnerName,
		"owner_phone":         entity.OwnerPhone,
		"maintainer_staff_id": entity.MaintainerStaffID,
		"service_staff_id":    entity.ServiceStaffID,
		"relation_status":     entity.RelationStatus,
		"effective_from":      entity.EffectiveFrom,
		"effective_to":        entity.EffectiveTo,
	}
}

func activeEntrustRelationFilter(fields bson.M) bson.M {
	filter := activeFilter(fields)
	filter["relation_status"] = hpdmodel.HpdRelationStatusActive
	return filter
}

func activeEntrustAccessFilter(listingID bson.ObjectID, staffID bson.ObjectID, ownerPhone string) (bson.M, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("listingID is required")
	}
	ownerPhone = strings.TrimSpace(ownerPhone)
	clauses := make(bson.A, 0, 3)
	if !staffID.IsZero() {
		clauses = append(clauses,
			bson.M{"maintainer_staff_id": staffID},
			bson.M{"service_staff_id": staffID},
		)
	}
	if ownerPhone != "" {
		clauses = append(clauses, bson.M{"owner_phone": ownerPhone})
	}
	if len(clauses) == 0 {
		return nil, fmt.Errorf("staffID or ownerPhone is required")
	}
	return activeEntrustRelationFilter(bson.M{
		"listing_id": listingID,
		"$or":        clauses,
	}), nil
}
