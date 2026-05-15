package publishauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
)

type LoginLogRepository struct {
	*common.Repository[authmodel.AdmLoginLog]
}

func NewLoginLogRepository(client *dbmongo.Client) *LoginLogRepository {
	return &LoginLogRepository{
		Repository: common.NewRepository[authmodel.AdmLoginLog](client.Collection(authmodel.CollectionAdmLoginLog)),
	}
}

func (r *LoginLogRepository) Create(ctx context.Context, log *authmodel.AdmLoginLog) error {
	if err := log.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm login log: %w", err)
	}
	return r.Insert(ctx, log)
}
