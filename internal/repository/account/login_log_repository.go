package account

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
)

type LoginLogRepository struct {
	*common.Repository[model.AdmLoginLog]
}

func NewLoginLogRepository(client *dbmongo.Client) *LoginLogRepository {
	return &LoginLogRepository{
		Repository: common.NewRepository[model.AdmLoginLog](client.Collection(model.CollectionAdmLoginLog)),
	}
}

func (r *LoginLogRepository) Create(ctx context.Context, log *model.AdmLoginLog) error {
	if err := log.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm login log: %w", err)
	}
	return r.Insert(ctx, log)
}
