package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TxManager MongoDB 事务管理器
type TxManager struct {
	client *mongo.Client
}

// NewTxManager 创建事务管理器
func NewTxManager(client *mongo.Client) *TxManager {
	return &TxManager{client: client}
}

// RunInTransaction 在 MongoDB 事务中执行 fn
// fn 内的 repository 操作需使用传入的 ctx 以加入同一事务
func (m *TxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		return nil, fn(txCtx)
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}
	return nil
}
