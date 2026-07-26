package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/cloudwego/eino/adk"

	agenttool "github.com/lgc202/gateway-agent/internal/chatsvc/service/agent/tool"
)

// ApprovalDecision 表示用户对一次网关写操作的决定。
type ApprovalDecision struct {
	Approved bool
	Reason   string
}

// ResumeApproval 从已保存的运行状态恢复 Agent，并将用户决定发送给原来的写 Tool。
func (a *Agent) ResumeApproval(
	ctx context.Context,
	runtimeState []byte,
	resumeTarget string,
	decision ApprovalDecision,
) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		runCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// 恢复状态必须先装入 Store，Runner 才能从原来的中断位置继续执行。
		a.checkpointStore.restore(a.checkpointID, runtimeState)
		iterator, err := a.runner.ResumeWithParams(runCtx, a.checkpointID, &adk.ResumeParams{
			Targets: map[string]any{
				resumeTarget: &agenttool.ApprovalDecision{
					Approved: decision.Approved,
					Reason:   decision.Reason,
				},
			},
		})
		if err != nil {
			yield(Event{}, err)
			return
		}

		a.streamEvents(runCtx, iterator, yield)
	}
}

// approvalEvent 将 Eino 的根中断转换为 Chat Service 可以持久化的审批事件。
func (a *Agent) approvalEvent(ctx context.Context, interrupt *adk.InterruptInfo) (Event, error) {
	for _, interruptContext := range interrupt.InterruptContexts {
		if !interruptContext.IsRootCause {
			continue
		}

		request, ok := interruptContext.Info.(*agenttool.Approval)
		if !ok {
			return Event{}, fmt.Errorf("unexpected interrupt info %T", interruptContext.Info)
		}

		arguments, err := json.Marshal(request.Arguments)
		if err != nil {
			return Event{}, fmt.Errorf("marshal approval arguments: %w", err)
		}

		// Runner 会先保存 Checkpoint，再发出中断事件；此处读取到的状态可以直接持久化。
		runtimeState, exists, err := a.checkpointStore.Get(ctx, a.checkpointID)
		if err != nil {
			return Event{}, fmt.Errorf("get agent checkpoint: %w", err)
		}
		if !exists {
			return Event{}, fmt.Errorf("agent checkpoint not found")
		}

		return Event{
			Type: EventTypeApprovalRequired,
			Approval: &Approval{
				Operation:    request.Operation,
				Arguments:    arguments,
				ResumeTarget: interruptContext.ID,
				RuntimeState: runtimeState,
			},
		}, nil
	}

	return Event{}, fmt.Errorf("approval interrupt not found")
}
