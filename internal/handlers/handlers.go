package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/generateurl"
	"github.com/AvdzhiV/urlShort/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	HeaderContentType = "Content-Type"
	TextPlain         = "text/plain"
	ApplicationJSON   = "application/json"
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
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	origURL := string(body)
	if origURL == "" {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	shortURL := generateurl.GenerateShortURL()
	existingShortURL, err := h.Store.Put(shortURL, origURL)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			// Возвращаем существующий shortURL с статусом 409 Conflict
			fullShortURL := h.Config.BaseURL + "/" + existingShortURL
			w.Header().Set(HeaderContentType, TextPlain)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write([]byte(fullShortURL)); err != nil {
				zap.L().Error("failed to write response", zap.Error(err))
				return
			}
			return
		}

		zap.L().Error("Failed to save URL")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Вставка успешна, возвращаем новый shortURL
	fullShortURL := h.Config.BaseURL + "/" + existingShortURL
	w.Header().Set(HeaderContentType, TextPlain)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(fullShortURL)); err != nil {
		zap.L().Error("failed to write response", zap.Error(err))
		return
	}
}

func (h *Handler) ShorterHandlerGet(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")

	origURL, ok := h.Store.Get(shortURL)
	if !ok {
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
	existingShortURL, err := h.Store.Put(shortURL, req.URL)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			fullShortURL := h.Config.BaseURL + "/" + existingShortURL
			resp := ShortenResponse{Result: fullShortURL}
			w.Header().Set(HeaderContentType, ApplicationJSON)
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				zap.L().Error("Failed to encode JSON response", zap.Error(err))
				return
			}
			return
		}
		zap.L().Error("Failed to save URL", zap.Error(err))
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	fullShortURL := h.Config.BaseURL + "/" + existingShortURL
	resp := ShortenResponse{Result: fullShortURL}
	w.Header().Set(HeaderContentType, ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		zap.L().Error("Failed to encode JSON response", zap.Error(err))
		return
	}
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if dbStorage, ok := h.Store.(*storage.DBStorage); ok {
		if err := dbStorage.Pool.Ping(r.Context()); err != nil {
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

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqItems); err != nil {
		zap.L().Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(reqItems) == 0 {
		zap.L().Error("Empty batch received")
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	records := make([]storage.BatchRecord, 0, len(reqItems))
	for _, item := range reqItems {
		if item.OriginalURL == "" || item.CorrelationID == "" {
			zap.L().Error("Invalid request data",
				zap.String("correlation_id", item.CorrelationID),
				zap.String("original_url", item.OriginalURL))
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		// Проверка существующих URL перед вставкой
		existingShortURL, found := h.Store.GetShortURLByOriginalURL(item.OriginalURL)
		if found {
			records = append(records, storage.BatchRecord{
				ShortURL:    existingShortURL,
				OriginalURL: item.OriginalURL,
			})
		} else {
			records = append(records, storage.BatchRecord{
				ShortURL:    generateurl.GenerateShortURL(),
				OriginalURL: item.OriginalURL,
			})
		}
	}

	// Вставка записей в хранилище
	shortURLs, err := h.Store.PutBatch(records)
	if err != nil {
		zap.L().Error("Failed to save batch records", zap.Error(err))
		http.Error(w, "Failed to save URLs", http.StatusInternalServerError)
		return
	}

	respItems := make([]BatchResponseItem, 0, len(reqItems))
	for i, item := range reqItems {
		respItems = append(respItems, BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      h.Config.BaseURL + "/" + shortURLs[i],
		})
	}

	w.Header().Set(HeaderContentType, ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(respItems); err != nil {
		zap.L().Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	zap.L().Info("Batch response sent successfully", zap.Int("count", len(respItems)))
}
