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

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gz.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

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
		return nil, err
	}

	return &compressReader{
		r:  r,
		gz: gz,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.gz.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("failed to close compress reader: %w", err)
	}

	return c.gz.Close()
}

func RequestCompressorMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		compressionFn := func(w http.ResponseWriter, r *http.Request) {
			ow := w

			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				cw := newCompressWriter(w)

				defer cw.Close()
				ow = cw
			}

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Error("failed to create gzip reader", "error", err)
					ow.WriteHeader(http.StatusInternalServerError)
					return
				}
				
				defer cr.Close()
				r.Body = cr
			}
			h.ServeHTTP(ow, r)
		}

		return http.HandlerFunc(compressionFn)
	}
}