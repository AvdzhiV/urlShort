package main

import (
	"log"
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

	zap.ReplaceGlobals(logger)
	cfg := configs.ParseParts()
	if cfg == nil {
		logger.Fatal("Failed to parse configuration")
	}

	store, err := storage.StoreInit(*cfg, logger)
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

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r); err != nil {
		logger.Error("Error", zap.Error(err))
	}
}
