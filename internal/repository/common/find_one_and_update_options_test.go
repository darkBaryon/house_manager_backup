package common

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestBuildFindOneAndUpdateOptionsDefaultBefore(t *testing.T) {
	if got := buildFindOneAndUpdateOptions(); got != nil {
		t.Fatalf("expected nil options for default before behavior, got %#v", got)
	}
}

func TestBuildFindOneAndUpdateOptionsReturnAfter(t *testing.T) {
	listers := buildFindOneAndUpdateOptions(ReturnAfter())
	if len(listers) != 1 {
		t.Fatalf("expected one options lister, got %d", len(listers))
	}

	findOptions := options.FindOneAndUpdateOptions{}
	for _, setter := range listers[0].List() {
		if err := setter(&findOptions); err != nil {
			t.Fatalf("apply find one and update option: %v", err)
		}
	}
	if findOptions.ReturnDocument == nil || *findOptions.ReturnDocument != options.After {
		t.Fatalf("expected ReturnDocument=After, got %v", findOptions.ReturnDocument)
	}
}
