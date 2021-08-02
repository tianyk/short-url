### 短地址服务

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