package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		zap.L().Info(
			"Response",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("ip", r.RemoteAddr),
			zap.String("duration", time.Since(start).String()),
		)
		
	})
}

/*
Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
Сведения об ответах должны содержать код статуса и размер содержимого ответа.
*/