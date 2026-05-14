package publish

import (
	"context"
	"fmt"
	"testing"

	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/internal/model"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPublishServiceAppliesHmdChangesAfterWrite(t *testing.T) {
	entity := &model.HmdCentralized{CommonFields: model.CommonFields{ID: bson.NewObjectID()}}
	change := hmddomain.HmdChange{
		Action:     hmddomain.HmdChangeCreated,
		EntityType: hmddomain.HmdEntityCentralizedProject,
		EntityID:   entity.ID,
		Scope:      hmddomain.HmdScopeCentralizedProject,
		ProjectID:  entity.ID,
	}
	hmd := &fakeHmdService{
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[model.HmdCentralized]{
			Entity:  entity,
			Changes: []hmddomain.HmdChange{change},
		},
	}
	hpd := &fakeHpdApplier{}
	service := newTestPublishService(hmd, hpd)

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
	service := newTestPublishService(hmd, hpd)

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
	service := newTestPublishService(hmd, hpd)

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
		createCentralizedProjectResult: &hmddomain.HmdMutationResult[model.HmdCentralized]{
			Entity:  entity,
			Changes: []hmddomain.HmdChange{{EntityID: entity.ID}},
		},
	}
	hpd := &fakeHpdApplier{err: applyErr}
	service := newTestPublishService(hmd, hpd)

	_, err := service.CreateCentralizedProject(context.Background(), CreateCentralizedProjectInput{})
	if err != applyErr {
		t.Fatalf("expected apply error, got %v", err)
	}
}

func newTestPublishService(hmd *fakeHmdService, hpd *fakeHpdApplier) *PublishService {
	return &PublishService{
		centralizedProjectService: newCentralizedProjectService(hmd, mutationPublisher{hpd: hpd}),
	}
}

type fakeHpdApplier struct {
	calls   int
	changes []hmddomain.HmdChange
	err     error
}

func (f *fakeHpdApplier) Apply(ctx context.Context, changes []hmddomain.HmdChange) error {
	f.calls++
	f.changes = append([]hmddomain.HmdChange(nil), changes...)
	return f.err
}

type fakeHmdService struct {
	createCentralizedProjectResult *hmddomain.HmdMutationResult[model.HmdCentralized]
	createCentralizedProjectErr    error
	getCentralizedProjectResult    *model.HmdCentralized
	getCentralizedProjectErr       error
}

func (f *fakeHmdService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmddomain.HmdMutationResult[model.HmdCentralized], error) {
	return f.createCentralizedProjectResult, f.createCentralizedProjectErr
}

func (f *fakeHmdService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return f.getCentralizedProjectResult, f.getCentralizedProjectErr
}

func (f *fakeHmdService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	return nil, nil
}

func (f *fakeHmdService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmddomain.HmdMutationResult[model.HmdCentralized], error) {
	return nil, nil
}
