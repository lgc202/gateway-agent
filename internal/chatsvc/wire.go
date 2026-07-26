//go:build wireinject

// Package chatsvc 负责装配 Chat Service 的运行依赖
package chatsvc

//go:generate wire

import (
	"github.com/google/wire"

	"github.com/lgc202/gateway-agent/internal/chatsvc/config"
	chathandler "github.com/lgc202/gateway-agent/internal/chatsvc/handler/chat"
	"github.com/lgc202/gateway-agent/internal/chatsvc/handler/health"
	modelconfighandler "github.com/lgc202/gateway-agent/internal/chatsvc/handler/modelconfig"
	"github.com/lgc202/gateway-agent/internal/chatsvc/server"
	agentservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/agent"
	chatservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/chat"
	gatewayservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/gateway"
	modelconfigservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/modelconfig"
	higressstore "github.com/lgc202/gateway-agent/internal/chatsvc/store/higress"
	mysqlstore "github.com/lgc202/gateway-agent/internal/chatsvc/store/mysql"
)

// InitializeServer 装配 Chat Service 当前运行所需的全部依赖
func InitializeServer(configFile string) (*server.Server, error) {
	wire.Build(
		config.Load,
		mysqlstore.NewDB,
		mysqlstore.NewStore,
		higressstore.NewClient,
		wire.Bind(new(gatewayservice.RouteReader), new(*higressstore.Client)),
		modelconfigservice.New,
		agentservice.NewFactory,
		chatservice.New,
		chathandler.New,
		modelconfighandler.New,
		health.New,
		server.New,
	)
	return nil, nil
}
