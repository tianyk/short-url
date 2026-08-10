package httpapi

import (
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tianyk/short-url/internal/web"
)

// staticMaxAge 静态资源缓存时长：一年，配合内容 hash 的 ETag 使用
const staticMaxAge = 365 * 24 * time.Hour

// Static 服务静态资源：支持 ETag 条件请求（304）与 Range 请求
func Static(ctx *gin.Context) {
	name := strings.TrimPrefix(ctx.Param("filepath"), "/")

	etag, ok := web.StaticETag(name)
	if !ok {
		ctx.Status(http.StatusNotFound)
		return
	}

	file, err := http.FS(web.StaticFS).Open(name)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.IsDir() {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.Header("Cache-Control", "public, max-age="+strconv.FormatInt(int64(staticMaxAge/time.Second), 10)+", immutable")
	ctx.Header("ETag", etag)
	ctx.Header("Content-Type", mime.TypeByExtension(filepath.Ext(name)))

	// 处理 If-None-Match（304）与 Range（206）
	http.ServeContent(ctx.Writer, ctx.Request, stat.Name(), stat.ModTime(), file)
}
