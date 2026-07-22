package adm

import (
	"context"
	"fmt"
	"strings"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PermissionRepository struct {
	*common.Repository[authmodel.AdmPermission]
}

const (
	permissionFieldPermissionCode common.Field = "permission_code"
	permissionFieldModule         common.Field = "module"
	permissionFieldStatus         common.Field = "status"
)

func NewPermissionRepository(client *dbmongo.Client) *PermissionRepository {
	return &PermissionRepository{
		Repository: common.NewRepository[authmodel.AdmPermission](client.Collection(authmodel.CollectionAdmPermission)),
	}
}

func (r *PermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error) {
	objectIDs := common.CompactObjectIDs(ids)
	if len(objectIDs) == 0 {
		return []authmodel.AdmPermission{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": objectIDs},
		"status": commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find permissions by ids: %w", err)
	}
	return items, nil
}

func (r *PermissionRepository) FindActiveByCodes(ctx context.Context, codes []string) ([]authmodel.AdmPermission, error) {
	values := compactCodes(codes)
	if len(values) == 0 {
		return []authmodel.AdmPermission{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"permission_code": bson.M{"$in": values},
		"status":          commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find permissions by codes: %w", err)
	}
	return items, nil
}

func compactCodes(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
