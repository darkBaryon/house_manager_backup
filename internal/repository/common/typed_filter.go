package common

import "go.mongodb.org/mongo-driver/v2/bson"

// Field marks repository-owned Mongo field names used by typed builders.
type Field string

// Filter is a typed wrapper around a Mongo filter document.
type Filter struct {
	doc bson.M
}

// BSON returns a copy of the underlying Mongo filter document.
func (f Filter) BSON() bson.M {
	return cloneBsonM(f.doc)
}

// EmptyFilter matches all documents.
func EmptyFilter() Filter {
	return Filter{doc: bson.M{}}
}

// Eq builds an equality condition.
func Eq(field Field, value any) Filter {
	return Filter{doc: bson.M{string(field): value}}
}

// Ne builds a not-equal condition.
func Ne(field Field, value any) Filter {
	return Filter{doc: bson.M{string(field): bson.M{"$ne": value}}}
}

// In builds an $in condition.
func In(field Field, values any) Filter {
	return Filter{doc: bson.M{string(field): bson.M{"$in": values}}}
}

// Regex builds a regex condition.
func Regex(field Field, pattern string, opts string) Filter {
	return Filter{doc: bson.M{string(field): bson.Regex{Pattern: pattern, Options: opts}}}
}

// And combines filters with $and.
func And(filters ...Filter) Filter {
	parts := make([]bson.M, 0, len(filters))
	for _, filter := range filters {
		parts = append(parts, filter.BSON())
	}
	return Filter{doc: bson.M{"$and": parts}}
}

// Or combines filters with $or.
func Or(filters ...Filter) Filter {
	parts := make([]bson.M, 0, len(filters))
	for _, filter := range filters {
		parts = append(parts, filter.BSON())
	}
	return Filter{doc: bson.M{"$or": parts}}
}
