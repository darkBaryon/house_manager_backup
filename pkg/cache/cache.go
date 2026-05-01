package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"time"

	"golang.org/x/sync/singleflight"
)

// ErrNil 缓存 key 不存在
var ErrNil = errors.New("cache: key not found")

// Client 缓存操作依赖的最小 Redis 接口（与 go-redis 解耦）
type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Cache 通用缓存（Redis 通过构造器注入）
type Cache struct {
	client Client
	sf     singleflight.Group
}

// New 创建缓存实例
func New(client Client) *Cache {
	return &Cache{client: client}
}

// Get 从缓存获取值并反序列化到 out
func (c *Cache) Get(ctx context.Context, key string, out any) error {
	val, err := c.client.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), out)
}

// Set 将值序列化后写入缓存
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, string(data), ttl)
}

// Delete 删除缓存 key
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...)
}

// GetSet Cache-Aside with singleflight：并发同一 key 只穿透一次
// Go 不支持泛型方法，因此用包级泛型函数
func GetSet[T any](c *Cache, ctx context.Context, key string, ttl time.Duration, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	var result T

	// 快速路径：尝试缓存
	if err := c.Get(ctx, key, &result); err == nil {
		return result, nil
	}

	// 慢路径：singleflight 防穿透
	v, err, _ := c.sf.Do(key, func() (any, error) {
		// singleflight 内二次检查缓存
		val, err := c.client.Get(ctx, key)
		if err == nil {
			return []byte(val), nil
		}

		data, err := fn(ctx)
		if err != nil {
			return nil, err
		}

		raw, merr := json.Marshal(data)
		if merr != nil {
			return nil, merr
		}

		_ = c.client.Set(ctx, key, string(raw), jitter(ttl))
		return raw, nil
	})
	if err != nil {
		return zero, err
	}

	if err := json.Unmarshal(v.([]byte), &result); err != nil {
		return zero, err
	}
	return result, nil
}

// jitter 给 TTL 加 ±10% 随机抖动，防止大量 key 同时过期引发雪崩
func jitter(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return ttl
	}
	delta := int64(float64(ttl) * 0.1)
	return ttl + time.Duration(rand.Int64N(2*delta+1)-delta)
}
