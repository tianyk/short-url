package server

import (
    "fmt"

    "github.com/gin-gonic/gin"

    "short-url/config"
    "short-url/middleware"
    "short-url/routes"
)

func Run(httpServer *gin.Engine) {
    httpServer = gin.New()
    // 404
    httpServer.NoRoute(middleware.NotFoundHandler)

    httpServer.Use(middleware.Recovery)
    httpServer.Use(middleware.ErrorHandler)
    httpServer.Use(middleware.RequestLog)

    // 注册路由
    routes.RegisterRoutes(httpServer)

    // 启动
    addr := fmt.Sprintf(":%d", config.Config.Port)
    err := httpServer.Run(addr)
    if err != nil {
        panic(fmt.Errorf("start error: %s", err.Error()))
    }
}
