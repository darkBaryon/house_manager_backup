package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionFavorite = "hs_usr_favorite"
	CollectionHistory  = "hs_usr_history"
)

// Favorite 对应 hs_usr_favorite，小程序用户收藏。
type Favorite struct {
	CommonFields `bson:",inline"`

	UserID    bson.ObjectID `bson:"user_id" json:"userId"`
	ListingID bson.ObjectID `bson:"listing_id" json:"listingId"`
}

// History 对应 hs_usr_history，小程序用户浏览足迹。
type History struct {
	CommonFields `bson:",inline"`

	UserID    bson.ObjectID `bson:"user_id" json:"userId"`
	ListingID bson.ObjectID `bson:"listing_id" json:"listingId"`
	Source    string        `bson:"source" json:"source"`
	ViewedAt  int64         `bson:"viewed_at" json:"viewedAt"`
}
