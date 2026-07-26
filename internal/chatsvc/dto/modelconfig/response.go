package modelconfig

import (
	"time"

	modelconfigservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/modelconfig"
)

// ModelConfigResp 是返回给客户端的模型配置，不包含 API Key 明文或密文
type ModelConfigResp struct {
	ID               uint64    `json:"id"`
	Name             string    `json:"name"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	BaseURL          string    `json:"base_url"`
	MaxTokens        uint32    `json:"max_tokens"`
	APIKeyConfigured bool      `json:"api_key_configured"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NewModelConfigResp 根据用例层模型构造 HTTP 响应
func NewModelConfigResp(config modelconfigservice.ModelConfig) ModelConfigResp {
	return ModelConfigResp{
		ID:               config.ID,
		Name:             config.Name,
		Provider:         config.Provider,
		Model:            config.Model,
		BaseURL:          config.BaseURL,
		MaxTokens:        config.MaxTokens,
		APIKeyConfigured: config.APIKeyConfigured,
		CreatedAt:        config.CreatedAt,
		UpdatedAt:        config.UpdatedAt,
	}
}

// NewModelConfigListResp 构造模型配置列表响应
func NewModelConfigListResp(configs []modelconfigservice.ModelConfig) []ModelConfigResp {
	resp := make([]ModelConfigResp, 0, len(configs))
	for _, config := range configs {
		resp = append(resp, NewModelConfigResp(config))
	}
	return resp
}
