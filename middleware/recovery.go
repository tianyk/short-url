package middleware

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

var Recovery = gin.CustomRecovery(func(ctx *gin.Context, recovered interface{}) {
    log.Printf("%s %s %s\n", ctx.Request.Method, ctx.Request.URL, recovered)
    if err, ok := recovered.(string); ok {
        ctx.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
    }
    ctx.AbortWithStatus(http.StatusInternalServerError)
})