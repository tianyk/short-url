package routes

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/controllers"
)

func heatbeart(ctx *gin.Context) {
    ctx.String(http.StatusOK, "health")
}

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine) {
    router.GET("/", func(context *gin.Context) {
        context.String(http.StatusOK, "home")
    })

    // 心跳
    router.GET("/_heatbeart", heatbeart)

    // 跳转
    router.GET("/s/:urlId", controllers.OpenShortUrl)

    // api group
    apiRouter := router.Group("/api")
    {
        // 短地址
        apiRouter.POST("/short-url", controllers.CreateShortUrl)
    }
}
