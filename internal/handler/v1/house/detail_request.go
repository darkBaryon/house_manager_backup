package house

import (
	"fmt"

	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type detailRequest struct {
	ListingID string `json:"listing_id" binding:"required"`
}

func bindDetailRequest(c *gin.Context) (bson.ObjectID, bool) {
	var req detailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return bson.NilObjectID, false
	}
	listingID, err := bson.ObjectIDFromHex(req.ListingID)
	if err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("invalid listing_id %q", req.ListingID)))
		return bson.NilObjectID, false
	}
	return listingID, true
}
