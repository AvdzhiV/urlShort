package main

import (
	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/pkg"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func main() {
	cfg := configs.ParseParts()
	r := chi.NewRouter()
	r.Get("/{shortURL}", pkg.ShorterHandlerGet)
	r.Post("/", pkg.ShorterHandlerPost)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
