// Package chat 提供 Chat 和 Message HTTP 接口
package chat

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	chatdto "github.com/lgc202/gateway-agent/internal/chatsvc/dto/chat"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/response"
	chatservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/chat"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// Handler 处理 Chat 和 Message HTTP 请求
type Handler struct {
	service *chatservice.Service
}

// New 创建 Chat Handler
func New(service *chatservice.Service) *Handler {
	return &Handler{service: service}
}

// Register 注册 Chat 和 Message 路由
func (h *Handler) Register(router *gin.RouterGroup) {
	router.POST("/chats", h.createChat)
	router.GET("/chats/:chat_id", h.getChat)
	router.POST("/chats/:chat_id/messages", h.sendMessage)
	router.GET("/chats/:chat_id/messages", h.listMessages)
}

// createChat 创建一段使用指定模型配置的新对话
func (h *Handler) createChat(ctx *gin.Context) {
	var req chatdto.CreateChatReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	chat, err := h.service.CreateChat(ctx.Request.Context(), req.ModelConfigID)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusCreated, chatdto.NewChatResp(chat))
}

// getChat 查询对话基本信息
func (h *Handler) getChat(ctx *gin.Context) {
	var req chatdto.GetChatReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "路径参数格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	chat, err := h.service.GetChat(ctx.Request.Context(), req.ChatID)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, chatdto.NewChatResp(chat))
}

// sendMessage 接收用户消息，并通过 SSE 流式返回 Agent 回复
func (h *Handler) sendMessage(ctx *gin.Context) {
	var req chatdto.SendMessageReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "路径参数格式错误"))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	// Server 的普通响应写超时是 30 秒，SSE 请求需要在模型调用期间保持可写。
	if err := http.NewResponseController(ctx.Writer).SetWriteDeadline(time.Time{}); err != nil {
		response.WriteError(ctx, fmt.Errorf("disable SSE write deadline: %w", err))
		return
	}

	events, err := h.service.StreamReply(ctx.Request.Context(), req.ChatID, req.Content)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	streamReply(ctx, events)
}

// listMessages 按消息 ID 游标查询对话历史
func (h *Handler) listMessages(ctx *gin.Context) {
	var req chatdto.ListMessagesReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "路径参数格式错误"))
		return
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "查询参数格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	messages, err := h.service.ListMessages(ctx.Request.Context(), req.ChatID, req.AfterID, req.Limit)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, chatdto.NewMessageListResp(messages, req.AfterID))
}
