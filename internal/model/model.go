package model

import "go.mongodb.org/mongo-driver/v2/bson"

type BaseModel struct {
	Id        bson.ObjectID `bson:"id" json:"id"`
	CreatedBy string        `bson:"createdby" json:"createdBy"`
	CreatedAt int64         `bson:"createdat" json:"createdAt"`
	UpdatedBy string        `bson:"updatedby" json:"updatedBy"`
	UpdatedAt int64         `bson:"updatedat" json:"updatedAt"`
	Version   int64         `bson:"version" json:"version"`
}

type LiteBaseModel struct {
	Id        bson.ObjectID `bson:"id" json:"id"`
	CreatedBy string        `bson:"createdby" json:"createdBy"`
	CreatedAt int64         `bson:"createdat" json:"createdAt"`
}

type PageReq struct {
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
	Sort   []string `json:"sort"`
}
