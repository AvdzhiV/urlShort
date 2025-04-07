package main

import (
	"context"

	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/handlers"
	"github.com/AvdzhiV/urlShort/internal/middleware"
	"github.com/AvdzhiV/urlShort/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Println("Error, ", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Println("Error, ", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zap.ReplaceGlobals(logger)
	cfg := configs.ParseParts()
	if cfg == nil {
		logger.Fatal("Failed to parse configuration")
	}

	store, err := storage.NewStore(ctx, *cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize storage: ", zap.Error(err))
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)

	handler := handlers.NewHandler(store, cfg)

	r.Get("/ping", handler.PingHandler)
	r.Get("/{shortURL}", handler.ShorterHandlerGet)
	r.Post("/", handler.ShorterHandlerPost)
	r.Post("/api/shorten", handler.ShorterHandlerAPI)
	r.Post("/api/shorten/batch", handler.ShorterHandlerBatch)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: r,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Error", zap.Error(err))
	}
}
