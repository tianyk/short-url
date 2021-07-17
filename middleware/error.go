package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "short-url/errors"
)

func ErrorHandler(ctx *gin.Context) {
    ctx.Next()

    if length := len(ctx.Errors); length > 0 {
        err := ctx.Errors[length-1].Err

        if err != nil {
            switch e := err.(type) {
            case *gin.Error:
                ctx.Status(http.StatusInternalServerError)
                ctx.Writer.Write([]byte(e.Error()))
            case errors.HttpError:
                ctx.Status(e.Status)
                ctx.Writer.Write([]byte(e.Message))
            case error:
                ctx.Writer.Write([]byte(e.Error()))
            }
        }
    }
}
