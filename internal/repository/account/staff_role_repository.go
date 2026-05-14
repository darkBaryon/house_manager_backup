package account

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StaffRoleRepository struct {
	*common.Repository[model.AdmStaffRole]
}

func NewStaffRoleRepository(client *dbmongo.Client) *StaffRoleRepository {
	return &StaffRoleRepository{
		Repository: common.NewRepository[model.AdmStaffRole](client.Collection(model.CollectionAdmStaffRole)),
	}
}

func (r *StaffRoleRepository) Create(ctx context.Context, staffRole *model.AdmStaffRole) error {
	if err := staffRole.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm staff role: %w", err)
	}
	return r.Insert(ctx, staffRole)
}

func (r *StaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]model.AdmStaffRole, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("list staff roles: staffID is required")
	}
	return r.FindMany(ctx, bson.M{
		"staff_id": staffID,
		"status":   model.StatusActive,
	})
}
