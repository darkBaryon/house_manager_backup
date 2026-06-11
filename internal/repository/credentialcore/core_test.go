package credentialcore

import (
	"context"
	"errors"
	"testing"

	authmodel "house-manager/internal/model/auth"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateRunsNormalizeValidateAndInsert(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	calledNormalize := false
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{
		OwnerField: common.Field("staff_id"),
		Normalize: func(entity *authmodel.AdmStaffAuth) {
			calledNormalize = true
			entity.PasswordHash = "trimmed"
		},
		Validate: func(entity *authmodel.AdmStaffAuth) error {
			if entity.PasswordHash != "trimmed" {
				t.Fatalf("expected validate after normalize, got %q", entity.PasswordHash)
			}
			return nil
		},
	})

	entity := &authmodel.AdmStaffAuth{}
	if err := core.Create(context.Background(), entity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !calledNormalize {
		t.Fatal("expected normalize to be called")
	}
	if repo.inserted != entity {
		t.Fatal("expected entity inserted")
	}
}

func TestCreateReturnsValidateError(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	validateErr := errors.New("invalid")
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{
		OwnerField: common.Field("staff_id"),
		Validate: func(entity *authmodel.AdmStaffAuth) error {
			return validateErr
		},
	})

	if err := core.Create(context.Background(), &authmodel.AdmStaffAuth{}); !errors.Is(err, validateErr) {
		t.Fatalf("expected validate error, got %v", err)
	}
	if repo.inserted != nil {
		t.Fatal("expected insert not called")
	}
}

func TestFindActivePasswordByOwnerBuildsFilter(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{OwnerField: common.Field("staff_id")})
	staffID := bson.NewObjectID()

	if _, err := core.FindActivePasswordByOwner(context.Background(), staffID); err != nil {
		t.Fatalf("FindActivePasswordByOwner() error = %v", err)
	}
	filter := repo.findOneFilter.BSON()
	parts := filter["$and"].([]bson.M)
	assertFilterPart(t, parts, "staff_id", staffID)
	assertFilterPart(t, parts, "auth_type", authmodel.PasswordAuthTypePassword)
	assertFilterPart(t, parts, "status", 1)
}

func TestFindActivePasswordByOwnersShortCircuitsEmptyIDs(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{OwnerField: common.Field("staff_id")})

	items, err := core.FindActivePasswordByOwners(context.Background(), []bson.ObjectID{bson.NilObjectID})
	if err != nil {
		t.Fatalf("FindActivePasswordByOwners() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty result, got %#v", items)
	}
	if repo.findManyCalled {
		t.Fatal("expected empty ids to short-circuit before DB")
	}
}

func TestFindActivePasswordByOwnersCompactsIDs(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{OwnerField: common.Field("staff_id")})
	staffID := bson.NewObjectID()

	if _, err := core.FindActivePasswordByOwners(context.Background(), []bson.ObjectID{staffID, bson.NilObjectID, staffID}); err != nil {
		t.Fatalf("FindActivePasswordByOwners() error = %v", err)
	}
	filter := repo.findManyFilter.BSON()
	parts := filter["$and"].([]bson.M)
	inPart := findFilterPart(t, parts, "staff_id")
	values := inPart["staff_id"].(bson.M)["$in"].([]bson.ObjectID)
	if len(values) != 1 || values[0] != staffID {
		t.Fatalf("expected compacted owner ids, got %#v", values)
	}
}

func TestTouchLastLoginSetsFields(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{OwnerField: common.Field("staff_id")})
	authID := bson.NewObjectID()

	if err := core.TouchLastLogin(context.Background(), authID, " 127.0.0.1 "); err != nil {
		t.Fatalf("TouchLastLogin() error = %v", err)
	}
	if repo.updateID != authID {
		t.Fatalf("expected update id %v, got %v", authID, repo.updateID)
	}
	update := repo.update.BSON()
	set := update["$set"].(bson.M)
	if set["last_login_at"] == nil || set["updated_at"] == nil {
		t.Fatalf("expected last_login_at and updated_at set, got %#v", set)
	}
	if got := set["last_login_ip"]; got != " 127.0.0.1 " {
		t.Fatalf("expected caller-normalized login ip, got %#v", got)
	}
	inc := update["$inc"].(bson.M)
	if got := inc["version"]; got != 1 {
		t.Fatalf("expected version increment, got %#v", got)
	}
}

func TestRollbackCreateByOwnerDeletesOwnerScope(t *testing.T) {
	repo := &spyRepository[authmodel.AdmStaffAuth]{}
	core := newCore[authmodel.AdmStaffAuth](repo, Config[authmodel.AdmStaffAuth]{OwnerField: common.Field("staff_id")})
	staffID := bson.NewObjectID()

	if err := core.RollbackCreateByOwner(context.Background(), staffID); err != nil {
		t.Fatalf("RollbackCreateByOwner() error = %v", err)
	}
	filter := repo.deleteFilter.BSON()
	if got := filter["staff_id"]; got != staffID {
		t.Fatalf("expected delete scoped to staff_id %v, got %#v", staffID, got)
	}
}

type spyRepository[T any] struct {
	inserted       *T
	findOneFilter  common.Filter
	findManyFilter common.Filter
	findManyCalled bool
	updateID       bson.ObjectID
	update         common.UpdateDoc
	deleteFilter   common.Filter
}

func (s *spyRepository[T]) Insert(ctx context.Context, entity *T) error {
	s.inserted = entity
	return nil
}

func (s *spyRepository[T]) FindOneBy(ctx context.Context, filter common.Filter) (*T, error) {
	s.findOneFilter = filter
	return nil, nil
}

func (s *spyRepository[T]) FindManyBy(ctx context.Context, filter common.Filter, opts ...common.QueryOption) ([]T, error) {
	s.findManyCalled = true
	s.findManyFilter = filter
	return nil, nil
}

func (s *spyRepository[T]) UpdateByID(ctx context.Context, id bson.ObjectID, update common.UpdateDoc) error {
	s.updateID = id
	s.update = update
	return nil
}

func (s *spyRepository[T]) DeleteAllBy(ctx context.Context, filter common.Filter) error {
	s.deleteFilter = filter
	return nil
}

func assertFilterPart(t *testing.T, parts []bson.M, field string, want any) {
	t.Helper()
	part := findFilterPart(t, parts, field)
	if got := part[field]; got != want {
		t.Fatalf("expected %s=%#v, got %#v", field, want, got)
	}
}

func findFilterPart(t *testing.T, parts []bson.M, field string) bson.M {
	t.Helper()
	for _, part := range parts {
		if _, ok := part[field]; ok {
			return part
		}
	}
	t.Fatalf("missing filter part %q in %#v", field, parts)
	return nil
}
