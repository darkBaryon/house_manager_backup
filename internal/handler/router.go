package handler

import "github.com/gin-gonic/gin"

// RouteRegistrar 路由注册接口，各 handler 实现此接口自注册路由
type RouteRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}
