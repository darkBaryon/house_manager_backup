package hmd

import (
	"strings"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func roomTypeID(roomType *model.HmdRoomTypeCentralized) bson.ObjectID {
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

func toGeoPoint(input *GeoPointInput) *model.GeoPoint {
	if input == nil {
		return nil
	}
	return &model.GeoPoint{
		Lng: input.Lng,
		Lat: input.Lat,
	}
}

func toTaggedImages(inputs []TaggedImageInput) []model.TaggedImage {
	if len(inputs) == 0 {
		return nil
	}
	images := make([]model.TaggedImage, 0, len(inputs))
	for _, item := range inputs {
		images = append(images, model.TaggedImage{
			URL: strings.TrimSpace(item.URL),
			Tag: model.ImageTag(strings.TrimSpace(item.Tag)),
		})
	}
	return images
}

func toRoomFacilities(items []string) []model.RoomFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.RoomFacility, 0, len(items))
	for _, item := range items {
		out = append(out, model.RoomFacility(strings.TrimSpace(item)))
	}
	return out
}

func toListingFacilities(items []string) []model.ListingFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.ListingFacility, 0, len(items))
	for _, item := range items {
		out = append(out, model.ListingFacility(strings.TrimSpace(item)))
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
