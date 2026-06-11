package adm

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoleRepository struct {
	*common.Repository[authmodel.AdmRole]
}

const (
	roleFieldID          common.Field = "_id"
	roleFieldRoleName    common.Field = "role_name"
	roleFieldRoleCode    common.Field = "role_code"
	roleFieldDescription common.Field = "description"
	roleFieldStatus      common.Field = "status"
	roleFieldCreatedAt   common.Field = "created_at"
	roleFieldUpdatedAt   common.Field = "updated_at"
	roleFieldVersion     common.Field = "version"
)

type RoleListFilter struct {
	Keyword string
	Skip    int64
	Limit   int64
}

type RoleUpdate struct {
	RoleName    *string
	Description *string
}

func NewRoleRepository(client *dbmongo.Client) *RoleRepository {
	return &RoleRepository{
		Repository: common.NewRepository[authmodel.AdmRole](client.Collection(authmodel.CollectionAdmRole)),
	}
}

func (r *RoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error) {
	objectIDs := common.CompactObjectIDs(ids)
	if len(objectIDs) == 0 {
		return []authmodel.AdmRole{}, nil
	}
	items, err := r.FindManyBy(ctx, common.And(
		common.In(roleFieldID, objectIDs),
		common.Eq(roleFieldStatus, commonmodel.StatusActive),
	))
	if err != nil {
		return nil, fmt.Errorf("find roles by ids: %w", err)
	}
	return items, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *authmodel.AdmRole) error {
	normalizeRole(role)
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return r.Insert(ctx, role)
}

func (r *RoleRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmRole, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find role by id: id is required")
	}
	role, err := r.FindOneBy(ctx, common.And(
		common.Eq(roleFieldID, id),
		common.Eq(roleFieldStatus, commonmodel.StatusActive),
	))
	if err != nil {
		return nil, fmt.Errorf("find role by id: %w", err)
	}
	return role, nil
}

func (r *RoleRepository) FindByCode(ctx context.Context, roleCode string) (*authmodel.AdmRole, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil, fmt.Errorf("find role by code: roleCode is required")
	}
	items, err := r.FindManyBy(ctx, common.Eq(roleFieldRoleCode, roleCode), common.Limit(2))
	if err != nil {
		return nil, fmt.Errorf("find role by code: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find role by code: multiple role records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *RoleRepository) List(ctx context.Context, input RoleListFilter) ([]authmodel.AdmRole, int64, error) {
	filter := common.Eq(roleFieldStatus, commonmodel.StatusActive)
	if input.Keyword = strings.TrimSpace(input.Keyword); input.Keyword != "" {
		pattern := regexp.QuoteMeta(input.Keyword)
		filter = common.And(filter, common.Or(
			common.Regex(roleFieldRoleName, pattern, "i"),
			common.Regex(roleFieldRoleCode, pattern, "i"),
		))
	}
	total, err := r.CountBy(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count role list: %w", err)
	}
	opts := []common.QueryOption{
		common.SortBy(roleFieldCreatedAt, common.SortDesc),
		common.SortBy(roleFieldID, common.SortDesc),
	}
	if input.Skip > 0 {
		opts = append(opts, common.Skip(input.Skip))
	}
	if input.Limit > 0 {
		opts = append(opts, common.Limit(input.Limit))
	}
	items, err := r.FindManyBy(ctx, filter, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}
	return items, total, nil
}

func (r *RoleRepository) Update(ctx context.Context, id bson.ObjectID, input RoleUpdate) error {
	if id.IsZero() {
		return fmt.Errorf("update role: id is required")
	}
	if input.RoleName == nil && input.Description == nil {
		return fmt.Errorf("update role: fields is required")
	}
	update := common.NewUpdateDoc().
		Set(roleFieldUpdatedAt, time.Now().Unix()).
		Inc(roleFieldVersion, 1)
	if input.RoleName != nil {
		update = update.Set(roleFieldRoleName, *input.RoleName)
	}
	if input.Description != nil {
		update = update.Set(roleFieldDescription, *input.Description)
	}
	matched, err := r.UpdateOneBy(ctx, common.And(
		common.Eq(roleFieldID, id),
		common.Eq(roleFieldStatus, commonmodel.StatusActive),
	), update)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	if !matched {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *RoleRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("rollback role create: id is required")
	}
	if err := r.DeleteOneBy(ctx, common.Eq(roleFieldID, id)); err != nil {
		return fmt.Errorf("rollback role create: %w", err)
	}
	return nil
}

func normalizeRole(role *authmodel.AdmRole) {
	if role == nil {
		return
	}
	role.RoleName = strings.TrimSpace(role.RoleName)
	role.RoleCode = strings.TrimSpace(role.RoleCode)
	role.Description = strings.TrimSpace(role.Description)
}
