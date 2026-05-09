package publish

import (
	"fmt"

	publishsvc "house-manager/internal/service/publish"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type idRequest struct {
	ID string `json:"id" binding:"required"`
}

type projectIDRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
}

type buildingIDRequest struct {
	BuildingID string `json:"building_id" binding:"required"`
}

type decentralizedIDRequest struct {
	DecentralizedID string `json:"decentralized_id" binding:"required"`
}

type roomStatusRequest struct {
	ID         string `json:"id" binding:"required"`
	RoomStatus int    `json:"room_status"`
}

type listByCityRequest struct {
	City     string `json:"city" binding:"required"`
	District string `json:"district"`
}

type geoPointRequest struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func (r *geoPointRequest) toServiceInput() *publishsvc.GeoPointInput {
	if r == nil {
		return nil
	}
	return &publishsvc.GeoPointInput{Lng: r.Lng, Lat: r.Lat}
}

type taggedImageRequest struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func taggedImageInputs(images []taggedImageRequest) []publishsvc.TaggedImageInput {
	if len(images) == 0 {
		return nil
	}
	inputs := make([]publishsvc.TaggedImageInput, 0, len(images))
	for _, image := range images {
		inputs = append(inputs, publishsvc.TaggedImageInput{
			URL: image.URL,
			Tag: image.Tag,
		})
	}
	return inputs
}

func bindPublishJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return false
	}
	return true
}

func bindIDRequest(c *gin.Context) (bson.ObjectID, bool) {
	var req idRequest
	if !bindPublishJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return objectIDFromHex(c, req.ID)
}

func bindProjectIDRequest(c *gin.Context) (bson.ObjectID, bool) {
	var req projectIDRequest
	if !bindPublishJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return objectIDFromHex(c, req.ProjectID)
}

func bindBuildingIDRequest(c *gin.Context) (bson.ObjectID, bool) {
	var req buildingIDRequest
	if !bindPublishJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return objectIDFromHex(c, req.BuildingID)
}

func bindRoomStatusRequest(c *gin.Context) (bson.ObjectID, int, bool) {
	var req roomStatusRequest
	if !bindPublishJSON(c, &req) {
		return bson.NilObjectID, 0, false
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return bson.NilObjectID, 0, false
	}
	return id, req.RoomStatus, true
}

func parseCentralizedRoomIDs(c *gin.Context, projectIDText, buildingIDText, roomTypeIDText string) (bson.ObjectID, bson.ObjectID, bson.ObjectID, bool) {
	projectID, ok := objectIDFromHex(c, projectIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	buildingID, ok := objectIDFromHex(c, buildingIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	roomTypeID, ok := optionalObjectIDFromHex(c, roomTypeIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	return projectID, buildingID, roomTypeID, true
}

func objectIDFromHex(c *gin.Context, value string) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("invalid object id %q", value)))
		return bson.NilObjectID, false
	}
	return id, true
}

func optionalObjectIDFromHex(c *gin.Context, value string) (bson.ObjectID, bool) {
	if value == "" {
		return bson.NilObjectID, true
	}
	return objectIDFromHex(c, value)
}
