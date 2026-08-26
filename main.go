package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tianyk/short-url/internal/config"
	"github.com/tianyk/short-url/internal/httpapi"
	"github.com/tianyk/short-url/internal/server"
	"github.com/tianyk/short-url/internal/shorturl"
	"github.com/tianyk/short-url/internal/staticmap"
	"github.com/tianyk/short-url/internal/store"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	storage, err := store.Open(cfg.StorePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := storage.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	staticMapPath := filepath.Join(cfg.StorePath, staticmap.Filename)
	staticMappings, err := staticmap.New(staticMapPath)
	if err != nil {
		return err
	}

	shortURLService, err := shorturl.New(storage, shorturl.WithStaticMapping(staticMappings))
	if err != nil {
		return err
	}
	shortURLController := httpapi.NewShortURLController(cfg.Prefix, shortURLService)

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	cleanupCtx, stopCleanup := context.WithCancel(signalCtx)
	defer stopCleanup()
	staticMapDone, err := staticMappings.StartWatching(cleanupCtx)
	if err != nil {
		return err
	}
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		shortURLService.RunExpirationCleanup(cleanupCtx, time.Hour)
	}()

	serverErr := server.Run(signalCtx, cfg, shortURLController)
	stopCleanup()
	<-cleanupDone
	if watchErr := <-staticMapDone; watchErr != nil && serverErr == nil {
		serverErr = watchErr
	}

	return serverErr
}
