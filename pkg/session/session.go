package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"house-manager/pkg/cache"
)

const keyPrefix = "session:"

// Store Redis session 存储
type Store struct {
	client cache.Client
	ttl    time.Duration
}

// NewStore 创建 session store（复用 cache.Client 接口）
func NewStore(client cache.Client, ttl time.Duration) *Store {
	return &Store{client: client, ttl: ttl}
}

// Create 创建 session，返回 token
func (s *Store) Create(ctx context.Context, userId string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	if err := s.client.Set(ctx, keyPrefix+token, userId, s.ttl); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}
	return token, nil
}

// Get 根据 token 获取 userId，空字符串表示 session 不存在
func (s *Store) Get(ctx context.Context, token string) (string, error) {
	val, err := s.client.Get(ctx, keyPrefix+token)
	if err != nil {
		return "", nil
	}
	return val, nil
}

// Delete 删除 session
func (s *Store) Delete(ctx context.Context, token string) error {
	return s.client.Del(ctx, keyPrefix+token)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
