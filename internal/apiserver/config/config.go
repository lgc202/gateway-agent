// Package config 负责读取并校验 apiserver 启动配置
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 是 apiserver 当前运行所需的最小配置
type Config struct {
	HTTP  HTTPConfig  `mapstructure:"http"`
	MySQL MySQLConfig `mapstructure:"mysql"`
	Model ModelConfig `mapstructure:"model"`
	Agent AgentConfig `mapstructure:"agent"`
}

// HTTPConfig 定义 HTTP Server 配置
type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// MySQLConfig 定义 MySQL 连接配置
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// ModelConfig 定义默认模型的 Provider、连接信息和生成参数
type ModelConfig struct {
	Provider  string `mapstructure:"provider"`
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	Name      string `mapstructure:"name"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// AgentConfig 定义 Agent 的运行参数
type AgentConfig struct {
	MaxIterations int `mapstructure:"max_iterations"`
}

// Load 从指定 YAML 和环境变量加载配置，并在进程启动阶段完成必要校验
func Load(configFile string) (*Config, error) {
	v := viper.New()
	v.AllowEmptyEnv(true)

	if strings.TrimSpace(configFile) == "" {
		return nil, fmt.Errorf("config file is required")
	}
	v.SetConfigFile(configFile)
	if err := v.BindEnv("http.host", "HTTP_HOST"); err != nil {
		return nil, fmt.Errorf("bind HTTP_HOST: %w", err)
	}
	if err := v.BindEnv("http.port", "HTTP_PORT"); err != nil {
		return nil, fmt.Errorf("bind HTTP_PORT: %w", err)
	}
	if err := v.BindEnv("mysql.host", "MYSQL_HOST"); err != nil {
		return nil, fmt.Errorf("bind MYSQL_HOST: %w", err)
	}
	if err := v.BindEnv("mysql.port", "MYSQL_PORT"); err != nil {
		return nil, fmt.Errorf("bind MYSQL_PORT: %w", err)
	}
	if err := v.BindEnv("mysql.database", "MYSQL_DATABASE"); err != nil {
		return nil, fmt.Errorf("bind MYSQL_DATABASE: %w", err)
	}
	if err := v.BindEnv("mysql.username", "MYSQL_USERNAME"); err != nil {
		return nil, fmt.Errorf("bind MYSQL_USERNAME: %w", err)
	}
	if err := v.BindEnv("mysql.password", "MYSQL_PASSWORD"); err != nil {
		return nil, fmt.Errorf("bind MYSQL_PASSWORD: %w", err)
	}
	if err := v.BindEnv("model.provider", "MODEL_PROVIDER"); err != nil {
		return nil, fmt.Errorf("bind MODEL_PROVIDER: %w", err)
	}
	if err := v.BindEnv("model.base_url", "MODEL_BASE_URL"); err != nil {
		return nil, fmt.Errorf("bind MODEL_BASE_URL: %w", err)
	}
	if err := v.BindEnv("model.api_key", "MODEL_API_KEY"); err != nil {
		return nil, fmt.Errorf("bind MODEL_API_KEY: %w", err)
	}
	if err := v.BindEnv("model.name", "MODEL_NAME"); err != nil {
		return nil, fmt.Errorf("bind MODEL_NAME: %w", err)
	}
	if err := v.BindEnv("model.max_tokens", "MODEL_MAX_TOKENS"); err != nil {
		return nil, fmt.Errorf("bind MODEL_MAX_TOKENS: %w", err)
	}
	if err := v.BindEnv("agent.max_iterations", "AGENT_MAX_ITERATIONS"); err != nil {
		return nil, fmt.Errorf("bind AGENT_MAX_ITERATIONS: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file %q: %w", configFile, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config file %q: %w", configFile, err)
	}
	if strings.TrimSpace(cfg.HTTP.Host) == "" {
		return nil, fmt.Errorf("http.host is required")
	}
	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 {
		return nil, fmt.Errorf("http.port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.MySQL.Host) == "" {
		return nil, fmt.Errorf("mysql.host is required")
	}
	if cfg.MySQL.Port < 1 || cfg.MySQL.Port > 65535 {
		return nil, fmt.Errorf("mysql.port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.MySQL.Database) == "" {
		return nil, fmt.Errorf("mysql.database is required")
	}
	if strings.TrimSpace(cfg.MySQL.Username) == "" {
		return nil, fmt.Errorf("mysql.username is required")
	}
	if strings.TrimSpace(cfg.MySQL.Password) == "" {
		return nil, fmt.Errorf("mysql.password is required")
	}

	return &cfg, nil
}
