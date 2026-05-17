package adminauth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"house-manager/internal/config"
	adminauthhandler "house-manager/internal/handler/v1/admin/auth"
	"house-manager/internal/middleware"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	admrepo "house-manager/internal/repository/adm"
	adminauthsvc "house-manager/internal/service/admin/auth"
	dbmongo "house-manager/pkg/database/mongo"
	dbredis "house-manager/pkg/database/redis"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type integrationFixture struct {
	ctx         context.Context
	cancel      context.CancelFunc
	mongoClient *dbmongo.Client
	redisClient *dbredis.Client
	store       *session.Store
	router      *gin.Engine

	staffID        bson.ObjectID
	staffAuthID    bson.ObjectID
	roleID         bson.ObjectID
	permissionID   bson.ObjectID
	phone          string
	password       string
	roleCode       string
	permissionCode string
}

func newIntegrationFixture(t *testing.T) *integrationFixture {
	t.Helper()
	if os.Getenv("ADMIN_AUTH_INTEGRATION") != "1" {
		t.Skip("set ADMIN_AUTH_INTEGRATION=1 to run Mongo+Redis-backed admin auth integration tests")
	}

	root := repoRoot(t)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cfg := testConfig(t, root)
	if !strings.Contains(strings.ToLower(cfg.MongoDB.Database), "test") && os.Getenv("ADMIN_AUTH_TEST_ALLOW_NON_TEST_DB") != "1" {
		cancel()
		t.Fatalf("refusing to run admin auth integration tests against %q; use a test database or set ADMIN_AUTH_TEST_ALLOW_NON_TEST_DB=1 explicitly", cfg.MongoDB.Database)
	}

	mongoClient, err := newMongoClient(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("connect integration mongo: %v", err)
	}
	redisClient, err := newRedisClient(ctx, cfg)
	if err != nil {
		_ = mongoClient.Close(context.Background())
		cancel()
		t.Fatalf("connect integration redis: %v", err)
	}

	store := session.NewStore(redisClient, 2*time.Hour)
	service := adminauthsvc.NewService(
		admrepo.NewStaffRepository(mongoClient),
		admrepo.NewStaffAuthRepository(mongoClient),
		admrepo.NewStaffRoleRepository(mongoClient),
		admrepo.NewRoleRepository(mongoClient),
		admrepo.NewRolePermissionRepository(mongoClient),
		admrepo.NewPermissionRepository(mongoClient),
		admrepo.NewLoginLogRepository(mongoClient),
		store,
	)
	publicHandler := adminauthhandler.NewPublicHandler(service)
	protectedHandler := adminauthhandler.NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	publicGroup := router.Group("/api/v1")
	protectedGroup := router.Group("/api/v1", middleware.AdminAuth(store))
	publicHandler.RegisterRoutes(publicGroup)
	protectedHandler.RegisterRoutes(protectedGroup)

	fixture := &integrationFixture{
		ctx:            ctx,
		cancel:         cancel,
		mongoClient:    mongoClient,
		redisClient:    redisClient,
		store:          store,
		router:         router,
		password:       "secret123",
		phone:          fmt.Sprintf("itest_admin_%d", time.Now().UnixNano()),
		roleCode:       fmt.Sprintf("itest_super_admin_%d", time.Now().UnixNano()),
		permissionCode: fmt.Sprintf("itest_staff_view_%d", time.Now().UnixNano()),
	}

	fixture.seedAdminAuthData(t)
	t.Cleanup(func() {
		fixture.cleanup(t)
	})
	return fixture
}

func (f *integrationFixture) seedAdminAuthData(t *testing.T) {
	t.Helper()
	now := time.Now().Unix()
	f.staffID = bson.NewObjectID()
	f.staffAuthID = bson.NewObjectID()
	f.roleID = bson.NewObjectID()
	f.permissionID = bson.NewObjectID()

	passwordHash, err := bcryptPassword(f.password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	staffColl := f.mongoClient.Collection(authmodel.CollectionAdmStaff)
	staffAuthColl := f.mongoClient.Collection(authmodel.CollectionAdmStaffAuth)
	roleColl := f.mongoClient.Collection(authmodel.CollectionAdmRole)
	permissionColl := f.mongoClient.Collection(authmodel.CollectionAdmPermission)
	staffRoleColl := f.mongoClient.Collection(authmodel.CollectionAdmStaffRole)
	rolePermissionColl := f.mongoClient.Collection(authmodel.CollectionAdmRolePermission)

	_, err = staffColl.InsertOne(f.ctx, authmodel.AdmStaff{
		CommonFields: commonmodel.CommonFields{
			ID:        f.staffID,
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		Name:          "itest_admin_staff",
		Phone:         f.phone,
		Email:         "itest@example.com",
		Department:    "测试部",
		JobTitle:      "测试管理员",
		ContactQRCode: "",
	})
	if err != nil {
		t.Fatalf("insert staff: %v", err)
	}

	_, err = staffAuthColl.InsertOne(f.ctx, authmodel.AdmStaffAuth{
		CommonFields: commonmodel.CommonFields{
			ID:        f.staffAuthID,
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		StaffID:           f.staffID,
		AuthType:          authmodel.PasswordAuthTypePassword,
		PasswordHash:      passwordHash,
		PasswordUpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("insert staff auth: %v", err)
	}

	_, err = roleColl.InsertOne(f.ctx, authmodel.AdmRole{
		CommonFields: commonmodel.CommonFields{
			ID:        f.roleID,
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		RoleName: "集成测试管理员",
		RoleCode: f.roleCode,
		IsSystem: 0,
	})
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}

	_, err = permissionColl.InsertOne(f.ctx, authmodel.AdmPermission{
		CommonFields: commonmodel.CommonFields{
			ID:        f.permissionID,
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		PermissionName: "查看员工",
		PermissionCode: f.permissionCode,
		Module:         "staff",
		Action:         "view",
	})
	if err != nil {
		t.Fatalf("insert permission: %v", err)
	}

	_, err = staffRoleColl.InsertOne(f.ctx, authmodel.AdmStaffRole{
		CommonFields: commonmodel.CommonFields{
			ID:        bson.NewObjectID(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		StaffID:    f.staffID,
		RoleID:     f.roleID,
		AssignedAt: now,
	})
	if err != nil {
		t.Fatalf("insert staff role: %v", err)
	}

	_, err = rolePermissionColl.InsertOne(f.ctx, authmodel.AdmRolePermission{
		CommonFields: commonmodel.CommonFields{
			ID:        bson.NewObjectID(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    commonmodel.StatusActive,
			Version:   1,
		},
		RoleID:       f.roleID,
		PermissionID: f.permissionID,
		AssignedAt:   now,
	})
	if err != nil {
		t.Fatalf("insert role permission: %v", err)
	}
}

func (f *integrationFixture) cleanup(t *testing.T) {
	t.Helper()
	collections := []string{
		authmodel.CollectionAdmLoginLog,
		authmodel.CollectionAdmRolePermission,
		authmodel.CollectionAdmStaffRole,
		authmodel.CollectionAdmPermission,
		authmodel.CollectionAdmRole,
		authmodel.CollectionAdmStaffAuth,
		authmodel.CollectionAdmStaff,
	}
	filters := map[string]bson.M{
		authmodel.CollectionAdmLoginLog:       {"staff_id": f.staffID},
		authmodel.CollectionAdmRolePermission: {"role_id": f.roleID},
		authmodel.CollectionAdmStaffRole:      {"staff_id": f.staffID},
		authmodel.CollectionAdmPermission:     {"_id": f.permissionID},
		authmodel.CollectionAdmRole:           {"_id": f.roleID},
		authmodel.CollectionAdmStaffAuth:      {"staff_id": f.staffID},
		authmodel.CollectionAdmStaff:          {"_id": f.staffID},
	}
	for _, name := range collections {
		if _, err := f.mongoClient.Collection(name).DeleteMany(context.Background(), filters[name]); err != nil {
			t.Logf("cleanup %s: %v", name, err)
		}
	}
	if err := f.redisClient.Close(); err != nil {
		t.Logf("close integration redis: %v", err)
	}
	if err := f.mongoClient.Close(context.Background()); err != nil {
		t.Logf("close integration mongo: %v", err)
	}
	f.cancel()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("locate repo root from %s: %v", wd, err)
	}
	return root
}

func testConfig(t *testing.T, root string) *config.Config {
	t.Helper()
	cfg, err := config.Load(filepath.Join(root, "config", "config.test.yaml"))
	if err != nil {
		t.Fatalf("load config.test.yaml: %v", err)
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_AUTH_TEST_ADDRS")); value != "" {
		cfg.MongoDB.Addrs = splitCSV(value)
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_AUTH_TEST_DB")); value != "" {
		cfg.MongoDB.Database = value
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_AUTH_TEST_AUTH_SOURCE")); value != "" {
		cfg.MongoDB.AuthSource = value
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_AUTH_TEST_USERNAME")); value != "" {
		cfg.MongoDB.Username = value
	}
	if value := os.Getenv("ADMIN_AUTH_TEST_PASSWORD"); value != "" {
		cfg.MongoDB.Password = value
	}
	if value := strings.TrimSpace(os.Getenv("ADMIN_AUTH_TEST_REDIS_ADDRS")); value != "" {
		cfg.Redis.Addrs = splitCSV(value)
	}
	if value := os.Getenv("ADMIN_AUTH_TEST_REDIS_PASSWORD"); value != "" {
		cfg.Redis.Password = value
	}
	return cfg
}

func newMongoClient(ctx context.Context, cfg *config.Config) (*dbmongo.Client, error) {
	connectTimeout, err := cfg.MongoDB.ConnectTimeoutDuration()
	if err != nil {
		return nil, err
	}
	socketTimeout, err := cfg.MongoDB.SocketTimeoutDuration()
	if err != nil {
		return nil, err
	}
	serverSelectionTimeout, err := cfg.MongoDB.ServerSelectionTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return dbmongo.NewClient(ctx, dbmongo.Config{
		Addrs:                  cfg.MongoDB.Addrs,
		Database:               cfg.MongoDB.Database,
		AuthSource:             cfg.MongoDB.AuthSource,
		Username:               cfg.MongoDB.Username,
		Password:               cfg.MongoDB.Password,
		PoolSize:               cfg.MongoDB.PoolSize,
		MinPoolSize:            cfg.MongoDB.MinPoolSize,
		ConnectTimeout:         connectTimeout,
		SocketTimeout:          socketTimeout,
		ServerSelectionTimeout: serverSelectionTimeout,
		EnableRetryReads:       cfg.MongoDB.RetryReads,
		EnableRetryWrites:      cfg.MongoDB.RetryWrites,
		ReplicaSet:             cfg.MongoDB.ReplicaSet,
	})
}

func newRedisClient(ctx context.Context, cfg *config.Config) (*dbredis.Client, error) {
	connTimeout, err := cfg.Redis.ConnTimeoutDuration()
	if err != nil {
		return nil, err
	}
	readTimeout, err := cfg.Redis.ReadTimeoutDuration()
	if err != nil {
		return nil, err
	}
	writeTimeout, err := cfg.Redis.WriteTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return dbredis.NewClient(ctx, dbredis.Config{
		Addrs:        cfg.Redis.Addrs,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		ConnTimeout:  connTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		MaxRetries:   cfg.Redis.MaxRetries,
		ClusterMode:  cfg.Redis.ClusterMode,
	})
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func bcryptPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
