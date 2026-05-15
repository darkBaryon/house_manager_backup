package common

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
)

type ListResponse[T any] struct {
	List []T `json:"list"`
}

type EntityMetaResponse struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Status    int    `json:"status"`
	Version   int64  `json:"version"`
}

type GeoPointResponse struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type TaggedImageResponse struct {
	URL string `json:"url"`
	Tag string `json:"tag,omitempty"`
}

func EntityMeta(fields commonmodel.CommonFields) EntityMetaResponse {
	return EntityMetaResponse{
		ID:        ObjectIDHex(fields.ID),
		CreatedAt: fields.CreatedAt,
		UpdatedAt: fields.UpdatedAt,
		Status:    fields.Status,
		Version:   fields.Version,
	}
}

func GeoPoint(point *hmdmodel.GeoPoint) *GeoPointResponse {
	if point == nil {
		return nil
	}
	return &GeoPointResponse{Lng: point.Lng, Lat: point.Lat}
}

func TaggedImages(images []hmdmodel.TaggedImage) []TaggedImageResponse {
	if len(images) == 0 {
		return []TaggedImageResponse{}
	}
	out := make([]TaggedImageResponse, 0, len(images))
	for _, image := range images {
		out = append(out, TaggedImageResponse{URL: image.URL, Tag: string(image.Tag)})
	}
	return out
}

func ListingFacilities(items []hmdmodel.ListingFacility) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, string(item))
	}
	return out
}

func RoomFacilities(items []hmdmodel.RoomFacility) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, string(item))
	}
	return out
}

func StringsOrEmpty(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}

func ObjectIDHex(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}
