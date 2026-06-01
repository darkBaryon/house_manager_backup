package chat

import (
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSessionValidateForCreate(t *testing.T) {
	session := &Session{
		UserID:        bson.NewObjectID(),
		SessionStatus: SessionStatusActive,
		StartedAt:     1775451571,
		LastActiveAt:  1775451571,
	}
	if err := session.ValidateForCreate(); err != nil {
		t.Fatalf("ValidateForCreate() error = %v", err)
	}
}

func TestSessionValidateForCreateRejectsInvalidStatus(t *testing.T) {
	session := &Session{
		UserID:        bson.NewObjectID(),
		SessionStatus: 99,
		StartedAt:     1775451571,
		LastActiveAt:  1775451571,
	}
	err := session.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "sessionStatus is invalid") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestMessageValidateForCreate(t *testing.T) {
	message := &Message{
		SessionID: bson.NewObjectID(),
		UserID:    bson.NewObjectID(),
		Seq:       1,
		Role:      MessageRoleUser,
		Content:   "hello",
	}
	if err := message.ValidateForCreate(); err != nil {
		t.Fatalf("ValidateForCreate() error = %v", err)
	}
}

func TestMessageValidateForCreateRejectsBlankContent(t *testing.T) {
	message := &Message{
		SessionID: bson.NewObjectID(),
		UserID:    bson.NewObjectID(),
		Seq:       1,
		Role:      MessageRoleUser,
		Content:   " ",
	}
	err := message.ValidateForCreate()
	if err == nil || !strings.Contains(err.Error(), "content is required") {
		t.Fatalf("expected content error, got %v", err)
	}
}
