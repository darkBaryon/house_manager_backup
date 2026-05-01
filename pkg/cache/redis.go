package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter 将 go-redis 客户端适配为 cache.Client 接口
// 支持 *redis.Client 和 *redis.ClusterClient（均实现 redis.Cmdable）
type RedisAdapter struct {
	client redis.Cmdable
}

// NewRedisAdapter 创建 Redis 适配器（注入 go-redis 客户端）
func NewRedisAdapter(client redis.Cmdable) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	val, err := a.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNil
	}
	return val, err
}

func (a *RedisAdapter) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return a.client.Set(ctx, key, value, ttl).Err()
}

func (a *RedisAdapter) Del(ctx context.Context, keys ...string) error {
	return a.client.Del(ctx, keys...).Err()
}
