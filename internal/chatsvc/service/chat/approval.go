package chat

import (
	"context"
	"encoding/json"
	"iter"
	"time"

	agentservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/agent"
	mysqlstore "github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql"
)

// ApprovalStatus 表示网关写操作的审批状态。
type ApprovalStatus string

const (
	// ApprovalStatusPending 表示写操作正在等待用户决定。
	ApprovalStatusPending ApprovalStatus = "pending"
	// ApprovalStatusApproved 表示用户已经批准写操作。
	ApprovalStatusApproved ApprovalStatus = "approved"
	// ApprovalStatusRejected 表示用户已经拒绝写操作。
	ApprovalStatusRejected ApprovalStatus = "rejected"
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

// ListPendingApprovals 查询指定 Chat 中等待用户决定的审批记录。
func (s *Service) ListPendingApprovals(ctx context.Context, chatID uint64) ([]Approval, error) {
	if _, err := s.store.GetChat(ctx, chatID); err != nil {
		return nil, err
	}

	records, err := s.store.ListPendingApprovals(ctx, chatID)
	if err != nil {
		return nil, err
	}

	approvals := make([]Approval, 0, len(records))
	for _, record := range records {
		approvals = append(approvals, toApproval(record))
	}
	return approvals, nil
}

// DecideApproval 保存用户决定，并恢复被该审批暂停的 Agent 执行。
func (s *Service) DecideApproval(
	ctx context.Context,
	chatID uint64,
	approvalID uint64,
	status ApprovalStatus,
	reason string,
) (iter.Seq2[ReplyEvent, error], error) {
	chat, err := s.store.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	gatewayAgent, err := s.agentFactory.New(ctx, chat.ModelConfigID)
	if err != nil {
		return nil, err
	}

	record, err := s.store.DecideApproval(ctx, chatID, approvalID, string(status))
	if err != nil {
		return nil, err
	}

	events := gatewayAgent.ResumeApproval(
		ctx,
		record.RuntimeState,
		record.ResumeTarget,
		agentservice.ApprovalDecision{
			Approved: status == ApprovalStatusApproved,
			Reason:   reason,
		},
	)
	return s.streamAgentEvents(ctx, chatID, events), nil
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
