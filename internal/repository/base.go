package repository

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Repository 泛型数据仓库，提供基于公共字段约定的最小通用操作。
type Repository[T any] struct {
	Collection *mongo.Collection
}

// NewRepository 创建泛型仓库实例。
func NewRepository[T any](coll *mongo.Collection) *Repository[T] {
	return &Repository[T]{Collection: coll}
}

// FindById 根据 ID 查询单条记录。
func (r *Repository[T]) FindById(ctx context.Context, id bson.ObjectID) (*T, error) {
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

// Insert 插入单条文档，并初始化公共字段。
func (r *Repository[T]) Insert(ctx context.Context, entity *T) error {
	if entity == nil {
		return fmt.Errorf("insert: entity is nil")
	}

	now := time.Now().Unix()
	commonAware, ok := any(entity).(interface{ Common() *model.CommonFields })
	if !ok {
		return fmt.Errorf("insert: entity does not expose common fields")
	}
	common := commonAware.Common()
	if common.CreatedAt == 0 {
		common.CreatedAt = now
	}
	common.UpdatedAt = now
	if common.Status == model.StatusUnspecified {
		common.Status = model.StatusActive
	}
	if common.Version == 0 {
		common.Version = 1
	}

	res, err := r.Collection.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	if common.ID.IsZero() {
		if insertedID, ok := res.InsertedID.(bson.ObjectID); ok {
			common.ID = insertedID
		}
	}
	return nil
}

// UpsertFields 按条件局部更新，不存在则插入，并维护公共字段。
// matched=true 表示命中已有文档；matched=false 表示触发了 upsert 插入路径。
func (r *Repository[T]) UpsertFields(ctx context.Context, filter bson.M, fields bson.M) (matched bool, err error) {
	if fields == nil {
		return false, fmt.Errorf("upsert fields: fields is nil")
	}

	now := time.Now().Unix()
	update := bson.M{
		"$set": fields,
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at": now,
			"status":     model.StatusActive,
			"version":    1,
		},
	}
	update["$set"].(bson.M)["updated_at"] = now

	opts := options.UpdateOne().SetUpsert(true)
	res, err := r.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, fmt.Errorf("upsert fields: %w", err)
	}
	return res.MatchedCount > 0, nil
}

// UpdateFieldsById 根据 ID 局部更新，并刷新 updated_at / version。
func (r *Repository[T]) UpdateFieldsById(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if fields == nil {
		return fmt.Errorf("update fields by id: fields is nil")
	}

	now := time.Now().Unix()
	fields["updated_at"] = now
	update := bson.M{
		"$set": fields,
		"$inc": bson.M{"version": 1},
	}

	res, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id, "status": bson.M{"$ne": model.StatusDeleted}}, update)
	if err != nil {
		return fmt.Errorf("update fields by id: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// SoftDeleteById 按公共字段约定执行软删除。
func (r *Repository[T]) SoftDeleteById(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"status":     model.StatusDeleted,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
	}

	res, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id, "status": bson.M{"$ne": model.StatusDeleted}}, update)
	if err != nil {
		return fmt.Errorf("soft delete by id: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
