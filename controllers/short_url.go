package controllers

import (
    "fmt"
    "net/http"

    "github.com/PuerkitoBio/purell"
    validator "github.com/asaskevich/govalidator"
    "github.com/gin-gonic/gin"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"

    "short-url/config"
    "short-url/service"
    "short-url/vo"
)

func CreateShortUrl(ctx *gin.Context) {
    body := new(vo.ShortUrlVo)
    err := ctx.ShouldBind(&body)
    if err != nil {
        panic(fmt.Errorf("invalid request body: %s", err.Error()))
    }

    urlId, err := service.CreateShortUrl(body.LongUrl)
    if err != nil {
        panic(fmt.Errorf("create short url: %s", err.Error()))
    }

    ctx.String(http.StatusOK, purell.MustNormalizeURLString(fmt.Sprintf("%s/%s", config.Config.Prefix, urlId), purell.FlagRemoveDuplicateSlashes))
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

    if validator.IsURL(longUrl) {
        ctx.Redirect(http.StatusMovedPermanently, longUrl)
    } else {
        ctx.String(http.StatusOK, longUrl)
    }
}
