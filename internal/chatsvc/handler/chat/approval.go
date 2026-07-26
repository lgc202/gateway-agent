package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"

	chatdto "github.com/lgc202/gateway-agent/internal/chatsvc/dto/chat"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/response"
	chatservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/chat"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// listApprovals 查询对话中等待用户决定的审批记录
func (h *Handler) listApprovals(ctx *gin.Context) {
	var req chatdto.ListApprovalsReq
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

	approvals, err := h.service.ListPendingApprovals(ctx.Request.Context(), req.ChatID)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, chatdto.NewApprovalListResp(approvals))
}

// decideApproval 保存用户的审批决定，并通过 SSE 继续返回 Agent 回复
func (h *Handler) decideApproval(ctx *gin.Context) {
	var req chatdto.DecideApprovalReq
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
	if err := disableStreamWriteDeadline(ctx); err != nil {
		response.WriteError(ctx, err)
		return
	}

	events, err := h.service.DecideApproval(
		ctx.Request.Context(),
		req.ChatID,
		req.ApprovalID,
		chatservice.ApprovalStatus(req.Decision),
		req.Reason,
	)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	streamReply(ctx, events)
}
