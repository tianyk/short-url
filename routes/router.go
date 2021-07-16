package routes

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/config"
    "short-url/controllers"
    "short-url/middleware"
)

func heartbeat(ctx *gin.Context) {
    ctx.String(http.StatusOK, "health")
}

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine) {
    router.GET("/", func(context *gin.Context) {
        context.String(http.StatusOK, "short-url")
    })

    // 心跳
    router.GET("/_heartbeat", heartbeat)

    // 跳转
    router.GET("/s/:urlId", controllers.OpenShortUrl)

    // api group
    apiRouter := router.Group("/api")
    {
        // 短地址
        apiRouter.POST("/short-url", middleware.Token(config.Config.Token), controllers.CreateShortUrl)
    }
}
