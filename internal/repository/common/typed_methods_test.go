package common

import (
	"context"
	"os"
	"strings"
	"testing"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateByIDRejectsZeroID(t *testing.T) {
	repo := &Repository[authmodel.AdmRole]{}
	err := repo.UpdateByID(context.Background(), bson.ObjectID{}, NewUpdateDoc().Set(Field("role_name"), "管理员"))
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected zero id error, got %v", err)
	}
}

func TestUpdateByIDDefaultPredicateContractComment(t *testing.T) {
	source, err := os.ReadFile("typed_methods.go")
	if err != nil {
		t.Fatalf("read typed_methods.go: %v", err)
	}
	if !strings.Contains(string(source), "Common default predicate = not deleted") {
		t.Fatal("expected UpdateByID contract comment to document not-deleted default predicate")
	}
}

func TestUpdateOneByRejectsEmptyFilter(t *testing.T) {
	repo := &Repository[authmodel.AdmRole]{}
	_, err := repo.UpdateOneBy(context.Background(), EmptyFilter(), NewUpdateDoc().Set(Field("role_name"), "管理员"))
	if err == nil || !strings.Contains(err.Error(), "filter is required") {
		t.Fatalf("expected empty filter error, got %v", err)
	}
}

func TestUpdateOneByRejectsEmptyUpdate(t *testing.T) {
	repo := &Repository[authmodel.AdmRole]{}
	_, err := repo.UpdateOneBy(context.Background(), Eq(Field("_id"), bson.NewObjectID()), NewUpdateDoc())
	if err == nil || !strings.Contains(err.Error(), "update is required") {
		t.Fatalf("expected empty update error, got %v", err)
	}
}

func TestUpsertOneByRejectsEmptyUpdate(t *testing.T) {
	repo := &Repository[authmodel.AdmRole]{}
	_, err := repo.UpsertOneBy(context.Background(), Eq(Field("role_code"), "admin"), NewUpdateDoc())
	if err == nil || !strings.Contains(err.Error(), "update is required") {
		t.Fatalf("expected empty update error, got %v", err)
	}
}

func TestAggregateRejectsEmptyPipeline(t *testing.T) {
	repo := &Repository[authmodel.AdmRole]{}
	err := repo.Aggregate(context.Background(), nil, &[]authmodel.AdmRole{})
	if err == nil || !strings.Contains(err.Error(), "pipeline is required") {
		t.Fatalf("expected empty pipeline error, got %v", err)
	}
}

func TestOptimisticLockBuilderShape(t *testing.T) {
	id := bson.NewObjectID()
	filter := And(
		Eq(Field("_id"), id),
		Eq(Field("status"), commonmodel.StatusActive),
	).BSON()
	update := NewUpdateDoc().
		Set(Field("role_name"), "管理员").
		Set(Field("updated_at"), int64(123)).
		Inc(Field("version"), 1).
		BSON()

	parts := filter["$and"].([]bson.M)
	if got := parts[0]["_id"]; got != id {
		t.Fatalf("expected id condition, got %v", got)
	}
	if got := parts[1]["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status condition, got %v", got)
	}
	setFields := update["$set"].(bson.M)
	if got := setFields["role_name"]; got != "管理员" {
		t.Fatalf("expected role_name set, got %v", got)
	}
	if _, ok := setFields["description"]; ok {
		t.Fatal("expected unspecified description to be absent from $set")
	}
	incFields := update["$inc"].(bson.M)
	if got := incFields["version"]; got != 1 {
		t.Fatalf("expected version increment, got %v", got)
	}
}
