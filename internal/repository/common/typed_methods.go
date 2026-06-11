package common

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// FindOneBy queries one document with a typed filter.
func (r *Repository[T]) FindOneBy(ctx context.Context, filter Filter) (*T, error) {
	return r.FindOne(ctx, filter.BSON())
}

// FindManyBy queries documents with a typed filter and typed options.
func (r *Repository[T]) FindManyBy(ctx context.Context, filter Filter, opts ...QueryOption) ([]T, error) {
	return r.FindMany(ctx, filter.BSON(), buildFindOptions(opts...)...)
}

// CountBy counts documents matching a typed filter.
func (r *Repository[T]) CountBy(ctx context.Context, filter Filter) (int64, error) {
	total, err := r.Collection.CountDocuments(ctx, filter.BSON())
	if err != nil {
		return 0, fmt.Errorf("count by: %w", err)
	}
	return total, nil
}

// ExistsBy reports whether at least one document matches a typed filter.
func (r *Repository[T]) ExistsBy(ctx context.Context, filter Filter) (bool, error) {
	total, err := r.Collection.CountDocuments(ctx, filter.BSON(), options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("exists by: %w", err)
	}
	return total > 0, nil
}

// UpdateByID updates one non-deleted document by ID.
// Common default predicate = not deleted; stricter status constraints must be explicit at call sites.
func (r *Repository[T]) UpdateByID(ctx context.Context, id bson.ObjectID, update UpdateDoc) error {
	if id.IsZero() {
		return fmt.Errorf("update by id: id is required")
	}
	matched, err := r.UpdateOneBy(ctx, And(
		Eq(Field("_id"), id),
		NotDeleted(),
	), update)
	if err != nil {
		return err
	}
	if !matched {
		return mongo.ErrNoDocuments
	}
	return nil
}

// UpdateOneBy updates one document matching a typed filter.
func (r *Repository[T]) UpdateOneBy(ctx context.Context, filter Filter, update UpdateDoc) (matched bool, err error) {
	if len(filter.BSON()) == 0 {
		return false, fmt.Errorf("update one by: filter is required")
	}
	if update.IsEmpty() {
		return false, fmt.Errorf("update one by: update is required")
	}
	res, err := r.Collection.UpdateOne(ctx, filter.BSON(), update.BSON())
	if err != nil {
		return false, fmt.Errorf("update one by: %w", err)
	}
	return res.MatchedCount > 0, nil
}

// UpdateManyBy updates many documents matching a typed filter.
func (r *Repository[T]) UpdateManyBy(ctx context.Context, filter Filter, update UpdateDoc) (matched int64, err error) {
	if len(filter.BSON()) == 0 {
		return 0, fmt.Errorf("update many by: filter is required")
	}
	if update.IsEmpty() {
		return 0, fmt.Errorf("update many by: update is required")
	}
	res, err := r.Collection.UpdateMany(ctx, filter.BSON(), update.BSON())
	if err != nil {
		return 0, fmt.Errorf("update many by: %w", err)
	}
	return res.MatchedCount, nil
}

// FindOneAndUpdateBy updates one document and returns the matched document.
func (r *Repository[T]) FindOneAndUpdateBy(ctx context.Context, filter Filter, update UpdateDoc) (*T, error) {
	if len(filter.BSON()) == 0 {
		return nil, fmt.Errorf("find one and update by: filter is required")
	}
	if update.IsEmpty() {
		return nil, fmt.Errorf("find one and update by: update is required")
	}
	var entity T
	if err := r.Collection.FindOneAndUpdate(ctx, filter.BSON(), update.BSON()).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find one and update by: %w", err)
	}
	return &entity, nil
}

// UpsertOneBy updates one document or inserts it when absent.
// matched=true means an existing document matched the filter.
func (r *Repository[T]) UpsertOneBy(ctx context.Context, filter Filter, update UpdateDoc) (matched bool, err error) {
	if len(filter.BSON()) == 0 {
		return false, fmt.Errorf("upsert one by: filter is required")
	}
	if update.IsEmpty() {
		return false, fmt.Errorf("upsert one by: update is required")
	}
	res, err := r.Collection.UpdateOne(ctx, filter.BSON(), update.BSON(), options.UpdateOne().SetUpsert(true))
	if err != nil {
		return false, fmt.Errorf("upsert one by: %w", err)
	}
	return res.MatchedCount > 0, nil
}

// DeleteOneBy hard-deletes one document matching a typed filter.
func (r *Repository[T]) DeleteOneBy(ctx context.Context, filter Filter) error {
	if len(filter.BSON()) == 0 {
		return fmt.Errorf("delete one by: filter is required")
	}
	if _, err := r.Collection.DeleteOne(ctx, filter.BSON()); err != nil {
		return fmt.Errorf("delete one by: %w", err)
	}
	return nil
}

// DeleteAllBy hard-deletes all documents matching a typed filter.
func (r *Repository[T]) DeleteAllBy(ctx context.Context, filter Filter) error {
	if len(filter.BSON()) == 0 {
		return fmt.Errorf("delete all by: filter is required")
	}
	if _, err := r.Collection.DeleteMany(ctx, filter.BSON()); err != nil {
		return fmt.Errorf("delete all by: %w", err)
	}
	return nil
}

// EnsureIndexes creates indexes from typed declarations.
func (r *Repository[T]) EnsureIndexes(ctx context.Context, decls ...IndexDecl) error {
	if len(decls) == 0 {
		return nil
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, indexModels(decls)); err != nil {
		return fmt.Errorf("ensure indexes: %w", err)
	}
	return nil
}

// Aggregate runs a controlled aggregation pipeline into the provided output.
func (r *Repository[T]) Aggregate(ctx context.Context, pipeline mongo.Pipeline, into any) error {
	if len(pipeline) == 0 {
		return fmt.Errorf("aggregate: pipeline is required")
	}
	if into == nil {
		return fmt.Errorf("aggregate: output is required")
	}
	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("aggregate: %w", err)
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, into); err != nil {
		return fmt.Errorf("decode aggregate: %w", err)
	}
	return nil
}
