package handlers

import (
	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/generateurl"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

var urlMap = make(map[string]string)

func ShorterHandlerPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		origURL := string(body)
		shortURL := generateurl.GenerateShortURL()

		urlMap[shortURL] = origURL

		fullShortURL := configs.ParseParts().BaseURL + "/" + shortURL

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fullShortURL))
	}
}

func ShorterHandlerGet(w http.ResponseWriter, r *http.Request) {
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
