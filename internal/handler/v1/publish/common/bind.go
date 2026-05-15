package common

import (
	"fmt"
	"reflect"
	"strings"

	hmdmodel "house-manager/internal/model/hmd"
	publishsvc "house-manager/internal/service/publish"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
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
	requestlog.AddFields(c, summarizeRequest(req))
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
	if req.RoomStatus == nil || !hmdmodel.IsValidRoomStatusUpdateTarget(*req.RoomStatus) {
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
		requestlog.AddField(c, "action", logMessage)
		response.Err(c, err)
		return
	}
	response.Success(c, data)
}

func summarizeRequest(req any) map[string]any {
	value := reflect.ValueOf(req)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}

	typ := value.Type()
	fields := make(map[string]any)
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		key := jsonFieldName(field)
		if key == "" {
			continue
		}
		appendFieldSummary(fields, key, value.Field(i))
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func jsonFieldName(field reflect.StructField) string {
	tag := strings.TrimSpace(field.Tag.Get("json"))
	if tag == "" || tag == "-" {
		return ""
	}
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return ""
	}
	return name
}

func appendFieldSummary(fields map[string]any, key string, value reflect.Value) {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		text := strings.TrimSpace(value.String())
		if text == "" {
			return
		}
		if strings.Contains(key, "phone") {
			fields[key] = middlewareMaskPhone(text)
			return
		}
		fields[key] = truncateText(text, 64)
	case reflect.Bool:
		fields[key] = value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fields[key] = value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fields[key] = value.Uint()
	case reflect.Float32, reflect.Float64:
		fields[key] = value.Float()
	case reflect.Slice, reflect.Array:
		if value.Len() > 0 {
			fields[key+"_count"] = value.Len()
		}
	case reflect.Struct:
		fields[key] = "provided"
	}
}

func truncateText(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}

func middlewareMaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
