FROM golang:1.16 AS builder
# PROXY 
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /usr/src/app/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shorturl

FROM alpine:latest
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
#    && apk --no-cache add ca-certificates

ENV GIN_MODE=release
ENV APP_PORT=3000

WORKDIR /usr/src/app/
COPY --from=builder /usr/src/app/shorturl /usr/src/app/.env ./

EXPOSE 3000

CMD ["./shorturl"]
