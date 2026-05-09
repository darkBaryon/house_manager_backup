package publish

import (
	"context"
	"fmt"
	"testing"

	"house-manager/internal/model"
	hmdsvc "house-manager/internal/service/publish/hmd"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPublishServiceAppliesHmdChangesAfterWrite(t *testing.T) {
	entity := &model.HmdCentralized{CommonFields: model.CommonFields{ID: bson.NewObjectID()}}
	change := hmdsvc.HmdChange{
		Action:     hmdsvc.HmdChangeCreated,
		EntityType: hmdsvc.HmdEntityCentralizedProject,
		EntityID:   entity.ID,
		Scope:      hmdsvc.HmdScopeCentralizedProject,
		ProjectID:  entity.ID,
	}
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmdsvc.HmdMutationResult[model.HmdCentralized]{
			Entity:  entity,
			Changes: []hmdsvc.HmdChange{change},
		},
	}
	hpd := &fakeHpdApplier{}
	service := newPublishService(hmd, hpd)

	got, err := service.CreateCentralizedProject(context.Background(), CreateCentralizedProjectInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != entity {
		t.Fatalf("expected entity returned from mutation result")
	}
	if hpd.calls != 1 {
		t.Fatalf("expected hpd Apply to be called once, got %d", hpd.calls)
	}
	if len(hpd.changes) != 1 || hpd.changes[0] != change {
		t.Fatalf("unexpected applied changes: %#v", hpd.changes)
	}
}

func TestPublishServiceSkipsApplyWhenHmdWriteFails(t *testing.T) {
	hmd := &fakeHmdService{
		createCentralizedProjectErr: errcode.InvalidParam.WithError(fmt.Errorf("projectName is required")),
	}
	hpd := &fakeHpdApplier{}
	service := newPublishService(hmd, hpd)

	_, err := service.CreateCentralizedProject(context.Background(), CreateCentralizedProjectInput{})
	if err == nil {
		t.Fatal("expected HMD error")
	}
	if hpd.calls != 0 {
		t.Fatalf("expected hpd Apply not to be called, got %d calls", hpd.calls)
	}
}

func TestPublishServiceReadDoesNotApplyHpdChanges(t *testing.T) {
	entity := &model.HmdCentralized{CommonFields: model.CommonFields{ID: bson.NewObjectID()}}
	hmd := &fakeHmdService{getCentralizedProjectResult: entity}
	hpd := &fakeHpdApplier{}
	service := newPublishService(hmd, hpd)

	got, err := service.GetCentralizedProject(context.Background(), entity.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != entity {
		t.Fatalf("expected get entity")
	}
	if hpd.calls != 0 {
		t.Fatalf("expected read not to call hpd Apply, got %d calls", hpd.calls)
	}
}

func TestPublishServiceReturnsApplyError(t *testing.T) {
	entity := &model.HmdCentralized{CommonFields: model.CommonFields{ID: bson.NewObjectID()}}
	applyErr := errcode.InternalError.WithError(fmt.Errorf("apply failed"))
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmdsvc.HmdMutationResult[model.HmdCentralized]{
			Entity:  entity,
			Changes: []hmdsvc.HmdChange{{EntityID: entity.ID}},
		},
	}
	hpd := &fakeHpdApplier{err: applyErr}
	service := newPublishService(hmd, hpd)

	_, err := service.CreateCentralizedProject(context.Background(), CreateCentralizedProjectInput{})
	if err != applyErr {
		t.Fatalf("expected apply error, got %v", err)
	}
}

type fakeHpdApplier struct {
	calls   int
	changes []hmdsvc.HmdChange
	err     error
}

func (f *fakeHpdApplier) Apply(ctx context.Context, changes []hmdsvc.HmdChange) error {
	f.calls++
	f.changes = append([]hmdsvc.HmdChange(nil), changes...)
	return f.err
}

type fakeHmdService struct {
	createCentralizedProjectResult *hmdsvc.HmdMutationResult[model.HmdCentralized]
	createCentralizedProjectErr    error
	getCentralizedProjectResult    *model.HmdCentralized
	getCentralizedProjectErr       error
}

func (f *fakeHmdService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error) {
	return f.createCentralizedProjectResult, f.createCentralizedProjectErr
}

func (f *fakeHmdService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return f.getCentralizedProjectResult, f.getCentralizedProjectErr
}

func (f *fakeHmdService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error) {
	return nil, nil
}

func (f *fakeHmdService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return nil, nil
}

func (f *fakeHmdService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error) {
	return nil, nil
}
