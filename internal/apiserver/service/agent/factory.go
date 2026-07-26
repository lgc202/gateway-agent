package agent

import (
	"context"
	"fmt"

	"github.com/lgc202/gateway-agent/internal/apiserver/config"
	modelconfigservice "github.com/lgc202/gateway-agent/internal/apiserver/service/modelconfig"
)

// Factory 根据 Chat 选择的模型配置创建 Agent
type Factory struct {
	defaultModel  config.ModelConfig
	maxIterations int
	modelConfigs  *modelconfigservice.Service
}

// NewFactory 创建 Agent Factory
func NewFactory(cfg *config.Config, modelConfigs *modelconfigservice.Service) *Factory {
	return &Factory{
		defaultModel:  cfg.Model,
		maxIterations: cfg.Agent.MaxIterations,
		modelConfigs:  modelConfigs,
	}
}

// New 创建使用指定模型配置的 Agent；modelConfigID 为空时使用系统默认配置
func (f *Factory) New(ctx context.Context, modelConfigID *uint64) (*Agent, error) {
	modelConfig := f.defaultModel
	if modelConfigID != nil {
		var err error
		modelConfig, err = f.modelConfigs.GetRuntimeConfig(ctx, *modelConfigID)
		if err != nil {
			return nil, err
		}
	}

	chatModel, err := NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	gatewayAgent, err := New(ctx, chatModel, f.maxIterations)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	return gatewayAgent, nil
}
