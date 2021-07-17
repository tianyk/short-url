package middleware

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/errors"
)


func Token(token string) gin.HandlerFunc {
    if token == "" {
        log.Println("WARNING: No token is set")
    }
    return func(ctx *gin.Context) {
        if token == "" {
            ctx.Next()
            return
        }

        if xToken := ctx.GetHeader("x-token"); xToken != token {
            ctx.Error(&errors.HttpError{
                Status: http.StatusForbidden,
                Message: "invalid token",
            })
            ctx.Abort()
        } else {
            ctx.Next()
        }
    }
}