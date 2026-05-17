package common

import (
	commonmodel "house-manager/internal/model/common"

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
	setOnInsert := bson.M{
		"created_at": now,
		"status":     commonmodel.StatusActive,
	}
	for key := range setFields {
		delete(setOnInsert, key)
	}
	return bson.M{
		"$set":         setFields,
		"$inc":         bson.M{"version": 1},
		"$setOnInsert": setOnInsert,
	}
}
