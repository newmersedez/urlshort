package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	gz *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		gz: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки оригинального ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные через gzip-компрессор.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gz.Write(p)
}

// WriteHeader устанавливает заголовок Content-Encoding: gzip и записывает статус ответа.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close завершает запись и сбрасывает буфер gzip-компрессора.
func (c *compressWriter) Close() error {
	return c.gz.Close()
}

type compressReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gzip reader object: %w", err)
	}

	return &compressReader{
		r:  r,
		gz: gz,
	}, nil
}

// Read читает и распаковывает данные из gzip-потока.
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.gz.Read(p)
}

// Close закрывает оригинальный ReadCloser и gzip-ридер.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("failed to close compress reader object: %w", err)
	}

	return c.gz.Close()
}

// RequestCompressorMiddleware обрабатывает gzip-сжатие запросов и ответов.
// Если клиент принимает gzip (Accept-Encoding: gzip), ответ сжимается.
// Если тело запроса сжато (Content-Encoding: gzip), оно автоматически распаковывается.
func RequestCompressorMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				cw := newCompressWriter(w)

				defer cw.Close()
				ow = cw
			}

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Error("failed to initialize gzip reader object", "error", err)
					ow.WriteHeader(http.StatusInternalServerError)
					return
				}

				defer cr.Close()
				r.Body = cr
			}
			h.ServeHTTP(ow, r)
		})
	}
}
