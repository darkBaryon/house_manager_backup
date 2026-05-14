package common

import (
	"fmt"
	"log/slog"

	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IDRequest struct {
	ID string `json:"id" binding:"required"`
}

type ProjectIDRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
}

type BuildingIDRequest struct {
	BuildingID string `json:"building_id" binding:"required"`
}

type DecentralizedIDRequest struct {
	DecentralizedID string `json:"decentralized_id" binding:"required"`
}

type RoomStatusRequest struct {
	ID         string `json:"id" binding:"required"`
	RoomStatus *int   `json:"room_status" binding:"required"`
}

type ListByCityRequest struct {
	City     string `json:"city" binding:"required"`
	District string `json:"district"`
}

type GeoPointRequest struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func (r *GeoPointRequest) ToServiceInput() *publishsvc.GeoPointInput {
	if r == nil {
		return nil
	}
	return &publishsvc.GeoPointInput{Lng: r.Lng, Lat: r.Lat}
}

type TaggedImageRequest struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func TaggedImageInputs(images []TaggedImageRequest) []publishsvc.TaggedImageInput {
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

func BindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return false
	}
	return true
}

func BindID(c *gin.Context) (bson.ObjectID, bool) {
	var req IDRequest
	if !BindJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return ObjectIDFromHex(c, req.ID)
}

func BindProjectID(c *gin.Context) (bson.ObjectID, bool) {
	var req ProjectIDRequest
	if !BindJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return ObjectIDFromHex(c, req.ProjectID)
}

func BindBuildingID(c *gin.Context) (bson.ObjectID, bool) {
	var req BuildingIDRequest
	if !BindJSON(c, &req) {
		return bson.NilObjectID, false
	}
	return ObjectIDFromHex(c, req.BuildingID)
}

func BindRoomStatus(c *gin.Context) (bson.ObjectID, int, bool) {
	var req RoomStatusRequest
	if !BindJSON(c, &req) {
		return bson.NilObjectID, 0, false
	}
	id, ok := ObjectIDFromHex(c, req.ID)
	if !ok {
		return bson.NilObjectID, 0, false
	}
	if req.RoomStatus == nil || !model.IsValidRoomStatusUpdateTarget(*req.RoomStatus) {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("room_status must be one of -1, 1, 2, 3")))
		return bson.NilObjectID, 0, false
	}
	return id, *req.RoomStatus, true
}

func ObjectIDFromHex(c *gin.Context, value string) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("invalid object id %q", value)))
		return bson.NilObjectID, false
	}
	return id, true
}

func OptionalObjectIDFromHex(c *gin.Context, value string) (bson.ObjectID, bool) {
	if value == "" {
		return bson.NilObjectID, true
	}
	return ObjectIDFromHex(c, value)
}

func WriteResult(c *gin.Context, logMessage string, data any, err error) {
	if err != nil {
		slog.Error(logMessage, "error", err)
		response.Err(c, err)
		return
	}
	response.Success(c, data)
}
