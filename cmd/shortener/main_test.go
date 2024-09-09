package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShorterHandlerPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://localhost:8080"))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(shorterHandlerPost)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if rr.Body.String() == "" {
		t.Errorf("emty body")
	}
}

func TestShorterHandlerGet(t *testing.T) {
	urlMap["safqwe"] = "http://ya.ru"
	req := httptest.NewRequest(http.MethodGet, "/safqwe", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(shorterHandlerGet)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("wrong status code: ")
	}
	if location := rr.Header().Get("location"); location != "http://ya.ru" {
		t.Errorf("wrong Location header")
	}

}
