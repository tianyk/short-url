package middleware

import (
    "log"
    "net"
    "net/http"
    "os"
    "reflect"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "github.com/pkg/errors"

    errors2 "short-url/errors"
)

func Recovery(ctx *gin.Context) {
    // recovery
    defer func() {
        if recovered := recover(); recovered != nil {
            // 连接已经关闭
            if ne, ok := recovered.(*net.OpError); ok {
                if se, ok := ne.Err.(*os.SyscallError); ok {
                    if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
                        strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
                        // brokenPipe
                        // If the connection is dead, we can't write a status to it.
                        log.Printf("BrokenPipe: %s", se)
                        ctx.Abort()
                        return
                    }
                }
            }

            log.Printf("%s %s [%s], %+v ", ctx.Request.Method, ctx.Request.URL, reflect.TypeOf(recovered).String(), recovered)
            switch err := recovered.(type) {
            case error:
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
                ctx.String(http.StatusInternalServerError, err)
            default:
                ctx.String(http.StatusInternalServerError, "Unknown error")
            }
        }
    }()

    ctx.Next()

    if length := len(ctx.Errors); length > 0 {
        log.Printf("Errors: %s", ctx.Errors)

        if !ctx.Writer.Written() {
            panic(ctx.Errors.Last().Err)
        }
    }
}
