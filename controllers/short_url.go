package controllers

import (
    "fmt"
    "net/http"

    "github.com/PuerkitoBio/purell"
    validator "github.com/asaskevich/govalidator"
    "github.com/gin-gonic/gin"
    "github.com/pkg/errors"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"

    "short-url/config"
    "short-url/service"
    "short-url/vo"
)

// CreateShortUrl 生成短地址
func CreateShortUrl(ctx *gin.Context) {
    body := new(vo.ShortUrlVo)
    err := ctx.ShouldBind(&body)
    if err != nil {
        panic(errors.Wrap(err, "Invalid request body"))
    }

    urlId, err := service.CreateShortUrl(body.LongUrl)
    if err != nil {
        panic(errors.Wrap(err, "Create short url error"))
    }

    shortUrl := purell.MustNormalizeURLString(fmt.Sprintf("%s/%s", config.Config.Prefix, urlId), purell.FlagRemoveDuplicateSlashes)
    ctx.String(http.StatusOK, shortUrl)
}

// OpenShortUrl 访问原页面
func OpenShortUrl(ctx *gin.Context) {
    urlId := ctx.Param("urlId")
    longUrl, err := service.FindLongUrl(urlId)
    if err != nil {
        if err == leveldbErrors.ErrNotFound {
            ctx.String(http.StatusNotFound, "NotFound")
            return
        } else {
            panic(errors.Wrap(err, "Open short url error"))
        }
    }

    if validator.IsURL(longUrl) {
        ctx.Redirect(http.StatusMovedPermanently, longUrl)
    } else {
        ctx.String(http.StatusOK, longUrl)
    }
}
