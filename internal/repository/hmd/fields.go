package hmd

import (
	"fmt"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	centralizedBaseInfoFields = allowedFields(
		"project_name",
		"city",
		"district",
		"address_text",
		"geo",
		"brand_name",
	)

	buildingBaseInfoFields = allowedFields(
		"building_name",
		"floor_total",
		"manager_name",
		"manager_phone",
		"photos",
		"listing_facilities",
	)

	decentralizedBaseInfoFields = allowedFields(
		"community_name",
		"city",
		"district",
		"biz_area",
		"address_text",
		"geo",
		"subway_station",
	)

	roomTypeCentralizedBaseInfoFields = allowedFields(
		"room_type_name",
		"room_count",
		"hall_count",
		"bathroom_count",
		"kitchen_count",
		"area_size",
		"orientation",
		"decoration_level",
		"payment_cycle",
		"rent",
		"deposit",
		"service_fee",
		"agency_fee_mode",
		"agency_fee_value",
		"images",
		"room_facilities",
	)

	roomBaseInfoFields = allowedFields(
		"room_no",
		"floor_no",
		"rent_mode",
		"layout_text",
		"area_size",
		"orientation",
		"decoration_level",
		"payment_cycle",
		"rent",
		"deposit",
		"service_fee",
		"agency_fee_mode",
		"agency_fee_value",
		"viewing_time_rule",
		"start_rent_rule",
		"images",
		"room_facilities",
		"listing_facilities",
	)

	roomCentralizedBaseInfoFields   = roomBaseInfoFields
	roomDecentralizedBaseInfoFields = roomBaseInfoFields
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
	filter["status"] = model.StatusActive
	return filter
}

func hmdListFindOptions() *options.FindOptionsBuilder {
	return options.Find().SetSort(bson.D{
		{Key: "updated_at", Value: -1},
		{Key: "_id", Value: -1},
	})
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

func roomStatusUpdateFields(roomStatus int) bson.M {
	return bson.M{"room_status": roomStatus}
}
