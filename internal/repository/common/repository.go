package common

import "go.mongodb.org/mongo-driver/v2/mongo"

// Repository 泛型数据仓库，提供基于公共字段约定的最小通用操作。
type Repository[T any] struct {
	collection *mongo.Collection
}

// NewRepository 创建泛型仓库实例。
func NewRepository[T any](coll *mongo.Collection) *Repository[T] {
	return &Repository[T]{collection: coll}
}
