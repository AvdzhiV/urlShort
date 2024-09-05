package main

import (
	"io"
	"math/rand"
	"net/http"
	"time"
)

var urlMap = make(map[string]string)

func generateShortUrl() string {
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
		origUrl := string(body)
		shortUrl := generateShortUrl()

		urlMap[shortUrl] = origUrl

		fullShortUrl := "http://localhost:8080/" + shortUrl

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fullShortUrl))
	}

	if r.Method == http.MethodGet {
		shortUrl := r.URL.Path[1:]
		origUrl, exists := urlMap[shortUrl]
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, origUrl, http.StatusTemporaryRedirect)

	}

}
func main() {
	http.HandleFunc("/", shorterHandler)
	http.ListenAndServe(":8080", nil)

}
