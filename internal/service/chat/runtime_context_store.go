package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dbredis "house-manager/pkg/database/redis"

	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const runtimeContextKeyPrefix = "hs:chat:runtime:"
const defaultRuntimeContextTTL = 24 * time.Hour

type RuntimeContextStore struct {
	redis *dbredis.Client
	ttl   time.Duration
}

func NewRuntimeContextStore(redis *dbredis.Client, ttl time.Duration) *RuntimeContextStore {
	if ttl <= 0 {
		ttl = defaultRuntimeContextTTL
	}
	return &RuntimeContextStore{redis: redis, ttl: ttl}
}

func (s *RuntimeContextStore) Get(ctx context.Context, sessionID bson.ObjectID) (map[string]any, error) {
	if sessionID.IsZero() {
		return nil, fmt.Errorf("get runtime context: sessionID is required")
	}
	if s == nil || s.redis == nil {
		return nil, fmt.Errorf("get runtime context: redis client is nil")
	}
	raw, err := s.redis.Get(ctx, runtimeContextKey(sessionID))
	if err == goredis.Nil {
		return defaultRuntimeContext(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get runtime context: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("decode runtime context: %w", err)
	}
	if len(out) == 0 {
		return defaultRuntimeContext(), nil
	}
	return out, nil
}

func (s *RuntimeContextStore) Set(ctx context.Context, sessionID bson.ObjectID, runtimeContext map[string]any) error {
	if sessionID.IsZero() {
		return fmt.Errorf("set runtime context: sessionID is required")
	}
	if s == nil || s.redis == nil {
		return fmt.Errorf("set runtime context: redis client is nil")
	}
	if runtimeContext == nil {
		runtimeContext = defaultRuntimeContext()
	}
	body, err := json.Marshal(runtimeContext)
	if err != nil {
		return fmt.Errorf("encode runtime context: %w", err)
	}
	if err := s.redis.Set(ctx, runtimeContextKey(sessionID), string(body), s.ttl); err != nil {
		return fmt.Errorf("set runtime context: %w", err)
	}
	return nil
}

func runtimeContextKey(sessionID bson.ObjectID) string {
	return runtimeContextKeyPrefix + sessionID.Hex()
}

func defaultRuntimeContext() map[string]any {
	return map[string]any{
		"dialogue_state":                "IDLE",
		"requirement":                   map[string]any{},
		"last_recommended_listing_refs": []any{},
		"focused_listing_ref":           nil,
	}
}
