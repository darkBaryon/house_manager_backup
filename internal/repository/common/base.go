package common

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// FindUniqueBy returns zero or one document for a typed filter and rejects duplicates.
func (r *Repository[T]) FindUniqueBy(ctx context.Context, filter Filter) (*T, error) {
	if len(filter.BSON()) == 0 {
		return nil, fmt.Errorf("find unique by: filter is required")
	}
	items, err := r.FindManyBy(ctx, filter, Limit(2))
	if err != nil {
		return nil, fmt.Errorf("find unique by: %w", err)
	}
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find unique by: multiple documents found")
	}
	return &items[0], nil
}

// CompactObjectIDs removes zero ObjectIDs while preserving the original order.
func CompactObjectIDs(ids []bson.ObjectID) []bson.ObjectID {
	if len(ids) == 0 {
		return nil
	}
	result := make([]bson.ObjectID, 0, len(ids))
	seen := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// CloneBSONMap returns a shallow copy of a BSON map.
func CloneBSONMap(src bson.M) bson.M {
	return cloneBsonM(src)
}
