package hmd

import (
	"strings"
	"testing"

	hmdmodel "house-manager/internal/model/hmd"
)

func TestBuildingBaseInfoUpdateFields(t *testing.T) {
	fields, err := buildingBaseInfoUpdateFields(BuildingBaseInfoUpdate{
		BuildingName:      "A座",
		FloorTotal:        18,
		ManagerName:       "测试管家",
		ManagerPhone:      "18800000000",
		Photos:            []string{"https://example.com/a.jpg"},
		ListingFacilities: []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{
		"building_name":      "A座",
		"floor_total":        18,
		"manager_name":       "测试管家",
		"manager_phone":      "18800000000",
		"photos":             []string{"https://example.com/a.jpg"},
		"listing_facilities": []hmdmodel.ListingFacility{hmdmodel.ListingFacilityElevator},
	}
	for key, value := range want {
		if got := fields[key]; !equalValue(got, value) {
			t.Fatalf("expected %s=%#v, got %#v", key, value, got)
		}
	}
	if _, ok := fields["project_id"]; ok {
		t.Fatalf("project_id must not be part of building base info update: %#v", fields)
	}
}

func TestBuildingBaseInfoUpdateRejectsEmptyName(t *testing.T) {
	_, err := buildingBaseInfoUpdateFields(BuildingBaseInfoUpdate{BuildingName: ""})
	if err == nil || !strings.Contains(err.Error(), "building_name 不能为空") {
		t.Fatalf("expected empty building_name to be rejected, got %v", err)
	}
}

func TestBuildingBaseInfoUpdateRejectsNegativeFloorTotal(t *testing.T) {
	_, err := buildingBaseInfoUpdateFields(BuildingBaseInfoUpdate{BuildingName: "A座", FloorTotal: -1})
	if err == nil || !strings.Contains(err.Error(), "floor_total 不能小于 0") {
		t.Fatalf("expected negative floor_total to be rejected, got %v", err)
	}
}

func equalValue(got any, want any) bool {
	switch want := want.(type) {
	case []string:
		got, ok := got.([]string)
		if !ok || len(got) != len(want) {
			return false
		}
		for i := range want {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	case []hmdmodel.ListingFacility:
		got, ok := got.([]hmdmodel.ListingFacility)
		if !ok || len(got) != len(want) {
			return false
		}
		for i := range want {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	default:
		return got == want
	}
}
