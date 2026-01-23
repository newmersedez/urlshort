package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	acceptEncodingHeader  = "Accept-Encoding"
	contentEncodingHeader = "Content-Encoding"
	contentEncodingGZip   = "gzip"
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
	c.w.Header().Set(contentEncodingHeader, contentEncodingGZip)
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
		return err
	}

	return c.gz.Close()
}

func RequestCompressorMiddleware(next http.Handler) http.Handler {
	compressionFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if !strings.Contains(r.Header.Get(acceptEncodingHeader), contentEncodingGZip) {
			cw := newCompressWriter(w)

			defer cw.Close()
			ow = cw
		}

		if strings.Contains(r.Header.Get(contentEncodingHeader), contentEncodingGZip) {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				ow.WriteHeader(http.StatusInternalServerError)
				return
			}
			
			defer cr.Close()
			r.Body = cr
		}
		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(compressionFn)
}