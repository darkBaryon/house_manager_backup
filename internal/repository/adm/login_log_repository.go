package adm

import (
	"context"
	"fmt"
	"strings"

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

func (r *LoginLogRepository) Create(ctx context.Context, logEntry *authmodel.AdmLoginLog) error {
	normalizeLoginLog(logEntry)
	if err := logEntry.ValidateForCreate(); err != nil {
		return fmt.Errorf("create admin login log: %w", err)
	}
	return r.Insert(ctx, logEntry)
}

func normalizeLoginLog(logEntry *authmodel.AdmLoginLog) {
	if logEntry == nil {
		return
	}
	logEntry.LoginIP = strings.TrimSpace(logEntry.LoginIP)
	logEntry.UserAgent = strings.TrimSpace(logEntry.UserAgent)
	logEntry.Remark = strings.TrimSpace(logEntry.Remark)
}
