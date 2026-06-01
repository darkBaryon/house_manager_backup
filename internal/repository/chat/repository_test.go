package chat

import (
	"context"
	"strings"
	"testing"

	chatmodel "house-manager/internal/model/chat"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSessionRepositoryRejectsInvalidInputBeforeDB(t *testing.T) {
	repo := &SessionRepository{}
	_, err := repo.FindActiveByUserID(context.Background(), bson.NilObjectID, 0)
	if err == nil || !strings.Contains(err.Error(), "userID is required") {
		t.Fatalf("expected userID error, got %v", err)
	}
	err = repo.TouchActive(context.Background(), bson.NilObjectID, 1)
	if err == nil || !strings.Contains(err.Error(), "sessionID is required") {
		t.Fatalf("expected sessionID error, got %v", err)
	}
	err = repo.End(context.Background(), bson.NewObjectID(), 0)
	if err == nil || !strings.Contains(err.Error(), "endedAt is required") {
		t.Fatalf("expected endedAt error, got %v", err)
	}
	_, err = repo.AllocateMessageSeq(context.Background(), bson.NilObjectID)
	if err == nil || !strings.Contains(err.Error(), "sessionID is required") {
		t.Fatalf("expected sessionID error, got %v", err)
	}
}

func TestMessageRepositoryRejectsInvalidInputBeforeDB(t *testing.T) {
	repo := &MessageRepository{}
	_, err := repo.ListRecentBySessionID(context.Background(), bson.NilObjectID, 10)
	if err == nil || !strings.Contains(err.Error(), "sessionID is required") {
		t.Fatalf("expected sessionID error, got %v", err)
	}
}

func TestReverseMessages(t *testing.T) {
	items := []chatmodel.Message{
		{Seq: 3},
		{Seq: 2},
		{Seq: 1},
	}
	reverseMessages(items)
	if items[0].Seq != 1 || items[1].Seq != 2 || items[2].Seq != 3 {
		t.Fatalf("items not reversed: %+v", items)
	}
}
