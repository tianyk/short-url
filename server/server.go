package server

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/config"
    "short-url/middleware"
    "short-url/routes"
)

func notFoundHandler(ctx *gin.Context) {
    ctx.String(http.StatusNotFound, "NotFound")
}

func Run(httpServer *gin.Engine) {
    httpServer = gin.New()

    httpServer.Use(gin.Logger())
    httpServer.Use(middleware.Recovery)
    httpServer.Use(middleware.ErrorHandler)

    // 注册路由
    routes.RegisterRoutes(httpServer)

    // 404
    httpServer.NoRoute(notFoundHandler)

    // 启动
    addr := fmt.Sprintf(":%d", config.Config.Port)
    err := httpServer.Run(addr)
    if err != nil {
        panic(fmt.Errorf("start error: %s", err.Error()))
    }
}
