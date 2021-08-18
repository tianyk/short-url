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
    "github.com/xhit/go-str2duration/v2"

    "short-url/config"
    errors2 "short-url/errors"
    "short-url/proto"
    "short-url/service"
    "short-url/utils"
    "short-url/vo"
)

// CreateShortUrl 生成短地址
func CreateShortUrl(ctx *gin.Context) {
    body := new(vo.ShortUrlVo)
    err := ctx.ShouldBind(body)
    if err != nil {
        panic(errors.Wrap(err, "Invalid request body"))
    }

    duration, _ := str2duration.ParseDuration("1d")
    if body.MaxAge != "" {
        duration, err = str2duration.ParseDuration(body.MaxAge)
        if err != nil {
            panic(errors.Wrap(err, fmt.Sprintf("Invalid MaxAge %s", body.MaxAge)))
        }
    }

    var password string
    if body.Scope == "private" {
        password = utils.RandomString(4)
    }

    urlId, err := service.CreateShortUrl(&proto.ShortUrlMessage{LongUrl: body.LongUrl, Expire: time.Now().Add(duration).Unix(), Password: password})
    if err != nil {
        panic(errors.Wrap(err, "Create short url error"))
    }

    shortUrl := purell.MustNormalizeURLString(fmt.Sprintf("%s/%s", config.Config.Prefix, urlId), purell.FlagRemoveDuplicateSlashes)
    ctx.JSON(http.StatusOK, &vo.ShortUrlVo{
        LongUrl:  body.LongUrl,
        ShortUrl: shortUrl,
        Password: password,
    })
}

// OpenShortUrl 访问原页面
func OpenShortUrl(ctx *gin.Context) {
    urlId := ctx.Param("urlId")

    // query 前四位为密码
    var password string
    rawQuery := ctx.Request.URL.RawQuery
    if len(rawQuery) >= 4 {
        password = rawQuery[0:4]
    }

    longUrl, err := service.FindLongUrl(urlId, password)
    if originErr := errors.Cause(err); originErr != nil {
        if originErr == leveldbErrors.ErrNotFound {
            // 没有记录错误
            ctx.String(http.StatusNotFound, "NotFound")
            return
        } else if httpErr, ok := originErr.(errors2.HttpError); ok && httpErr.Status == http.StatusForbidden {
            // HTTP 403错误
            ctx.HTML(httpErr.Status, "authorization.html", gin.H{})
        } else {
            // 原始错误
            panic(errors.WithMessage(err, "Open short url error"))
        }
    }

    if validator.IsURL(longUrl) {
        ctx.Redirect(http.StatusMovedPermanently, longUrl)
    } else {
        ctx.String(http.StatusOK, longUrl)
    }
}

func GC(ctx *gin.Context) {
    err := service.GC()
    if err != nil {
        panic(errors.WithMessage(err, "gc error"))
    }

    ctx.String(http.StatusOK, "gc ok")
}
