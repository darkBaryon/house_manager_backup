package common

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestBuildFindOptions(t *testing.T) {
	listers := buildFindOptions(
		SortBy(Field("created_at"), SortDesc),
		SortBy(Field("_id"), SortDesc),
		Skip(20),
		Limit(10),
	)
	if len(listers) != 1 {
		t.Fatalf("expected one options lister, got %d", len(listers))
	}

	findOptions := options.FindOptions{}
	for _, setter := range listers[0].List() {
		if err := setter(&findOptions); err != nil {
			t.Fatalf("apply find option: %v", err)
		}
	}

	sort := findOptions.Sort.(bson.D)
	if sort[0].Key != "created_at" || sort[0].Value != SortDesc {
		t.Fatalf("unexpected first sort key: %#v", sort[0])
	}
	if sort[1].Key != "_id" || sort[1].Value != SortDesc {
		t.Fatalf("unexpected second sort key: %#v", sort[1])
	}
	if findOptions.Skip == nil || *findOptions.Skip != 20 {
		t.Fatalf("expected skip=20, got %v", findOptions.Skip)
	}
	if findOptions.Limit == nil || *findOptions.Limit != 10 {
		t.Fatalf("expected limit=10, got %v", findOptions.Limit)
	}
}

func TestBuildFindOptionsEmpty(t *testing.T) {
	if got := buildFindOptions(); got != nil {
		t.Fatalf("expected nil find options for empty input, got %#v", got)
	}
}
