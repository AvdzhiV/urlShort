package main

import (
	"github.com/AvdzhiV/urlShort/cmd/shortener/config"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

var urlMap = make(map[string]string)

func generateShortURL() string {
	letterBytes := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func shorterHandlerPost(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		origURL := string(body)
		shortURL := generateShortURL()

		urlMap[shortURL] = origURL

		fullShortURL := cfg.BaseURL + "/" + shortURL

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fullShortURL))
	}
}

func shorterHandlerGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		shortURL := chi.URLParam(r, "shortURL")

		origURL, exists := urlMap[shortURL]
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
	}
}

func main() {
	cfg := config.ParseParts()
	r := chi.NewRouter()
	r.Get("/{shortURL}", shorterHandlerGet)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		shorterHandlerPost(w, r, cfg)
	})
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
