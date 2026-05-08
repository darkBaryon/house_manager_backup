package common

import (
	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

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

func buildUpdateFieldsByIDDoc(fields bson.M, now int64) bson.M {
	setFields := cloneBsonM(fields)
	setFields["updated_at"] = now
	return bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
	}
}

func buildUpsertFieldsDoc(fields bson.M, now int64) bson.M {
	setFields := cloneBsonM(fields)
	setFields["updated_at"] = now
	return bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at": now,
			"status":     model.StatusActive,
		},
	}
}
