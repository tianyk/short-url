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

    // 注册路由
    routes.RegisterRoutes(httpServer)

    // 404
    httpServer.NoRoute(notFoundHandler)

    // 启动
    httpServer.Run(fmt.Sprintf(":%d", config.Config.Port))
}
