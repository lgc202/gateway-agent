package chat

import (
	"time"

	chatservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/chat"
)

// ChatResp 是对话响应
type ChatResp struct {
	ID            uint64    `json:"id"`
	ModelConfigID *uint64   `json:"model_config_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MessageResp 是对话消息响应
type MessageResp struct {
	ID        uint64                  `json:"id"`
	ChatID    uint64                  `json:"chat_id"`
	Role      chatservice.MessageRole `json:"role"`
	Content   string                  `json:"content"`
	CreatedAt time.Time               `json:"created_at"`
}

// MessageListResp 是对话消息列表响应
type MessageListResp struct {
	Items       []MessageResp `json:"items"`
	NextAfterID uint64        `json:"next_after_id"`
}

// TextDeltaResp 是模型新增文本片段的 SSE 响应
type TextDeltaResp struct {
	Content string `json:"content"`
}

// NewChatResp 根据用例层对话构造 HTTP 响应
func NewChatResp(value chatservice.Chat) ChatResp {
	return ChatResp{
		ID:            value.ID,
		ModelConfigID: value.ModelConfigID,
		CreatedAt:     value.CreatedAt,
		UpdatedAt:     value.UpdatedAt,
	}
}

// NewMessageResp 根据用例层消息构造 HTTP 响应
func NewMessageResp(value chatservice.Message) MessageResp {
	return MessageResp{
		ID:        value.ID,
		ChatID:    value.ChatID,
		Role:      value.Role,
		Content:   value.Content,
		CreatedAt: value.CreatedAt,
	}
}

// NewMessageListResp 构造消息列表响应
func NewMessageListResp(messages []chatservice.Message, afterID uint64) MessageListResp {
	resp := MessageListResp{
		Items:       make([]MessageResp, 0, len(messages)),
		NextAfterID: afterID,
	}
	for _, message := range messages {
		resp.Items = append(resp.Items, NewMessageResp(message))
		resp.NextAfterID = message.ID
	}
	return resp
}

// NewTextDeltaResp 构造模型新增文本片段响应
func NewTextDeltaResp(content string) TextDeltaResp {
	return TextDeltaResp{Content: content}
}
