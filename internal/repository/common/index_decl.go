package common

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexField describes one typed field in an index key.
type IndexField struct {
	Field Field
	Dir   int
}

// IndexKey creates one index key part.
func IndexKey(field Field, dir int) IndexField {
	if dir >= 0 {
		dir = SortAsc
	} else {
		dir = SortDesc
	}
	return IndexField{Field: field, Dir: dir}
}

// IndexDecl describes a Mongo index without exposing Collection.Indexes.
type IndexDecl struct {
	Name   string
	Keys   []IndexField
	Unique bool
}

// NewIndex creates a named index declaration.
func NewIndex(name string, keys ...IndexField) IndexDecl {
	return IndexDecl{Name: name, Keys: keys}
}

// WithUnique marks the index as unique.
func (d IndexDecl) WithUnique() IndexDecl {
	d.Unique = true
	return d
}

func (d IndexDecl) model() mongo.IndexModel {
	keys := make(bson.D, 0, len(d.Keys))
	for _, key := range d.Keys {
		keys = append(keys, bson.E{Key: string(key.Field), Value: key.Dir})
	}

	opts := options.Index().SetName(d.Name)
	if d.Unique {
		opts.SetUnique(true)
	}
	return mongo.IndexModel{Keys: keys, Options: opts}
}

func indexModels(decls []IndexDecl) []mongo.IndexModel {
	models := make([]mongo.IndexModel, 0, len(decls))
	for _, decl := range decls {
		models = append(models, decl.model())
	}
	return models
}
