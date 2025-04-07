package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	_ "github.com/AvdzhiV/urlShort/internal/storage"
	"go.uber.org/zap"
)

// GzipMiddleware добавляет поддержку GZIP-сжатия.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Декомпрессия тела запроса
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer func() {
				err := reader.Close()
				if err != nil {
					zap.L().Error("Failed to close gzip reader", zap.Error(err))
				}
			}()
			r.Body = io.NopCloser(reader)
		}

		// Обёртка для сжатия ответа
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// ответ будет сжат
		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(w)
		defer func() {
			err := gzipWriter.Close()
			if err != nil {
				zap.L().Error("Failed to close gzip writer", zap.Error(err))
			}
		}()
		wrappedWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gzipWriter,
		}
		next.ServeHTTP(wrappedWriter, r)
	})
}

// обертка для GZIP-сжатия.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	n, err := w.Writer.Write(data)
	if err != nil {
		return n, fmt.Errorf("gzipResponseWriter write error: %w", err)
	}
	return n, nil
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}
