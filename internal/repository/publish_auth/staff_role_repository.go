package publishauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StaffRoleRepository struct {
	*common.Repository[authmodel.AdmStaffRole]
}

func NewStaffRoleRepository(client *dbmongo.Client) *StaffRoleRepository {
	return &StaffRoleRepository{
		Repository: common.NewRepository[authmodel.AdmStaffRole](client.Collection(authmodel.CollectionAdmStaffRole)),
	}
}

func (r *StaffRoleRepository) Create(ctx context.Context, staffRole *authmodel.AdmStaffRole) error {
	if err := staffRole.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm staff role: %w", err)
	}
	return r.Insert(ctx, staffRole)
}

func (r *StaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("list staff roles: staffID is required")
	}
	return r.FindMany(ctx, bson.M{
		"staff_id": staffID,
		"status":   commonmodel.StatusActive,
	})
}
