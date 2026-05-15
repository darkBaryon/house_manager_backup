package publishauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StaffRepository struct {
	*common.Repository[authmodel.AdmStaff]
}

func NewStaffRepository(client *dbmongo.Client) *StaffRepository {
	return &StaffRepository{
		Repository: common.NewRepository[authmodel.AdmStaff](client.Collection(authmodel.CollectionAdmStaff)),
	}
}

func (r *StaffRepository) Create(ctx context.Context, staff *authmodel.AdmStaff) error {
	if err := staff.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm staff: %w", err)
	}
	return r.Insert(ctx, staff)
}

func (r *StaffRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error) {
	if phone == "" {
		return nil, fmt.Errorf("find staff by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find staff by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find staff by phone: multiple active staff records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *StaffRepository) TouchLastLogin(ctx context.Context, staffID bson.ObjectID, lastLoginAt int64, lastLoginIP string) error {
	if staffID.IsZero() {
		return fmt.Errorf("touch staff last login: staffID is required")
	}
	if lastLoginAt <= 0 {
		return fmt.Errorf("touch staff last login: lastLoginAt is required")
	}
	return r.UpdateFieldsByID(ctx, staffID, bson.M{
		"last_login_at": lastLoginAt,
		"last_login_ip": lastLoginIP,
	})
}
