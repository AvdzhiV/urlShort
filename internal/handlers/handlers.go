package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/generateurl"
	"github.com/go-chi/chi/v5"
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

func ShorterHandlerAPI(w http.ResponseWriter, r *http.Request) {
	type ShortenRequest struct {
		URL string `json:"url"` //orig 
	}
	type ShortenResponse struct {
		Result string `json:"result"` // short 
	}

	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	shortURL := generateurl.GenerateShortURL()
	urlMap[shortURL] = req.URL

	fullShortURL := configs.ParseParts().BaseURL + "/" + shortURL

	resp := ShortenResponse{Result: fullShortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
