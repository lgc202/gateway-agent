// Package errorsx 扩展标准错误能力，定义跨协议使用的稳定应用错误
package errorsx

import "fmt"

// Code 是可用于协议和审计的稳定错误码
type Code string

const (
	CodeOK                    Code = "OK"
	CodeInvalidRequest        Code = "INVALID_REQUEST"
	CodeChatNotFound          Code = "CHAT_NOT_FOUND"
	CodeInvalidMessageContent Code = "INVALID_MESSAGE_CONTENT"
	CodeDependencyUnavailable Code = "DEPENDENCY_UNAVAILABLE"
	CodeInternal              Code = "INTERNAL_ERROR"
)

// UserError 表示错误码和消息可以直接返回给 API 调用方
type UserError struct {
	Code    Code
	Message string
}

// NewUser 创建可安全展示的应用错误
func NewUser(code Code, message string) *UserError {
	return &UserError{Code: code, Message: message}
}

func (e *UserError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

// Error 保存稳定错误码和仅供服务端诊断的内部原因
type Error struct {
	Code    Code
	Message string
	Cause   error
}

// New 创建不应直接向调用方展示的应用错误
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap 创建包含底层原因的应用错误
func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

// Unwrap 支持 errors.Is 和 errors.As 分类
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
