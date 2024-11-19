package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)
type ResponseWriterWrapper struct {
	http.ResponseWriter
	Status int
	Size   int
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.Status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriterWrapper) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.Size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter
		wrappedWriter := &ResponseWriterWrapper{ResponseWriter: w, Status: http.StatusOK}

		// Передаём обработку следующему обработчику
		next.ServeHTTP(wrappedWriter, r)

		// Логируем информацию о запросе и ответе
		zap.L().Info("Request and Response",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", wrappedWriter.Status),
			zap.Int("size", wrappedWriter.Size),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

/*
Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
Сведения об ответах должны содержать код статуса и размер содержимого ответа.
*/