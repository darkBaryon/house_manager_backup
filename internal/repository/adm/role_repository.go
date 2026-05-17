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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RoleRepository struct {
	*common.Repository[authmodel.AdmRole]
}

type RoleListFilter struct {
	Keyword string
	Skip    int64
	Limit   int64
}

func NewRoleRepository(client *dbmongo.Client) *RoleRepository {
	return &RoleRepository{
		Repository: common.NewRepository[authmodel.AdmRole](client.Collection(authmodel.CollectionAdmRole)),
	}
}

func (r *RoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error) {
	objectIDs := compactObjectIDs(ids)
	if len(objectIDs) == 0 {
		return []authmodel.AdmRole{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": objectIDs},
		"status": commonmodel.StatusActive,
	})
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
	role, err := r.FindOne(ctx, bson.M{
		"_id":    id,
		"status": commonmodel.StatusActive,
	})
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
	items, err := r.FindMany(ctx, bson.M{"role_code": roleCode}, options.Find().SetLimit(2))
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
	filter := bson.M{"status": commonmodel.StatusActive}
	if input.Keyword = strings.TrimSpace(input.Keyword); input.Keyword != "" {
		pattern := regexp.QuoteMeta(input.Keyword)
		filter["$or"] = []bson.M{
			{"role_name": bson.Regex{Pattern: pattern, Options: "i"}},
			{"role_code": bson.Regex{Pattern: pattern, Options: "i"}},
		}
	}
	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count role list: %w", err)
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}})
	if input.Skip > 0 {
		opts.SetSkip(input.Skip)
	}
	if input.Limit > 0 {
		opts.SetLimit(input.Limit)
	}
	items, err := r.FindMany(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}
	return items, total, nil
}

func (r *RoleRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update role fields: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update role fields: fields is required")
	}
	fields = cloneBsonM(fields)
	fields["updated_at"] = time.Now().Unix()
	res, err := r.Collection.UpdateOne(ctx, bson.M{"_id": id, "status": commonmodel.StatusActive}, bson.M{
		"$set": fields,
		"$inc": bson.M{"version": 1},
	})
	if err != nil {
		return fmt.Errorf("update role fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *RoleRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return fmt.Errorf("rollback role create: id is required")
	}
	if _, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
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
