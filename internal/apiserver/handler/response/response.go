// Package response 定义 apiserver 的统一 HTTP 响应
package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
	"github.com/lgc202/gateway-agent/internal/pkg/requestctx"
)

// Envelope 是所有 JSON 响应的统一外层结构
type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// WriteSuccess 写入成功响应
func WriteSuccess(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, Envelope{
		Code:    string(errorsx.CodeOK),
		Message: "success",
		Data:    data,
	})
}

// WriteFailure 写入统一且不包含内部细节的失败响应
func WriteFailure(ctx *gin.Context, status int, code errorsx.Code, message string) {
	ctx.JSON(status, Envelope{
		Code:    string(code),
		Message: message,
		Data:    nil,
	})
}

// WriteError 根据错误类型写入安全的错误响应
func WriteError(ctx *gin.Context, err error) {
	if userErr, ok := errors.AsType[*errorsx.UserError](err); ok {
		WriteFailure(ctx, userErrorStatus(userErr.Code), userErr.Code, userErr.Message)
		return
	}

	status := http.StatusInternalServerError
	message := "服务器内部错误"
	if appErr, ok := errors.AsType[*errorsx.Error](err); ok && appErr.Code == errorsx.CodeDependencyUnavailable {
		status = http.StatusServiceUnavailable
		message = "服务暂时不可用"
	}

	slog.ErrorContext(
		ctx.Request.Context(),
		"HTTP request failed",
		"request_id", requestctx.RequestID(ctx.Request.Context()),
		"method", ctx.Request.Method,
		"path", ctx.Request.URL.Path,
		"error", err,
	)
	WriteFailure(ctx, status, errorsx.CodeInternal, message)
}

func userErrorStatus(code errorsx.Code) int {
	switch code {
	case errorsx.CodeChatNotFound, errorsx.CodeModelConfigNotFound:
		return http.StatusNotFound
	case errorsx.CodeModelConfigNameConflict, errorsx.CodeModelConfigInUse:
		return http.StatusConflict
	case errorsx.CodeInvalidRequest, errorsx.CodeInvalidMessageContent:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
