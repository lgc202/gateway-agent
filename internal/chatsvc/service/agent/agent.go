package agent

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	agentName        = "gateway_agent"
	agentDescription = "Query and manage the configured API gateway."
	agentInstruction = `你是当前 Gateway 的运维 Agent。
Gateway 实时状态必须通过 Tool 查询。
任何写操作只能生成待审批变更，不能声称已经执行。
只有收到确定的执行成功结果后，才能告诉用户变更已经生效。
不得展示内部思考过程。`
)

// Agent 使用 Eino 执行模型调用，并隐藏 Eino 的运行时协议。
type Agent struct {
	runner          *adk.Runner
	checkpointID    string
	checkpointStore *checkpointStore
}

// New 创建可流式输出的 Gateway Agent。
func New(
	ctx context.Context,
	chatModel model.ToolCallingChatModel,
	tools []tool.BaseTool,
	maxIterations int,
) (*Agent, error) {
	chatModelAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        agentName,
		Description: agentDescription,
		Instruction: agentInstruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
		MaxIterations: maxIterations,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model agent: %w", err)
	}

	checkpointStore := newCheckpointStore()
	return &Agent{
		runner: adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           chatModelAgent,
			EnableStreaming: true,
			CheckPointStore: checkpointStore,
		}),
		checkpointID:    rand.Text(),
		checkpointStore: checkpointStore,
	}, nil
}

// Stream 执行一次 Agent 调用，依次输出文本增量和完整回复。
func (a *Agent) Stream(ctx context.Context, messages []Message) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		// yield 将事件逐条交给调用方；返回 false 表示调用方已经停止读取。
		einoMessages, err := toEinoMessages(messages)
		if err != nil {
			yield(Event{}, err)
			return
		}

		// 消费方提前结束迭代时，需要同时取消仍在进行的模型请求。
		runCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		iterator := a.runner.Run(runCtx, einoMessages, adk.WithCheckPointID(a.checkpointID))
		a.streamEvents(runCtx, iterator, yield)
	}
}

// streamEvents 消费 Eino Runner 事件，并输出 Agent 的稳定事件。
func (a *Agent) streamEvents(
	ctx context.Context,
	iterator *adk.AsyncIterator[*adk.AgentEvent],
	yield func(Event, error) bool,
) {
	var finalContent string
	var completed bool

	// Runner 按 ReAct 执行顺序输出模型消息和 Tool 结果。
	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			yield(Event{}, event.Err)
			return
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			approvalEvent, err := a.approvalEvent(ctx, event.Action.Interrupted)
			if err != nil {
				yield(Event{}, err)
				return
			}
			yield(approvalEvent, nil)
			return
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		message, keepGoing, err := streamAssistantMessage(event.Output.MessageOutput, yield)
		if err != nil {
			yield(Event{}, err)
			return
		}
		if !keepGoing {
			return
		}
		if message == nil || len(message.ToolCalls) > 0 {
			continue
		}

		finalContent = message.Content
		completed = true
	}

	if completed {
		yield(Event{Type: EventTypeCompleted, Content: finalContent}, nil)
	}
}

func streamAssistantMessage(
	output *adk.MessageVariant,
	yield func(Event, error) bool,
) (*schema.Message, bool, error) {
	if output.Role != schema.Assistant {
		return nil, true, nil
	}
	if !output.IsStreaming {
		if output.Message != nil && output.Message.Content != "" {
			if !yield(Event{Type: EventTypeTextDelta, Content: output.Message.Content}, nil) {
				return nil, false, nil
			}
		}
		return output.Message, true, nil
	}

	stream := output.MessageStream
	defer stream.Close()

	var chunks []*schema.Message
	// 一条流式 Assistant 消息由多个文本分片组成，需要逐个交给上层写入 SSE。
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, false, fmt.Errorf("read assistant message stream: %w", err)
		}

		chunks = append(chunks, chunk)
		if chunk.Content != "" {
			if !yield(Event{Type: EventTypeTextDelta, Content: chunk.Content}, nil) {
				return nil, false, nil
			}
		}
	}

	message, err := schema.ConcatMessages(chunks)
	if err != nil {
		return nil, false, fmt.Errorf("concatenate assistant message: %w", err)
	}
	return message, true, nil
}
