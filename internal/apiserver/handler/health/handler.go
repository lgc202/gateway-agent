// Package health 提供进程存活和依赖就绪检查
package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/apiserver/handler/response"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const readinessTimeout = time.Second

type statusResponse struct {
	Status string `json:"status"`
}

// Handler 处理健康检查请求
type Handler struct {
	db *sql.DB
}

// New 创建健康检查 Handler
func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// Register 注册健康检查路由
func (h *Handler) Register(router *gin.Engine) {
	router.GET("/healthz", h.healthz)
	router.GET("/readyz", h.readyz)
}

func (h *Handler) healthz(ctx *gin.Context) {
	response.WriteSuccess(ctx, http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) readyz(ctx *gin.Context) {
	pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), readinessTimeout)
	defer cancel()
	if err := h.db.PingContext(pingCtx); err != nil {
		response.WriteError(ctx, errorsx.Wrap(errorsx.CodeDependencyUnavailable, "ping MySQL", err))
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, statusResponse{Status: "ready"})
}
