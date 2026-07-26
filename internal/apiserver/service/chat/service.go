// Package chat 实现 Chat 和 Message 用例
package chat

import (
	"context"
	"strings"
	"time"

	mysqlstore "github.com/lgc202/gateway-agent/internal/apiserver/store/mysql"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const maxMessageContentBytes = 60 * 1024

// MessageRole 表示消息在对话中的角色
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

// Chat 是 Chat API 使用的对话模型
type Chat struct {
	ID            uint64
	ModelConfigID *uint64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Message 是 Chat API 使用的消息模型
type Message struct {
	ID        uint64
	ChatID    uint64
	Role      MessageRole
	Content   string
	CreatedAt time.Time
}

// Service 承载 Chat 和 Message 用例
type Service struct {
	store *mysqlstore.Store
}

// New 创建 Chat Service
func New(store *mysqlstore.Store) *Service {
	return &Service{store: store}
}

// Valid 判断消息角色是否属于当前协议定义
func (r MessageRole) Valid() bool {
	return r == MessageRoleUser || r == MessageRoleAssistant
}

// CreateChat 创建对话；modelConfigID 为空时使用系统默认模型配置
func (s *Service) CreateChat(ctx context.Context, modelConfigID *uint64) (Chat, error) {
	record, err := s.store.CreateChat(ctx, modelConfigID)
	if err != nil {
		return Chat{}, err
	}
	return toChat(record), nil
}

// GetChat 查询指定对话
func (s *Service) GetChat(ctx context.Context, chatID uint64) (Chat, error) {
	record, err := s.store.GetChat(ctx, chatID)
	if err != nil {
		return Chat{}, err
	}
	return toChat(record), nil
}

// AppendUserMessage 向对话追加一条原始用户消息
func (s *Service) AppendUserMessage(ctx context.Context, chatID uint64, content string) (Message, error) {
	if strings.TrimSpace(content) == "" || len(content) > maxMessageContentBytes {
		return Message{}, errorsx.NewUser(
			errorsx.CodeInvalidMessageContent,
			"消息内容不能为空且不能超过 60 KiB",
		)
	}

	record, err := s.store.AppendMessage(ctx, chatID, string(MessageRoleUser), content)
	if err != nil {
		return Message{}, err
	}

	return toMessage(record), nil
}

// ListMessages 按游标正序查询对话消息
func (s *Service) ListMessages(ctx context.Context, chatID, afterID uint64, limit int32) ([]Message, error) {
	if limit < 1 || limit > 200 {
		return nil, errorsx.NewUser(errorsx.CodeInvalidRequest, "limit 必须在 1 到 200 之间")
	}

	records, err := s.store.ListMessages(ctx, chatID, afterID, limit)
	if err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(records))
	for _, record := range records {
		messages = append(messages, toMessage(record))
	}

	return messages, nil
}

func toChat(record mysqlstore.Chat) Chat {
	return Chat{
		ID:            record.ID,
		ModelConfigID: record.ModelConfigID,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

func toMessage(record mysqlstore.Message) Message {
	return Message{
		ID:        record.ID,
		ChatID:    record.ChatID,
		Role:      MessageRole(record.Role),
		Content:   record.Content,
		CreatedAt: record.CreatedAt,
	}
}
