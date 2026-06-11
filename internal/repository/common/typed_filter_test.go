package common

import (
	"testing"

	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestTypedFilterBuilders(t *testing.T) {
	id := bson.NewObjectID()
	filter := And(
		Eq(Field("_id"), id),
		Ne(Field("status"), 0),
		Or(
			Regex(Field("role_name"), "admin", "i"),
			Regex(Field("role_code"), "admin", "i"),
		),
	)

	doc := filter.BSON()
	parts := doc["$and"].([]bson.M)
	if got := parts[0]["_id"]; got != id {
		t.Fatalf("expected id condition, got %v", got)
	}
	status := parts[1]["status"].(bson.M)
	if got := status["$ne"]; got != 0 {
		t.Fatalf("expected status $ne 0, got %v", got)
	}
	orParts := parts[2]["$or"].([]bson.M)
	roleName := orParts[0]["role_name"].(bson.Regex)
	if roleName.Pattern != "admin" || roleName.Options != "i" {
		t.Fatalf("unexpected role_name regex: %#v", roleName)
	}
}

func TestNinFilterBuilder(t *testing.T) {
	values := []string{"a", "b"}
	filter := Nin(Field("role_code"), values).BSON()
	condition := filter["role_code"].(bson.M)
	if got := condition["$nin"]; len(got.([]string)) != 2 {
		t.Fatalf("expected $nin values, got %#v", got)
	}
}

func TestFilterFromBSONDoesNotMutateSource(t *testing.T) {
	src := bson.M{"status": commonmodel.StatusActive}
	filter := FilterFromBSON(src)
	src["status"] = commonmodel.StatusDeleted

	if got := filter.BSON()["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected wrapped filter source to remain unchanged, got %v", got)
	}
}

func TestTypedFilterBSONDoesNotMutateSource(t *testing.T) {
	filter := Eq(Field("role_code"), "admin")
	doc := filter.BSON()
	doc["role_code"] = "manager"

	if got := filter.BSON()["role_code"]; got != "admin" {
		t.Fatalf("expected filter source to remain unchanged, got %v", got)
	}
}

func TestStatusPredicateFragments(t *testing.T) {
	active := Active().BSON()
	if got := active["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status predicate, got %v", got)
	}

	notDeleted := NotDeleted().BSON()
	status := notDeleted["status"].(bson.M)
	if got := status["$ne"]; got != commonmodel.StatusDeleted {
		t.Fatalf("expected not-deleted status predicate, got %v", got)
	}
}
