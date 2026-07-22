package landlord

import (
	"context"
	"fmt"

	"house-manager/internal/repository/common"
)

func (r *LandlordRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("phone_1", common.IndexKey(landlordFieldPhone, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"status_1_updated_at_-1",
			common.IndexKey(landlordFieldStatus, common.SortAsc),
			common.IndexKey(landlordFieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure publish landlord indexes: %w", err)
	}
	return nil
}

func (r *LandlordAuthRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("landlord_id_1", common.IndexKey(landlordAuthFieldLandlordID, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"auth_type_1_status_1",
			common.IndexKey(landlordAuthFieldAuthType, common.SortAsc),
			common.IndexKey(landlordAuthFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure publish landlord auth indexes: %w", err)
	}
	return nil
}
