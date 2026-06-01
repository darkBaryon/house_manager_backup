package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInternalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(InternalAuth("token-1"))
	router.POST("/internal/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/ping", nil)
	req.Header.Set("Internal-Token", "token-1")
	req.Header.Set("Request-ID", "req-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
	if resp.Header().Get("Request-ID") != "req-1" {
		t.Fatalf("request id header = %q", resp.Header().Get("Request-ID"))
	}
}

func TestInternalAuthRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(InternalAuth("token-1"))
	router.POST("/internal/ping", func(c *gin.Context) {
		t.Fatal("handler should not run")
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/ping", nil)
	req.Header.Set("Internal-Token", "wrong")
	req.Header.Set("Request-ID", "req-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
}

func TestInternalAuthRequiresRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(InternalAuth("token-1"))
	router.POST("/internal/ping", func(c *gin.Context) {
		t.Fatal("handler should not run")
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/ping", nil)
	req.Header.Set("Internal-Token", "token-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", resp.Code, resp.Body.String())
	}
}
