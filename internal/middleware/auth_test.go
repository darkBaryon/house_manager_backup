package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

func TestPublishAuthRejectsMiniappSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewStore(newMiddlewareFakeCache(), time.Minute)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeUser,
		PrincipalID:   "user-id",
		Terminal:      session.TerminalMiniapp,
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	router := gin.New()
	router.POST("/protected", PublishAuth(store), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestPublishAuthAcceptsPublishSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewStore(newMiddlewareFakeCache(), time.Minute)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   "staff-id",
		Terminal:      session.TerminalPublish,
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	router := gin.New()
	router.POST("/protected", PublishAuth(store), func(c *gin.Context) {
		value, ok := c.Get(ContextPrincipal)
		if !ok {
			t.Fatalf("expected principal in context")
		}
		principal := value.(session.Principal)
		if principal.PrincipalID != "staff-id" {
			t.Fatalf("unexpected principal: %#v", principal)
		}
		requestPrincipal, ok := session.PrincipalFromContext(c.Request.Context())
		if !ok || requestPrincipal.PrincipalID != "staff-id" || requestPrincipal.Terminal != session.TerminalPublish {
			t.Fatalf("unexpected request context principal: %#v ok=%v", requestPrincipal, ok)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMiniappAuthRejectsPublishSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewStore(newMiddlewareFakeCache(), time.Minute)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   "staff-id",
		Terminal:      session.TerminalPublish,
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	router := gin.New()
	router.POST("/protected", MiniappAuth(store), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMiniappAuthAcceptsMiniappSessionAndSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := session.NewStore(newMiddlewareFakeCache(), time.Minute)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeUser,
		PrincipalID:   "user-id",
		Terminal:      session.TerminalMiniapp,
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	router := gin.New()
	router.POST("/protected", MiniappAuth(store), func(c *gin.Context) {
		userID, ok := c.Get(ContextUserID)
		if !ok || userID != "user-id" {
			t.Fatalf("unexpected user id: %#v ok=%v", userID, ok)
		}
		requestPrincipal, ok := session.PrincipalFromContext(c.Request.Context())
		if !ok || requestPrincipal.PrincipalID != "user-id" || requestPrincipal.Terminal != session.TerminalMiniapp {
			t.Fatalf("unexpected request context principal: %#v ok=%v", requestPrincipal, ok)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d body=%s", resp.Code, resp.Body.String())
	}
}

type middlewareFakeCache struct {
	values map[string]string
	ttls   map[string]time.Duration
}

func newMiddlewareFakeCache() *middlewareFakeCache {
	return &middlewareFakeCache{values: map[string]string{}, ttls: map[string]time.Duration{}}
}

func (f *middlewareFakeCache) Get(ctx context.Context, key string) (string, error) {
	return f.values[key], nil
}

func (f *middlewareFakeCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.values[key] = value
	f.ttls[key] = ttl
	return nil
}

func (f *middlewareFakeCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.values, key)
		delete(f.ttls, key)
	}
	return nil
}

func TestPublishAuthRefreshesSessionTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newMiddlewareFakeCache()
	store := session.NewStore(cache, 2*time.Hour)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeLandlord,
		PrincipalID:   "landlord-id",
		Terminal:      session.TerminalPublish,
	})
	if err != nil {
		t.Fatalf("create principal: %v", err)
	}

	key := "hs:sess:" + token
	cache.ttls[key] = time.Minute

	router := gin.New()
	router.POST("/protected", PublishAuth(store), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d body=%s", resp.Code, resp.Body.String())
	}
	if got := cache.ttls[key]; got != 2*time.Hour {
		t.Fatalf("expected ttl refreshed to 2h, got %v", got)
	}
}
