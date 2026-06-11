package common

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateDocBuilders(t *testing.T) {
	update := NewUpdateDoc().
		Set(Field("role_name"), "管理员").
		Unset(Field("description")).
		Inc(Field("version"), 1).
		SetOnInsert(Field("created_at"), int64(123))

	doc := update.BSON()
	if got := doc["$set"].(bson.M)["role_name"]; got != "管理员" {
		t.Fatalf("expected role_name set, got %v", got)
	}
	if got := doc["$unset"].(bson.M)["description"]; got != "" {
		t.Fatalf("expected description unset marker, got %v", got)
	}
	if got := doc["$inc"].(bson.M)["version"]; got != 1 {
		t.Fatalf("expected version increment, got %v", got)
	}
	if got := doc["$setOnInsert"].(bson.M)["created_at"]; got != int64(123) {
		t.Fatalf("expected created_at setOnInsert, got %v", got)
	}
}

func TestUpdateDocBSONDoesNotMutateSource(t *testing.T) {
	update := NewUpdateDoc().Set(Field("role_name"), "管理员")
	doc := update.BSON()
	doc["$set"].(bson.M)["role_name"] = "运营"

	if got := update.BSON()["$set"].(bson.M)["role_name"]; got != "管理员" {
		t.Fatalf("expected update source to remain unchanged, got %v", got)
	}
}

func TestUpdateDocIsEmpty(t *testing.T) {
	if !NewUpdateDoc().IsEmpty() {
		t.Fatal("expected new update doc to be empty")
	}
	if NewUpdateDoc().Set(Field("role_name"), "管理员").IsEmpty() {
		t.Fatal("expected update doc with set field to be non-empty")
	}
}
