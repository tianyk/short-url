package middleware

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
)


func Token(token string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        if xToken := ctx.GetHeader("x-token"); xToken != token {
            ctx.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid token"))
        } else {
            ctx.Next()
        }
    }
}