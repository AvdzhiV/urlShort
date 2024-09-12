package pkg

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShorterHandlerPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://localhost:8080"))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(ShorterHandlerPost)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if rr.Body.String() == "" {
		t.Errorf("emty body")
	}
}

func TestShorterHandlerGet(t *testing.T) {
	// Инициализация тестовых данных
	urlMap["safqwe"] = "http://example.com"

	// Создание запроса
	req := httptest.NewRequest(http.MethodGet, "/safqwe", nil)
	rr := httptest.NewRecorder()

	// Настройка маршрутизатора с обработчиком
	r := chi.NewRouter()
	r.Get("/{shortURL}", ShorterHandlerGet)

	// Вызов маршрутизатора с запросом
	r.ServeHTTP(rr, req)

	// Проверка статуса ответа
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	// Проверка заголовка Location
	if location := rr.Header().Get("Location"); location != "http://example.com" {
		t.Errorf("wrong Location header: got %v want %v", location, "http://example.com")
	}
}
