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
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Println("Error, ", err)
		return
	}

	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	cfg := configs.ParseParts()
	if cfg == nil {
		logger.Fatal("Failed to parse configuration")
	}

	//TODO Переместить логику в internal
	var store storage.Storage

	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to connect to the database", zap.Error(err))
		} else {
			logger.Info("Successfully connected to the database")
		}
		dbStore := storage.NewDBStorage(db)
		if err := dbStore.Init(); err != nil {
			logger.Fatal("Failed to initialize database", zap.Error(err))
		} else {
			logger.Info("Database initialized successfully")
		}
		store = dbStore
	} else if cfg.FileStoragePath != "" {
		fileStore := storage.NewFileStorage(cfg.FileStoragePath)
		if err := fileStore.Init(); err != nil {
			logger.Fatal("Failed to initialize file storage", zap.Error(err))
		}
		store = fileStore
	} else {
		logger.Warn("No storage configuration provided, using in-memory storage")
		memStore := storage.NewMemoryStorage()
		if err := memStore.Init(); err != nil {
			logger.Fatal("Failed to initialize memory storage", zap.Error(err))
		}
		store = memStore
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
