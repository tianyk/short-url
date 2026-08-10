package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tianyk/short-url/internal/config"
	"github.com/tianyk/short-url/internal/httpapi"
)

func setupRouter(cfg config.Config, shortURLController *httpapi.ShortURLController) *gin.Engine {
	app := gin.New()
	// 404
	app.NoRoute(httpapi.NotFoundHandler)

	app.Use(httpapi.Recovery)
	app.Use(httpapi.RequestLog)

	// 注册路由
	httpapi.RegisterRoutes(app, cfg, shortURLController)

	return app
}

func Run(ctx context.Context, cfg config.Config, shortURLController *httpapi.ShortURLController) error {
	app := setupRouter(cfg, shortURLController)

	// 启动
	addr := fmt.Sprintf(":%d", cfg.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: app,
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("Listening on port %s", addr)
		serverErrCh <- httpServer.Serve(listener)
	}()

	select {
	case err := <-serverErrCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	log.Println("Shutdown Server ...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErrCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	log.Println("Server exiting")
	return nil
}
