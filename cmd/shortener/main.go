package main

import (
	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/app"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func main() {
	cfg := configs.ParseParts()
	r := chi.NewRouter()
	r.Get("/{shortURL}", app.ShorterHandlerGet)
	r.Post("/", app.ShorterHandlerPost)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
