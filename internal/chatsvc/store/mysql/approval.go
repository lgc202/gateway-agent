package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql/sqlc"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// Approval 是 Store 返回给用例层的审批记录
type Approval struct {
	ID           uint64
	ChatID       uint64
	Status       string
	Operation    string
	Arguments    json.RawMessage
	ResumeTarget string
	RuntimeState []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateApproval 保存等待用户决定的写操作和恢复执行所需的运行状态
func (s *Store) CreateApproval(ctx context.Context, approval Approval) (Approval, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	result, err := s.queries.InsertApproval(ctx, sqlc.InsertApprovalParams{
		ChatID:       approval.ChatID,
		Status:       approval.Status,
		Operation:    approval.Operation,
		Arguments:    approval.Arguments,
		ResumeTarget: approval.ResumeTarget,
		RuntimeState: approval.RuntimeState,
	})
	if err != nil {
		return Approval{}, databaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Approval{}, databaseError(err)
	}

	return s.GetApproval(ctx, approval.ChatID, uint64(id))
}

// GetApproval 查询指定 Chat 中的一条审批记录
func (s *Store) GetApproval(ctx context.Context, chatID, approvalID uint64) (Approval, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	record, err := s.queries.GetApproval(ctx, sqlc.GetApprovalParams{
		ID:     approvalID,
		ChatID: chatID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Approval{}, errorsx.NewUser(errorsx.CodeApprovalNotFound, "审批不存在")
	}
	if err != nil {
		return Approval{}, databaseError(err)
	}

	return toApproval(record), nil
}

// DecideApproval 将待审批记录原子更新为批准或拒绝
func (s *Store) DecideApproval(ctx context.Context, chatID, approvalID uint64, status string) (Approval, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	rowsAffected, err := s.queries.DecideApproval(ctx, sqlc.DecideApprovalParams{
		Status: status,
		ID:     approvalID,
		ChatID: chatID,
	})
	if err != nil {
		return Approval{}, databaseError(err)
	}
	if rowsAffected == 0 {
		if _, err := s.GetApproval(ctx, chatID, approvalID); err != nil {
			return Approval{}, err
		}
		return Approval{}, errorsx.NewUser(errorsx.CodeApprovalAlreadyDecided, "审批已经处理，不能重复决定")
	}

	return s.GetApproval(ctx, chatID, approvalID)
}

func toApproval(record sqlc.Approval) Approval {
	return Approval{
		ID:           record.ID,
		ChatID:       record.ChatID,
		Status:       record.Status,
		Operation:    record.Operation,
		Arguments:    record.Arguments,
		ResumeTarget: record.ResumeTarget,
		RuntimeState: record.RuntimeState,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}
