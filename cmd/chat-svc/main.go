// chat-svc 提供 Gateway Agent HTTP API
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lgc202/gateway-agent/internal/chatsvc"
)

const (
	defaultConfigFile = "configs/chat-svc.yaml"
	shutdownTimeout   = 10 * time.Second
)

func main() {
	configFile := flag.String("config", defaultConfigFile, "chat-svc 配置文件路径")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(*configFile); err != nil {
		logger.Error("chat service exited", "error", err)
		os.Exit(1)
	}
}

func run(configFile string) error {
	server, err := chatsvc.InitializeServer(configFile)
	if err != nil {
		return fmt.Errorf("initialize chat service: %w", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Run()
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown chat service: %w", err)
	}

	return <-serveErr
}
