package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/apiserver/handler/response"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
	"github.com/lgc202/gateway-agent/internal/pkg/requestctx"
)

// Recovery 捕获未处理的 panic，避免记录请求头和其他敏感信息
func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(
					ctx.Request.Context(),
					"HTTP request panic recovered",
					"request_id", requestctx.RequestID(ctx.Request.Context()),
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				if !ctx.Writer.Written() {
					response.WriteFailure(ctx, http.StatusInternalServerError, errorsx.CodeInternal, "服务器内部错误")
				}
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}
