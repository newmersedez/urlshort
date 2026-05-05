package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write записывает тело ответа и накапливает количество записанных байт.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader фиксирует код статуса ответа для последующего логирования.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// RequestLoggerMiddleware логирует начало и завершение каждого HTTP-запроса:
// метод, URI, статус ответа, длительность и размер тела.
func RequestLoggerMiddleware(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("Request starting",
				"method", r.Method,
				"uri", r.RequestURI)

			responseData := new(responseData)

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)
			logger.Info("Request finished",
				"method", r.Method,
				"uri", r.RequestURI,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size)
		}
		return http.HandlerFunc(logFn)
	}
}
