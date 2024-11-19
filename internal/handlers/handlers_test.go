package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	// Создаем временный файл для хранилища
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Удаляем файл после теста

	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewStorage(tmpFile.Name())
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
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewStorage(tmpFile.Name())

	// Добавляем тестовые данные в хранилище
	shortURL := "safqwe"
	originalURL := "http://example.com"
	err = store.Put(shortURL, originalURL)
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
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewStorage(tmpFile.Name())
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
