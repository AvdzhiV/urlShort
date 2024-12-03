package main

import (
	"fmt"
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
		fmt.Println("Error, ", err)
		return
	}

	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	cfg := configs.ParseParts()
	if cfg == nil {
		logger.Fatal("Failed to parse configuration")
	}

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


/*
	store := storage.NewStorage(cfg.FileStoragePath, db)
	if db == nil {
		if err := store.Load(); err != nil {
			logger.Fatal("Failed to load storage", zap.Error(err))
		}
	} else {
		if err := store.InitDB(); err != nil {
			logger.Fatal("Failed to initialize database", zap.Error(err))
		} else {
			logger.Info("Database initialized successfully")
		}
	}
*/
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.GzipMiddleware)

	handler := handlers.NewHandler(store, cfg)

	r.Get("/ping", handler.PingHandler)
	r.Get("/{shortURL}", handler.ShorterHandlerGet)
	r.Post("/", handler.ShorterHandlerPost)
	r.Post("/api/shorten", handler.ShorterHandlerAPI)

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r); err != nil {
		fmt.Println("Error")
	}
}

/*
	Задание по треку «Сервис сокращения URL»
	Добавьте поддержку gzip в ваш сервис. Научите его:
	Принимать запросы в сжатом формате (с HTTP-заголовком Content-Encoding).
	Отдавать сжатый ответ клиенту, который поддерживает обработку сжатых ответов (с HTTP-заголовком Accept-Encoding).
	Функция сжатия должна работать для контента с типами application/json и text/html.
	Вспомните middleware из урока про HTTP-сервер, это может вам помочь.
*/
