package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/generateurl"
	"github.com/AvdzhiV/urlShort/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Store  storage.Storage
	Config *configs.Config
}

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewHandler(store storage.Storage, cfg *configs.Config) *Handler {
	return &Handler{
		Store:  store,
		Config: cfg,
	}
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
		if err.Error() == "url_exists" {
            existingShortURL, _ := h.Store.GetShortURLByOriginalURL(origURL)
            fullShortURL := h.Config.BaseURL + "/" + existingShortURL
            w.Header().Set("Content-Type", "text/plain")
            w.WriteHeader(http.StatusConflict)
            w.Write([]byte(fullShortURL))
            return
		}
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
		if err.Error() == "url_exists" {
			existingShortURL, _ := h.Store.GetShortURLByOriginalURL(req.URL)
            fullShortURL := h.Config.BaseURL + "/" + existingShortURL
            resp := ShortenResponse{Result: fullShortURL}
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusConflict)
            json.NewEncoder(w).Encode(resp)
            return
		}
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	fullShortURL := h.Config.BaseURL + "/" + shortURL

	resp := ShortenResponse{Result: fullShortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Обновление PingHandler для проверки базы данных
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if dbStorage, ok := h.Store.(*storage.DBStorage); ok {
		err := dbStorage.DB.Ping()
		if err != nil {
			http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
	}
}

func (h *Handler) ShorterHandlerBatch(w http.ResponseWriter, r *http.Request) {
	var reqItems []BatchRequestItem

	// Читаем тело запроса
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqItems); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(reqItems) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	// Обрабатываем каждый элемент
	var respItems []BatchResponseItem
	var records []storage.BatchRecord

	for _, item := range reqItems {
		if item.OriginalURL == "" || item.CorrelationID == "" {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}
		shortURL := generateurl.GenerateShortURL()
		fullShortURL := h.Config.BaseURL + "/" + shortURL

		respItems = append(respItems, BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fullShortURL,
		})

		records = append(records, storage.BatchRecord{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: item.OriginalURL,
		})
	}

	// Сохраняем в хранилище
	err := h.Store.PutBatch(records)
	if err != nil {
		http.Error(w, "Failed to save URLs", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(respItems); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
