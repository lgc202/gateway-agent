package agent

import "encoding/json"

// EventType 表示 Agent 向 Chat Service 输出的事件类型。
type EventType string

const (
	// EventTypeTextDelta 表示模型新生成的一段文本。
	EventTypeTextDelta EventType = "text_delta"
	// EventTypeCompleted 表示本次 Agent 调用已经得到完整回复。
	EventTypeCompleted EventType = "completed"
	// EventTypeApprovalRequired 表示写操作正在等待用户审批。
	EventTypeApprovalRequired EventType = "approval_required"
)

// Approval 描述一次等待用户决定的网关写操作。
type Approval struct {
	Operation    string
	Arguments    json.RawMessage
	ResumeTarget string
	RuntimeState []byte
}

// Event 是 Agent 对 Eino 内部事件的稳定投影。
type Event struct {
	Type     EventType
	Content  string
	Approval *Approval
}
