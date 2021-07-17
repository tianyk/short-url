package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/errors"
)

func ErrorHandler(ctx *gin.Context) {
    ctx.Next()

    if length := len(ctx.Errors); length > 0 {
        err := ctx.Errors.Last().Err

        if err != nil && !ctx.Writer.Written() {
            switch e := err.(type) {
            case *errors.HttpError:
                ctx.String(e.Status, err.Error())
            case *gin.Error:
                ctx.String(http.StatusInternalServerError, err.Error())
            case error:
                ctx.Writer.Write([]byte(e.Error()))
            }
        }
    }
}
