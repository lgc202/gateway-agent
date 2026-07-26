// Package chat 定义 Chat 和 Message HTTP 接口的数据结构
package chat

import (
	"strings"

	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const maxMessageContentBytes = 60 * 1024

// GetChatReq 是查询对话的请求
type GetChatReq struct {
	ChatID uint64 `uri:"chat_id"`
}

// CreateChatReq 是创建对话的请求
type CreateChatReq struct {
	ModelConfigID *uint64 `json:"model_config_id"`
}

// SendMessageReq 是发送用户消息的请求
type SendMessageReq struct {
	ChatID  uint64 `uri:"chat_id"`
	Content string `json:"content"`
}

// ListMessagesReq 是查询对话消息的请求
type ListMessagesReq struct {
	ChatID  uint64 `uri:"chat_id"`
	AfterID uint64 `form:"after_id"`
	Limit   int32  `form:"limit,default=50"`
}

// Validate 校验创建对话请求
func (r CreateChatReq) Validate() error {
	if r.ModelConfigID != nil && *r.ModelConfigID == 0 {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "model_config_id 必须是大于 0 的整数")
	}
	return nil
}

// Validate 校验查询对话请求
func (r GetChatReq) Validate() error {
	return validateChatID(r.ChatID)
}

// Validate 校验用户消息内容
func (r SendMessageReq) Validate() error {
	if err := validateChatID(r.ChatID); err != nil {
		return err
	}
	if strings.TrimSpace(r.Content) == "" || len(r.Content) > maxMessageContentBytes {
		return errorsx.NewUser(
			errorsx.CodeInvalidMessageContent,
			"消息内容不能为空且不能超过 60 KiB",
		)
	}
	return nil
}

// Validate 校验消息查询参数
func (r ListMessagesReq) Validate() error {
	if err := validateChatID(r.ChatID); err != nil {
		return err
	}
	if r.Limit < 1 || r.Limit > 200 {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "limit 必须在 1 到 200 之间")
	}
	return nil
}

func validateChatID(chatID uint64) error {
	if chatID == 0 {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "chat_id 必须是大于 0 的整数")
	}
	return nil
}
