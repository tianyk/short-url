FROM golang:1.16 AS builder
# PROXY 
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /usr/src/app/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /usr/src/app/
COPY --from=builder /usr/src/app/app .

EXPOSE 4000

CMD ["./app"]
