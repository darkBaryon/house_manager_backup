package common

import (
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func CurrentUserID(c *gin.Context) (bson.ObjectID, bool) {
	value, ok := c.Get("userId")
	if !ok {
		response.Err(c, errcode.Unauthorized)
		return bson.NilObjectID, false
	}
	userID, ok := value.(string)
	if !ok || userID == "" {
		response.Err(c, errcode.Unauthorized)
		return bson.NilObjectID, false
	}
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		response.Err(c, errcode.Unauthorized.WithError(err))
		return bson.NilObjectID, false
	}
	return id, true
}

func OptionalUserID(c *gin.Context) (bson.ObjectID, bool) {
	value, ok := c.Get("userId")
	if !ok {
		return bson.NilObjectID, false
	}
	userID, ok := value.(string)
	if !ok || userID == "" {
		return bson.NilObjectID, false
	}
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return bson.NilObjectID, false
	}
	return id, true
}

func ParseObjectID(c *gin.Context, value string) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return bson.NilObjectID, false
	}
	return id, true
}
