# Short URL Service

[English](README.md) | [简体中文](README.zh-CN.md)

A short URL service built with Go, Gin, and LevelDB. It supports custom expiration times and private short links protected by access passwords.

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `APP_PREFIX` | Yes | - | The short-link domain prefix, such as `https://example.com/` |
| `APP_TOKEN` | Required in production | Empty (authentication is skipped when empty) | Authentication token for the short-link creation API |
| `APP_PORT` | No | `5000` | Listening port |
| `APP_STORE_PATH` | No | `short-url-store` | LevelDB data directory |
| `APP_ENV` | No | `development` | Environment name used to load `.env.<env>` |

The application supports `.env`, `.env.local`, and `.env.<APP_ENV>` files. Actual environment variables take precedence over values from these files.

## Fixed Short URL Mappings

The application looks for `short-url-map.conf` in the LevelDB data directory. Its full path is `<APP_STORE_PATH>/short-url-map.conf`. If `APP_STORE_PATH` is not configured, the default path is `short-url-store/short-url-map.conf`. This file is optional. If it does not exist, the application queries LevelDB as usual. The file is not included in the Docker image.

The file must use UTF-8 encoding, with one `key=value` mapping per line:

```conf
# Blank lines and full-line comments are ignored
docs=https://example.com/docs
notice=The system is under maintenance. Please try again later.
search=https://example.com/search?q=test&type=all#result
```

Parsing rules:

- Keys must be between 1 and 64 characters long. They may contain ASCII letters, digits, and the characters `-`, `.`, `_`, and `~`. Keys are case-sensitive.
- `.`, `..`, `api`, `static`, and `_heartbeat` are reserved path or route names and cannot be used as keys.
- Each line is split only at the first `=`, so the value may contain additional `=` characters.
- Leading and trailing whitespace is removed from both the key and value. Empty values are allowed.
- Only full-line comments are supported. A line is treated as a comment only when its first non-whitespace character is `#`; any `#` characters in a value are preserved.
- Duplicate keys, lines without `=`, or invalid keys cause the current load to fail.

Static mappings are processed before the standard short-code validation and LevelDB lookup. If the same short code exists in LevelDB, the value from this file takes precedence even if the LevelDB entry has a password or expiration time. After a mapping is removed from the file, the LevelDB entry becomes active again. When no static mapping matches, the request continues through the standard validation, which accepts only 4- to 8-character short codes made up of lowercase letters or digits, before querying LevelDB. Absolute `http` or `https` URLs still return a `301` redirect; all other values are returned as UTF-8 plain text.

The application watches the file for creation, modification, replacement, and deletion, and reloads it automatically. If a format error occurs while the application is running, the error is logged and the most recently loaded valid mappings remain active. Deleting the file clears all static mappings. If the file already exists at startup but contains invalid data, the application refuses to start.

## Development

```bash
go test -v ./...
```

```bash
go run main.go
```

## Regenerating Protobuf Code

Both `protoc` and `protoc-gen-gofast` (gogo/protobuf) must be available in your `PATH`:

```bash
protoc --gofast_out=. proto/short_url.proto
```

## Docker

```bash
docker build -t shorturl:release .
```

To use fixed mappings, first create the file on the host:

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

The `short-url-map.conf` file and LevelDB data are provided through the same mounted directory, so no additional single-file mount is required. You can edit `/data/short-url-store/short-url-map.conf` directly on the host, including with editors that save by writing a temporary file and then renaming it.

Alternatively, use `docker-compose.yml`, which maps `127.0.0.1:5050` to container port `3000` by default.

## API

Create a short link:

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

- `maxAge`: Expiration time in `go-str2duration` syntax, such as `1d` or `30m`. The default is `1d`, and the maximum is 3 years.
- `visibility`: Either `public` (the default, no password required) or `private` (requires a 4-character access password).

When opening a private short link, include its 4-character password at the beginning of the query string: `http://example.com/AbCd12ef?P4ss`.

## Reverse Proxy Example

```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
    }
}
```

Health check endpoint: `GET /_heartbeat`.
