package hmd

import "go.mongodb.org/mongo-driver/v2/bson"

type ListCentralizedProjectsInput struct {
	City     string
	District string
}

type CreateCentralizedProjectInput struct {
	ProjectName string
	City        string
	District    string
	AddressText string
	Geo         *GeoPointInput
	BrandName   string
}

type UpdateCentralizedProjectInput struct {
	ID          bson.ObjectID
	ProjectName string
	City        string
	District    string
	AddressText string
	Geo         *GeoPointInput
	BrandName   string
}

type CreateBuildingInput struct {
	ProjectID         bson.ObjectID
	BuildingName      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []string
}

type UpdateBuildingInput struct {
	ID                bson.ObjectID
	BuildingName      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []string
}
