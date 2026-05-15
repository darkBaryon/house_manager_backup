package hmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"house-manager/internal/config"
	hmdmodel "house-manager/internal/model/hmd"
	repohmd "house-manager/internal/repository/hmd"
	dbmongo "house-manager/pkg/database/mongo"
	"house-manager/pkg/errcode"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type integrationFixture struct {
	ctx    context.Context
	cancel context.CancelFunc
	client *dbmongo.Client
	svc    *Service
	prefix string
}

func TestServiceIntegrationCentralizedFlow(t *testing.T) {
	f := newIntegrationFixture(t)

	_, err := f.svc.CreateCentralizedProject(f.ctx, CreateCentralizedProjectInput{
		ProjectCode: f.prefix + "_missing_project_name",
		City:        "深圳",
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	project := mustCreateCentralizedProject(t, f, f.prefix+"_project_a")
	assertChange(t, project.Changes, HmdChangeCreated, HmdEntityCentralizedProject, HmdScopeCentralizedProject)

	_, err = f.svc.CreateCentralizedProject(f.ctx, CreateCentralizedProjectInput{
		ProjectName: project.Entity.ProjectName,
		ProjectCode: project.Entity.ProjectCode,
		City:        project.Entity.City,
		District:    project.Entity.District,
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	projects, err := f.svc.ListCentralizedProjects(f.ctx, ListCentralizedProjectsInput{City: "深圳"})
	requireNoError(t, err)
	assertContainsCentralizedProject(t, projects, project.Entity.ID)

	detail, err := f.svc.GetCentralizedProject(f.ctx, project.Entity.ID)
	requireNoError(t, err)
	if detail.ProjectCode != project.Entity.ProjectCode {
		t.Fatalf("expected project code %q, got %q", project.Entity.ProjectCode, detail.ProjectCode)
	}

	updatedProject, err := f.svc.UpdateCentralizedProject(f.ctx, UpdateCentralizedProjectInput{
		ID:          project.Entity.ID,
		ProjectName: f.prefix + "_project_a_updated",
		City:        "深圳",
		District:    "福田",
		AddressText: "测试地址更新",
		BrandName:   "测试品牌",
	})
	requireNoError(t, err)
	if updatedProject.Entity.ProjectName != f.prefix+"_project_a_updated" {
		t.Fatalf("expected updated project name, got %q", updatedProject.Entity.ProjectName)
	}
	assertChange(t, updatedProject.Changes, HmdChangeUpdated, HmdEntityCentralizedProject, HmdScopeCentralizedProject)

	building := mustCreateBuilding(t, f, project.Entity.ID, f.prefix+"_building_a")
	assertChange(t, building.Changes, HmdChangeCreated, HmdEntityBuilding, HmdScopeBuilding)

	_, err = f.svc.CreateBuilding(f.ctx, CreateBuildingInput{
		ProjectID:    project.Entity.ID,
		BuildingName: f.prefix + "_building_dup",
		BuildingCode: building.Entity.BuildingCode,
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	_, err = f.svc.CreateBuilding(f.ctx, CreateBuildingInput{
		ProjectID:    bson.NewObjectID(),
		BuildingName: f.prefix + "_building_missing_project",
	})
	assertErrCode(t, err, errcode.NotFound.Code)

	buildings, err := f.svc.ListBuildingsByProject(f.ctx, project.Entity.ID)
	requireNoError(t, err)
	assertContainsBuilding(t, buildings, building.Entity.ID)

	_, err = f.svc.UpdateBuilding(f.ctx, UpdateBuildingInput{
		ID:           building.Entity.ID,
		BuildingName: "",
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	updatedBuilding, err := f.svc.UpdateBuilding(f.ctx, UpdateBuildingInput{
		ID:                building.Entity.ID,
		BuildingName:      f.prefix + "_building_a_updated",
		FloorTotal:        18,
		ManagerName:       "测试管家",
		ManagerPhone:      "18800000000",
		ListingFacilities: []string{string(hmdmodel.ListingFacilityElevator)},
	})
	requireNoError(t, err)
	if updatedBuilding.Entity.FloorTotal != 18 {
		t.Fatalf("expected floor total 18, got %d", updatedBuilding.Entity.FloorTotal)
	}

	roomType := mustCreateRoomType(t, f, project.Entity.ID, building.Entity.ID, f.prefix+"_room_type_a")
	assertChange(t, roomType.Changes, HmdChangeCreated, HmdEntityRoomTypeCentralized, HmdScopeRoomTypeCentralized)

	_, err = f.svc.CreateRoomType(f.ctx, CreateRoomTypeInput{
		ProjectID:    project.Entity.ID,
		BuildingID:   building.Entity.ID,
		RoomTypeName: roomType.Entity.RoomTypeName,
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	otherProject := mustCreateCentralizedProject(t, f, f.prefix+"_project_b")
	otherBuilding := mustCreateBuilding(t, f, otherProject.Entity.ID, f.prefix+"_building_b")
	otherRoomType := mustCreateRoomType(t, f, otherProject.Entity.ID, otherBuilding.Entity.ID, f.prefix+"_room_type_b")

	_, err = f.svc.CreateRoomType(f.ctx, CreateRoomTypeInput{
		ProjectID:    project.Entity.ID,
		BuildingID:   otherBuilding.Entity.ID,
		RoomTypeName: f.prefix + "_room_type_bad_owner",
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	roomTypes, err := f.svc.ListRoomTypesByBuilding(f.ctx, building.Entity.ID)
	requireNoError(t, err)
	assertContainsRoomType(t, roomTypes, roomType.Entity.ID)

	centralizedRoom := mustCreateCentralizedRoom(t, f, project.Entity.ID, building.Entity.ID, roomType.Entity.ID, f.prefix+"_room_1201")
	assertChange(t, centralizedRoom.Changes, HmdChangeCreated, HmdEntityRoomCentralized, HmdScopeCentralizedRoom)
	if centralizedRoom.Entity.RoomStatus != hmdmodel.RoomStatusAvailable {
		t.Fatalf("expected default available room status, got %d", centralizedRoom.Entity.RoomStatus)
	}

	_, err = f.svc.CreateCentralizedRoom(f.ctx, CreateCentralizedRoomInput{
		ProjectID:  project.Entity.ID,
		BuildingID: building.Entity.ID,
		RoomTypeID: roomType.Entity.ID,
		RoomNo:     centralizedRoom.Entity.RoomNo,
		RentMode:   string(hmdmodel.RentModeWhole),
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	_, err = f.svc.CreateCentralizedRoom(f.ctx, CreateCentralizedRoomInput{
		ProjectID:  project.Entity.ID,
		BuildingID: building.Entity.ID,
		RoomTypeID: otherRoomType.Entity.ID,
		RoomNo:     f.prefix + "_room_bad_type",
		RentMode:   string(hmdmodel.RentModeWhole),
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	_, err = f.svc.UpdateCentralizedRoom(f.ctx, UpdateCentralizedRoomInput{
		ID:       centralizedRoom.Entity.ID,
		RoomNo:   centralizedRoom.Entity.RoomNo,
		RentMode: "daily",
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	updatedRoom, err := f.svc.UpdateCentralizedRoomStatus(f.ctx, UpdateCentralizedRoomStatusInput{
		ID:         centralizedRoom.Entity.ID,
		RoomStatus: int(hmdmodel.RoomStatusRented),
	})
	requireNoError(t, err)
	if updatedRoom.Entity.RoomStatus != hmdmodel.RoomStatusRented {
		t.Fatalf("expected rented room status, got %d", updatedRoom.Entity.RoomStatus)
	}

	_, err = f.svc.UpdateCentralizedRoomStatus(f.ctx, UpdateCentralizedRoomStatusInput{
		ID:         centralizedRoom.Entity.ID,
		RoomStatus: 99,
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)
}

func TestServiceIntegrationDecentralizedFlow(t *testing.T) {
	f := newIntegrationFixture(t)

	_, err := f.svc.CreateDecentralizedCommunity(f.ctx, CreateDecentralizedCommunityInput{
		City: "深圳",
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	community := mustCreateDecentralizedCommunity(t, f, f.prefix+"_community_a", "")
	assertChange(t, community.Changes, HmdChangeCreated, HmdEntityDecentralizedCommunity, HmdScopeDecentralizedCommunity)
	if community.Entity.District != "" {
		t.Fatalf("expected empty district to be allowed, got %q", community.Entity.District)
	}

	_, err = f.svc.CreateDecentralizedCommunity(f.ctx, CreateDecentralizedCommunityInput{
		CommunityName: community.Entity.CommunityName,
		City:          community.Entity.City,
		District:      community.Entity.District,
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	communities, err := f.svc.ListDecentralizedCommunities(f.ctx, ListDecentralizedCommunitiesInput{City: "深圳"})
	requireNoError(t, err)
	assertContainsDecentralizedCommunity(t, communities, community.Entity.ID)

	updatedCommunity, err := f.svc.UpdateDecentralizedCommunity(f.ctx, UpdateDecentralizedCommunityInput{
		ID:            community.Entity.ID,
		CommunityName: f.prefix + "_community_a_updated",
		City:          "深圳",
		District:      "",
		BizArea:       "科技园",
		AddressText:   "测试分散式地址",
		SubwayStation: "高新园",
	})
	requireNoError(t, err)
	if updatedCommunity.Entity.CommunityName != f.prefix+"_community_a_updated" {
		t.Fatalf("expected updated community name, got %q", updatedCommunity.Entity.CommunityName)
	}

	_, err = f.svc.CreateDecentralizedRoom(f.ctx, CreateDecentralizedRoomInput{
		DecentralizedID: bson.NewObjectID(),
		RoomNo:          f.prefix + "_de_room_missing_community",
		RentMode:        string(hmdmodel.RentModeWhole),
	})
	assertErrCode(t, err, errcode.NotFound.Code)

	room := mustCreateDecentralizedRoom(t, f, community.Entity.ID, f.prefix+"_de_room_1")
	assertChange(t, room.Changes, HmdChangeCreated, HmdEntityRoomDecentralized, HmdScopeDecentralizedRoom)
	assertDecentralizedRoomHasNoRoomTypeID(t, f, room.Entity.ID)

	_, err = f.svc.CreateDecentralizedRoom(f.ctx, CreateDecentralizedRoomInput{
		DecentralizedID: community.Entity.ID,
		RoomNo:          room.Entity.RoomNo,
		RentMode:        string(hmdmodel.RentModeWhole),
	})
	assertErrCode(t, err, errcode.AlreadyExists.Code)

	_, err = f.svc.UpdateDecentralizedRoom(f.ctx, UpdateDecentralizedRoomInput{
		ID:       room.Entity.ID,
		RoomNo:   room.Entity.RoomNo,
		RentMode: string(hmdmodel.RentModeWhole),
		Rent:     -1,
	})
	assertErrCode(t, err, errcode.InvalidParam.Code)

	updatedRoom, err := f.svc.UpdateDecentralizedRoomStatus(f.ctx, UpdateDecentralizedRoomStatusInput{
		ID:         room.Entity.ID,
		RoomStatus: int(hmdmodel.RoomStatusOffline),
	})
	requireNoError(t, err)
	if updatedRoom.Entity.RoomStatus != hmdmodel.RoomStatusOffline {
		t.Fatalf("expected offline room status, got %d", updatedRoom.Entity.RoomStatus)
	}
}

func newIntegrationFixture(t *testing.T) *integrationFixture {
	t.Helper()
	if os.Getenv("PUBLISH_HMD_INTEGRATION") != "1" {
		t.Skip("set PUBLISH_HMD_INTEGRATION=1 to run Mongo-backed HMD service integration tests")
	}

	root := repoRoot(t)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cfg := testMongoConfig(t, root)
	if strings.EqualFold(cfg.Database, "rent-house") && os.Getenv("PUBLISH_HMD_TEST_ALLOW_RENT_HOUSE") != "1" {
		cancel()
		t.Fatalf("refusing to run integration tests against %q; use house_manager_test or set PUBLISH_HMD_TEST_ALLOW_RENT_HOUSE=1 explicitly", cfg.Database)
	}

	client, err := dbmongo.NewClient(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("connect integration mongo: %v", err)
	}

	prefix := fmt.Sprintf("itest_%d", time.Now().UnixNano())
	fixture := &integrationFixture{
		ctx:    ctx,
		cancel: cancel,
		client: client,
		svc: NewService(
			repohmd.NewCentralizedRepository(client),
			repohmd.NewBuildingRepository(client),
			repohmd.NewDecentralizedRepository(client),
			repohmd.NewRoomTypeCentralizedRepository(client),
			repohmd.NewRoomCentralizedRepository(client),
			repohmd.NewRoomDecentralizedRepository(client),
		),
		prefix: prefix,
	}

	cleanupIntegrationData(t, fixture)
	t.Cleanup(func() {
		cleanupIntegrationData(t, fixture)
		if err := client.Close(context.Background()); err != nil {
			t.Logf("close integration mongo: %v", err)
		}
		cancel()
	})
	return fixture
}

func testMongoConfig(t *testing.T, root string) dbmongo.Config {
	t.Helper()
	if uri := strings.TrimSpace(os.Getenv("PUBLISH_HMD_TEST_URI")); uri != "" {
		return dbmongo.Config{
			URI:                    uri,
			Database:               envOr("PUBLISH_HMD_TEST_DB", "house_manager_test"),
			ConnectTimeout:         10 * time.Second,
			SocketTimeout:          30 * time.Second,
			ServerSelectionTimeout: 10 * time.Second,
			EnableRetryReads:       true,
			EnableRetryWrites:      true,
		}
	}

	cfg, err := config.Load(filepath.Join(root, "config", "config.test.yaml"))
	if err != nil {
		t.Fatalf("load config.test.yaml: %v", err)
	}
	if value := strings.TrimSpace(os.Getenv("PUBLISH_HMD_TEST_ADDRS")); value != "" {
		cfg.MongoDB.Addrs = splitCSV(value)
	}
	if value := strings.TrimSpace(os.Getenv("PUBLISH_HMD_TEST_DB")); value != "" {
		cfg.MongoDB.Database = value
	}
	if value := strings.TrimSpace(os.Getenv("PUBLISH_HMD_TEST_AUTH_SOURCE")); value != "" {
		cfg.MongoDB.AuthSource = value
	}
	if value := strings.TrimSpace(os.Getenv("PUBLISH_HMD_TEST_USERNAME")); value != "" {
		cfg.MongoDB.Username = value
	}
	if value := os.Getenv("PUBLISH_HMD_TEST_PASSWORD"); value != "" {
		cfg.MongoDB.Password = value
	}

	connectTimeout, err := cfg.MongoDB.ConnectTimeoutDuration()
	if err != nil {
		t.Fatalf("parse mongo connect timeout: %v", err)
	}
	socketTimeout, err := cfg.MongoDB.SocketTimeoutDuration()
	if err != nil {
		t.Fatalf("parse mongo socket timeout: %v", err)
	}
	serverSelectionTimeout, err := cfg.MongoDB.ServerSelectionTimeoutDuration()
	if err != nil {
		t.Fatalf("parse mongo server selection timeout: %v", err)
	}

	return dbmongo.Config{
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
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}

func cleanupIntegrationData(t *testing.T, f *integrationFixture) {
	t.Helper()
	pattern := "^" + regexp.QuoteMeta(f.prefix)
	cases := []struct {
		collection string
		filter     bson.M
	}{
		{hmdmodel.CollectionHmdRoomCentralized, bson.M{"room_no": bson.M{"$regex": pattern}}},
		{hmdmodel.CollectionHmdRoomDecentralized, bson.M{"room_no": bson.M{"$regex": pattern}}},
		{hmdmodel.CollectionHmdRoomTypeCentralized, bson.M{"room_type_name": bson.M{"$regex": pattern}}},
		{hmdmodel.CollectionHmdBuilding, bson.M{"building_code": bson.M{"$regex": pattern}}},
		{hmdmodel.CollectionHmdDecentralized, bson.M{"community_name": bson.M{"$regex": pattern}}},
		{hmdmodel.CollectionHmdCentralized, bson.M{"project_code": bson.M{"$regex": pattern}}},
	}
	for _, tc := range cases {
		if _, err := f.client.Collection(tc.collection).DeleteMany(f.ctx, tc.filter); err != nil {
			t.Fatalf("cleanup %s: %v", tc.collection, err)
		}
	}
}

func mustCreateCentralizedProject(t *testing.T, f *integrationFixture, code string) *HmdMutationResult[hmdmodel.HmdCentralized] {
	t.Helper()
	result, err := f.svc.CreateCentralizedProject(f.ctx, CreateCentralizedProjectInput{
		ProjectName: code,
		ProjectCode: code,
		City:        "深圳",
		District:    "南山",
		AddressText: "测试集中式地址",
		BrandName:   "测试品牌",
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created centralized project entity with id, got %#v", result)
	}
	return result
}

func mustCreateBuilding(t *testing.T, f *integrationFixture, projectID bson.ObjectID, code string) *HmdMutationResult[hmdmodel.HmdBuilding] {
	t.Helper()
	result, err := f.svc.CreateBuilding(f.ctx, CreateBuildingInput{
		ProjectID:         projectID,
		BuildingName:      code,
		BuildingCode:      code,
		FloorTotal:        12,
		ManagerName:       "测试管家",
		ManagerPhone:      "18800000000",
		ListingFacilities: []string{string(hmdmodel.ListingFacilityElevator)},
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created building entity with id, got %#v", result)
	}
	return result
}

func mustCreateRoomType(t *testing.T, f *integrationFixture, projectID, buildingID bson.ObjectID, name string) *HmdMutationResult[hmdmodel.HmdRoomTypeCentralized] {
	t.Helper()
	result, err := f.svc.CreateRoomType(f.ctx, CreateRoomTypeInput{
		ProjectID:       projectID,
		BuildingID:      buildingID,
		RoomTypeName:    name,
		RoomCount:       1,
		HallCount:       1,
		BathroomCount:   1,
		AreaSize:        35,
		Orientation:     string(hmdmodel.OrientationSouth),
		DecorationLevel: string(hmdmodel.DecorationLevelFine),
		PaymentCycle:    string(hmdmodel.PaymentCycleMonthly),
		Rent:            5800,
		Deposit:         5800,
		AgencyFeeMode:   string(hmdmodel.AgencyFeeModeNone),
		Images:          []TaggedImageInput{{URL: "https://example.com/room-type.jpg", Tag: string(hmdmodel.ImageTagBedroom)}},
		RoomFacilities:  []string{string(hmdmodel.RoomFacilityBed)},
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created room type entity with id, got %#v", result)
	}
	return result
}

func mustCreateCentralizedRoom(t *testing.T, f *integrationFixture, projectID, buildingID, roomTypeID bson.ObjectID, roomNo string) *HmdMutationResult[hmdmodel.HmdRoomCentralized] {
	t.Helper()
	result, err := f.svc.CreateCentralizedRoom(f.ctx, CreateCentralizedRoomInput{
		ProjectID:         projectID,
		BuildingID:        buildingID,
		RoomTypeID:        roomTypeID,
		RoomNo:            roomNo,
		FloorNo:           12,
		RentMode:          string(hmdmodel.RentModeWhole),
		LayoutText:        "一室一厅",
		AreaSize:          35,
		Orientation:       string(hmdmodel.OrientationSouth),
		DecorationLevel:   string(hmdmodel.DecorationLevelFine),
		PaymentCycle:      string(hmdmodel.PaymentCycleMonthly),
		Rent:              5800,
		Deposit:           5800,
		AgencyFeeMode:     string(hmdmodel.AgencyFeeModeNone),
		ViewingTimeRule:   string(hmdmodel.ViewingTimeRuleAnytime),
		StartRentRule:     string(hmdmodel.StartRentRuleLongOneYear),
		Images:            []TaggedImageInput{{URL: "https://example.com/room.jpg", Tag: string(hmdmodel.ImageTagBedroom)}},
		RoomFacilities:    []string{string(hmdmodel.RoomFacilityBed)},
		ListingFacilities: []string{string(hmdmodel.ListingFacilityElevator)},
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created centralized room entity with id, got %#v", result)
	}
	return result
}

func mustCreateDecentralizedCommunity(t *testing.T, f *integrationFixture, name, district string) *HmdMutationResult[hmdmodel.HmdDecentralized] {
	t.Helper()
	result, err := f.svc.CreateDecentralizedCommunity(f.ctx, CreateDecentralizedCommunityInput{
		CommunityName: name,
		City:          "深圳",
		District:      district,
		BizArea:       "科技园",
		AddressText:   "测试分散式地址",
		SubwayStation: "高新园",
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created decentralized community entity with id, got %#v", result)
	}
	return result
}

func mustCreateDecentralizedRoom(t *testing.T, f *integrationFixture, communityID bson.ObjectID, roomNo string) *HmdMutationResult[hmdmodel.HmdRoomDecentralized] {
	t.Helper()
	result, err := f.svc.CreateDecentralizedRoom(f.ctx, CreateDecentralizedRoomInput{
		DecentralizedID:   communityID,
		RoomNo:            roomNo,
		FloorNo:           8,
		RentMode:          string(hmdmodel.RentModeWhole),
		LayoutText:        "两室一厅",
		AreaSize:          72,
		Orientation:       string(hmdmodel.OrientationSouth),
		DecorationLevel:   string(hmdmodel.DecorationLevelFine),
		PaymentCycle:      string(hmdmodel.PaymentCycleMonthly),
		Rent:              7600,
		Deposit:           7600,
		AgencyFeeMode:     string(hmdmodel.AgencyFeeModeNone),
		ViewingTimeRule:   string(hmdmodel.ViewingTimeRuleAnytime),
		StartRentRule:     string(hmdmodel.StartRentRuleLongOneYear),
		Images:            []TaggedImageInput{{URL: "https://example.com/de-room.jpg", Tag: string(hmdmodel.ImageTagBedroom)}},
		RoomFacilities:    []string{string(hmdmodel.RoomFacilityBed)},
		ListingFacilities: []string{string(hmdmodel.ListingFacilitySubway)},
	})
	requireNoError(t, err)
	if result == nil || result.Entity == nil || result.Entity.ID.IsZero() {
		t.Fatalf("expected created decentralized room entity with id, got %#v", result)
	}
	return result
}

func assertDecentralizedRoomHasNoRoomTypeID(t *testing.T, f *integrationFixture, roomID bson.ObjectID) {
	t.Helper()
	var doc bson.M
	err := f.client.Collection(hmdmodel.CollectionHmdRoomDecentralized).FindOne(f.ctx, bson.M{"_id": roomID}).Decode(&doc)
	requireNoError(t, err)
	if _, ok := doc["room_type_id"]; ok {
		t.Fatalf("expected decentralized room document not to contain room_type_id, got %#v", doc)
	}
}

func assertChange(t *testing.T, changes []HmdChange, action HmdChangeAction, entityType HmdEntityType, scope HmdProjectionScope) {
	t.Helper()
	if len(changes) != 1 {
		t.Fatalf("expected exactly one change, got %d", len(changes))
	}
	change := changes[0]
	if change.Action != action || change.EntityType != entityType || change.Scope != scope || change.EntityID.IsZero() {
		t.Fatalf("unexpected change: %#v", change)
	}
}

func assertContainsCentralizedProject(t *testing.T, items []hmdmodel.HmdCentralized, id bson.ObjectID) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	t.Fatalf("expected centralized project %s in list", id.Hex())
}

func assertContainsBuilding(t *testing.T, items []hmdmodel.HmdBuilding, id bson.ObjectID) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	t.Fatalf("expected building %s in list", id.Hex())
}

func assertContainsRoomType(t *testing.T, items []hmdmodel.HmdRoomTypeCentralized, id bson.ObjectID) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	t.Fatalf("expected room type %s in list", id.Hex())
}

func assertContainsDecentralizedCommunity(t *testing.T, items []hmdmodel.HmdDecentralized, id bson.ObjectID) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	t.Fatalf("expected decentralized community %s in list", id.Hex())
}

func assertErrCode(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %d, got nil", code)
	}
	e := errcode.FromError(err)
	if e == nil {
		t.Fatalf("expected errcode %d, got plain error: %v", code, err)
	}
	if e.Code != code {
		t.Fatalf("expected errcode %d, got %d: %v", code, e.Code, err)
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
