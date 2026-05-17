package session

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestStoreCreatePrincipalStoresJSON(t *testing.T) {
	cache := newFakeCache()
	store := NewStore(cache, time.Minute)

	token, err := store.CreatePrincipal(context.Background(), Principal{
		PrincipalType:   PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        TerminalPublish,
		Phone:           "13800000000",
		RoleCodes:       []string{"super_admin"},
		PermissionCodes: []string{"house.manage"},
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token")
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(cache.values[keyPrefix+token]), &raw); err != nil {
		t.Fatalf("session payload must be json: %v payload=%q", err, cache.values[keyPrefix+token])
	}
	if raw["principal_type"] != PrincipalTypeStaff || raw["terminal"] != TerminalPublish {
		t.Fatalf("unexpected payload: %#v", raw)
	}

	principal, err := store.GetPrincipal(context.Background(), token)
	if err != nil {
		t.Fatalf("get principal: %v", err)
	}
	if principal == nil || principal.PrincipalID != "staff-id" || principal.Phone != "13800000000" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}

func TestStoreCreateWritesMiniappUserPrincipal(t *testing.T) {
	cache := newFakeCache()
	store := NewStore(cache, time.Minute)

	token, err := store.CreateMiniappUser(context.Background(), "user-id", "13800000000")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	principal, err := store.GetPrincipal(context.Background(), token)
	if err != nil {
		t.Fatalf("get principal: %v", err)
	}
	if principal == nil || principal.PrincipalType != PrincipalTypeUser || principal.Terminal != TerminalMiniapp {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if principal.Phone != "13800000000" {
		t.Fatalf("unexpected phone: %q", principal.Phone)
	}
	userID, err := store.GetUserID(context.Background(), token)
	if err != nil {
		t.Fatalf("get user id: %v", err)
	}
	if userID != "user-id" {
		t.Fatalf("unexpected userID: %q", userID)
	}
}

func TestPrincipalContextRoundTrip(t *testing.T) {
	principal := Principal{
		PrincipalType:   PrincipalTypeStaff,
		PrincipalID:     "staff-id",
		Terminal:        TerminalPublish,
		RoleCodes:       []string{"super_admin"},
		PermissionCodes: []string{"house.manage"},
	}
	ctx := ContextWithPrincipal(context.Background(), principal)
	got, ok := PrincipalFromContext(ctx)
	if !ok {
		t.Fatalf("expected principal from context")
	}
	if got.PrincipalID != principal.PrincipalID || got.Terminal != TerminalPublish {
		t.Fatalf("unexpected principal: %#v", got)
	}
}

type fakeCache struct {
	values map[string]string
	ttls   map[string]time.Duration
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: map[string]string{}, ttls: map[string]time.Duration{}}
}

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	return f.values[key], nil
}

func (f *fakeCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.values[key] = value
	f.ttls[key] = ttl
	return nil
}

func (f *fakeCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.values, key)
		delete(f.ttls, key)
	}
	return nil
}

func TestStoreGetPrincipalRefreshesTTL(t *testing.T) {
	cache := newFakeCache()
	store := NewStore(cache, 2*time.Hour)

	token, err := store.CreatePrincipal(context.Background(), Principal{
		PrincipalType: PrincipalTypeLandlord,
		PrincipalID:   "landlord-id",
		Terminal:      TerminalPublish,
		Phone:         "18002584637",
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	key := keyPrefix + token
	if got := cache.ttls[key]; got != 2*time.Hour {
		t.Fatalf("unexpected initial ttl: %v", got)
	}

	cache.ttls[key] = time.Minute
	principal, err := store.GetPrincipal(context.Background(), token)
	if err != nil {
		t.Fatalf("get principal: %v", err)
	}
	if principal == nil || principal.PrincipalID != "landlord-id" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
	if got := cache.ttls[key]; got != 2*time.Hour {
		t.Fatalf("ttl not refreshed, got %v", got)
	}
}

func TestStoreInvalidatePrincipalRejectsOldSession(t *testing.T) {
	cache := newFakeCache()
	store := NewStore(cache, time.Hour)
	principal := Principal{
		PrincipalType: PrincipalTypeStaff,
		PrincipalID:   "staff-id",
		Terminal:      TerminalAdmin,
	}
	token, err := store.CreatePrincipal(context.Background(), principal)
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}
	if got, err := store.GetPrincipal(context.Background(), token); err != nil || got == nil {
		t.Fatalf("expected session before invalidation, principal=%#v err=%v", got, err)
	}

	if err := store.InvalidatePrincipal(context.Background(), principal); err != nil {
		t.Fatalf("invalidate principal: %v", err)
	}
	if got, err := store.GetPrincipal(context.Background(), token); err != nil || got != nil {
		t.Fatalf("expected old session invalidated, principal=%#v err=%v", got, err)
	}

	newToken, err := store.CreatePrincipal(context.Background(), principal)
	if err != nil {
		t.Fatalf("create new principal: %v", err)
	}
	if got, err := store.GetPrincipal(context.Background(), newToken); err != nil || got == nil || got.PrincipalID != principal.PrincipalID {
		t.Fatalf("expected new session valid, principal=%#v err=%v", got, err)
	}
}
