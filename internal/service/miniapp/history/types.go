package history

import (
	"house-manager/internal/service/miniapp/listingview"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AddInput struct {
	UserID    bson.ObjectID
	ListingID bson.ObjectID
	Source    string
}

type ListInput struct {
	UserID   bson.ObjectID
	Page     int
	PageSize int
}

type AddResult struct {
	ListingID string
	ViewedAt  int64
}

type ListItem struct {
	listingview.Item
	ViewedAt int64
}

type ListResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}
