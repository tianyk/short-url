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

## 固定短地址映射

程序会在 LevelDB 数据目录中查找 `short-url-map.conf`，完整路径为 `<APP_STORE_PATH>/short-url-map.conf`。`APP_STORE_PATH` 未配置时，默认路径是 `short-url-store/short-url-map.conf`。该文件是可选的，不存在时完全按照原有逻辑查询 LevelDB，也不会被打包进 Docker 镜像。

文件使用 UTF-8 编码，每行格式为 `key=value`：

```conf
# 空行和整行注释会被忽略
docs=https://example.com/docs
notice=系统维护中，请稍后再试
search=https://example.com/search?q=test&type=all#result
```

解析规则：

- key 必须是 4～8 位小写字母或数字，与短地址路由规则一致。
- 每行只按第一个 `=` 分隔，因此 value 中可以继续包含 `=`。
- key 和 value 两侧的空白会被去除；允许空 value。
- 只支持整行注释：一行中第一个非空字符为 `#` 时，该行才是注释；value 中的 `#` 会原样保留。
- 重复 key、无 `=` 或非法 key 会导致本次加载失败。

静态映射的优先级高于 LevelDB。同名短码即使在 LevelDB 中设置了密码或有效期，也会由该文件中的 value 覆盖；删除文件中的映射后，LevelDB 数据会重新生效。绝对 `http`/`https` URL 仍返回 301 重定向，其他内容按 UTF-8 纯文本返回。

程序会监听该文件的创建、写入、替换和删除并自动重载。运行期间遇到格式错误时会记录日志并保留上一次成功加载的映射；文件被删除后会清空静态映射。若文件在启动时已经存在但格式错误，程序会拒绝启动。

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

如需使用固定映射，先在宿主机创建文件：

```bash
touch /data/short-url-store/short-url-map.conf
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

`short-url-map.conf` 和 LevelDB 数据由同一个目录挂载提供，不需要额外增加单文件挂载。可以在宿主机直接编辑 `/data/short-url-store/short-url-map.conf`，包括采用“写临时文件后重命名”的编辑器保存方式。

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
