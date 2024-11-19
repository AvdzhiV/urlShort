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
	r.Use(middleware.GzipMiddleware)

	r.Get("/{shortURL}", handlers.ShorterHandlerGet)
	r.Post("/", handlers.ShorterHandlerPost)
	r.Post("/api/shorten", handlers.ShorterHandlerAPI)

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
