package middleware

import (
    "github.com/gin-gonic/gin"
)

type ConditionFunc func(*gin.Context) bool

func When(when ConditionFunc, then gin.HandlerFunc, fallback gin.HandlerFunc) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        if ok := when(ctx); ok {
            then(ctx)
        } else {
            // httpRouter 路由不像 express 能同时匹配多个，
            // 如果有多个类似的会 `conflicts with existing wildcard`
            // 这里只能转发到404，继续 next 没有意义
            fallback(ctx)
        }
    }
}
