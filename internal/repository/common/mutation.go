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
	if len(filter) == 0 {
		return false, fmt.Errorf("upsert fields: filter is required")
	}
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

// UpdateFieldsByID 根据 ID 局部更新，并刷新 updated_at / version。
func (r *Repository[T]) UpdateFieldsByID(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update fields by id: id is required")
	}
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

// SoftDeleteByID 按公共字段约定执行软删除。
func (r *Repository[T]) SoftDeleteByID(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("soft delete by id: id is required")
	}

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
