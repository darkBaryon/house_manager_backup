package common

import (
	"testing"

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

func TestTypedFilterBSONDoesNotMutateSource(t *testing.T) {
	filter := Eq(Field("role_code"), "admin")
	doc := filter.BSON()
	doc["role_code"] = "manager"

	if got := filter.BSON()["role_code"]; got != "admin" {
		t.Fatalf("expected filter source to remain unchanged, got %v", got)
	}
}
