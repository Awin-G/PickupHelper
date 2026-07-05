package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	"pickup-helper/internal/config"
	"pickup-helper/internal/handler"
	"pickup-helper/internal/log"
	"pickup-helper/internal/router"
	"pickup-helper/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("load config: " + err.Error())
	}

	log.Init(cfg.Log.Level)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	container, err := BuildContainer(cfg)
	if err != nil {
		panic("build container: " + err.Error())
	}
	defer container.Close()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	healthH := handler.NewHealthHandler(container.DB, container.RDB)
	router.Register(engine, healthH)

	srv := server.New(cfg, engine)
	log.From(ctx).InfoContext(ctx, "server starting", "addr", srv.Addr())
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("server run: " + err.Error())
	}
	log.From(ctx).InfoContext(ctx, "server stopped")
}
