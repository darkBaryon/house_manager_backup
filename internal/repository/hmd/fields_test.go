package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
	"testing"
)

func TestActiveFilter(t *testing.T) {
	filter := activeFilter(bson.M{"city": "深圳"})

	if got := filter["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status filter, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city preserved, got %v", got)
	}
}

func TestActiveFilterOverridesIncomingStatus(t *testing.T) {
	filter := activeFilter(bson.M{"city": "深圳", "status": commonmodel.StatusDeleted})

	if got := filter["status"]; got != commonmodel.StatusActive {
		t.Fatalf("expected active status filter to override incoming status, got %v", got)
	}
	if got := filter["city"]; got != "深圳" {
		t.Fatalf("expected city preserved, got %v", got)
	}
}

func TestHmdListFindOptionsSortsByUpdatedAtAndIDDesc(t *testing.T) {
	var findOptions options.FindOptions
	for _, setter := range hmdListFindOptions().List() {
		if err := setter(&findOptions); err != nil {
			t.Fatalf("apply find option: %v", err)
		}
	}

	sort, ok := findOptions.Sort.(bson.D)
	if !ok {
		t.Fatalf("expected bson.D sort, got %#v", findOptions.Sort)
	}
	want := bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}
	if len(sort) != len(want) {
		t.Fatalf("expected sort %v, got %v", want, sort)
	}
	for i := range want {
		if sort[i] != want[i] {
			t.Fatalf("expected sort %v, got %v", want, sort)
		}
	}
}

func TestPickAllowedFieldsRejectsSystemField(t *testing.T) {
	_, err := pickAllowedFields(bson.M{"updated_at": 123}, centralizedBaseInfoFields)
	if err == nil || !strings.Contains(err.Error(), "updated_at") {
		t.Fatalf("expected updated_at to be rejected, got %v", err)
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

func TestIsValidRoomStatus(t *testing.T) {
	validStatuses := []int{-1, 0, 1, 2, 3}
	for _, status := range validStatuses {
		if !hmdmodel.IsValidRoomStatus(status) {
			t.Fatalf("expected status %d to be valid", status)
		}
	}

	invalidStatuses := []int{-2, 4, 99}
	for _, status := range invalidStatuses {
		if hmdmodel.IsValidRoomStatus(status) {
			t.Fatalf("expected status %d to be invalid", status)
		}
	}
}

func TestIsValidRoomStatusUpdateTarget(t *testing.T) {
	validStatuses := []int{-1, 1, 2, 3}
	for _, status := range validStatuses {
		if !hmdmodel.IsValidRoomStatusUpdateTarget(status) {
			t.Fatalf("expected status %d to be a valid update target", status)
		}
	}

	invalidStatuses := []int{-2, 0, 4, 99}
	for _, status := range invalidStatuses {
		if hmdmodel.IsValidRoomStatusUpdateTarget(status) {
			t.Fatalf("expected status %d to be an invalid update target", status)
		}
	}
}

func TestRoomCentralizedUpdateStatusRejectsInvalidRoomStatus(t *testing.T) {
	repo := &RoomCentralizedRepository{}
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), 99)
	if err == nil || !strings.Contains(err.Error(), "roomStatus is invalid") {
		t.Fatalf("expected invalid roomStatus error, got %v", err)
	}
}

func TestRoomCentralizedUpdateStatusRejectsUnspecifiedRoomStatus(t *testing.T) {
	repo := &RoomCentralizedRepository{}
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), int(hmdmodel.RoomStatusUnspecified))
	if err == nil || !strings.Contains(err.Error(), "roomStatus is invalid") {
		t.Fatalf("expected invalid roomStatus error, got %v", err)
	}
}

func TestRoomDecentralizedUpdateStatusRejectsInvalidRoomStatus(t *testing.T) {
	repo := &RoomDecentralizedRepository{}
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), 99)
	if err == nil || !strings.Contains(err.Error(), "roomStatus is invalid") {
		t.Fatalf("expected invalid roomStatus error, got %v", err)
	}
}

func TestRoomDecentralizedUpdateStatusRejectsUnspecifiedRoomStatus(t *testing.T) {
	repo := &RoomDecentralizedRepository{}
	err := repo.UpdateStatus(context.Background(), bson.NewObjectID(), int(hmdmodel.RoomStatusUnspecified))
	if err == nil || !strings.Contains(err.Error(), "roomStatus is invalid") {
		t.Fatalf("expected invalid roomStatus error, got %v", err)
	}
}
