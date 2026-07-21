// Package middleware 提供 apiserver 使用的 HTTP 中间件
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/apiserver/handler/response"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
	"github.com/lgc202/gateway-agent/internal/pkg/requestctx"
)

const (
	requestIDHeader    = "X-Request-ID"
	requestIDBytes     = 16
	maxRequestIDLength = 128
)

// RequestID 确保每个请求都携带可贯穿日志和响应的 Request ID
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID, valid := requestIDFromHeader(ctx.Request.Header)
		if !valid {
			var err error
			requestID, err = newRequestID()
			if err != nil {
				response.WriteError(ctx, errorsx.Wrap(errorsx.CodeInternal, "generate request id", err))
				ctx.Abort()
				return
			}
		}

		ctx.Header(requestIDHeader, requestID)
		ctx.Request = ctx.Request.WithContext(requestctx.WithRequestID(ctx.Request.Context(), requestID))
		ctx.Next()
	}
}

func requestIDFromHeader(header http.Header) (string, bool) {
	var values []string
	for name, currentValues := range header {
		if strings.EqualFold(name, requestIDHeader) {
			values = append(values, currentValues...)
		}
	}
	if len(values) != 1 || !validRequestID(values[0]) {
		return "", false
	}
	return values[0], true
}

func validRequestID(value string) bool {
	if len(value) < 1 || len(value) > maxRequestIDLength {
		return false
	}
	for i := range len(value) {
		character := value[i]
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func newRequestID() (string, error) {
	value := make([]byte, requestIDBytes)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
