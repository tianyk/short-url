package httpapi

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xhit/go-str2duration/v2"

	"github.com/tianyk/short-url/internal/shorturl"
	"github.com/tianyk/short-url/proto"
)

type ShortURLController struct {
	prefix  string
	service *shorturl.ShortURLService
}

func NewShortURLController(prefix string, shortURLService *shorturl.ShortURLService) *ShortURLController {
	return &ShortURLController{prefix: prefix, service: shortURLService}
}

// CreateShortURL generates a short URL.
func (c *ShortURLController) CreateShortURL(ctx *gin.Context) {
	body := new(CreateShortURLRequest)
	err := ctx.ShouldBind(body)
	if err != nil {
		panic(errors.Wrap(err, "Invalid request body"))
	}

	duration, _ := str2duration.ParseDuration("1d")
	if body.MaxAge != "" {
		duration, err = str2duration.ParseDuration(body.MaxAge)
		if err != nil {
			panic(HTTPError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid MaxAge %s", body.MaxAge)})
		}

		// 限制最大有效期 3 年
		if duration > 1095*24*time.Hour {
			panic(HTTPError{Status: http.StatusBadRequest, Message: "MaxAge must be less than 3 years"})
		}
	}

	var password string
	if body.Visibility == "private" {
		password = shorturl.RandomString(4)
	}

	urlID, err := c.service.CreateShortURL(&proto.ShortUrlMessage{LongUrl: body.LongUrl, Expire: time.Now().Add(duration).Unix(), Password: password})
	if err != nil {
		panic(errors.Wrap(err, "Create short url error"))
	}

	shortUrl := purell.MustNormalizeURLString(fmt.Sprintf("%s/%s", c.prefix, urlID), purell.FlagRemoveDuplicateSlashes)
	ctx.JSON(http.StatusOK, &CreateShortURLResponse{
		LongUrl:  body.LongUrl,
		ShortUrl: shortUrl,
		Password: password,
	})
}

// OpenShortURL 访问原页面
func (c *ShortURLController) OpenShortURL(ctx *gin.Context) {
	urlID := ctx.Param("urlID")

	// query 前四位为密码
	var password string
	rawQuery := ctx.Request.URL.RawQuery
	if len(rawQuery) >= 4 {
		password = rawQuery[0:4]
	}

	longUrl, err := c.service.FindLongURL(urlID, password)
	if err != nil {
		if stderrors.Is(err, shorturl.ErrNotFound) {
			ctx.String(http.StatusNotFound, "NotFound")
			return
		}
		if stderrors.Is(err, shorturl.ErrForbidden) {
			panic(HTTPError{Status: http.StatusForbidden})
		}
		panic(errors.WithMessage(err, "Open short url error"))
	}

	if isRedirectableURL(longUrl) {
		ctx.Redirect(http.StatusMovedPermanently, longUrl)
	} else {
		ctx.String(http.StatusOK, longUrl)
	}
}

// isRedirectableURL reports whether content is an absolute http(s) URL that can be redirected to.
func isRedirectableURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
