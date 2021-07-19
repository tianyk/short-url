package middleware

import (
    "log"
    "net/http"
    "reflect"

    "github.com/gin-gonic/gin"
    "github.com/pkg/errors"

    validator "github.com/go-playground/validator/v10"

    errors2 "short-url/errors"
)

var Recovery = gin.CustomRecovery(func(ctx *gin.Context, recovered interface{}) {
    switch err := recovered.(type) {
    case error:
        log.Printf("%s %s %s, %+v ", ctx.Request.Method, ctx.Request.URL, reflect.TypeOf(err).String(), err)

        switch e := errors.Cause(err).(type) {
        // 数据校验错误
        case validator.ValidationErrors:
            ctx.String(http.StatusBadRequest, e.Error())
        case errors2.HttpError:
            ctx.String(e.Status, e.Message)
        case error:
            ctx.String(http.StatusInternalServerError, e.Error())
        }
    case validator.InvalidValidationError:
        // 参数类型错误
        ctx.String(http.StatusBadRequest, err.Error())
    case string:
        log.Printf("%s %s %s", ctx.Request.Method, ctx.Request.URL, err)
        ctx.String(http.StatusInternalServerError, err)
    default:
        log.Printf("%s %s %s", ctx.Request.Method, ctx.Request.URL, recovered)
        ctx.String(http.StatusInternalServerError, "Unknown error")
    }
})
