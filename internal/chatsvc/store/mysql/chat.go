package mysql

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql/sqlc"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// Chat 是 Store 返回给用例层的对话记录
type Chat struct {
	ID            uint64
	ModelConfigID *uint64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Message 是 Store 返回给用例层的消息记录
type Message struct {
	ID        uint64
	ChatID    uint64
	Role      string
	Content   string
	CreatedAt time.Time
}

// CreateChat 创建对话，并保存本次对话固定使用的模型配置
func (s *Store) CreateChat(ctx context.Context, modelConfigID *uint64) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	if modelConfigID != nil {
		if _, err := s.GetModelConfig(ctx, *modelConfigID); err != nil {
			return Chat{}, err
		}
	}

	result, err := s.queries.InsertChat(ctx, modelConfigID)
	if err != nil {
		return Chat{}, databaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Chat{}, databaseError(err)
	}
	if id <= 0 {
		return Chat{}, errorsx.New(errorsx.CodeInternal, "invalid chat id returned by database")
	}

	return s.GetChat(ctx, uint64(id))
}

// GetChat 按 ID 查询对话
func (s *Store) GetChat(ctx context.Context, chatID uint64) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	record, err := s.queries.GetChat(ctx, chatID)
	if err != nil {
		return Chat{}, mapChatQueryError(err)
	}

	return toChat(record), nil
}

// AppendMessage 原子地写入消息并更新对话最后活跃时间
func (s *Store) AppendMessage(ctx context.Context, chatID uint64, role, content string) (Message, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	var message Message
	err := s.WithinTransaction(ctx, func(ctx context.Context, queries *sqlc.Queries) error {
		// 锁定 Chat 行，使同一 Chat 的 Message ID 分配顺序与提交顺序一致，防止游标漏消息。
		if _, err := queries.GetChatForUpdate(ctx, chatID); err != nil {
			return mapChatQueryError(err)
		}

		result, err := queries.InsertChatMessage(ctx, sqlc.InsertChatMessageParams{
			ChatID:  chatID,
			Role:    role,
			Content: content,
		})
		if err != nil {
			return databaseError(err)
		}

		messageID, err := result.LastInsertId()
		if err != nil {
			return databaseError(err)
		}
		if messageID <= 0 {
			return errorsx.New(errorsx.CodeInternal, "invalid message id returned by database")
		}

		rowsAffected, err := queries.TouchChat(ctx, chatID)
		if err != nil {
			return databaseError(err)
		}
		if rowsAffected != 1 {
			return errorsx.New(errorsx.CodeInternal, "unexpected chat update result")
		}

		record, err := queries.GetChatMessage(ctx, uint64(messageID))
		if err != nil {
			return databaseError(err)
		}
		message = toMessage(record)

		return nil
	})
	if err != nil {
		return Message{}, err
	}

	return message, nil
}

// ListMessages 按消息 ID 游标正序查询对话消息
func (s *Store) ListMessages(ctx context.Context, chatID, afterID uint64, limit int32) ([]Message, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	if _, err := s.GetChat(ctx, chatID); err != nil {
		return nil, err
	}

	records, err := s.queries.ListChatMessagesAfter(ctx, sqlc.ListChatMessagesAfterParams{
		ChatID: chatID,
		ID:     afterID,
		Limit:  limit,
	})
	if err != nil {
		return nil, databaseError(err)
	}

	messages := make([]Message, 0, len(records))
	for _, record := range records {
		messages = append(messages, toMessage(record))
	}

	return messages, nil
}

// ListRecentMessages 查询最近的消息，并按对话发生顺序返回
func (s *Store) ListRecentMessages(ctx context.Context, chatID uint64, limit int32) ([]Message, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	records, err := s.queries.ListRecentChatMessages(ctx, sqlc.ListRecentChatMessagesParams{
		ChatID: chatID,
		Limit:  limit,
	})
	if err != nil {
		return nil, databaseError(err)
	}

	messages := make([]Message, 0, len(records))
	for _, record := range slices.Backward(records) {
		messages = append(messages, toMessage(record))
	}

	return messages, nil
}

func mapChatQueryError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errorsx.NewUser(errorsx.CodeChatNotFound, "Chat 不存在")
	}
	return databaseError(err)
}

func toChat(record sqlc.Chat) Chat {
	return Chat{
		ID:            record.ID,
		ModelConfigID: record.ModelConfigID,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

func toMessage(record sqlc.ChatMessage) Message {
	return Message{
		ID:        record.ID,
		ChatID:    record.ChatID,
		Role:      record.Role,
		Content:   record.Content,
		CreatedAt: record.CreatedAt,
	}
}
