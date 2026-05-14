package common

import (
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func CurrentUserID(c *gin.Context) (bson.ObjectID, bool) {
	userID, ok := userIDFromGinOrRequestContext(c)
	if !ok {
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
	userID, ok := userIDFromGinOrRequestContext(c)
	if !ok {
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

func userIDFromGinOrRequestContext(c *gin.Context) (string, bool) {
	value, ok := c.Get("userId")
	if ok {
		userID, ok := value.(string)
		if ok && userID != "" {
			return userID, true
		}
	}
	if value, ok := c.Get("principal"); ok {
		if principal, ok := value.(session.Principal); ok && isMiniappUserPrincipal(principal) {
			return principal.PrincipalID, true
		}
	}
	principal, ok := session.PrincipalFromContext(c.Request.Context())
	if ok && isMiniappUserPrincipal(principal) {
		return principal.PrincipalID, true
	}
	return "", false
}

func isMiniappUserPrincipal(principal session.Principal) bool {
	return principal.PrincipalType == session.PrincipalTypeUser && principal.Terminal == session.TerminalMiniapp && principal.PrincipalID != ""
}
