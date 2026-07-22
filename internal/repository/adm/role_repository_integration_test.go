package adm

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestRoleRepositoryUpdateIntegration(t *testing.T) {
	if os.Getenv("ADM_ROLE_REPOSITORY_INTEGRATION") != "1" {
		t.Skip("set ADM_ROLE_REPOSITORY_INTEGRATION=1 to run Mongo-backed role repository integration tests")
	}

	cfg := roleRepositoryTestMongoConfig(t)
	if !strings.Contains(strings.ToLower(cfg.Database), "test") && os.Getenv("ADM_ROLE_REPOSITORY_TEST_ALLOW_NON_TEST_DB") != "1" {
		t.Fatalf("refusing to run role repository integration tests against %q; use a test database or set ADM_ROLE_REPOSITORY_TEST_ALLOW_NON_TEST_DB=1 explicitly", cfg.Database)
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

	repo := NewRoleRepository(client)
	coll := client.Collection(authmodel.CollectionAdmRole)
	codePrefix := "it_role_update_" + bson.NewObjectID().Hex()
	cleanup := func() {
		if _, err := coll.DeleteMany(context.Background(), bson.M{"role_code": bson.M{"$regex": "^" + codePrefix}}); err != nil {
			t.Logf("cleanup role docs: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	active := &authmodel.AdmRole{
		RoleName:    "旧角色",
		RoleCode:    codePrefix + "_active",
		Description: "保留说明",
	}
	if err := repo.Create(ctx, active); err != nil {
		t.Fatalf("create active role: %v", err)
	}
	oldVersion := active.Version

	newName := "新角色"
	if err := repo.Update(ctx, active.ID, RoleUpdate{RoleName: &newName}); err != nil {
		t.Fatalf("update active role: %v", err)
	}
	updated, err := repo.FindActiveByID(ctx, active.ID)
	if err != nil {
		t.Fatalf("find updated role: %v", err)
	}
	if updated.RoleName != newName {
		t.Fatalf("expected role_name updated, got %q", updated.RoleName)
	}
	if updated.Description != "保留说明" {
		t.Fatalf("expected unspecified description untouched, got %q", updated.Description)
	}
	if updated.Version != oldVersion+1 {
		t.Fatalf("expected version increment from %d to %d, got %d", oldVersion, oldVersion+1, updated.Version)
	}

	inactive := authmodel.AdmRole{
		CommonFields: commonmodel.CommonFields{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			Status:    commonmodel.StatusDeleted,
			Version:   7,
		},
		RoleName:    "停用角色",
		RoleCode:    codePrefix + "_inactive",
		Description: "不可更新",
	}
	if _, err := coll.InsertOne(ctx, inactive); err != nil {
		t.Fatalf("insert inactive role: %v", err)
	}

	blockedName := "不应写入"
	err = repo.Update(ctx, inactive.ID, RoleUpdate{RoleName: &blockedName})
	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("expected ErrNoDocuments for inactive role update, got %v", err)
	}

	var afterInactive authmodel.AdmRole
	if err := coll.FindOne(ctx, bson.M{"_id": inactive.ID}).Decode(&afterInactive); err != nil {
		t.Fatalf("find inactive role: %v", err)
	}
	if afterInactive.RoleName != inactive.RoleName {
		t.Fatalf("expected inactive role_name unchanged, got %q", afterInactive.RoleName)
	}
	if afterInactive.Version != inactive.Version {
		t.Fatalf("expected inactive version unchanged, got %d", afterInactive.Version)
	}
}

func roleRepositoryTestMongoConfig(t *testing.T) dbmongo.Config {
	t.Helper()
	database := envOr("ADM_ROLE_REPOSITORY_TEST_DB", "house_manager_test")
	authSource := envOr("ADM_ROLE_REPOSITORY_TEST_AUTH_SOURCE", database)
	return dbmongo.Config{
		Addrs:                  splitCSV(envOr("ADM_ROLE_REPOSITORY_TEST_ADDRS", "127.0.0.1:27018")),
		Database:               database,
		AuthSource:             authSource,
		Username:               strings.TrimSpace(os.Getenv("ADM_ROLE_REPOSITORY_TEST_USERNAME")),
		Password:               os.Getenv("ADM_ROLE_REPOSITORY_TEST_PASSWORD"),
		ConnectTimeout:         5 * time.Second,
		SocketTimeout:          5 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
		EnableRetryReads:       true,
		EnableRetryWrites:      true,
	}
}

func envOr(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
