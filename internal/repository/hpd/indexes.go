package hpd

import (
	"context"
	"fmt"

	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (r *RootScopeRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex(
			"root_type_1_root_id_1_active_unique",
			common.IndexKey(hpdFieldRootType, common.SortAsc),
			common.IndexKey(hpdFieldRootID, common.SortAsc),
		).WithUnique().WithPartialFilter(bson.M{
			"status":          commonmodel.StatusActive,
			"relation_status": hpdmodel.HpdRelationStatusActive,
		}),
		common.NewIndex(
			"root_type_1_owner_landlord_id_1_relation_status_1_status_1",
			common.IndexKey(hpdFieldRootType, common.SortAsc),
			common.IndexKey(hpdFieldOwnerLandlordID, common.SortAsc),
			common.IndexKey(hpdFieldRelationStatus, common.SortAsc),
			common.IndexKey(hpdFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure hpd root scope relation indexes: %w", err)
	}
	return nil
}

func (r *PublisherListingRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("listing_id_1", common.IndexKey(hpdFieldListingID, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"source_type_1_source_id_1",
			common.IndexKey(hpdFieldSourceType, common.SortAsc),
			common.IndexKey(hpdFieldSourceID, common.SortAsc),
		),
		common.NewIndex(
			"owner_landlord_id_1_updated_at_-1",
			common.IndexKey(hpdFieldOwnerLandlordID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"root_type_1_root_id_1_updated_at_-1",
			common.IndexKey(hpdFieldRootType, common.SortAsc),
			common.IndexKey(hpdFieldRootID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"project_id_1_updated_at_-1",
			common.IndexKey(hpdFieldProjectID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"building_id_1_updated_at_-1",
			common.IndexKey(hpdFieldBuildingID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"decentralized_id_1_updated_at_-1",
			common.IndexKey(hpdFieldDecentralizedID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"room_status_1_listing_status_1_updated_at_-1",
			common.IndexKey(hpdFieldRoomStatus, common.SortAsc),
			common.IndexKey(hpdFieldListingStatus, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure hpd publisher listing indexes: %w", err)
	}
	return nil
}

func (r *AdminListingRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("listing_id_1", common.IndexKey(hpdFieldListingID, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"source_type_1_source_id_1",
			common.IndexKey(hpdFieldSourceType, common.SortAsc),
			common.IndexKey(hpdFieldSourceID, common.SortAsc),
		),
		common.NewIndex(
			"owner_landlord_id_1_updated_at_-1",
			common.IndexKey(hpdFieldOwnerLandlordID, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"asset_mode_1_city_1_district_1_updated_at_-1",
			common.IndexKey(hpdFieldAssetMode, common.SortAsc),
			common.IndexKey(hpdFieldCity, common.SortAsc),
			common.IndexKey(hpdFieldDistrict, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"listing_status_1_audit_status_1_updated_at_-1",
			common.IndexKey(hpdFieldListingStatus, common.SortAsc),
			common.IndexKey(hpdFieldAuditStatus, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
		common.NewIndex(
			"room_status_1_is_online_1_updated_at_-1",
			common.IndexKey(hpdFieldRoomStatus, common.SortAsc),
			common.IndexKey(hpdFieldIsOnline, common.SortAsc),
			common.IndexKey(hpdFieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure hpd admin listing indexes: %w", err)
	}
	return nil
}
