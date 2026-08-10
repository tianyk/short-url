package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/fs"
)

//go:embed index.html
var indexHTML string

//go:embed static
var staticFS embed.FS

// StaticFS 静态资源文件系统
var StaticFS fs.FS = func() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}()

// staticETags 静态文件内容 hash 的 ETag 表（编译期内容固定，启动时计算一次）
var staticETags = func() map[string]string {
	tags := make(map[string]string)
	fs.WalkDir(StaticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		data, readErr := fs.ReadFile(StaticFS, path)
		if readErr != nil {
			return nil
		}
		sum := sha256.Sum256(data)
		tags[path] = fmt.Sprintf(`"%x"`, sum[:16])
		return nil
	})
	return tags
}()

// StaticETag 返回静态文件的 ETag，不存在时返回 false
func StaticETag(name string) (string, bool) {
	etag, ok := staticETags[name]
	return etag, ok
}

var tmpl = template.Must(template.New("index").Parse(indexHTML))

// Render 渲染首页，注入 base64 编码的 API token
func Render(token string) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Token string }{
		Token: base64.StdEncoding.EncodeToString([]byte(token)),
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
