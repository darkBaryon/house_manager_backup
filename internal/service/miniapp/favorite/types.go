package favorite

import (
	"house-manager/internal/service/miniapp/listingview"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AddInput struct {
	UserID    bson.ObjectID
	ListingID bson.ObjectID
}

type RemoveInput struct {
	UserID    bson.ObjectID
	ListingID bson.ObjectID
}

type ExistsInput struct {
	UserID    bson.ObjectID
	ListingID bson.ObjectID
}

type ListInput struct {
	UserID   bson.ObjectID
	Page     int
	PageSize int
}

type MutationResult struct {
	ListingID   string
	IsFavorited bool
}

type ListResult struct {
	List     []listingview.Item
	Page     int
	PageSize int
	Total    int64
}
