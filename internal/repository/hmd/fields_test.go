package hmd

import (
	"context"
	"strings"
	"testing"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestActiveFilter(t *testing.T) {
	filter := activeFilter(bson.M{"city": "深圳"})

	if got := filter["status"]; got != model.StatusActive {
		t.Fatalf("expected active status filter, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city preserved, got %v", got)
	}
}

func TestNotDeletedFilter(t *testing.T) {
	filter := notDeletedFilter(bson.M{"project_code": "CENTRAL-001"})

	statusFilter, ok := filter["status"].(bson.M)
	if !ok {
		t.Fatalf("expected nested status filter, got %T", filter["status"])
	}
	if got := statusFilter["$ne"]; got != model.StatusDeleted {
		t.Fatalf("expected status != deleted, got %v", got)
	}
}

func TestPickAllowedFieldsRejectsSystemField(t *testing.T) {
	_, err := pickAllowedFields(bson.M{"updated_at": 123}, centralizedBaseInfoFields)
	if err == nil || !strings.Contains(err.Error(), "updated_at") {
		t.Fatalf("expected updated_at to be rejected, got %v", err)
	}
}

func TestBuildingUpdateBaseInfoRejectsOwnershipField(t *testing.T) {
	repo := &BuildingRepository{}
	err := repo.UpdateBaseInfo(context.Background(), bson.NewObjectID(), bson.M{"project_id": bson.NewObjectID()})
	if err == nil || !strings.Contains(err.Error(), "project_id") {
		t.Fatalf("expected project_id to be rejected, got %v", err)
	}
}

func TestRoomCentralizedUpdateBaseInfoRejectsStatusField(t *testing.T) {
	repo := &RoomCentralizedRepository{}
	err := repo.UpdateBaseInfo(context.Background(), bson.NewObjectID(), bson.M{"room_status": 2})
	if err == nil || !strings.Contains(err.Error(), "room_status") {
		t.Fatalf("expected room_status to be rejected, got %v", err)
	}
}

func TestRoomDecentralizedUpdateBaseInfoRejectsOwnershipField(t *testing.T) {
	repo := &RoomDecentralizedRepository{}
	err := repo.UpdateBaseInfo(context.Background(), bson.NewObjectID(), bson.M{"decentralized_id": bson.NewObjectID()})
	if err == nil || !strings.Contains(err.Error(), "decentralized_id") {
		t.Fatalf("expected decentralized_id to be rejected, got %v", err)
	}
}

func TestRoomStatusUpdateFields(t *testing.T) {
	fields := roomStatusUpdateFields(3)
	if len(fields) != 1 {
		t.Fatalf("expected exactly one field, got %d", len(fields))
	}
	if got := fields["room_status"]; got != 3 {
		t.Fatalf("expected room_status=3, got %v", got)
	}
}
