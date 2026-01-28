package middleware

import (
	"net/http"
	"time"
)

type Logger interface {
	Debug(msg string, args...any)
	Info(msg string, args...any)
	Warn(msg string, args...any)
	Error(msg string, args...any)
	Dispose()
}

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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLoggerMiddleware(logger Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("Request starting %s %s", r.Method, r.RequestURI)

			responseData := new(responseData)

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)
			logger.Info("Request finished %s %s - %d - %s - %d bytes", r.Method, r.RequestURI, responseData.status, duration, responseData.size)
		}
		return http.HandlerFunc(logFn)
	}
}
