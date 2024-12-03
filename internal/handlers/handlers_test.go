package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AvdzhiV/urlShort/configs"
	"github.com/AvdzhiV/urlShort/internal/storage"
	"github.com/go-chi/chi/v5"
)

/*

Изменить тесты для обработчиков

*/

func TestShorterHandlerPost(t *testing.T) {
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}

	// Используем MemoryStorage для тестирования
	store := storage.NewMemoryStorage()
	if err := store.Init(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	handler := NewHandler(store, cfg)

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ShorterHandlerPost(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusCreated)
	}

	responseBody := rr.Body.String()
	if responseBody == "" {
		t.Errorf("empty body")
	}

	expectedPrefix := cfg.BaseURL + "/"
	if !strings.HasPrefix(responseBody, expectedPrefix) {
		t.Errorf("expected response to start with %v, got %v", expectedPrefix, responseBody)
	}

	shortURL := strings.TrimPrefix(responseBody, expectedPrefix)
	originalURL, exists := store.Get(shortURL)
	if !exists {
		t.Errorf("short URL not found in storage")
	}
	if originalURL != reqBody {
		t.Errorf("expected original URL %v, got %v", reqBody, originalURL)
	}
}

func TestShorterHandlerGet(t *testing.T) {
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}

	store := storage.NewMemoryStorage()
	if err := store.Init(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Добавляем тестовые данные в хранилище
	shortURL := "safqwe"
	originalURL := "http://example.com"
	_, err := store.Put(shortURL, originalURL)
	if err != nil {
		t.Fatalf("Failed to put data in store: %v", err)
	}

	handler := NewHandler(store, cfg)

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/{shortURL}", handler.ShorterHandlerGet)

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	if location := rr.Header().Get("Location"); location != originalURL {
		t.Errorf("wrong Location header: got %v want %v", location, originalURL)
	}
}

func TestShorterHandlerAPI(t *testing.T) {
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}

	store := storage.NewMemoryStorage()
	if err := store.Init(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	handler := NewHandler(store, cfg)

	reqBody := `{"url":"https://practicum.yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ShorterHandlerAPI(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unable to parse response: %v", err)
	}
	if resp["result"] == "" {
		t.Errorf("Expected non-empty result, got %v", resp["result"])
	}

	expectedPrefix := cfg.BaseURL + "/"
	if !strings.HasPrefix(resp["result"], expectedPrefix) {
		t.Errorf("Expected result to start with %v, got %v", expectedPrefix, resp["result"])
	}

	shortURL := strings.TrimPrefix(resp["result"], expectedPrefix)
	originalURL, exists := store.Get(shortURL)
	if !exists {
		t.Errorf("short URL not found in storage")
	}
	if originalURL != "https://practicum.yandex.ru" {
		t.Errorf("expected original URL %v, got %v", "https://practicum.yandex.ru", originalURL)
	}
}
