package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tianyk/short-url/internal/config"
	"github.com/tianyk/short-url/internal/shorturl"
	"github.com/tianyk/short-url/internal/web"
)

func heartbeat(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ok")
}

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine, cfg config.Config, shortURLController *ShortURLController) {
	router.GET("/", func(context *gin.Context) {
		html, err := web.Render(cfg.Token)
		if err != nil {
			panic(err)
		}
		context.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})

	// 静态资源
	staticGroup := router.Group("/static")
	staticGroup.GET("/*filepath", Static)
	staticGroup.HEAD("/*filepath", Static)

	// 心跳
	router.GET("/_heartbeat", heartbeat)

	// 跳转
	router.GET("/:urlID", When(func(ctx *gin.Context) bool {
		urlID := ctx.Param("urlID")
		return len(urlID) >= 4 && len(urlID) <= 8 && shorturl.URLIDRegexp.MatchString(urlID)
	}, shortURLController.OpenShortURL, NotFoundHandler))

	// api group
	apiRouter := router.Group("/api")
	{
		// 短地址
		apiRouter.POST("/short-url", Token(cfg.Token), shortURLController.CreateShortURL)
	}
}
