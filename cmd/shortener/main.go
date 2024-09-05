package main

import (
	"io"
	"math/rand"
	"net/http"
	"time"
)

var urlMap = make(map[string]string)

func generateShortURL() string {
	rand.Seed(time.Now().UnixNano())
	letterBytes := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func shorterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		origURL := string(body)
		shortURL := generateShortURL()

		urlMap[shortURL] = origURL

		fullShortURL := "http://localhost:8080/" + shortURL

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fullShortURL))
	}

	if r.Method == http.MethodGet {
		shortURL := r.URL.Path[1:]
		origURL, exists := urlMap[shortURL]
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)

	}

}
func main() {
	http.HandleFunc("/", shorterHandler)
	http.ListenAndServe(":8080", nil)

}
