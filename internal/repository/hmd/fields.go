package hmd

import (
	"fmt"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func activeFilter(fields bson.M) bson.M {
	filter := bson.M{"status": model.StatusActive}
	for k, v := range fields {
		filter[k] = v
	}
	return filter
}

func notDeletedFilter(fields bson.M) bson.M {
	filter := bson.M{"status": bson.M{"$ne": model.StatusDeleted}}
	for k, v := range fields {
		filter[k] = v
	}
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

func roomStatusUpdateFields(roomStatus int) bson.M {
	return bson.M{"room_status": roomStatus}
}
