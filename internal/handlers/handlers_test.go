package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

/*

Изменить тесты для обработчиков

 */
 
func TestShorterHandlerPost(t *testing.T) {
	//req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://localhost:8080"))
	rr := httptest.NewRecorder()

	//handler := http.HandlerFunc(ShorterHandlerPost)
	//handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if rr.Body.String() == "" {
		t.Errorf("emty body")
	}
}

func TestShorterHandlerGet(t *testing.T) {
	// Инициализация тестовых данных
	//urlMap["safqwe"] = "http://example.com"

	// Создание запроса
	req := httptest.NewRequest(http.MethodGet, "/safqwe", nil)
	rr := httptest.NewRecorder()

	// Настройка маршрутизатора с обработчиком
	r := chi.NewRouter()
	//r.Get("/{shortURL}", ShorterHandlerGet)

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

func TestShorterHandlerAPI(t *testing.T) {

	reqBody := `{"url":"https://practicum.yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	//вызов обработчкиа
	//ShorterHandlerAPI(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	//check res body
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unable to parse response: %v", err)
	}
	if resp["result"] == "" {
		t.Errorf("Expected non-empty result, got %v", resp["result"])
	}
}
