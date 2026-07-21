// Package requestctx 保存与传输框架无关的请求身份信息
package requestctx

import "context"

type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// WithRequestID 返回带 Request ID 的 Context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID 从 Context 读取 Request ID
func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}
