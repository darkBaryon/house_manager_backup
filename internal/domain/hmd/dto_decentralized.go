package hmd

import "go.mongodb.org/mongo-driver/v2/bson"

type ListDecentralizedCommunitiesInput struct {
	City     string
	District string
}

type CreateDecentralizedCommunityInput struct {
	CommunityName string
	City          string
	District      string
	BizArea       string
	AddressText   string
	Geo           *GeoPointInput
	SubwayStation string
}

type UpdateDecentralizedCommunityInput struct {
	ID            bson.ObjectID
	CommunityName string
	City          string
	District      string
	BizArea       string
	AddressText   string
	Geo           *GeoPointInput
	SubwayStation string
}
