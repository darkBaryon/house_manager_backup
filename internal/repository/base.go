package repository

import (
	"context"
	"fmt"
	"strings"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Repository 泛型数据仓库，提供通用 CRUD 操作
type Repository[T any] struct {
	Collection *mongo.Collection
}

// NewRepository 创建泛型仓库实例
func NewRepository[T any](coll *mongo.Collection) *Repository[T] {
	return &Repository[T]{Collection: coll}
}

// =================== 查询 ===================

// FindById 根据 ID 查询单条记录
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

// FindOne 根据条件查询单条记录
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

// FindList 查询列表（分页 + 条件过滤 + 排序）
func (r *Repository[T]) FindList(ctx context.Context, filter bson.M, req model.PageReq) ([]T, int64, error) {
	opts := options.Find()
	if len(req.Sort) > 0 {
		sortDoc := bson.D{}
		for _, s := range req.Sort {
			sortDoc = append(sortDoc, parseSort(s))
		}
		opts.SetSort(sortDoc)
	}
	if req.Limit > 0 {
		opts.SetLimit(int64(req.Limit))
	}
	if req.Offset > 0 {
		opts.SetSkip(int64(req.Offset))
	}

	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var entities []T
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, 0, fmt.Errorf("decode documents: %w", err)
	}

	return entities, total, nil
}

// Count 根据条件统计文档数
func (r *Repository[T]) Count(ctx context.Context, filter bson.M) (int64, error) {
	n, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return n, nil
}

// Exists 判断是否存在匹配文档
func (r *Repository[T]) Exists(ctx context.Context, filter bson.M) (bool, error) {
	n, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("exists check: %w", err)
	}
	return n > 0, nil
}

// =================== 写入 ===================

// Insert 插入单条文档
func (r *Repository[T]) Insert(ctx context.Context, entity *T) error {
	if _, err := r.Collection.InsertOne(ctx, entity); err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	return nil
}

// InsertMany 批量插入文档
func (r *Repository[T]) InsertMany(ctx context.Context, entities []T) error {
	docs := make([]any, len(entities))
	for i, e := range entities {
		docs[i] = e
	}
	if _, err := r.Collection.InsertMany(ctx, docs); err != nil {
		return fmt.Errorf("insert many: %w", err)
	}
	return nil
}

// =================== 更新 ===================

// Upsert 按条件更新，不存在则插入。matched=true 表示命中已有文档
func (r *Repository[T]) Upsert(ctx context.Context, filter bson.M, update any) (matched bool, err error) {
	opts := options.UpdateOne().SetUpsert(true)
	res, err := r.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, fmt.Errorf("upsert: %w", err)
	}
	return res.MatchedCount > 0, nil
}

// UpdateById 根据 ID 局部更新（传入 bson.M/D 等更新表达式）
func (r *Repository[T]) UpdateById(ctx context.Context, id bson.ObjectID, update any) error {
	res, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("update by id: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// ReplaceById 根据 ID 整体替换文档
func (r *Repository[T]) ReplaceById(ctx context.Context, id bson.ObjectID, entity *T) error {
	res, err := r.Collection.ReplaceOne(ctx, bson.M{"_id": id}, entity)
	if err != nil {
		return fmt.Errorf("replace by id: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// =================== 删除 ===================

// DeleteById 根据 ID 删除单条文档
func (r *Repository[T]) DeleteById(ctx context.Context, id bson.ObjectID) error {
	res, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete by id: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// DeleteMany 根据条件批量删除
func (r *Repository[T]) DeleteMany(ctx context.Context, filter bson.M) (int64, error) {
	res, err := r.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete many: %w", err)
	}
	return res.DeletedCount, nil
}

// parseSort 解析排序字符串 "field:desc" → bson.E
func parseSort(s string) bson.E {
	parts := strings.SplitN(s, ":", 2)
	field := parts[0]
	dir := 1
	if len(parts) > 1 && parts[1] == "desc" {
		dir = -1
	}
	return bson.E{Key: field, Value: dir}
}
