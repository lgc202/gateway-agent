package chat

import (
	"context"
	"errors"
	"iter"
	"net/http"

	"github.com/gin-gonic/gin"

	chatdto "github.com/lgc202/gateway-agent/internal/apiserver/dto/chat"
	"github.com/lgc202/gateway-agent/internal/apiserver/handler/response"
	chatservice "github.com/lgc202/gateway-agent/internal/apiserver/service/chat"
)

const streamErrorEvent = "error"

// streamReply 通过同一条 HTTP 响应持续写入 Agent 事件
func streamReply(ctx *gin.Context, events iter.Seq2[chatservice.ReplyEvent, error]) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)
	ctx.Writer.Flush()

	for event, err := range events {
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			_, resp := response.NewErrorResp(ctx, err)
			ctx.SSEvent(streamErrorEvent, resp)
			ctx.Writer.Flush()
			return
		}

		switch event.Type {
		case chatservice.ReplyEventTypeTextDelta:
			ctx.SSEvent(string(event.Type), chatdto.NewTextDeltaResp(event.Content))
		case chatservice.ReplyEventTypeCompleted:
			ctx.SSEvent(string(event.Type), chatdto.NewMessageResp(*event.Message))
		}
		ctx.Writer.Flush()
	}
}
