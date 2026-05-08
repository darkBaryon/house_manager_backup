package common

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
	return r.FindOne(ctx, notDeletedByIDFilter(id))
}

// FindByIdIncludingDeleted 根据 ID 查询单条记录，允许返回软删除记录。
func (r *Repository[T]) FindByIdIncludingDeleted(ctx context.Context, id bson.ObjectID) (*T, error) {
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
	update := buildUpsertFieldsDoc(fields, now)

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
	update := buildUpdateFieldsByIDDoc(fields, now)

	res, err := r.Collection.UpdateOne(ctx, notDeletedByIDFilter(id), update)
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

	res, err := r.Collection.UpdateOne(ctx, notDeletedByIDFilter(id), update)
	if err != nil {
		return fmt.Errorf("soft delete by id: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func notDeletedByIDFilter(id bson.ObjectID) bson.M {
	return bson.M{"_id": id, "status": bson.M{"$ne": model.StatusDeleted}}
}

func cloneBsonM(src bson.M) bson.M {
	if src == nil {
		return nil
	}

	cloned := make(bson.M, len(src))
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
}

func buildUpdateFieldsByIDDoc(fields bson.M, now int64) bson.M {
	setFields := cloneBsonM(fields)
	setFields["updated_at"] = now
	return bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
	}
}

func buildUpsertFieldsDoc(fields bson.M, now int64) bson.M {
	setFields := cloneBsonM(fields)
	setFields["updated_at"] = now
	return bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at": now,
			"status":     model.StatusActive,
		},
	}
}
