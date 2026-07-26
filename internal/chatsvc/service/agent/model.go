package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"github.com/lgc202/gateway-agent/internal/chatsvc/config"
	"google.golang.org/genai"
)

// ModelProvider 表示 Eino 原生支持的模型服务。
type ModelProvider string

const (
	ModelProviderOpenAI   ModelProvider = "openai"
	ModelProviderDeepSeek ModelProvider = "deepseek"
	ModelProviderQwen     ModelProvider = "qwen"
	ModelProviderClaude   ModelProvider = "claude"
	ModelProviderGemini   ModelProvider = "gemini"
	ModelProviderOllama   ModelProvider = "ollama"
	ModelProviderArk      ModelProvider = "ark"
)

const (
	defaultQwenBaseURL   = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultOllamaBaseURL = "http://localhost:11434"
)

// Valid 判断 Provider 是否已有可用的 Eino 模型实现。
func (p ModelProvider) Valid() bool {
	switch p {
	case ModelProviderOpenAI,
		ModelProviderDeepSeek,
		ModelProviderQwen,
		ModelProviderClaude,
		ModelProviderGemini,
		ModelProviderOllama,
		ModelProviderArk:
		return true
	default:
		return false
	}
}

// NewChatModel 根据 Provider 创建供 Agent 使用的 Eino 模型客户端。
func NewChatModel(ctx context.Context, cfg config.ModelConfig) (model.ToolCallingChatModel, error) {
	var maxTokens *int
	if cfg.MaxTokens > 0 {
		maxTokens = &cfg.MaxTokens
	}

	switch ModelProvider(cfg.Provider) {
	case ModelProviderOpenAI:
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL:             cfg.BaseURL,
			APIKey:              cfg.APIKey,
			Model:               cfg.Model,
			MaxCompletionTokens: maxTokens,
		})
	case ModelProviderDeepSeek:
		return deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			BaseURL:   cfg.BaseURL,
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			MaxTokens: cfg.MaxTokens,
		})
	case ModelProviderQwen:
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = defaultQwenBaseURL
		}
		return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			BaseURL:   baseURL,
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			MaxTokens: maxTokens,
		})
	case ModelProviderClaude:
		var baseURL *string
		if cfg.BaseURL != "" {
			baseURL = &cfg.BaseURL
		}
		return claude.NewChatModel(ctx, &claude.Config{
			BaseURL:   baseURL,
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			MaxTokens: cfg.MaxTokens,
		})
	case ModelProviderGemini:
		return newGeminiChatModel(ctx, cfg, maxTokens)
	case ModelProviderOllama:
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = defaultOllamaBaseURL
		}
		return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: baseURL,
			Model:   cfg.Model,
			Options: &ollama.Options{NumPredict: cfg.MaxTokens},
		})
	case ModelProviderArk:
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			BaseURL:   cfg.BaseURL,
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			MaxTokens: maxTokens,
		})
	default:
		return nil, fmt.Errorf("unsupported model provider %q", cfg.Provider)
	}
}

// Gemini 的 Eino 适配器接收 Google 官方客户端，需要先完成客户端配置。
func newGeminiChatModel(
	ctx context.Context,
	cfg config.ModelConfig,
	maxTokens *int,
) (model.ToolCallingChatModel, error) {
	clientConfig := &genai.ClientConfig{APIKey: cfg.APIKey}
	if cfg.BaseURL != "" {
		clientConfig.HTTPOptions.BaseURL = cfg.BaseURL
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create Gemini client: %w", err)
	}

	return gemini.NewChatModel(ctx, &gemini.Config{
		Client:    client,
		Model:     cfg.Model,
		MaxTokens: maxTokens,
	})
}
