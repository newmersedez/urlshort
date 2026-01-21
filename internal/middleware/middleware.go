package middleware

import (
	"net/http"
	"time"

	"github.com/newmersedez/urlshort/internal/logger"
	"go.uber.org/zap"
)

type (
    responseData struct {
        status int
        size int
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

func RequestLogger(h http.Handler) http.Handler {
    logFn := func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        logger.Log.Info(
			"Request starting",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
        )

        responseData := new(responseData)
		
        lw := loggingResponseWriter {
            ResponseWriter: w,
            responseData: responseData,
        }
		
        h.ServeHTTP(&lw, r)

        duration := time.Since(start)

        logger.Log.Info(
			"Request finished",
            zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
            zap.Int("status", responseData.status),
            zap.Duration("duration", duration),
            zap.Int("size", responseData.size),
        )
    }
    return http.HandlerFunc(logFn)
}