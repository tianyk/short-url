# 短地址服务

基于 Go + Gin + LevelDB 的短地址服务，支持自定义有效期与私密短链（访问密码）。

## 环境变量

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `APP_PREFIX` | 是 | - | 短链域名前缀，如 `https://example.com/` |
| `APP_TOKEN` | 生产环境必填 | 空（空则跳过鉴权） | 创建短链 API 的鉴权 token |
| `APP_PORT` | 否 | `5000` | 监听端口 |
| `APP_STORE_PATH` | 否 | `short-url-store` | LevelDB 数据目录 |
| `APP_ENV` | 否 | `development` | 环境名，用于加载 `.env.<env>` |

支持 `.env` / `.env.local` / `.env.<APP_ENV>` 文件，真实环境变量优先。

## 开发

```bash
go test -v ./...
```

```bash
go run main.go
```

## 重新生成 protobuf 代码

需要 `protoc` 与 `protoc-gen-gofast`（gogo/protobuf）已在 `PATH` 中：

```bash
protoc --gofast_out=. proto/short_url.proto
```

## Docker

```bash
docker build -t shorturl:release .
```

```bash
docker run --name=short-url \
	-d \
	--restart=always \
	-p 127.0.0.1:3000:3000/tcp \
	--env APP_PREFIX="http://example.com/" \
	--env APP_TOKEN="YOUR_TOKEN" \
	--env APP_STORE_PATH="/usr/src/app/short-url-store" \
	-v /data/short-url-store:/usr/src/app/short-url-store \
	doog/short-url:latest
```

或使用 `docker-compose.yml`（默认映射 `127.0.0.1:5050:3000`）。

## API

创建短链：

```bash
curl -X POST \
  http://127.0.0.1:3000/api/short-url \
  -H 'content-type: application/json' \
  -H 'x-token: YOUR_TOKEN' \
  -d '{
	"longUrl": "http://www.example.com",
	"maxAge": "1d",
	"visibility": "public"
}'
```

- `maxAge`：有效期，格式为 `go-str2duration` 语法（如 `1d`、`30m`），默认 `1d`，最长 3 年
- `visibility`：`public`（默认，无需密码）或 `private`（访问需带 4 位密码）

访问短链时，私密短链需在 query 前 4 位携带密码：`http://example.com/AbCd12ef?P4ss`。

## 反向代理示例

```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
    }
}
```

健康检查端点：`GET /_heartbeat`。
