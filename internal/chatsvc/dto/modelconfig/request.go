// Package modelconfig 定义模型配置 HTTP 接口的数据结构
package modelconfig

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/lgc202/gateway-agent/internal/chatsvc/service/agent"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const (
	maxNameLength    = 64
	maxModelLength   = 255
	maxBaseURLLength = 2048
	maxAPIKeyBytes   = 2020 // 密文列 2048 字节减去 GCM 的 28 字节 Nonce 和认证标签
)

// CreateModelConfigReq 是创建模型配置的请求
type CreateModelConfigReq struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	MaxTokens uint32 `json:"max_tokens"`
}

// UpdateModelConfigReq 是更新模型配置基本信息的请求
type UpdateModelConfigReq struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	MaxTokens uint32 `json:"max_tokens"`
}

// UpdateModelConfigAPIKeyReq 是替换或清除模型 API Key 的请求
type UpdateModelConfigAPIKeyReq struct {
	APIKey *string `json:"api_key"`
}

// Validate 校验创建模型配置所需的全部字段
func (r CreateModelConfigReq) Validate() error {
	if err := validateModelConfig(r.Name, r.Provider, r.Model, r.BaseURL, r.MaxTokens); err != nil {
		return err
	}
	if len(r.APIKey) > maxAPIKeyBytes {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "api_key 不能超过 2020 字节")
	}
	return nil
}

// Validate 校验更新模型配置所需的全部字段
func (r UpdateModelConfigReq) Validate() error {
	return validateModelConfig(r.Name, r.Provider, r.Model, r.BaseURL, r.MaxTokens)
}

// Validate 要求请求明确提供 api_key 字段，空字符串表示清除已有密钥
func (r UpdateModelConfigAPIKeyReq) Validate() error {
	if r.APIKey == nil {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "api_key 字段不能为空")
	}
	if len(*r.APIKey) > maxAPIKeyBytes {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "api_key 不能超过 2020 字节")
	}
	return nil
}

func validateModelConfig(name, provider, model, baseURL string, maxTokens uint32) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLength {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "name 不能为空且不能超过 64 个字符")
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	if !agent.ModelProvider(provider).Valid() {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "provider 不受支持")
	}

	model = strings.TrimSpace(model)
	if model == "" || utf8.RuneCountInString(model) > maxModelLength {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "model 不能为空且不能超过 255 个字符")
	}

	baseURL = strings.TrimSpace(baseURL)
	if utf8.RuneCountInString(baseURL) > maxBaseURLLength {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "base_url 不能超过 2048 个字符")
	}
	if baseURL != "" {
		parsedURL, err := url.ParseRequestURI(baseURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
			return errorsx.NewUser(errorsx.CodeInvalidRequest, "base_url 必须是有效的 HTTP 地址")
		}
	}

	if maxTokens == 0 {
		return errorsx.NewUser(errorsx.CodeInvalidRequest, "max_tokens 必须大于 0")
	}

	return nil
}
