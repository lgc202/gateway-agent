package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"

	"github.com/lgc202/gateway-agent/internal/apiserver/store/mysql/sqlc"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// mysqlDuplicateEntryErrorNumber 对应 MySQL ER_DUP_ENTRY。
const mysqlDuplicateEntryErrorNumber uint16 = 1062

// ModelConfig 是 Store 返回给用例层的模型配置记录
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

// CreateModelConfig 创建模型配置，apiKey 只接收已经加密的内容
func (s *Store) CreateModelConfig(ctx context.Context, config ModelConfig, apiKey []byte) (ModelConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	result, err := s.queries.InsertModelConfig(ctx, sqlc.InsertModelConfigParams{
		Name:      config.Name,
		Provider:  config.Provider,
		Model:     config.Model,
		BaseURL:   config.BaseURL,
		APIKey:    apiKey,
		MaxTokens: config.MaxTokens,
	})
	if err != nil {
		if isDuplicateModelConfigName(err) {
			return ModelConfig{}, errorsx.NewUser(errorsx.CodeModelConfigNameConflict, "模型配置名称已存在")
		}
		return ModelConfig{}, databaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ModelConfig{}, databaseError(err)
	}
	if id <= 0 {
		return ModelConfig{}, errorsx.New(errorsx.CodeInternal, "invalid model config id returned by database")
	}

	return s.GetModelConfig(ctx, uint64(id))
}

// GetModelConfig 按 ID 查询模型配置，不读取 API Key
func (s *Store) GetModelConfig(ctx context.Context, id uint64) (ModelConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	record, err := s.queries.GetModelConfig(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ModelConfig{}, errorsx.NewUser(errorsx.CodeModelConfigNotFound, "模型配置不存在")
	}
	if err != nil {
		return ModelConfig{}, databaseError(err)
	}

	return modelConfigFromGetRow(record), nil
}

// ListModelConfigs 查询全部模型配置，不读取 API Key
func (s *Store) ListModelConfigs(ctx context.Context) ([]ModelConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	records, err := s.queries.ListModelConfigs(ctx)
	if err != nil {
		return nil, databaseError(err)
	}

	configs := make([]ModelConfig, 0, len(records))
	for _, record := range records {
		configs = append(configs, modelConfigFromListRow(record))
	}
	return configs, nil
}

// UpdateModelConfig 更新模型配置基本信息，不修改 API Key
func (s *Store) UpdateModelConfig(ctx context.Context, config ModelConfig) (ModelConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	_, err := s.queries.UpdateModelConfig(ctx, sqlc.UpdateModelConfigParams{
		Name:      config.Name,
		Provider:  config.Provider,
		Model:     config.Model,
		BaseURL:   config.BaseURL,
		MaxTokens: config.MaxTokens,
		ID:        config.ID,
	})
	if err != nil {
		if isDuplicateModelConfigName(err) {
			return ModelConfig{}, errorsx.NewUser(errorsx.CodeModelConfigNameConflict, "模型配置名称已存在")
		}
		return ModelConfig{}, databaseError(err)
	}

	return s.GetModelConfig(ctx, config.ID)
}

// UpdateModelConfigAPIKey 更新已经加密的 API Key，nil 表示清除
func (s *Store) UpdateModelConfigAPIKey(ctx context.Context, id uint64, apiKey []byte) (ModelConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	if _, err := s.queries.UpdateModelConfigAPIKey(ctx, sqlc.UpdateModelConfigAPIKeyParams{
		APIKey: apiKey,
		ID:     id,
	}); err != nil {
		return ModelConfig{}, databaseError(err)
	}

	return s.GetModelConfig(ctx, id)
}

// DeleteModelConfig 删除未被 Chat 使用的模型配置
func (s *Store) DeleteModelConfig(ctx context.Context, id uint64) error {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	chatCount, err := s.queries.CountChatsByModelConfigID(ctx, &id)
	if err != nil {
		return databaseError(err)
	}
	if chatCount > 0 {
		return errorsx.NewUser(errorsx.CodeModelConfigInUse, "模型配置正在被对话使用，不能删除")
	}

	rowsAffected, err := s.queries.DeleteModelConfig(ctx, id)
	if err != nil {
		return databaseError(err)
	}
	if rowsAffected == 0 {
		return errorsx.NewUser(errorsx.CodeModelConfigNotFound, "模型配置不存在")
	}

	return nil
}

func isDuplicateModelConfigName(err error) bool {
	mysqlError, ok := errors.AsType[*mysqldriver.MySQLError](err)
	return ok && mysqlError.Number == mysqlDuplicateEntryErrorNumber
}

func modelConfigFromGetRow(record sqlc.GetModelConfigRow) ModelConfig {
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

func modelConfigFromListRow(record sqlc.ListModelConfigsRow) ModelConfig {
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
