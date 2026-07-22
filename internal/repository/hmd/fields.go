package hmd

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	commonmodel "house-manager/internal/model/common"
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

func hmdListFindOptions() *options.FindOptionsBuilder {
	return options.Find().SetSort(bson.D{
		{Key: "updated_at", Value: -1},
		{Key: "_id", Value: -1},
	})
}

func roomStatusUpdateFields(roomStatus int) bson.M {
	return bson.M{"room_status": roomStatus}
}
