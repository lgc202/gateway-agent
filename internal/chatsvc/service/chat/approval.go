package chat

import (
	"context"
	"encoding/json"
	"time"

	agentservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/agent"
	mysqlstore "github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql"
)

// ApprovalStatus 表示网关写操作的审批状态。
type ApprovalStatus string

const (
	// ApprovalStatusPending 表示写操作正在等待用户决定。
	ApprovalStatusPending ApprovalStatus = "pending"
)

// Approval 是 Chat API 使用的审批记录。
type Approval struct {
	ID        uint64
	ChatID    uint64
	Status    ApprovalStatus
	Operation string
	Arguments json.RawMessage
	CreatedAt time.Time
}

// createApproval 保存 Agent 暂停执行时产生的待审批写操作。
func (s *Service) createApproval(
	ctx context.Context,
	chatID uint64,
	approval agentservice.Approval,
) (Approval, error) {
	record, err := s.store.CreateApproval(ctx, mysqlstore.Approval{
		ChatID:       chatID,
		Status:       string(ApprovalStatusPending),
		Operation:    approval.Operation,
		Arguments:    approval.Arguments,
		ResumeTarget: approval.ResumeTarget,
		RuntimeState: approval.RuntimeState,
	})
	if err != nil {
		return Approval{}, err
	}

	return toApproval(record), nil
}

func toApproval(record mysqlstore.Approval) Approval {
	return Approval{
		ID:        record.ID,
		ChatID:    record.ChatID,
		Status:    ApprovalStatus(record.Status),
		Operation: record.Operation,
		Arguments: record.Arguments,
		CreatedAt: record.CreatedAt,
	}
}
