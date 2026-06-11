package chat

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	chatmodel "house-manager/internal/model/chat"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestSessionRepositoryAllocateMessageSeqIntegration(t *testing.T) {
	if os.Getenv("CHAT_SESSION_REPOSITORY_INTEGRATION") != "1" {
		t.Skip("set CHAT_SESSION_REPOSITORY_INTEGRATION=1 to run Mongo-backed chat session repository integration tests")
	}

	cfg := chatSessionRepositoryTestMongoConfig(t)
	if !strings.Contains(strings.ToLower(cfg.Database), "test") && os.Getenv("CHAT_SESSION_REPOSITORY_TEST_ALLOW_NON_TEST_DB") != "1" {
		t.Fatalf("refusing to run chat session repository integration tests against %q; use a test database or set CHAT_SESSION_REPOSITORY_TEST_ALLOW_NON_TEST_DB=1 explicitly", cfg.Database)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := dbmongo.NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("connect integration mongo: %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			t.Logf("close integration mongo: %v", err)
		}
	}()

	repo := NewSessionRepository(client)
	coll := client.Collection(chatmodel.CollectionSession)
	userID := bson.NewObjectID()
	cleanup := func() {
		if _, err := coll.DeleteMany(context.Background(), bson.M{"user_id": userID}); err != nil {
			t.Logf("cleanup chat session docs: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	now := time.Now().Unix()
	active := &chatmodel.Session{
		UserID:        userID,
		SessionStatus: chatmodel.SessionStatusActive,
		StartedAt:     now,
		LastActiveAt:  now,
	}
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("create active chat session: %v", err)
	}
	first, err := repo.AllocateMessageSeq(ctx, active.ID)
	if err != nil {
		t.Fatalf("allocate first seq: %v", err)
	}
	if first != 1 {
		t.Fatalf("expected first allocated seq 1, got %d", first)
	}
	second, err := repo.AllocateMessageSeq(ctx, active.ID)
	if err != nil {
		t.Fatalf("allocate second seq: %v", err)
	}
	if second != 2 {
		t.Fatalf("expected second allocated seq 2, got %d", second)
	}

	ended := &chatmodel.Session{
		UserID:        userID,
		SessionStatus: chatmodel.SessionStatusEnded,
		StartedAt:     now,
		LastActiveAt:  now,
	}
	if err := repo.Create(ctx, ended); err != nil {
		t.Fatalf("create ended chat session: %v", err)
	}
	_, err = repo.AllocateMessageSeq(ctx, ended.ID)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("expected ErrNoDocuments for ended session allocation, got %v", err)
	}
}

func chatSessionRepositoryTestMongoConfig(t *testing.T) dbmongo.Config {
	t.Helper()
	database := chatEnvOr("CHAT_SESSION_REPOSITORY_TEST_DB", "house_manager_test")
	authSource := chatEnvOr("CHAT_SESSION_REPOSITORY_TEST_AUTH_SOURCE", database)
	return dbmongo.Config{
		Addrs:                  chatSplitCSV(chatEnvOr("CHAT_SESSION_REPOSITORY_TEST_ADDRS", "127.0.0.1:27018")),
		Database:               database,
		AuthSource:             authSource,
		Username:               strings.TrimSpace(os.Getenv("CHAT_SESSION_REPOSITORY_TEST_USERNAME")),
		Password:               os.Getenv("CHAT_SESSION_REPOSITORY_TEST_PASSWORD"),
		ConnectTimeout:         5 * time.Second,
		SocketTimeout:          5 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
		EnableRetryReads:       true,
		EnableRetryWrites:      true,
	}
}

func chatEnvOr(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func chatSplitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
