package routes

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/config"
    "short-url/controllers"
    "short-url/middleware"
    "short-url/service"
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
    // GC
    router.GET("/_gc", middleware.Token(config.Config.Token), controllers.GC)

    // 跳转
    router.GET("/:urlId", middleware.When(func(ctx *gin.Context) bool {
        urlId := ctx.Param("urlId")
        return len(urlId) >= 4 && len(urlId) <= 8 && service.UrlIdRegexp.MatchString(urlId)
    }, controllers.OpenShortUrl, middleware.NotFoundHandler))

    // api group
    apiRouter := router.Group("/api")
    {
        // 短地址
        apiRouter.POST("/short-url", middleware.Token(config.Config.Token), controllers.CreateShortUrl)
    }
}
