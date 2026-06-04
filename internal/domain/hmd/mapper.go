package hmd

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func roomTypeID(roomType *hmdmodel.HmdRoomTypeCentralized) bson.ObjectID {
	if roomType == nil {
		return bson.NilObjectID
	}
	return roomType.ID
}

func bsonFields(kv ...any) bson.M {
	fields := make(bson.M, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok || key == "" {
			continue
		}
		fields[key] = kv[i+1]
	}
	return fields
}

func layoutCountValue(input *int) int {
	if input == nil {
		return hmdmodel.UnknownLayoutCount
	}
	return *input
}

func layoutCountValueOrFallback(input *int, fallback int) int {
	if input != nil {
		return *input
	}
	return fallback
}

func toGeoPoint(input *GeoPointInput) *hmdmodel.GeoPoint {
	if input == nil {
		return nil
	}
	return &hmdmodel.GeoPoint{
		Lng: input.Lng,
		Lat: input.Lat,
	}
}

func toTaggedImages(inputs []TaggedImageInput) []hmdmodel.TaggedImage {
	if len(inputs) == 0 {
		return nil
	}
	images := make([]hmdmodel.TaggedImage, 0, len(inputs))
	for _, item := range inputs {
		images = append(images, hmdmodel.TaggedImage{
			URL: strings.TrimSpace(item.URL),
			Tag: hmdmodel.ImageTag(strings.TrimSpace(item.Tag)),
		})
	}
	return images
}

func toRoomFacilities(items []string) []hmdmodel.RoomFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]hmdmodel.RoomFacility, 0, len(items))
	for _, item := range items {
		out = append(out, hmdmodel.RoomFacility(strings.TrimSpace(item)))
	}
	return out
}

func toListingFacilities(items []string) []hmdmodel.ListingFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]hmdmodel.ListingFacility, 0, len(items))
	for _, item := range items {
		out = append(out, hmdmodel.ListingFacility(strings.TrimSpace(item)))
	}
	return out
}

func cloneStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, strings.TrimSpace(item))
	}
	return out
}
