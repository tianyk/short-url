package controllers

import (
    "fmt"
    "net/http"
    "time"

    "github.com/PuerkitoBio/purell"
    validator "github.com/asaskevich/govalidator"
    "github.com/gin-gonic/gin"
    "github.com/pkg/errors"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
    str2duration "github.com/xhit/go-str2duration/v2"

    "short-url/config"
    "short-url/proto"
    "short-url/service"
    "short-url/vo"
)

// CreateShortUrl 生成短地址
func CreateShortUrl(ctx *gin.Context) {
    body := new(vo.ShortUrlVo)
    err := ctx.ShouldBind(body)
    if err != nil {
        panic(errors.Wrap(err, "Invalid request body"))
    }

    duration, _ := str2duration.ParseDuration("365d")
    if body.MaxAge != "" {
        duration, err = str2duration.ParseDuration(body.MaxAge)
        if err != nil {
            panic(errors.Wrap(err, fmt.Sprintf("Invalid MaxAge %s", body.MaxAge)))
        }
    }

    urlId, err := service.CreateShortUrl(&proto.ShortUrlMessage{LongUrl: body.LongUrl, Expire: time.Now().Add(duration).Unix()})
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
        if errors.Cause(err) == leveldbErrors.ErrNotFound {
            ctx.String(http.StatusNotFound, "NotFound")
            return
        } else {
            panic(errors.WithMessage(err, "Open short url error"))
        }
    }

    if validator.IsURL(longUrl) {
        ctx.Redirect(http.StatusMovedPermanently, longUrl)
    } else {
        ctx.String(http.StatusOK, longUrl)
    }
}
