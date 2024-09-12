package main

import (
	"net/http"
	"strconv"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := configs.ParseParts()
	r := chi.NewRouter()
	r.Get("/{shortURL}", shorterHandlerGet)
	r.Post("/", shorterHandlerPost)
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
