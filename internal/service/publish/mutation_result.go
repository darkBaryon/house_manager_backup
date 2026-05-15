package publish

import (
	hmddomain "house-manager/internal/domain/hmd"
	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func mutationEntityID[T any](result *hmddomain.HmdMutationResult[T]) bson.ObjectID {
	if result == nil || result.Entity == nil {
		return bson.ObjectID{}
	}
	commonAware, ok := any(result.Entity).(interface {
		Common() *commonmodel.CommonFields
	})
	if !ok {
		return bson.ObjectID{}
	}
	return commonAware.Common().ID
}
