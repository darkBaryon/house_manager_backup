package common

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"strings"
	"testing"
)

func TestCloneBsonM(t *testing.T) {
	src := bson.M{"name": "alpha", "count": 1}
	cloned := cloneBsonM(src)
	cloned["name"] = "beta"

	if src["name"] != "alpha" {
		t.Fatalf("expected source map to remain unchanged, got %v", src["name"])
	}
}

func TestCloneBsonMNil(t *testing.T) {
	if cloneBsonM(nil) != nil {
		t.Fatal("expected nil clone for nil source")
	}
}

func TestNotDeletedByIDFilter(t *testing.T) {
	id := bson.NewObjectID()
	filter := notDeletedByIDFilter(id)

	if got := filter["_id"]; got != id {
		t.Fatalf("expected _id=%v, got %v", id, got)
	}

	statusFilter, ok := filter["status"].(bson.M)
	if !ok {
		t.Fatalf("expected nested status filter, got %T", filter["status"])
	}
	if got := statusFilter["$ne"]; got != commonmodel.StatusDeleted {
		t.Fatalf("expected status != %d, got %v", commonmodel.StatusDeleted, got)
	}
}

func TestFindByIDRejectsZeroID(t *testing.T) {
	repo := &Repository[authmodel.User]{}
	_, err := repo.FindByID(context.Background(), bson.ObjectID{})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected zero id error, got %v", err)
	}
}

func TestFindByIDIncludingDeletedRejectsZeroID(t *testing.T) {
	repo := &Repository[authmodel.User]{}
	_, err := repo.FindByIDIncludingDeleted(context.Background(), bson.ObjectID{})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected zero id error, got %v", err)
	}
}

func TestUpdateFieldsByIDRejectsZeroID(t *testing.T) {
	repo := &Repository[authmodel.User]{}
	err := repo.UpdateFieldsByID(context.Background(), bson.ObjectID{}, bson.M{"nickname": "x"})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected zero id error, got %v", err)
	}
}

func TestSoftDeleteByIDRejectsZeroID(t *testing.T) {
	repo := &Repository[authmodel.User]{}
	err := repo.SoftDeleteByID(context.Background(), bson.ObjectID{})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected zero id error, got %v", err)
	}
}

func TestBuildUpdateFieldsByIDDocDoesNotMutateInput(t *testing.T) {
	fields := bson.M{"project_name": "泊寓南山"}
	update := buildUpdateFieldsByIDDoc(fields, 123)

	if _, ok := fields["updated_at"]; ok {
		t.Fatal("expected original fields to remain unchanged")
	}

	setFields := update["$set"].(bson.M)
	if got := setFields["updated_at"]; got != int64(123) {
		t.Fatalf("expected updated_at=123, got %v", got)
	}
	if got := setFields["project_name"]; got != "泊寓南山" {
		t.Fatalf("expected project_name preserved, got %v", got)
	}
}

func TestBuildUpsertFieldsDocDoesNotSetVersionOnInsert(t *testing.T) {
	fields := bson.M{"building_name": "A座"}
	update := buildUpsertFieldsDoc(fields, 456)

	if _, ok := fields["updated_at"]; ok {
		t.Fatal("expected original fields to remain unchanged")
	}

	setOnInsert := update["$setOnInsert"].(bson.M)
	if _, ok := setOnInsert["version"]; ok {
		t.Fatal("expected version to be absent from $setOnInsert")
	}
	if got := setOnInsert["created_at"]; got != int64(456) {
		t.Fatalf("expected created_at=456, got %v", got)
	}
}

func TestBuildUpsertFieldsDocAvoidsSetOnInsertConflicts(t *testing.T) {
	update := buildUpsertFieldsDoc(bson.M{
		"status":     commonmodel.StatusActive,
		"created_at": int64(123),
	}, 456)

	setOnInsert := update["$setOnInsert"].(bson.M)
	if _, ok := setOnInsert["status"]; ok {
		t.Fatal("expected status to be omitted from $setOnInsert when $set already contains it")
	}
	if _, ok := setOnInsert["created_at"]; ok {
		t.Fatal("expected created_at to be omitted from $setOnInsert when $set already contains it")
	}
}

func TestUpsertFieldsRejectsEmptyFilter(t *testing.T) {
	repo := &Repository[authmodel.User]{}
	_, err := repo.UpsertFields(context.Background(), bson.M{}, bson.M{"nickname": "x"})
	if err == nil || !strings.Contains(err.Error(), "filter is required") {
		t.Fatalf("expected empty filter error, got %v", err)
	}
}
