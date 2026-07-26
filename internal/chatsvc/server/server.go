// Package server 负责组装并运行 Chat Service 的 HTTP Server
package server

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/chatsvc/config"
	chathandler "github.com/lgc202/gateway-agent/internal/chatsvc/handler/chat"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/health"
	modelconfighandler "github.com/lgc202/gateway-agent/internal/chatsvc/handler/modelconfig"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/response"
	"github.com/lgc202/gateway-agent/internal/chatsvc/middleware"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

// Server 管理 HTTP Server 和数据库连接池的生命周期
type Server struct {
	httpServer *http.Server
	db         *sql.DB
}

// New 创建并注册 Chat Service 的全部当前路由
func New(
	cfg *config.Config,
	db *sql.DB,
	chatHandler *chathandler.Handler,
	modelConfigHandler *modelconfighandler.Handler,
	healthHandler *health.Handler,
) *Server {
	if _, exists := os.LookupEnv(gin.EnvGinMode); !exists {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery())
	router.NoRoute(func(ctx *gin.Context) {
		response.WriteFailure(ctx, http.StatusNotFound, errorsx.CodeInvalidRequest, "请求路径不存在")
	})
	router.NoMethod(func(ctx *gin.Context) {
		response.WriteFailure(ctx, http.StatusMethodNotAllowed, errorsx.CodeInvalidRequest, "请求方法不允许")
	})

	healthHandler.Register(router)
	api := router.Group("/api/v1")
	chatHandler.Register(api)
	modelConfigHandler.Register(api)

	return &Server{
		httpServer: &http.Server{
			Addr:              net.JoinHostPort(cfg.HTTP.Host, strconv.Itoa(cfg.HTTP.Port)),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		db: db,
	}
}

// Run 启动 HTTP Server，并把非正常退出错误返回给入口层
func (s *Server) Run() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown 在给定 Context 内优雅停止 HTTP Server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Close 关闭 Server 持有的数据库连接池
func (s *Server) Close() error {
	return s.db.Close()
}
