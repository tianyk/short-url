### 短地址服务

```bash
go test -v ./...
```

```bash
protoc --gofast_out=. **/*.proto
```

```bash
go run main.go
```

```bash
docker run --name=short-url \
	-d \
	--restart=always \
	-p 127.0.0.1:3000:3000/tcp \
	--env APP_PREFIX="http://example.com/" \
	--env APP_TOKEN="YOUR_TOKEN" \
	--env APP_STORE_PATH="/usr/src/app/short-url-store" \
	-v /data/short-url-store:/usr/src/app/short-url-store  \
	doog/short-url:latest
```

```bash
curl -X POST \
  http://127.0.0.1:3000/api/short-url \
  -H 'content-type: application/json' \
  -H 'x-token: YOUR_TOKEN' \
  -d '{
	"longUrl": "http://www.example.com",
	"maxAge": "1d"
}'
```

```nginx
server {
    listen 80;
    server_name example.com;
    
    location / {
        proxy_pass http://127.0.0.1:3000;
    }
}
```

```shell
docker build \
  -t shorturl:release \
  .
```

## 兼容性敏感的重命名 TODO(暂不修改)

以下命名问题因涉及已存数据或对外 API 契约，修改前需评估兼容性：

1. **proto 字段名** (proto/short_url.proto)：`longUrl` → `longURL`、`expire` → `expiresAt`（proto3 按 field number 序列化，改名不破坏已存数据；需要 protoc 重新生成 `short_url.pb.go`，并同步改 Go 侧引用）。
2. **proto 消息名/生成标识符**：`ShortUrlMessage` → `ShortURLMessage`（同依赖 protoc 重新生成）。
3. **存储 key 前缀** (internal/shorturl/service.go)：`shortURLPrefix = "s-"` 与 `urlIDCounterKey = "URL-ID"` 风格不统一，但改动会破坏已有 LevelDB 数据查找。
4. **ID 编码算法常量**：`digitMap`、`encodeChars`、`initialURLID` 的值已用于生成历史短码，改动会破坏旧链接解析。
5. **API JSON key**：`longUrl`/`shortUrl`/`maxAge`/`scope` 是对外契约（internal/httpapi/types.go、README 示例、前端 index.html 均使用），改名需前后端同步发版。

注：`Scope` 字段已改名 `Visibility`，但保留 `json:"scope"` tag 以维持 API 兼容；`CreateShortURLResponse` 同理保留原 JSON key。
