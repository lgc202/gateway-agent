//go:build wireinject

// Package apiserver 负责 apiserver 依赖装配
package apiserver

//go:generate wire

import (
	"github.com/google/wire"

	"github.com/lgc202/gateway-agent/internal/apiserver/config"
	chathandler "github.com/lgc202/gateway-agent/internal/apiserver/handler/chat"
	"github.com/lgc202/gateway-agent/internal/apiserver/handler/health"
	"github.com/lgc202/gateway-agent/internal/apiserver/server"
	chatservice "github.com/lgc202/gateway-agent/internal/apiserver/service/chat"
	mysqlstore "github.com/lgc202/gateway-agent/internal/apiserver/store/mysql"
)

// InitializeServer 装配 apiserver 当前运行所需的全部依赖
func InitializeServer(configFile string) (*server.Server, error) {
	wire.Build(
		config.Load,
		mysqlstore.NewDB,
		mysqlstore.NewStore,
		chatservice.New,
		chathandler.New,
		health.New,
		server.New,
	)
	return nil, nil
}
