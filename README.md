### 短地址服务

```bash
go run main.go
```
```bash
docker run --name=short-url \
	-d \
	--net=host \
	--restart=always \
	--env APP_PREFIX="http://example.com/" \
	--env APP_TOKEN="YOUR_TOKEN"
	-v /data/short-url-store:/usr/src/app/short-url-store  \
	short-url:latest
```

```bash
curl -X POST \
  http://localhost:3000/api/short-url \
  -H 'content-type: application/json' \
  -H 'x-token: YOUR_TOKEN' \
  -d '{
	"longUrl": "http://www.example.com"
}'
```

```nginx
server {
    listen 80;
    server_name d.kekek.cc;
    
    location / {
        proxy_pass http://127.0.0.1:3000;
    }
}
```