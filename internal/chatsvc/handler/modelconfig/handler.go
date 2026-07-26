// Package modelconfig 提供模型配置管理 HTTP 接口
package modelconfig

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelconfigdto "github.com/lgc202/gateway-agent/internal/chatsvc/dto/modelconfig"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/response"
	modelconfigservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/modelconfig"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const modelConfigIDParam = "model_config_id"

// Handler 处理模型配置管理 HTTP 请求
type Handler struct {
	service *modelconfigservice.Service
}

// New 创建模型配置 Handler
func New(service *modelconfigservice.Service) *Handler {
	return &Handler{service: service}
}

// Register 注册模型配置管理路由
func (h *Handler) Register(router *gin.RouterGroup) {
	router.POST("/model-configs", h.create)
	router.GET("/model-configs", h.list)
	router.GET("/model-configs/:"+modelConfigIDParam, h.get)
	router.PUT("/model-configs/:"+modelConfigIDParam, h.update)
	router.PUT("/model-configs/:"+modelConfigIDParam+"/api-key", h.updateAPIKey)
	router.DELETE("/model-configs/:"+modelConfigIDParam, h.delete)
}

// create 校验用户输入并创建模型配置
func (h *Handler) create(ctx *gin.Context) {
	var req modelconfigdto.CreateModelConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	config, err := h.service.Create(ctx.Request.Context(), modelconfigservice.CreateInput{
		Name:      req.Name,
		Provider:  req.Provider,
		Model:     req.Model,
		BaseURL:   req.BaseURL,
		APIKey:    req.APIKey,
		MaxTokens: req.MaxTokens,
	})
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusCreated, modelconfigdto.NewModelConfigResp(config))
}

// list 返回全部模型配置
func (h *Handler) list(ctx *gin.Context) {
	configs, err := h.service.List(ctx.Request.Context())
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, modelconfigdto.NewModelConfigListResp(configs))
}

// get 查询单个模型配置
func (h *Handler) get(ctx *gin.Context) {
	id, err := parseModelConfigID(ctx.Param(modelConfigIDParam))
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	config, err := h.service.Get(ctx.Request.Context(), id)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, modelconfigdto.NewModelConfigResp(config))
}

// update 更新模型配置基本信息
func (h *Handler) update(ctx *gin.Context) {
	id, err := parseModelConfigID(ctx.Param(modelConfigIDParam))
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	var req modelconfigdto.UpdateModelConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	config, err := h.service.Update(ctx.Request.Context(), modelconfigservice.UpdateInput{
		ID:        id,
		Name:      req.Name,
		Provider:  req.Provider,
		Model:     req.Model,
		BaseURL:   req.BaseURL,
		MaxTokens: req.MaxTokens,
	})
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, modelconfigdto.NewModelConfigResp(config))
}

// updateAPIKey 替换模型 API Key，空字符串表示清除
func (h *Handler) updateAPIKey(ctx *gin.Context) {
	id, err := parseModelConfigID(ctx.Param(modelConfigIDParam))
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	var req modelconfigdto.UpdateModelConfigAPIKeyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}
	if err := req.Validate(); err != nil {
		response.WriteError(ctx, err)
		return
	}

	config, err := h.service.UpdateAPIKey(ctx.Request.Context(), id, *req.APIKey)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, modelconfigdto.NewModelConfigResp(config))
}

// delete 删除未被 Chat 使用的模型配置
func (h *Handler) delete(ctx *gin.Context) {
	id, err := parseModelConfigID(ctx.Param(modelConfigIDParam))
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), id); err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, nil)
}

func parseModelConfigID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errorsx.NewUser(errorsx.CodeInvalidRequest, "model_config_id 必须是大于 0 的整数")
	}
	return id, nil
}
