package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/handlers"
	"github.com/AvdzhiV/urlShort/internal/middleware"
	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Get("/{shortURL}", handlers.ShorterHandlerGet)
	r.Post("/", handlers.ShorterHandlerPost)
	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r); err != nil {
		fmt.Println("Error")
	}
}
