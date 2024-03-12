package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/allmazz/clickhouse-protocol-proxy/clickhouse-protocol-proxy/internal/config"
	"github.com/allmazz/clickhouse-protocol-proxy/clickhouse-protocol-proxy/internal/controller"

	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	if len(os.Args) != 2 {
		panic("example usage: ./ch-p-proxy /config.yaml")
	}
	cfg, logger := config.New(os.Args[1])
	defer logger.Sync()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	ctrl := controller.New(cfg, logger.With(zap.Field{Type: zapcore.StringType, Key: "component", String: "controller"}))
	errCh := ctrl.Run()

	select {
	case <-ctx.Done():
		err := ctrl.Stop(context.Background())
		if err != nil {
			logger.Error("error occurred during stopping the server", zap.Field{Type: zapcore.ErrorType, Key: "error", Interface: err})
		}
		logger.Info("the server stopped")
	case err := <-errCh:
		logger.Fatal("the server stopped unexpectedly", zap.Field{Type: zapcore.ErrorType, Key: "error", Interface: err})
	}
}
