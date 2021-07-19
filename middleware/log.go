package middleware

import (
    "log"
    "time"

    "github.com/gin-gonic/gin"
)

func RequestLog(ctx *gin.Context) {
    start := time.Now()
    ctx.Next()
    end := time.Now()

    log.Printf("%s %s %s %d %d", ctx.ClientIP(), ctx.Request.Method, ctx.Request.RequestURI, ctx.Writer.Status(), end.Sub(start).Milliseconds())
}
