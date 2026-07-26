// Package agent 封装 Eino Agent，并向上层提供稳定的消息与流式事件。
package agent

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// MessageRole 表示持久消息进入 Agent 时的角色。
type MessageRole string

const (
	// MessageRoleUser 表示用户输入。
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant 表示 Agent 已完成的历史回复。
	MessageRoleAssistant MessageRole = "assistant"
)

// Message 是 Chat Service 传给 Agent 的历史消息。
type Message struct {
	Role    MessageRole
	Content string
}

func toEinoMessages(messages []Message) ([]*schema.Message, error) {
	result := make([]*schema.Message, 0, len(messages))
	for _, message := range messages {
		switch message.Role {
		case MessageRoleUser:
			result = append(result, schema.UserMessage(message.Content))
		case MessageRoleAssistant:
			result = append(result, schema.AssistantMessage(message.Content, nil))
		default:
			return nil, fmt.Errorf("unsupported message role %q", message.Role)
		}
	}

	return result, nil
}
