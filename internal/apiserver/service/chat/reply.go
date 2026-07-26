package chat

import (
	"context"
	"iter"

	agentservice "github.com/lgc202/gateway-agent/internal/apiserver/service/agent"
)

const recentMessageLimit = 100

// ReplyEventType 表示一次 Agent 回复过程中产生的事件类型
type ReplyEventType string

const (
	// ReplyEventTypeTextDelta 表示模型新生成的一段文本
	ReplyEventTypeTextDelta ReplyEventType = "text_delta"
	// ReplyEventTypeCompleted 表示完整回复已经写入 MySQL
	ReplyEventTypeCompleted ReplyEventType = "completed"
)

// ReplyEvent 是 Chat Service 输出给 HTTP SSE 的稳定事件
type ReplyEvent struct {
	Type    ReplyEventType
	Content string
	Message *Message
}

// StreamReply 保存用户消息，并基于最近的对话历史流式生成 Agent 回复
func (s *Service) StreamReply(
	ctx context.Context,
	chatID uint64,
	content string,
) (iter.Seq2[ReplyEvent, error], error) {
	chat, err := s.store.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	gatewayAgent, err := s.agentFactory.New(ctx, chat.ModelConfigID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.AppendMessage(ctx, chatID, string(MessageRoleUser), content); err != nil {
		return nil, err
	}
	records, err := s.store.ListRecentMessages(ctx, chatID, recentMessageLimit)
	if err != nil {
		return nil, err
	}

	messages := make([]agentservice.Message, 0, len(records))
	for _, record := range records {
		messages = append(messages, agentservice.Message{
			Role:    agentservice.MessageRole(record.Role),
			Content: record.Content,
		})
	}

	return s.streamReply(ctx, chatID, gatewayAgent, messages), nil
}

func (s *Service) streamReply(
	ctx context.Context,
	chatID uint64,
	gatewayAgent *agentservice.Agent,
	messages []agentservice.Message,
) iter.Seq2[ReplyEvent, error] {
	return func(yield func(ReplyEvent, error) bool) {
		for event, err := range gatewayAgent.Stream(ctx, messages) {
			if err != nil {
				yield(ReplyEvent{}, err)
				return
			}

			switch event.Type {
			case agentservice.EventTypeTextDelta:
				if !yield(ReplyEvent{Type: ReplyEventTypeTextDelta, Content: event.Content}, nil) {
					return
				}
			case agentservice.EventTypeCompleted:
				record, err := s.store.AppendMessage(
					ctx,
					chatID,
					string(MessageRoleAssistant),
					event.Content,
				)
				if err != nil {
					yield(ReplyEvent{}, err)
					return
				}

				message := toMessage(record)
				yield(ReplyEvent{Type: ReplyEventTypeCompleted, Message: &message}, nil)
				return
			}
		}
	}
}
