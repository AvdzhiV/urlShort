package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriterWrapper) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &ResponseWriterWrapper{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		zap.L().Info("Request and Response",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", wrappedWriter.status),
			zap.Int("size", wrappedWriter.size),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
