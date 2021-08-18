package server

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"

    "short-url/config"
    "short-url/middleware"
    "short-url/routes"
)

func setupRouter() *gin.Engine {
    app := gin.New()
    // 404
    app.NoRoute(middleware.NotFoundHandler)

    app.Use(middleware.Recovery)
    app.Use(middleware.RequestLog)

    // 注册路由
    routes.RegisterRoutes(app)

    return app
}

func Run() {
    app := setupRouter()

    // 启动
    addr := fmt.Sprintf(":%d", config.Config.Port)
    httpServer := &http.Server{
        Addr:    addr,
        Handler: app,
    }
    go func() {
        log.Printf("Listening on port %s", addr)
        // service connections
        if err := httpServer.ListenAndServe(); err != nil {
            log.Printf("listen: %s\n", err)
        }
    }()

    // 退出
    quit := make(chan os.Signal)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutdown Server ...")
    // 10秒后退出程序
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := httpServer.Shutdown(ctx); err != nil {
        log.Fatal("Server Shutdown:", err)
    }
    log.Println("Server exiting")
}
