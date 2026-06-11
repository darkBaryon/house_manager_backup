package common

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// SortAsc sorts a field in ascending order.
	SortAsc = 1
	// SortDesc sorts a field in descending order.
	SortDesc = -1
)

// QueryOption configures a typed find query.
type QueryOption func(*queryOptions)

type queryOptions struct {
	sort  bson.D
	limit *int64
	skip  *int64
}

// SortBy adds one sort field.
func SortBy(field Field, dir int) QueryOption {
	return func(q *queryOptions) {
		if dir >= 0 {
			dir = SortAsc
		} else {
			dir = SortDesc
		}
		q.sort = append(q.sort, bson.E{Key: string(field), Value: dir})
	}
}

// Limit sets the maximum number of documents to return.
func Limit(n int64) QueryOption {
	return func(q *queryOptions) {
		q.limit = &n
	}
}

// Skip sets the number of documents to skip.
func Skip(n int64) QueryOption {
	return func(q *queryOptions) {
		q.skip = &n
	}
}

func buildFindOptions(opts ...QueryOption) []options.Lister[options.FindOptions] {
	if len(opts) == 0 {
		return nil
	}

	typed := queryOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&typed)
		}
	}

	builder := options.Find()
	if len(typed.sort) > 0 {
		builder.SetSort(typed.sort)
	}
	if typed.limit != nil {
		builder.SetLimit(*typed.limit)
	}
	if typed.skip != nil {
		builder.SetSkip(*typed.skip)
	}
	return []options.Lister[options.FindOptions]{builder}
}
