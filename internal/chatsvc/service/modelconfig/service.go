// Package modelconfig 实现模型配置管理用例
package modelconfig

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lgc202/gateway-agent/internal/chatsvc/config"
	mysqlstore "github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql"
	"github.com/lgc202/gateway-agent/internal/pkg/cryptox"
)

// ModelConfig 是模型配置管理接口使用的模型
type ModelConfig struct {
	ID               uint64
	Name             string
	Provider         string
	Model            string
	BaseURL          string
	MaxTokens        uint32
	APIKeyConfigured bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CreateInput 包含创建模型配置所需的信息
type CreateInput struct {
	Name      string
	Provider  string
	Model     string
	BaseURL   string
	APIKey    string
	MaxTokens uint32
}

// UpdateInput 包含更新模型配置基本信息所需的信息
type UpdateInput struct {
	ID        uint64
	Name      string
	Provider  string
	Model     string
	BaseURL   string
	MaxTokens uint32
}

// Service 承载模型配置管理用例
type Service struct {
	store  *mysqlstore.Store
	cipher *cryptox.AESGCM
}

// New 创建模型配置 Service，并在启动阶段校验加密主密钥
func New(cfg *config.Config, store *mysqlstore.Store) (*Service, error) {
	cipher, err := cryptox.NewAESGCM(cfg.ModelConfigEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create model config cipher: %w", err)
	}

	return &Service{store: store, cipher: cipher}, nil
}

// Create 创建模型配置，并在进入 Store 前加密 API Key
func (s *Service) Create(ctx context.Context, input CreateInput) (ModelConfig, error) {
	record, err := s.store.CreateModelConfig(ctx, mysqlstore.ModelConfig{
		Name:      strings.TrimSpace(input.Name),
		Provider:  strings.ToLower(strings.TrimSpace(input.Provider)),
		Model:     strings.TrimSpace(input.Model),
		BaseURL:   strings.TrimSpace(input.BaseURL),
		MaxTokens: input.MaxTokens,
	}, s.encryptAPIKey(input.APIKey))
	if err != nil {
		return ModelConfig{}, err
	}

	return toModelConfig(record), nil
}

// Get 查询指定模型配置
func (s *Service) Get(ctx context.Context, id uint64) (ModelConfig, error) {
	record, err := s.store.GetModelConfig(ctx, id)
	if err != nil {
		return ModelConfig{}, err
	}

	return toModelConfig(record), nil
}

// GetRuntimeConfig 读取并解密 Agent 创建模型客户端所需的完整配置
func (s *Service) GetRuntimeConfig(ctx context.Context, id uint64) (config.ModelConfig, error) {
	record, err := s.Get(ctx, id)
	if err != nil {
		return config.ModelConfig{}, err
	}

	runtimeConfig := config.ModelConfig{
		Provider:  record.Provider,
		BaseURL:   record.BaseURL,
		Model:     record.Model,
		MaxTokens: int(record.MaxTokens),
	}
	if !record.APIKeyConfigured {
		return runtimeConfig, nil
	}

	encryptedAPIKey, err := s.store.GetModelConfigAPIKey(ctx, id)
	if err != nil {
		return config.ModelConfig{}, err
	}
	apiKey, err := s.cipher.Decrypt(encryptedAPIKey)
	if err != nil {
		return config.ModelConfig{}, fmt.Errorf("decrypt model config API key: %w", err)
	}
	runtimeConfig.APIKey = string(apiKey)

	return runtimeConfig, nil
}

// List 查询全部模型配置
func (s *Service) List(ctx context.Context) ([]ModelConfig, error) {
	records, err := s.store.ListModelConfigs(ctx)
	if err != nil {
		return nil, err
	}

	configs := make([]ModelConfig, 0, len(records))
	for _, record := range records {
		configs = append(configs, toModelConfig(record))
	}
	return configs, nil
}

// Update 更新模型配置基本信息，不修改已有 API Key
func (s *Service) Update(ctx context.Context, input UpdateInput) (ModelConfig, error) {
	record, err := s.store.UpdateModelConfig(ctx, mysqlstore.ModelConfig{
		ID:        input.ID,
		Name:      strings.TrimSpace(input.Name),
		Provider:  strings.ToLower(strings.TrimSpace(input.Provider)),
		Model:     strings.TrimSpace(input.Model),
		BaseURL:   strings.TrimSpace(input.BaseURL),
		MaxTokens: input.MaxTokens,
	})
	if err != nil {
		return ModelConfig{}, err
	}

	return toModelConfig(record), nil
}

// UpdateAPIKey 替换模型 API Key，传入空字符串时清除已有密钥
func (s *Service) UpdateAPIKey(ctx context.Context, id uint64, apiKey string) (ModelConfig, error) {
	record, err := s.store.UpdateModelConfigAPIKey(ctx, id, s.encryptAPIKey(apiKey))
	if err != nil {
		return ModelConfig{}, err
	}

	return toModelConfig(record), nil
}

// Delete 删除未被 Chat 使用的模型配置
func (s *Service) Delete(ctx context.Context, id uint64) error {
	return s.store.DeleteModelConfig(ctx, id)
}

func (s *Service) encryptAPIKey(apiKey string) []byte {
	if apiKey == "" {
		return nil
	}
	return s.cipher.Encrypt([]byte(apiKey))
}

func toModelConfig(record mysqlstore.ModelConfig) ModelConfig {
	return ModelConfig{
		ID:               record.ID,
		Name:             record.Name,
		Provider:         record.Provider,
		Model:            record.Model,
		BaseURL:          record.BaseURL,
		MaxTokens:        record.MaxTokens,
		APIKeyConfigured: record.APIKeyConfigured,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
	}
}
