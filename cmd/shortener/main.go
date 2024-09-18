package main

import (
	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func main() {
	cfg := configs.ParseParts()
	r := chi.NewRouter()
	r.Get("/{shortURL}", handlers.ShorterHandlerGet)
	r.Post("/", handlers.ShorterHandlerPost)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
