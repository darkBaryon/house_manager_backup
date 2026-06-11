package common

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestIndexDeclModel(t *testing.T) {
	decl := NewIndex(
		"status_1_updated_at_-1",
		IndexKey(Field("status"), SortAsc),
		IndexKey(Field("updated_at"), SortDesc),
	).WithUnique()

	model := decl.model()
	keys := model.Keys.(bson.D)
	if keys[0].Key != "status" || keys[0].Value != SortAsc {
		t.Fatalf("unexpected first index key: %#v", keys[0])
	}
	if keys[1].Key != "updated_at" || keys[1].Value != SortDesc {
		t.Fatalf("unexpected second index key: %#v", keys[1])
	}

	indexOptions := options.IndexOptions{}
	for _, setter := range model.Options.List() {
		if err := setter(&indexOptions); err != nil {
			t.Fatalf("apply index option: %v", err)
		}
	}
	if indexOptions.Name == nil || *indexOptions.Name != "status_1_updated_at_-1" {
		t.Fatalf("expected named index, got %v", indexOptions.Name)
	}
	if indexOptions.Unique == nil || !*indexOptions.Unique {
		t.Fatalf("expected unique index, got %v", indexOptions.Unique)
	}
}
