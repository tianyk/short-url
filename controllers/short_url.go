package controllers

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"

    "short-url/config"
    "short-url/service"
    "short-url/vo"
)

func ShortUrl(ctx *gin.Context) {
    body := new(vo.ShortUrlVo)
    err := ctx.MustBindWith(body, binding.FormPost)
    if err != nil {
        panic(fmt.Errorf("invalid request body: %s", err.Error()))
    }

    urlId, err := service.CreateShortUrl(body.LongUrl)
    if err != nil {
        panic(fmt.Errorf("create short url: %s", err.Error()))
    }

    ctx.String(200, fmt.Sprintf("%s/%s", config.Config.Prefix, urlId))
}

func OpenShortUrl(ctx *gin.Context) {
    urlId := ctx.Param("urlId")
    longUrl, err := service.FindLongUrl(urlId)
    if err != nil {
        if err == leveldbErrors.ErrNotFound {
            ctx.String(http.StatusNotFound, "NotFound")
            return
        } else {
            panic(fmt.Errorf("open short url %s", err.Error()))
        }
    }

    ctx.Redirect(http.StatusMovedPermanently, longUrl)
}
