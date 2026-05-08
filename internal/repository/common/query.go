package common

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// FindByID 根据 ID 查询单条记录。
func (r *Repository[T]) FindByID(ctx context.Context, id bson.ObjectID) (*T, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find by id: id is required")
	}
	return r.FindOne(ctx, notDeletedByIDFilter(id))
}

// FindByIDIncludingDeleted 根据 ID 查询单条记录，允许返回软删除记录。
func (r *Repository[T]) FindByIDIncludingDeleted(ctx context.Context, id bson.ObjectID) (*T, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find by id: id is required")
	}

	var entity T
	if err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return &entity, nil
}

// FindOne 根据条件查询单条记录。
func (r *Repository[T]) FindOne(ctx context.Context, filter bson.M) (*T, error) {
	var entity T
	if err := r.Collection.FindOne(ctx, filter).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find one: %w", err)
	}
	return &entity, nil
}

// FindMany 根据条件查询多条记录。
func (r *Repository[T]) FindMany(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]T, error) {
	cursor, err := r.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("find many: %w", err)
	}
	defer cursor.Close(ctx)

	var entities []T
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, fmt.Errorf("decode many: %w", err)
	}
	return entities, nil
}
