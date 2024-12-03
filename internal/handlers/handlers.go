package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/generateurl"
	"github.com/AvdzhiV/urlShort/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Store  *storage.Storage
	Config *configs.Config
}

func NewHandler(store *storage.Storage, cfg *configs.Config) *Handler {
	return &Handler{
		Store:  store,
		Config: cfg,
	}
}


	func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
		if h.Store.DB() == nil {
			http.Error(w, "Database not configured", http.StatusInternalServerError)
			return
		}
		err := h.Store.DB().Ping()
		if err != nil {
			http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

func (h *Handler) ShorterHandlerPost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	origURL := string(body)
	if origURL == "" {
        http.Error(w, "Request body is empty", http.StatusBadRequest)
        return
    }
	shortURL := generateurl.GenerateShortURL()

	err = h.Store.Put(shortURL, origURL)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	fullShortURL := h.Config.BaseURL + "/" + shortURL

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortURL))
}

func (h *Handler) ShorterHandlerGet(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")

	origURL, exists := h.Store.Get(shortURL)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
}

func (h *Handler) ShorterHandlerAPI(w http.ResponseWriter, r *http.Request) {
	type ShortenRequest struct {
		URL string `json:"url"`
	}
	type ShortenResponse struct {
		Result string `json:"result"`
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
	err := h.Store.Put(shortURL, req.URL)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	fullShortURL := h.Config.BaseURL + "/" + shortURL

	resp := ShortenResponse{Result: fullShortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
