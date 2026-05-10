package house

import "go.mongodb.org/mongo-driver/v2/bson"

type DetailInput struct {
	ListingID bson.ObjectID
}

type DetailResult struct {
	House       Detail
	IsFavorited bool
}

type Detail struct {
	ListItem
	AddressText   string
	Geo           *GeoPoint
	StartRentRule string
	CostItems     []CostItem
	Description   string
	RiskNotice    string
	ContactPhone  string
}

type GeoPoint struct {
	Lng float64
	Lat float64
}

type CostItem struct {
	Name   string
	Amount int
	Unit   string
	Remark string
}
