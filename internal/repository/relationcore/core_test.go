package relationcore

import (
	"context"
	"testing"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestReplaceByLeftKeepsDisablesRevivesInsertsAndHandlesEmptyRightSet(t *testing.T) {
	leftID := bson.NewObjectID()
	assignedBy := bson.NewObjectID()
	keptRightID := bson.NewObjectID()
	removedRightID := bson.NewObjectID()
	revivedRightID := bson.NewObjectID()
	newRightID := bson.NewObjectID()
	otherLeftID := bson.NewObjectID()

	keptID := bson.NewObjectID()
	removedID := bson.NewObjectID()
	revivedID := bson.NewObjectID()
	otherID := bson.NewObjectID()
	repo := &relationMemoryRepository{
		items: []authmodel.AdmStaffRole{
			{CommonFields: commonmodel.CommonFields{ID: keptID, Status: commonmodel.StatusActive, Version: 3}, StaffID: leftID, RoleID: keptRightID},
			{CommonFields: commonmodel.CommonFields{ID: removedID, Status: commonmodel.StatusActive, Version: 5}, StaffID: leftID, RoleID: removedRightID},
			{CommonFields: commonmodel.CommonFields{ID: revivedID, Status: commonmodel.StatusDeleted, Version: 7}, StaffID: leftID, RoleID: revivedRightID},
			{CommonFields: commonmodel.CommonFields{ID: otherID, Status: commonmodel.StatusActive, Version: 11}, StaffID: otherLeftID, RoleID: removedRightID},
		},
	}
	core := newStaffRoleCore(repo)

	if err := core.ReplaceByLeft(context.Background(), leftID, []bson.ObjectID{keptRightID, revivedRightID, newRightID}, assignedBy); err != nil {
		t.Fatalf("ReplaceByLeft() error = %v", err)
	}

	kept := repo.mustFind(leftID, keptRightID)
	if kept.ID != keptID {
		t.Fatalf("expected kept relation _id unchanged, got %v", kept.ID)
	}
	if kept.Version != 4 {
		t.Fatalf("expected kept relation version incremented to 4, got %d", kept.Version)
	}

	removed := repo.mustFind(leftID, removedRightID)
	if removed.ID != removedID {
		t.Fatalf("expected removed relation to remain physically present, got id %v", removed.ID)
	}
	if removed.Status != commonmodel.StatusDeleted || removed.Version != 6 {
		t.Fatalf("expected removed relation soft disabled with version 6, got status=%d version=%d", removed.Status, removed.Version)
	}

	revived := repo.mustFind(leftID, revivedRightID)
	if revived.ID != revivedID || revived.Status != commonmodel.StatusActive {
		t.Fatalf("expected old disabled pair revived with same _id, got id=%v status=%d", revived.ID, revived.Status)
	}

	inserted := repo.mustFind(leftID, newRightID)
	if inserted.ID.IsZero() || inserted.Status != commonmodel.StatusActive || inserted.AssignedBy != assignedBy {
		t.Fatalf("expected new active inserted relation, got %#v", inserted)
	}

	other := repo.mustFind(otherLeftID, removedRightID)
	if other.ID != otherID || other.Status != commonmodel.StatusActive {
		t.Fatalf("expected other left scope untouched, got %#v", other)
	}

	if err := core.ReplaceByLeft(context.Background(), leftID, nil, assignedBy); err != nil {
		t.Fatalf("ReplaceByLeft(empty) error = %v", err)
	}
	for _, item := range repo.items {
		if item.StaffID == leftID && item.Status == commonmodel.StatusActive {
			t.Fatalf("expected empty right set to disable all active relations for left, still active: %#v", item)
		}
	}
}

func TestCreateManyBuildsAndValidatesEntities(t *testing.T) {
	repo := &relationMemoryRepository{}
	core := newStaffRoleCore(repo)
	staffID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	assignedBy := bson.NewObjectID()

	if err := core.CreateMany(context.Background(), staffID, []bson.ObjectID{bson.NilObjectID, roleID, roleID}, assignedBy); err != nil {
		t.Fatalf("CreateMany() error = %v", err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one compacted relation inserted, got %d", len(repo.items))
	}
	item := repo.items[0]
	if item.StaffID != staffID || item.RoleID != roleID || item.AssignedBy != assignedBy || item.AssignedAt <= 0 {
		t.Fatalf("unexpected inserted relation: %#v", item)
	}
}

func TestListActiveByLeftsShortCircuitsEmptyIDs(t *testing.T) {
	repo := &relationMemoryRepository{}
	core := newStaffRoleCore(repo)

	items, err := core.ListActiveByLefts(context.Background(), []bson.ObjectID{bson.NilObjectID})
	if err != nil {
		t.Fatalf("ListActiveByLefts() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %#v", items)
	}
	if repo.findCalled {
		t.Fatal("expected empty left ids to short-circuit before DB")
	}
}

func newStaffRoleCore(repo *relationMemoryRepository) *Core[authmodel.AdmStaffRole] {
	return newCore[authmodel.AdmStaffRole](repo, Config[authmodel.AdmStaffRole]{
		LeftField:       common.Field("staff_id"),
		RightField:      common.Field("role_id"),
		AssignedByField: common.Field("assigned_by"),
		AssignedAtField: common.Field("assigned_at"),
		Validate:        (*authmodel.AdmStaffRole).ValidateForCreate,
	})
}

type relationMemoryRepository struct {
	items      []authmodel.AdmStaffRole
	findCalled bool
}

func (r *relationMemoryRepository) Insert(ctx context.Context, item *authmodel.AdmStaffRole) error {
	if item.ID.IsZero() {
		item.ID = bson.NewObjectID()
	}
	if item.Status == commonmodel.StatusUnspecified {
		item.Status = commonmodel.StatusActive
	}
	if item.Version == 0 {
		item.Version = 1
	}
	r.items = append(r.items, *item)
	return nil
}

func (r *relationMemoryRepository) FindManyBy(ctx context.Context, filter common.Filter, opts ...common.QueryOption) ([]authmodel.AdmStaffRole, error) {
	r.findCalled = true
	result := make([]authmodel.AdmStaffRole, 0, len(r.items))
	for _, item := range r.items {
		if relationMatches(item, filter.BSON()) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *relationMemoryRepository) UpdateManyBy(ctx context.Context, filter common.Filter, update common.UpdateDoc) (int64, error) {
	var matched int64
	updateDoc := update.BSON()
	for i := range r.items {
		if !relationMatches(r.items[i], filter.BSON()) {
			continue
		}
		matched++
		applyRelationUpdate(&r.items[i], updateDoc, false)
	}
	return matched, nil
}

func (r *relationMemoryRepository) UpsertOneBy(ctx context.Context, filter common.Filter, update common.UpdateDoc) (bool, error) {
	updateDoc := update.BSON()
	for i := range r.items {
		if !relationMatches(r.items[i], filter.BSON()) {
			continue
		}
		applyRelationUpdate(&r.items[i], updateDoc, false)
		return true, nil
	}
	item := authmodel.AdmStaffRole{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}}
	applyRelationUpdate(&item, updateDoc, true)
	if item.Status == commonmodel.StatusUnspecified {
		item.Status = commonmodel.StatusActive
	}
	r.items = append(r.items, item)
	return false, nil
}

func (r *relationMemoryRepository) DeleteAllBy(ctx context.Context, filter common.Filter) error {
	kept := r.items[:0]
	for _, item := range r.items {
		if !relationMatches(item, filter.BSON()) {
			kept = append(kept, item)
		}
	}
	r.items = kept
	return nil
}

func (r *relationMemoryRepository) mustFind(leftID, rightID bson.ObjectID) authmodel.AdmStaffRole {
	for _, item := range r.items {
		if item.StaffID == leftID && item.RoleID == rightID {
			return item
		}
	}
	panic("relation not found")
}

func relationMatches(item authmodel.AdmStaffRole, filter bson.M) bool {
	if parts, ok := filter["$and"].([]bson.M); ok {
		for _, part := range parts {
			if !relationMatches(item, part) {
				return false
			}
		}
		return true
	}
	for key, value := range filter {
		switch key {
		case "staff_id":
			if !matchObjectID(item.StaffID, value) {
				return false
			}
		case "role_id":
			if !matchObjectID(item.RoleID, value) {
				return false
			}
		case "status":
			if item.Status != value {
				return false
			}
		}
	}
	return true
}

func matchObjectID(got bson.ObjectID, want any) bool {
	switch value := want.(type) {
	case bson.ObjectID:
		return got == value
	case bson.M:
		if in, ok := value["$in"].([]bson.ObjectID); ok {
			for _, id := range in {
				if got == id {
					return true
				}
			}
			return false
		}
		if nin, ok := value["$nin"].([]bson.ObjectID); ok {
			for _, id := range nin {
				if got == id {
					return false
				}
			}
			return true
		}
	}
	return false
}

func applyRelationUpdate(item *authmodel.AdmStaffRole, update bson.M, inserting bool) {
	if setOnInsert, ok := update["$setOnInsert"].(bson.M); ok && inserting {
		applyRelationSet(item, setOnInsert)
	}
	if set, ok := update["$set"].(bson.M); ok {
		applyRelationSet(item, set)
	}
	if inc, ok := update["$inc"].(bson.M); ok {
		if value, ok := inc["version"].(int); ok {
			item.Version += int64(value)
		}
	}
}

func applyRelationSet(item *authmodel.AdmStaffRole, set bson.M) {
	for key, value := range set {
		switch key {
		case "staff_id":
			item.StaffID = value.(bson.ObjectID)
		case "role_id":
			item.RoleID = value.(bson.ObjectID)
		case "assigned_by":
			item.AssignedBy = value.(bson.ObjectID)
		case "assigned_at":
			item.AssignedAt = value.(int64)
		case "status":
			item.Status = value.(int)
		case "created_at":
			item.CreatedAt = value.(int64)
		case "updated_at":
			item.UpdatedAt = value.(int64)
		}
	}
}
