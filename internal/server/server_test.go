package server

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/tianyk/short-url/internal/config"
	"github.com/tianyk/short-url/internal/httpapi"
)

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, config.Config{Port: 0}, httpapi.NewShortURLController("", nil, nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = Run(ctx, config.Config{Port: port}, httpapi.NewShortURLController("", nil, nil))
	if err == nil {
		t.Fatal("expected listen error")
	}
}
