package publish

import (
	"context"
	"fmt"
	"testing"

	hmddomain "house-manager/internal/domain/hmd"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPublishServiceAppliesHmdChangesAfterWrite(t *testing.T) {
	entity := &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}}
	change := hmddomain.HmdChange{
		Action:     hmddomain.HmdChangeCreated,
		EntityType: hmddomain.HmdEntityCentralizedProject,
		EntityID:   entity.ID,
		Scope:      hmddomain.HmdScopeCentralizedProject,
		ProjectID:  entity.ID,
	}
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]{
			Entity:  entity,
			Changes: []hmddomain.HmdChange{change},
		},
	}
	projection := &fakeListingProjection{}
	service := newTestPublishService(hmd, projection)

	got, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != entity {
		t.Fatalf("expected entity returned from mutation result")
	}
	if projection.calls != 1 {
		t.Fatalf("expected listing projection Apply to be called once, got %d", projection.calls)
	}
	if len(projection.changes) != 1 || projection.changes[0] != change {
		t.Fatalf("unexpected applied changes: %#v", projection.changes)
	}
}

func TestPublishServiceSkipsApplyWhenHmdWriteFails(t *testing.T) {
	hmd := &fakeHmdService{
		createCentralizedProjectErr: errcode.InvalidParam.WithError(fmt.Errorf("projectName is required")),
	}
	projection := &fakeListingProjection{}
	service := newTestPublishService(hmd, projection)

	_, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if err == nil {
		t.Fatal("expected HMD error")
	}
	if projection.calls != 0 {
		t.Fatalf("expected listing projection Apply not to be called, got %d calls", projection.calls)
	}
}

func TestPublishServiceReadDoesNotApplyHpdChanges(t *testing.T) {
	entity := &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}}
	hmd := &fakeHmdService{getCentralizedProjectResult: entity}
	projection := &fakeListingProjection{}
	service := newTestPublishService(hmd, projection)

	got, err := service.GetCentralizedProject(publishGlobalContext(), entity.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != entity {
		t.Fatalf("expected get entity")
	}
	if projection.calls != 0 {
		t.Fatalf("expected read not to call listing projection Apply, got %d calls", projection.calls)
	}
}

func TestPublishServiceReturnsApplyError(t *testing.T) {
	entity := &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}}
	applyErr := errcode.InternalError.WithError(fmt.Errorf("apply failed"))
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]{
			Entity:  entity,
			Changes: []hmddomain.HmdChange{{EntityID: entity.ID}},
		},
	}
	projection := &fakeListingProjection{err: applyErr}
	service := newTestPublishService(hmd, projection)

	_, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if err != applyErr {
		t.Fatalf("expected apply error, got %v", err)
	}
}

func TestPublishServiceFailsClosedWithoutListingProjection(t *testing.T) {
	entity := &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}}
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]{
			Entity:  entity,
			Changes: []hmddomain.HmdChange{{EntityID: entity.ID}},
		},
	}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{}, nil),
	}

	_, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.InternalError.Code {
		t.Fatalf("expected internal error, got %v", err)
	}
}

func TestPublishServiceCreatesEntrustRelationAfterCentralizedRoomCreate(t *testing.T) {
	roomID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	change := hmddomain.HmdChange{
		Action:     hmddomain.HmdChangeCreated,
		EntityType: hmddomain.HmdEntityRoomCentralized,
		EntityID:   roomID,
		Scope:      hmddomain.HmdScopeCentralizedRoom,
	}
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity:  &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
			Changes: []hmddomain.HmdChange{change},
		},
	}
	projection := &fakeListingProjection{}
	entrust := &fakePublishAccess{listing: &hpdmodel.HpdListing{CommonFields: commonmodel.CommonFields{ID: listingID}}}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, entrust),
	}
	ctx := publishGlobalContextForStaff(staffID)

	got, err := service.CreateCentralizedRoom(ctx, CreateCentralizedRoomInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != roomID {
		t.Fatalf("unexpected room: %#v", got)
	}
	if projection.calls != 1 {
		t.Fatalf("expected listing projection Apply to be called once, got %d", projection.calls)
	}
	if entrust.findSourceType != hpdmodel.HpdSourceTypeCentralizedRoom || entrust.findSourceID != roomID {
		t.Fatalf("unexpected listing source lookup: %s %s", entrust.findSourceType, entrust.findSourceID.Hex())
	}
	if entrust.upsertListingID != listingID || entrust.upsertPrincipal.PrincipalID != staffID.Hex() {
		t.Fatalf("unexpected entrust upsert: listing=%s principal=%#v", entrust.upsertListingID.Hex(), entrust.upsertPrincipal)
	}
}

func TestPublishServiceCreatesEntrustRelationAfterDecentralizedRoomCreate(t *testing.T) {
	roomID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	change := hmddomain.HmdChange{
		Action:     hmddomain.HmdChangeCreated,
		EntityType: hmddomain.HmdEntityRoomDecentralized,
		EntityID:   roomID,
		Scope:      hmddomain.HmdScopeDecentralizedRoom,
	}
	hmd := &fakeHmdService{
		createDecentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]{
			Entity:  &hmdmodel.HmdRoomDecentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
			Changes: []hmddomain.HmdChange{change},
		},
	}
	projection := &fakeListingProjection{}
	entrust := &fakePublishAccess{listing: &hpdmodel.HpdListing{CommonFields: commonmodel.CommonFields{ID: listingID}}}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, entrust),
	}
	ctx := publishGlobalContextForStaff(staffID)

	got, err := service.CreateDecentralizedRoom(ctx, CreateDecentralizedRoomInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != roomID {
		t.Fatalf("unexpected room: %#v", got)
	}
	if projection.calls != 1 {
		t.Fatalf("expected listing projection Apply to be called once, got %d", projection.calls)
	}
	if entrust.findSourceType != hpdmodel.HpdSourceTypeDecentralizedRoom || entrust.findSourceID != roomID {
		t.Fatalf("unexpected listing source lookup: %s %s", entrust.findSourceType, entrust.findSourceID.Hex())
	}
	if entrust.upsertListingID != listingID || entrust.upsertPrincipal.PrincipalID != staffID.Hex() {
		t.Fatalf("unexpected entrust upsert: listing=%s principal=%#v", entrust.upsertListingID.Hex(), entrust.upsertPrincipal)
	}
}

func TestPublishServiceCreateCentralizedRoomRequiresPrincipalBeforeWrite(t *testing.T) {
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}},
		},
	}
	projection := &fakeListingProjection{}
	entrust := &fakePublishAccess{}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, entrust),
	}

	_, err := service.CreateCentralizedRoom(context.Background(), CreateCentralizedRoomInput{})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.Unauthorized.Code {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if hmd.createCentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD not to be called before principal")
	}
}

func TestPublishServiceCreateCentralizedRoomFailsClosedWithoutEntrust(t *testing.T) {
	roomID := bson.NewObjectID()
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: &fakeListingProjection{}}, nil),
	}
	ctx := session.ContextWithPrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   bson.NewObjectID().Hex(),
		Terminal:      session.TerminalPublish,
		RoleCodes:     []string{"super_admin"},
	})

	_, err := service.CreateCentralizedRoom(ctx, CreateCentralizedRoomInput{})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.InternalError.Code {
		t.Fatalf("expected internal error, got %v", err)
	}
}

func TestPublishServiceCreateCentralizedRoomPropagatesEntrustError(t *testing.T) {
	roomID := bson.NewObjectID()
	listingID := bson.NewObjectID()
	entrustErr := errcode.DatabaseError.WithError(fmt.Errorf("upsert entrust failed"))
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(
			hmd,
			mutationPublisher{listingProjection: &fakeListingProjection{}},
			&fakePublishAccess{
				listing:     &hpdmodel.HpdListing{CommonFields: commonmodel.CommonFields{ID: listingID}},
				registerErr: entrustErr,
			},
		),
	}
	ctx := session.ContextWithPrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   bson.NewObjectID().Hex(),
		Terminal:      session.TerminalPublish,
		RoleCodes:     []string{"super_admin"},
	})

	_, err := service.CreateCentralizedRoom(ctx, CreateCentralizedRoomInput{})
	if err != entrustErr {
		t.Fatalf("expected entrust error, got %v", err)
	}
}

func TestPublishServiceFiltersCentralizedProjectsByScope(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	roomB := bson.NewObjectID()
	hmd := &fakeHmdService{
		listCentralizedProjectsResult: []hmdmodel.HmdCentralized{
			{CommonFields: commonmodel.CommonFields{ID: projectA}, City: "杭州", District: "西湖"},
			{CommonFields: commonmodel.CommonFields{ID: projectB}, City: "杭州", District: "西湖"},
		},
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, ProjectID: projectA},
			roomB: {CommonFields: commonmodel.CommonFields{ID: roomB}, ProjectID: projectB},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeCentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedProjects(publishStaffContext(), ListCentralizedProjectsInput{
		City:     "杭州",
		District: "西湖",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != projectA {
		t.Fatalf("expected only scoped project A, got %#v", got)
	}
	if hmd.listCentralizedProjectsInput.City != "杭州" || hmd.listCentralizedProjectsInput.District != "西湖" {
		t.Fatalf("expected business filters to pass through, got %#v", hmd.listCentralizedProjectsInput)
	}
}

func TestPublishServiceGlobalScopeKeepsFullCentralizedProjectList(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	hmd := &fakeHmdService{
		listCentralizedProjectsResult: []hmdmodel.HmdCentralized{
			{CommonFields: commonmodel.CommonFields{ID: projectA}},
			{CommonFields: commonmodel.CommonFields{ID: projectB}},
		},
	}
	access := &fakePublishAccess{}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedProjects(publishGlobalContext(), ListCentralizedProjectsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected global scope to see all projects, got %#v", got)
	}
	if access.listAccessibleCalls != 0 {
		t.Fatalf("global scope should not query entrust relation, got %d calls", access.listAccessibleCalls)
	}
}

func TestPublishServiceFiltersCentralizedRoomsByRelation(t *testing.T) {
	roomA := bson.NewObjectID()
	roomB := bson.NewObjectID()
	hmd := &fakeHmdService{
		listCentralizedRoomsByProjectResult: []hmdmodel.HmdRoomCentralized{
			{CommonFields: commonmodel.CommonFields{ID: roomA}},
			{CommonFields: commonmodel.CommonFields{ID: roomB}},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeCentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedRoomsByProject(publishStaffContext(), bson.NewObjectID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != roomA {
		t.Fatalf("expected only scoped room A, got %#v", got)
	}
}

func TestPublishServiceRejectsCrossStaffCentralizedRoomUpdate(t *testing.T) {
	roomID := bson.NewObjectID()
	hmd := &fakeHmdService{
		updateCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	access := &fakePublishAccess{canAccessSource: false}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.UpdateCentralizedRoom(publishStaffContext(), UpdateCentralizedRoomInput{ID: roomID})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for cross-staff update, got %v", err)
	}
	if hmd.updateCentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD update not to be called, got %d calls", hmd.updateCentralizedRoomCalls)
	}
}

func TestPublishServiceFiltersDecentralizedCommunitiesByScope(t *testing.T) {
	communityA := bson.NewObjectID()
	communityB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	roomB := bson.NewObjectID()
	hmd := &fakeHmdService{
		listDecentralizedCommunitiesResult: []hmdmodel.HmdDecentralized{
			{CommonFields: commonmodel.CommonFields{ID: communityA}, City: "杭州"},
			{CommonFields: commonmodel.CommonFields{ID: communityB}, City: "杭州"},
		},
		decentralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, DecentralizedID: communityA},
			roomB: {CommonFields: commonmodel.CommonFields{ID: roomB}, DecentralizedID: communityB},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeDecentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListDecentralizedCommunities(publishStaffContext(), ListDecentralizedCommunitiesInput{City: "杭州"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != communityA {
		t.Fatalf("expected only scoped community A, got %#v", got)
	}
}

func TestPublishServiceListDecentralizedCommunitiesTreatsCityAsOptionalFilter(t *testing.T) {
	communityA := bson.NewObjectID()
	roomA := bson.NewObjectID()
	hmd := &fakeHmdService{
		listDecentralizedCommunitiesResult: []hmdmodel.HmdDecentralized{
			{CommonFields: commonmodel.CommonFields{ID: communityA}},
		},
		decentralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, DecentralizedID: communityA},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeDecentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListDecentralizedCommunities(publishStaffContext(), ListDecentralizedCommunitiesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != communityA {
		t.Fatalf("expected scoped community without city filter, got %#v", got)
	}
	if hmd.listDecentralizedCommunitiesInput.City != "" || hmd.listDecentralizedCommunitiesInput.District != "" {
		t.Fatalf("expected empty city/district to pass as optional filters, got %#v", hmd.listDecentralizedCommunitiesInput)
	}
}

func TestPublishServiceRejectsBuildingCreateWhenProjectOutOfScope(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	hmd := &fakeHmdService{
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, ProjectID: projectA},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeCentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		buildingService: newBuildingService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateBuilding(publishStaffContext(), CreateBuildingInput{ProjectID: projectB})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for out-of-scope project create, got %v", err)
	}
	if hmd.createBuildingCalls != 0 {
		t.Fatalf("expected HMD create building not to be called, got %d", hmd.createBuildingCalls)
	}
}

func TestPublishServiceRejectsCentralizedRoomCreateWhenBuildingOutOfScope(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	buildingB := bson.NewObjectID()
	hmd := &fakeHmdService{
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, ProjectID: projectA},
		},
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingB: {CommonFields: commonmodel.CommonFields{ID: buildingB}, ProjectID: projectB},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeCentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateCentralizedRoom(publishStaffContext(), CreateCentralizedRoomInput{BuildingID: buildingB})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for out-of-scope building create, got %v", err)
	}
	if hmd.createCentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD create room not to be called, got %d", hmd.createCentralizedRoomCalls)
	}
}

func TestPublishServiceRejectsRoomTypeCreateWhenBuildingOutOfScope(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	buildingB := bson.NewObjectID()
	hmd := &fakeHmdService{
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, ProjectID: projectA},
		},
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingB: {CommonFields: commonmodel.CommonFields{ID: buildingB}, ProjectID: projectB},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeCentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		roomTypeService: newRoomTypeService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateRoomType(publishStaffContext(), CreateRoomTypeInput{BuildingID: buildingB})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for out-of-scope room type create, got %v", err)
	}
	if hmd.createRoomTypeCalls != 0 {
		t.Fatalf("expected HMD create room type not to be called, got %d", hmd.createRoomTypeCalls)
	}
}

func TestPublishServiceRejectsDecentralizedRoomCreateWhenCommunityOutOfScope(t *testing.T) {
	communityA := bson.NewObjectID()
	communityB := bson.NewObjectID()
	roomA := bson.NewObjectID()
	hmd := &fakeHmdService{
		decentralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized{
			roomA: {CommonFields: commonmodel.CommonFields{ID: roomA}, DecentralizedID: communityA},
		},
	}
	access := &fakePublishAccess{
		accessibleListings: []hpdmodel.HpdListing{
			{SourceType: hpdmodel.HpdSourceTypeDecentralizedRoom, SourceID: roomA},
		},
	}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateDecentralizedRoom(publishStaffContext(), CreateDecentralizedRoomInput{DecentralizedID: communityB})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for out-of-scope community create, got %v", err)
	}
	if hmd.createDecentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD create decentralized room not to be called, got %d", hmd.createDecentralizedRoomCalls)
	}
}

func TestPublishServiceRejectsParentUpdatesForNonGlobalScope(t *testing.T) {
	hmd := &fakeHmdService{}
	service := &PublishService{
		centralizedProjectService:     newCentralizedProjectService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		buildingService:               newBuildingService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		roomTypeService:               newRoomTypeService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, &fakePublishAccess{}),
	}
	ctx := publishStaffContext()

	if _, err := service.UpdateCentralizedProject(ctx, UpdateCentralizedProjectInput{ID: bson.NewObjectID()}); errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected project update not found, got %v", err)
	}
	if _, err := service.UpdateBuilding(ctx, UpdateBuildingInput{ID: bson.NewObjectID()}); errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected building update not found, got %v", err)
	}
	if _, err := service.UpdateRoomType(ctx, UpdateRoomTypeInput{ID: bson.NewObjectID()}); errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected room type update not found, got %v", err)
	}
	if _, err := service.UpdateDecentralizedCommunity(ctx, UpdateDecentralizedCommunityInput{ID: bson.NewObjectID()}); errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected community update not found, got %v", err)
	}
	if hmd.updateCentralizedProjectCalls != 0 || hmd.updateBuildingCalls != 0 || hmd.updateRoomTypeCalls != 0 || hmd.updateDecentralizedCommunityCalls != 0 {
		t.Fatalf("expected no parent HMD updates, got project=%d building=%d roomType=%d community=%d",
			hmd.updateCentralizedProjectCalls,
			hmd.updateBuildingCalls,
			hmd.updateRoomTypeCalls,
			hmd.updateDecentralizedCommunityCalls,
		)
	}
}

func TestPublishServiceRejectsCrossStaffDecentralizedRoomUpdate(t *testing.T) {
	roomID := bson.NewObjectID()
	hmd := &fakeHmdService{
		updateDecentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]{
			Entity: &hmdmodel.HmdRoomDecentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	access := &fakePublishAccess{canAccessSource: false}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.UpdateDecentralizedRoom(publishStaffContext(), UpdateDecentralizedRoomInput{ID: roomID})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for cross-staff decentralized update, got %v", err)
	}
	if hmd.updateDecentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD decentralized room update not to be called, got %d", hmd.updateDecentralizedRoomCalls)
	}
}

func newTestPublishService(hmd *fakeHmdService, projection *fakeListingProjection) *PublishService {
	return &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{listingProjection: projection}, nil),
	}
}

func publishStaffContext() context.Context {
	return session.ContextWithPrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   bson.NewObjectID().Hex(),
		Terminal:      session.TerminalPublish,
	})
}

func publishGlobalContext() context.Context {
	return publishGlobalContextForStaff(bson.NewObjectID())
}

func publishGlobalContextForStaff(staffID bson.ObjectID) context.Context {
	return session.ContextWithPrincipal(context.Background(), session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     staffID.Hex(),
		Terminal:        session.TerminalPublish,
		RoleCodes:       []string{"super_admin"},
		PermissionCodes: []string{"house.manage"},
	})
}

type fakeListingProjection struct {
	calls   int
	changes []hmddomain.HmdChange
	err     error
}

func (f *fakeListingProjection) Apply(ctx context.Context, changes []hmddomain.HmdChange) error {
	f.calls++
	f.changes = append([]hmddomain.HmdChange(nil), changes...)
	return f.err
}

type fakePublishAccess struct {
	listing             *hpdmodel.HpdListing
	findSourceType      hpdmodel.HpdSourceType
	findSourceID        bson.ObjectID
	upsertListingID     bson.ObjectID
	upsertPrincipal     session.Principal
	registerErr         error
	accessibleListings  []hpdmodel.HpdListing
	listAccessibleCalls int
	canAccessSource     bool
}

func (f *fakePublishAccess) FindListingBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error) {
	f.findSourceType = sourceType
	f.findSourceID = sourceID
	return f.listing, nil
}

func (f *fakePublishAccess) UpsertEntrustForPrincipal(ctx context.Context, listingID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdEntrustRelation, error) {
	f.upsertListingID = listingID
	f.upsertPrincipal = principal
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return &hpdmodel.HpdEntrustRelation{ListingID: listingID}, nil
}

func (f *fakePublishAccess) ListAccessibleListings(ctx context.Context, principal session.Principal) ([]hpdmodel.HpdListing, error) {
	f.listAccessibleCalls++
	return f.accessibleListings, nil
}

func (f *fakePublishAccess) CanAccessSourceForPrincipal(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID, principal session.Principal) (bool, error) {
	f.findSourceType = sourceType
	f.findSourceID = sourceID
	return f.canAccessSource, nil
}

type fakeHmdService struct {
	createCentralizedProjectResult      *hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]
	createCentralizedProjectErr         error
	updateCentralizedProjectCalls       int
	getCentralizedProjectResult         *hmdmodel.HmdCentralized
	getCentralizedProjectErr            error
	listCentralizedProjectsResult       []hmdmodel.HmdCentralized
	listCentralizedProjectsInput        ListCentralizedProjectsInput
	createBuildingCalls                 int
	updateBuildingCalls                 int
	buildingsByID                       map[bson.ObjectID]*hmdmodel.HmdBuilding
	createRoomTypeCalls                 int
	updateRoomTypeCalls                 int
	centralizedRoomsByID                map[bson.ObjectID]*hmdmodel.HmdRoomCentralized
	listCentralizedRoomsByProjectResult []hmdmodel.HmdRoomCentralized
	createCentralizedRoomResult         *hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]
	createCentralizedRoomErr            error
	createCentralizedRoomCalls          int
	updateCentralizedRoomResult         *hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]
	updateCentralizedRoomCalls          int
	createDecentralizedRoomResult       *hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]
	createDecentralizedRoomErr          error
	createDecentralizedRoomCalls        int
	updateDecentralizedRoomResult       *hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]
	updateDecentralizedRoomCalls        int
	decentralizedRoomsByID              map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized
	listDecentralizedCommunitiesResult  []hmdmodel.HmdDecentralized
	listDecentralizedCommunitiesInput   ListDecentralizedCommunitiesInput
	updateDecentralizedCommunityCalls   int
}

func (f *fakeHmdService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdCentralized], error) {
	return f.createCentralizedProjectResult, f.createCentralizedProjectErr
}

func (f *fakeHmdService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	return f.getCentralizedProjectResult, f.getCentralizedProjectErr
}

func (f *fakeHmdService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	f.listCentralizedProjectsInput = input
	return f.listCentralizedProjectsResult, nil
}

func (f *fakeHmdService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdCentralized], error) {
	f.updateCentralizedProjectCalls++
	return nil, nil
}

func (f *fakeHmdService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdBuilding], error) {
	f.createBuildingCalls++
	return nil, nil
}

func (f *fakeHmdService) GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	if f.buildingsByID != nil {
		return f.buildingsByID[id], nil
	}
	return nil, nil
}

func (f *fakeHmdService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdBuilding], error) {
	f.updateBuildingCalls++
	return nil, nil
}

func (f *fakeHmdService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error) {
	f.createRoomTypeCalls++
	return nil, nil
}

func (f *fakeHmdService) GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error) {
	f.updateRoomTypeCalls++
	return nil, nil
}

func (f *fakeHmdService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	f.createCentralizedRoomCalls++
	return f.createCentralizedRoomResult, f.createCentralizedRoomErr
}

func (f *fakeHmdService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error) {
	if f.centralizedRoomsByID != nil {
		return f.centralizedRoomsByID[id], nil
	}
	return nil, nil
}

func (f *fakeHmdService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	return f.listCentralizedRoomsByProjectResult, nil
}

func (f *fakeHmdService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	f.updateCentralizedRoomCalls++
	return f.updateCentralizedRoomResult, nil
}

func (f *fakeHmdService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	f.createDecentralizedRoomCalls++
	return f.createDecentralizedRoomResult, f.createDecentralizedRoomErr
}

func (f *fakeHmdService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error) {
	if f.decentralizedRoomsByID != nil {
		return f.decentralizedRoomsByID[id], nil
	}
	return nil, nil
}

func (f *fakeHmdService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	f.updateDecentralizedRoomCalls++
	return f.updateDecentralizedRoomResult, nil
}

func (f *fakeHmdService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	f.listDecentralizedCommunitiesInput = input
	return f.listDecentralizedCommunitiesResult, nil
}

func (f *fakeHmdService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized], error) {
	f.updateDecentralizedCommunityCalls++
	return nil, nil
}
