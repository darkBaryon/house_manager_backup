package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	StatusUnspecified = 0
	StatusActive      = 1
	StatusDeleted     = -1
)

// CommonFields 所有集合通用字段。
type CommonFields struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt int64         `bson:"created_at" json:"createdAt"`
	UpdatedAt int64         `bson:"updated_at" json:"updatedAt"`
	Status    int           `bson:"status" json:"status"`
	Version   int64         `bson:"version" json:"version"`
}

// Common 返回公共字段指针，供通用仓库维护公共字段。
func (f *CommonFields) Common() *CommonFields {
	return f
}
