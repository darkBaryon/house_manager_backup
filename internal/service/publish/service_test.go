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
		createCentralizedProjectErr: errcode.InvalidParam.WithError(fmt.Errorf("项目名称不能为空")),
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
	projectID := bson.NewObjectID()
	roomID := bson.NewObjectID()
	entity := &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: projectID}}
	hmd := &fakeHmdService{
		getCentralizedProjectResult: entity,
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomID: {CommonFields: commonmodel.CommonFields{ID: roomID}, ProjectID: projectID},
		},
	}
	projection := &fakeListingProjection{}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{listingProjection: projection}, &fakePublishAccess{
			accessibleProjectIDs: []bson.ObjectID{projectID},
		}),
	}

	got, err := service.GetCentralizedProject(publishLandlordContext("13800000000"), entity.ID)
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
	applyErr := errcode.InternalError.WithError(fmt.Errorf("投影刷新失败"))
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

func TestPublishServiceRollsBackCentralizedProjectWhenApplyFails(t *testing.T) {
	projectID := bson.NewObjectID()
	applyErr := errcode.InternalError.WithError(fmt.Errorf("投影刷新失败"))
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]{
			Entity:  &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: projectID}},
			Changes: []hmddomain.HmdChange{{EntityID: projectID}},
		},
	}
	projection := &fakeListingProjection{err: applyErr}
	service := newTestPublishService(hmd, projection)

	_, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if err != applyErr {
		t.Fatalf("expected apply error, got %v", err)
	}
	if hmd.rollbackCentralizedProjectCalls != 1 {
		t.Fatalf("expected centralized project rollback once, got %d", hmd.rollbackCentralizedProjectCalls)
	}
	if hmd.rollbackCentralizedProjectID != projectID {
		t.Fatalf("expected rollback project id %s, got %s", projectID.Hex(), hmd.rollbackCentralizedProjectID.Hex())
	}
}

func TestPublishServiceRollsBackCentralizedProjectWhenRootScopeFails(t *testing.T) {
	projectID := bson.NewObjectID()
	rootScopeErr := errcode.InternalError.WithError(fmt.Errorf("归属关系写入失败"))
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]{
			Entity:  &hmdmodel.HmdCentralized{CommonFields: commonmodel.CommonFields{ID: projectID}},
			Changes: []hmddomain.HmdChange{{EntityID: projectID}},
		},
	}
	projection := &fakeListingProjection{}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{listingProjection: projection}, &fakePublishAccess{
			rootScopeErr: rootScopeErr,
		}),
	}

	_, err := service.CreateCentralizedProject(publishGlobalContext(), CreateCentralizedProjectInput{})
	if err != rootScopeErr {
		t.Fatalf("expected root scope error, got %v", err)
	}
	if hmd.rollbackCentralizedProjectCalls != 1 {
		t.Fatalf("expected centralized project rollback once, got %d", hmd.rollbackCentralizedProjectCalls)
	}
	if hmd.rollbackCentralizedProjectID != projectID {
		t.Fatalf("expected rollback project id %s, got %s", projectID.Hex(), hmd.rollbackCentralizedProjectID.Hex())
	}
}

func TestPublishServiceRollsBackDecentralizedCommunityWhenRootScopeFails(t *testing.T) {
	communityID := bson.NewObjectID()
	rootScopeErr := errcode.InternalError.WithError(fmt.Errorf("归属关系写入失败"))
	hmd := &fakeHmdService{
		createDecentralizedCommunityResult: &hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized]{
			Entity:  &hmdmodel.HmdDecentralized{CommonFields: commonmodel.CommonFields{ID: communityID}},
			Changes: []hmddomain.HmdChange{{EntityID: communityID}},
		},
	}
	projection := &fakeListingProjection{}
	service := &PublishService{
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{listingProjection: projection}, &fakePublishAccess{
			rootScopeErr: rootScopeErr,
		}),
	}

	_, err := service.CreateDecentralizedCommunity(publishGlobalContext(), CreateDecentralizedCommunityInput{})
	if err != rootScopeErr {
		t.Fatalf("expected root scope error, got %v", err)
	}
	if hmd.rollbackDecentralizedCommunityCalls != 1 {
		t.Fatalf("expected decentralized community rollback once, got %d", hmd.rollbackDecentralizedCommunityCalls)
	}
	if hmd.rollbackDecentralizedCommunityID != communityID {
		t.Fatalf("expected rollback community id %s, got %s", communityID.Hex(), hmd.rollbackDecentralizedCommunityID.Hex())
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

func TestPublishServiceCreatesCentralizedRoomWithinAccessibleProject(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	accessibleRoomID := bson.NewObjectID()
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
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingID: {CommonFields: commonmodel.CommonFields{ID: buildingID}, ProjectID: projectID},
		},
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			accessibleRoomID: {CommonFields: commonmodel.CommonFields{ID: accessibleRoomID}, ProjectID: projectID},
		},
	}
	projection := &fakeListingProjection{}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectID},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, access),
	}

	got, err := service.CreateCentralizedRoom(publishLandlordContext("13800000000"), CreateCentralizedRoomInput{BuildingID: buildingID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != roomID {
		t.Fatalf("unexpected room: %#v", got)
	}
	if projection.calls != 1 {
		t.Fatalf("expected listing projection Apply to be called once, got %d", projection.calls)
	}
}

func TestPublishServiceCreateCentralizedRoomInjectsRoomTypeTemplate(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	roomTypeID := bson.NewObjectID()
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity:  &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
			Changes: []hmddomain.HmdChange{{EntityID: roomID}},
		},
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingID: {CommonFields: commonmodel.CommonFields{ID: buildingID}, ProjectID: projectID},
		},
		roomTypesByID: map[bson.ObjectID]*hmdmodel.HmdRoomTypeCentralized{
			roomTypeID: {
				CommonFields:    commonmodel.CommonFields{ID: roomTypeID},
				ProjectID:       projectID,
				BuildingID:      buildingID,
				RoomCount:       2,
				HallCount:       1,
				BathroomCount:   1,
				AreaSize:        48,
				Orientation:     hmdmodel.OrientationSouth,
				DecorationLevel: hmdmodel.DecorationLevelFine,
				PaymentCycle:    hmdmodel.PaymentCycleMonthly,
				Rent:            4200,
				Deposit:         4200,
				ServiceFee:      200,
				AgencyFeeMode:   hmdmodel.AgencyFeeModeFixed,
				AgencyFeeValue:  500,
				Images: []hmdmodel.TaggedImage{
					{URL: "https://example.com/a.jpg", Tag: hmdmodel.ImageTagExterior},
				},
				RoomFacilities: []hmdmodel.RoomFacility{hmdmodel.RoomFacilityBed},
			},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: &fakeListingProjection{}}, &fakePublishAccess{
			accessibleProjectIDs: []bson.ObjectID{projectID},
		}),
	}

	_, err := service.CreateCentralizedRoom(publishLandlordContext("13800000000"), CreateCentralizedRoomInput{
		ProjectID:  projectID,
		BuildingID: buildingID,
		RoomTypeID: roomTypeID,
		RoomNo:     "1208",
		RentMode:   string(hmdmodel.RentModeWhole),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := hmd.createCentralizedRoomInput
	if got.LayoutText != "2室1厅1卫" {
		t.Fatalf("expected layout text from room type, got %q", got.LayoutText)
	}
	if got.AreaSize != 48 || got.Rent != 4200 || got.Deposit != 4200 || got.ServiceFee != 200 || got.AgencyFeeValue != 500 {
		t.Fatalf("expected price fields from room type, got %#v", got)
	}
	if got.Orientation != string(hmdmodel.OrientationSouth) || got.DecorationLevel != string(hmdmodel.DecorationLevelFine) || got.PaymentCycle != string(hmdmodel.PaymentCycleMonthly) || got.AgencyFeeMode != string(hmdmodel.AgencyFeeModeFixed) {
		t.Fatalf("expected enum fields from room type, got %#v", got)
	}
	if len(got.Images) != 1 || got.Images[0].URL != "https://example.com/a.jpg" {
		t.Fatalf("expected images from room type, got %#v", got.Images)
	}
	if len(got.RoomFacilities) != 1 || got.RoomFacilities[0] != string(hmdmodel.RoomFacilityBed) {
		t.Fatalf("expected room facilities from room type, got %#v", got.RoomFacilities)
	}
}

func TestPublishServiceUpdateCentralizedRoomPersistsRoomTypeBinding(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()
	roomTypeID := bson.NewObjectID()
	hmd := &fakeHmdService{
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomID: {
				CommonFields: commonmodel.CommonFields{ID: roomID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "1208",
				RentMode:     hmdmodel.RentModeWhole,
			},
		},
		roomTypesByID: map[bson.ObjectID]*hmdmodel.HmdRoomTypeCentralized{
			roomTypeID: {
				CommonFields: commonmodel.CommonFields{ID: roomTypeID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
			},
		},
		updateCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity:  &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}, ProjectID: projectID, BuildingID: buildingID, RoomTypeID: roomTypeID},
			Changes: []hmddomain.HmdChange{{EntityID: roomID}},
		},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: &fakeListingProjection{}}, &fakePublishAccess{
			accessibleProjectIDs: []bson.ObjectID{projectID},
		}),
	}

	_, err := service.UpdateCentralizedRoom(publishLandlordContext("13800000000"), UpdateCentralizedRoomInput{
		ID:         roomID,
		RoomTypeID: &roomTypeID,
		RoomNo:     "1208",
		RentMode:   string(hmdmodel.RentModeWhole),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hmd.updateCentralizedRoomCalls != 1 {
		t.Fatalf("expected update centralized room once, got %d", hmd.updateCentralizedRoomCalls)
	}
	if hmd.updateCentralizedRoomInput.RoomTypeID == nil || *hmd.updateCentralizedRoomInput.RoomTypeID != roomTypeID {
		t.Fatalf("expected room_type_id to be forwarded, got %#v", hmd.updateCentralizedRoomInput.RoomTypeID)
	}
}

func TestPublishServiceCreatesDecentralizedRoomWithinAccessibleCommunity(t *testing.T) {
	roomID := bson.NewObjectID()
	communityID := bson.NewObjectID()
	accessibleRoomID := bson.NewObjectID()
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
		decentralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized{
			accessibleRoomID: {CommonFields: commonmodel.CommonFields{ID: accessibleRoomID}, DecentralizedID: communityID},
		},
	}
	projection := &fakeListingProjection{}
	access := &fakePublishAccess{
		accessibleCommunityIDs: []bson.ObjectID{communityID},
	}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, access),
	}

	got, err := service.CreateDecentralizedRoom(publishLandlordContext("13800000000"), CreateDecentralizedRoomInput{DecentralizedID: communityID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != roomID {
		t.Fatalf("unexpected room: %#v", got)
	}
	if projection.calls != 1 {
		t.Fatalf("expected listing projection Apply to be called once, got %d", projection.calls)
	}
}

func TestPublishServiceCreateCentralizedRoomRequiresPrincipalBeforeWrite(t *testing.T) {
	hmd := &fakeHmdService{
		createCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}},
		},
	}
	projection := &fakeListingProjection{}
	access := &fakePublishAccess{}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{listingProjection: projection}, access),
	}

	_, err := service.CreateCentralizedRoom(context.Background(), CreateCentralizedRoomInput{})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.Unauthorized.Code {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if hmd.createCentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD not to be called before principal")
	}
}

func TestPublishServiceFiltersCentralizedProjectsByScope(t *testing.T) {
	projectA := bson.NewObjectID()
	hmd := &fakeHmdService{
		listCentralizedProjectsByIDsResult: []hmdmodel.HmdCentralized{
			{CommonFields: commonmodel.CommonFields{ID: projectA}, City: "杭州", District: "西湖"},
		},
	}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectA},
	}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedProjects(publishLandlordContext("13800000000"), ListCentralizedProjectsInput{
		City:     "杭州",
		District: "西湖",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != projectA {
		t.Fatalf("expected only scoped project A, got %#v", got)
	}
	if hmd.listCentralizedProjectsByIDsInput.City != "杭州" || hmd.listCentralizedProjectsByIDsInput.District != "西湖" {
		t.Fatalf("expected business filters to pass through, got %#v", hmd.listCentralizedProjectsByIDsInput)
	}
	if len(hmd.listCentralizedProjectsByIDsIDs) != 1 || hmd.listCentralizedProjectsByIDsIDs[0] != projectA {
		t.Fatalf("expected scoped ids to pass through, got %#v", hmd.listCentralizedProjectsByIDsIDs)
	}
}

func TestPublishServiceWithoutAccessibleProjectsReturnsEmptyCentralizedProjectList(t *testing.T) {
	hmd := &fakeHmdService{}
	access := &fakePublishAccess{}
	service := &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedProjects(publishLandlordContext("13800000000"), ListCentralizedProjectsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected landlord scope without listings to see no projects, got %#v", got)
	}
	if access.listAccessibleProjectCalls != 1 {
		t.Fatalf("expected landlord scope to query root scope once, got %d calls", access.listAccessibleProjectCalls)
	}
	if len(hmd.listCentralizedProjectsByIDsIDs) != 0 {
		t.Fatalf("expected no scoped ids query, got %#v", hmd.listCentralizedProjectsByIDsIDs)
	}
}

func TestPublishServiceListsAllCentralizedRoomsWithinAccessibleProject(t *testing.T) {
	roomA := bson.NewObjectID()
	roomB := bson.NewObjectID()
	projectID := bson.NewObjectID()
	hmd := &fakeHmdService{
		listCentralizedRoomsByProjectResult: []hmdmodel.HmdRoomCentralized{
			{CommonFields: commonmodel.CommonFields{ID: roomA}, ProjectID: projectID},
			{CommonFields: commonmodel.CommonFields{ID: roomB}, ProjectID: projectID},
		},
	}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectID},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListCentralizedRoomsByProject(publishLandlordContext("13800000000"), projectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected all project rooms, got %#v", got)
	}
}

func TestPublishServiceRejectsCrossOwnerCentralizedRoomUpdate(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	hmd := &fakeHmdService{
		centralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomID: {CommonFields: commonmodel.CommonFields{ID: roomID}, ProjectID: projectID},
		},
		updateCentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]{
			Entity: &hmdmodel.HmdRoomCentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	access := &fakePublishAccess{}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.UpdateCentralizedRoom(publishLandlordContext("13800000000"), UpdateCentralizedRoomInput{ID: roomID})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for cross-owner update, got %v", err)
	}
	if hmd.updateCentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD update not to be called, got %d calls", hmd.updateCentralizedRoomCalls)
	}
}

func TestPublishServiceFiltersDecentralizedCommunitiesByScope(t *testing.T) {
	communityA := bson.NewObjectID()
	hmd := &fakeHmdService{
		listDecentralizedCommunitiesByIDsResult: []hmdmodel.HmdDecentralized{
			{CommonFields: commonmodel.CommonFields{ID: communityA}, City: "杭州"},
		},
	}
	access := &fakePublishAccess{
		accessibleCommunityIDs: []bson.ObjectID{communityA},
	}
	service := &PublishService{
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListDecentralizedCommunities(publishLandlordContext("13800000000"), ListDecentralizedCommunitiesInput{City: "杭州"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != communityA {
		t.Fatalf("expected only scoped community A, got %#v", got)
	}
	if hmd.listDecentralizedCommunitiesByIDsInput.City != "杭州" {
		t.Fatalf("expected city filter to pass through, got %#v", hmd.listDecentralizedCommunitiesByIDsInput)
	}
	if len(hmd.listDecentralizedCommunitiesByIDsIDs) != 1 || hmd.listDecentralizedCommunitiesByIDsIDs[0] != communityA {
		t.Fatalf("expected scoped ids to pass through, got %#v", hmd.listDecentralizedCommunitiesByIDsIDs)
	}
}

func TestPublishServiceListDecentralizedCommunitiesTreatsCityAsOptionalFilter(t *testing.T) {
	communityA := bson.NewObjectID()
	hmd := &fakeHmdService{
		listDecentralizedCommunitiesByIDsResult: []hmdmodel.HmdDecentralized{
			{CommonFields: commonmodel.CommonFields{ID: communityA}},
		},
	}
	access := &fakePublishAccess{
		accessibleCommunityIDs: []bson.ObjectID{communityA},
	}
	service := &PublishService{
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, access),
	}

	got, err := service.ListDecentralizedCommunities(publishLandlordContext("13800000000"), ListDecentralizedCommunitiesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != communityA {
		t.Fatalf("expected scoped community without city filter, got %#v", got)
	}
	if hmd.listDecentralizedCommunitiesByIDsInput.City != "" || hmd.listDecentralizedCommunitiesByIDsInput.District != "" {
		t.Fatalf("expected empty city/district to pass as optional filters, got %#v", hmd.listDecentralizedCommunitiesByIDsInput)
	}
}

func TestPublishServiceRejectsBuildingCreateWhenProjectOutOfScope(t *testing.T) {
	projectA := bson.NewObjectID()
	projectB := bson.NewObjectID()
	hmd := &fakeHmdService{}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectA},
	}
	service := &PublishService{
		buildingService: newBuildingService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateBuilding(publishLandlordContext("13800000000"), CreateBuildingInput{ProjectID: projectB})
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
	buildingB := bson.NewObjectID()
	hmd := &fakeHmdService{
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingB: {CommonFields: commonmodel.CommonFields{ID: buildingB}, ProjectID: projectB},
		},
	}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectA},
	}
	service := &PublishService{
		centralizedRoomService: newCentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateCentralizedRoom(publishLandlordContext("13800000000"), CreateCentralizedRoomInput{BuildingID: buildingB})
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
	buildingB := bson.NewObjectID()
	hmd := &fakeHmdService{
		buildingsByID: map[bson.ObjectID]*hmdmodel.HmdBuilding{
			buildingB: {CommonFields: commonmodel.CommonFields{ID: buildingB}, ProjectID: projectB},
		},
	}
	access := &fakePublishAccess{
		accessibleProjectIDs: []bson.ObjectID{projectA},
	}
	service := &PublishService{
		roomTypeService: newRoomTypeService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateRoomType(publishLandlordContext("13800000000"), CreateRoomTypeInput{BuildingID: buildingB})
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
	hmd := &fakeHmdService{}
	access := &fakePublishAccess{
		accessibleCommunityIDs: []bson.ObjectID{communityA},
	}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.CreateDecentralizedRoom(publishLandlordContext("13800000000"), CreateDecentralizedRoomInput{DecentralizedID: communityB})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for out-of-scope community create, got %v", err)
	}
	if hmd.createDecentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD create decentralized room not to be called, got %d", hmd.createDecentralizedRoomCalls)
	}
}

func TestPublishServiceRejectsParentUpdatesOutOfScope(t *testing.T) {
	hmd := &fakeHmdService{}
	service := &PublishService{
		centralizedProjectService:     newCentralizedProjectService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		buildingService:               newBuildingService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		roomTypeService:               newRoomTypeService(hmd, mutationPublisher{}, &fakePublishAccess{}),
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, mutationPublisher{}, &fakePublishAccess{}),
	}
	ctx := publishLandlordContext("13800000000")

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

func TestPublishServiceRejectsCrossOwnerDecentralizedRoomUpdate(t *testing.T) {
	roomID := bson.NewObjectID()
	communityID := bson.NewObjectID()
	hmd := &fakeHmdService{
		decentralizedRoomsByID: map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized{
			roomID: {CommonFields: commonmodel.CommonFields{ID: roomID}, DecentralizedID: communityID},
		},
		updateDecentralizedRoomResult: &hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]{
			Entity: &hmdmodel.HmdRoomDecentralized{CommonFields: commonmodel.CommonFields{ID: roomID}},
		},
	}
	access := &fakePublishAccess{}
	service := &PublishService{
		decentralizedRoomService: newDecentralizedRoomService(hmd, mutationPublisher{}, access),
	}

	_, err := service.UpdateDecentralizedRoom(publishLandlordContext("13800000000"), UpdateDecentralizedRoomInput{ID: roomID})
	if errcode.FromError(err) == nil || errcode.FromError(err).Code != errcode.NotFound.Code {
		t.Fatalf("expected not found for cross-owner decentralized update, got %v", err)
	}
	if hmd.updateDecentralizedRoomCalls != 0 {
		t.Fatalf("expected HMD decentralized room update not to be called, got %d", hmd.updateDecentralizedRoomCalls)
	}
}

func newTestPublishService(hmd *fakeHmdService, projection *fakeListingProjection) *PublishService {
	return &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{listingProjection: projection}, &fakePublishAccess{}),
	}
}

func publishLandlordContext(phone string) context.Context {
	landlordID := bson.NewObjectID()
	return session.ContextWithPrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeLandlord,
		PrincipalID:   landlordID.Hex(),
		Terminal:      session.TerminalPublish,
		Phone:         phone,
	})
}

func publishGlobalContext() context.Context {
	return publishLandlordContext("13800000000")
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
	upsertRootID                 bson.ObjectID
	upsertRootType               hpdmodel.HpdRootScopeType
	upsertPrincipal              session.Principal
	rootScopeErr                 error
	accessibleProjectIDs         []bson.ObjectID
	accessibleCommunityIDs       []bson.ObjectID
	listAccessibleProjectCalls   int
	listAccessibleCommunityCalls int
}

func (f *fakePublishAccess) UpsertRootScopeForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdRootScopeRelation, error) {
	f.upsertRootType = rootType
	f.upsertRootID = rootID
	f.upsertPrincipal = principal
	if f.rootScopeErr != nil {
		return nil, f.rootScopeErr
	}
	landlordID, _ := bson.ObjectIDFromHex(principal.PrincipalID)
	return &hpdmodel.HpdRootScopeRelation{RootType: rootType, RootID: rootID, OwnerLandlordID: landlordID, OwnerPhone: principal.Phone}, nil
}

func (f *fakePublishAccess) ListAccessibleProjectIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error) {
	f.listAccessibleProjectCalls++
	return append([]bson.ObjectID(nil), f.accessibleProjectIDs...), nil
}

func (f *fakePublishAccess) ListAccessibleCommunityIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error) {
	f.listAccessibleCommunityCalls++
	return append([]bson.ObjectID(nil), f.accessibleCommunityIDs...), nil
}

func (f *fakePublishAccess) CanAccessProjectForPrincipal(ctx context.Context, projectID bson.ObjectID, principal session.Principal) (bool, error) {
	for _, id := range f.accessibleProjectIDs {
		if id == projectID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakePublishAccess) CanAccessCommunityForPrincipal(ctx context.Context, communityID bson.ObjectID, principal session.Principal) (bool, error) {
	for _, id := range f.accessibleCommunityIDs {
		if id == communityID {
			return true, nil
		}
	}
	return false, nil
}

type fakeHmdService struct {
	createCentralizedProjectResult          *hmddomain.HmdMutationResult[hmdmodel.HmdCentralized]
	createCentralizedProjectErr             error
	rollbackCentralizedProjectCalls         int
	rollbackCentralizedProjectID            bson.ObjectID
	rollbackCentralizedProjectErr           error
	updateCentralizedProjectCalls           int
	getCentralizedProjectResult             *hmdmodel.HmdCentralized
	getCentralizedProjectErr                error
	listCentralizedProjectsResult           []hmdmodel.HmdCentralized
	listCentralizedProjectsInput            ListCentralizedProjectsInput
	listCentralizedProjectsByIDsResult      []hmdmodel.HmdCentralized
	listCentralizedProjectsByIDsInput       ListCentralizedProjectsInput
	listCentralizedProjectsByIDsIDs         []bson.ObjectID
	createBuildingCalls                     int
	updateBuildingCalls                     int
	buildingsByID                           map[bson.ObjectID]*hmdmodel.HmdBuilding
	createRoomTypeCalls                     int
	roomTypesByID                           map[bson.ObjectID]*hmdmodel.HmdRoomTypeCentralized
	updateRoomTypeCalls                     int
	centralizedRoomsByID                    map[bson.ObjectID]*hmdmodel.HmdRoomCentralized
	listCentralizedRoomsByProjectResult     []hmdmodel.HmdRoomCentralized
	createCentralizedRoomResult             *hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]
	createCentralizedRoomErr                error
	createCentralizedRoomCalls              int
	createCentralizedRoomInput              CreateCentralizedRoomInput
	updateCentralizedRoomResult             *hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized]
	updateCentralizedRoomCalls              int
	updateCentralizedRoomInput              UpdateCentralizedRoomInput
	createDecentralizedRoomResult           *hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]
	createDecentralizedRoomErr              error
	createDecentralizedRoomCalls            int
	updateDecentralizedRoomResult           *hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized]
	updateDecentralizedRoomCalls            int
	decentralizedRoomsByID                  map[bson.ObjectID]*hmdmodel.HmdRoomDecentralized
	createDecentralizedCommunityResult      *hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized]
	createDecentralizedCommunityErr         error
	rollbackDecentralizedCommunityCalls     int
	rollbackDecentralizedCommunityID        bson.ObjectID
	rollbackDecentralizedCommunityErr       error
	listDecentralizedCommunitiesResult      []hmdmodel.HmdDecentralized
	listDecentralizedCommunitiesInput       ListDecentralizedCommunitiesInput
	listDecentralizedCommunitiesByIDsResult []hmdmodel.HmdDecentralized
	listDecentralizedCommunitiesByIDsInput  ListDecentralizedCommunitiesInput
	listDecentralizedCommunitiesByIDsIDs    []bson.ObjectID
	updateDecentralizedCommunityCalls       int
}

func (f *fakeHmdService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdCentralized], error) {
	return f.createCentralizedProjectResult, f.createCentralizedProjectErr
}

func (f *fakeHmdService) RollbackCentralizedProjectCreate(ctx context.Context, id bson.ObjectID) error {
	f.rollbackCentralizedProjectCalls++
	f.rollbackCentralizedProjectID = id
	return f.rollbackCentralizedProjectErr
}

func (f *fakeHmdService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	return f.getCentralizedProjectResult, f.getCentralizedProjectErr
}

func (f *fakeHmdService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	f.listCentralizedProjectsInput = input
	return f.listCentralizedProjectsResult, nil
}

func (f *fakeHmdService) ListCentralizedProjectsByIDs(ctx context.Context, ids []bson.ObjectID, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error) {
	f.listCentralizedProjectsByIDsIDs = ids
	f.listCentralizedProjectsByIDsInput = input
	return f.listCentralizedProjectsByIDsResult, nil
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
	if f.roomTypesByID != nil {
		return f.roomTypesByID[id], nil
	}
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
	f.createCentralizedRoomInput = input
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
	f.updateCentralizedRoomInput = input
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
	return f.createDecentralizedCommunityResult, f.createDecentralizedCommunityErr
}

func (f *fakeHmdService) RollbackDecentralizedCommunityCreate(ctx context.Context, id bson.ObjectID) error {
	f.rollbackDecentralizedCommunityCalls++
	f.rollbackDecentralizedCommunityID = id
	return f.rollbackDecentralizedCommunityErr
}

func (f *fakeHmdService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	f.listDecentralizedCommunitiesInput = input
	return f.listDecentralizedCommunitiesResult, nil
}

func (f *fakeHmdService) ListDecentralizedCommunitiesByIDs(ctx context.Context, ids []bson.ObjectID, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	f.listDecentralizedCommunitiesByIDsIDs = ids
	f.listDecentralizedCommunitiesByIDsInput = input
	return f.listDecentralizedCommunitiesByIDsResult, nil
}

func (f *fakeHmdService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized], error) {
	f.updateDecentralizedCommunityCalls++
	return nil, nil
}
