package common

import (
	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func notDeletedByIDFilter(id bson.ObjectID) bson.M {
	return bson.M{"_id": id, "status": bson.M{"$ne": model.StatusDeleted}}
}
