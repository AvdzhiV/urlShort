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
	// Создаем тестовую конфигурацию
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}

	// Создаем тестовое хранилище
	store := storage.NewStorage("") // Пустой путь означает, что данные не будут сохранены на диск
	handler := NewHandler(store, cfg)

	// Создаем тестовый запрос
	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler.ShorterHandlerPost(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Проверяем тело ответа
	responseBody := rr.Body.String()
	if responseBody == "" {
		t.Errorf("empty body")
	}

	// Дополнительно можем проверить, что URL корректно сокращен
	expectedPrefix := cfg.BaseURL + "/"
	if !strings.HasPrefix(responseBody, expectedPrefix) {
		t.Errorf("expected response to start with %v, got %v", expectedPrefix, responseBody)
	}

	// Проверяем, что URL сохранен в хранилище
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
	// Создаем тестовую конфигурацию
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}
	// Создаем тестовое хранилище и добавляем тестовые данные
	store := storage.NewStorage("")
	shortURL := "safqwe"
	originalURL := "http://example.com"
	store.Put(shortURL, originalURL)

	// Создаем экземпляр обработчика
	handler := NewHandler(store, cfg)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
	rr := httptest.NewRecorder()

	// Создаем маршрутизатор chi и регистрируем маршрут
	r := chi.NewRouter()
	r.Get("/{shortURL}", handler.ShorterHandlerGet)

	// Используем маршрутизатор для обработки запроса
	r.ServeHTTP(rr, req)

	// Проверка статуса ответа
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	// Проверка заголовка Location
	if location := rr.Header().Get("Location"); location != originalURL {
		t.Errorf("wrong Location header: got %v want %v", location, originalURL)
	}
}

func TestShorterHandlerAPI(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &configs.Config{
		BaseURL: "http://localhost:8080",
	}

	// Создаем тестовое хранилище
	store := storage.NewStorage("")
	handler := NewHandler(store, cfg)

	// Создаем тестовый запрос
	reqBody := `{"url":"https://practicum.yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler.ShorterHandlerAPI(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Проверяем тело ответа
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unable to parse response: %v", err)
	}
	if resp["result"] == "" {
		t.Errorf("Expected non-empty result, got %v", resp["result"])
	}

	// Проверяем, что результат содержит корректный сокращенный URL
	expectedPrefix := cfg.BaseURL + "/"
	if !strings.HasPrefix(resp["result"], expectedPrefix) {
		t.Errorf("Expected result to start with %v, got %v", expectedPrefix, resp["result"])
	}

	// Проверяем, что URL сохранен в хранилище
	shortURL := strings.TrimPrefix(resp["result"], expectedPrefix)
	originalURL, exists := store.Get(shortURL)
	if !exists {
		t.Errorf("short URL not found in storage")
	}
	if originalURL != "https://practicum.yandex.ru" {
		t.Errorf("expected original URL %v, got %v", "https://practicum.yandex.ru", originalURL)
	}
}
