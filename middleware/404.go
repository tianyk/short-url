package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func NotFoundHandler(ctx *gin.Context) {
    ctx.String(http.StatusNotFound, "NotFound")
}
