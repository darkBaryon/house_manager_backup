package v1

import (
	"log/slog"

	"house-manager/internal/handler"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

// PingFunc 健康检查函数签名
type PingFunc func() error

// HealthHandler 健康检查处理器
type HealthHandler struct {
	MongoPing PingFunc
	RedisPing PingFunc
}

// RegisterRoutes 注册健康检查路由
func (h *HealthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/health/check", h.Check)
}

// Check 检查基础设施连通性
func (h *HealthHandler) Check(c *gin.Context) {
	type status struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	var statuses []status
	var hasError bool

	if err := h.MongoPing(); err != nil {
		hasError = true
		statuses = append(statuses, status{Name: "mongodb", Status: "down", Error: err.Error()})
		slog.Error("health check: mongodb down", "error", err)
	} else {
		statuses = append(statuses, status{Name: "mongodb", Status: "up"})
	}

	if err := h.RedisPing(); err != nil {
		hasError = true
		statuses = append(statuses, status{Name: "redis", Status: "down", Error: err.Error()})
		slog.Error("health check: redis down", "error", err)
	} else {
		statuses = append(statuses, status{Name: "redis", Status: "up"})
	}

	if hasError {
		response.Err(c, errcode.InternalError)
		return
	}
	response.Success(c, statuses)
}

var _ handler.RouteRegistrar = (*HealthHandler)(nil)
