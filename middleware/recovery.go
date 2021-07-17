package middleware

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

var Recovery = gin.CustomRecovery(func(ctx *gin.Context, recovered interface{}) {
    log.Printf("%s %s %s\n", ctx.Request.Method, ctx.Request.URL, recovered)

    if err, ok := recovered.(error); ok {
        ctx.String(http.StatusInternalServerError, err.Error())
    } else if err, ok := recovered.(string); ok {
        ctx.String(http.StatusInternalServerError, err)
    } else {
        ctx.AbortWithStatus(http.StatusInternalServerError)
    }
})
