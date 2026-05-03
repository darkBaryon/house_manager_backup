package repository

import (
	"context"
	"fmt"

	dbmongo "house-manager/pkg/database/mongo"
)

// TxManager MongoDB 事务管理器
type TxManager struct {
	client *dbmongo.Client
}

// NewTxManager 创建事务管理器
func NewTxManager(client *dbmongo.Client) *TxManager {
	return &TxManager{client: client}
}

// RunInTransaction 在 MongoDB 事务中执行 fn
// fn 内的 repository 操作需使用传入的 ctx 以加入同一事务
func (m *TxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := m.client.RunInTransaction(ctx, fn); err != nil {
		return fmt.Errorf("transaction: %w", err)
	}
	return nil
}
