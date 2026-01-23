package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/newmersedez/urlshort/internal/logger"
)

const (
	acceptEncodingHeader	= "Accept-Encoding"
	contentEncodingHeader	= "Content-Encoding"
	contentEncodingGZip		= "gzip"
)


type compressWriter struct {
	w	http.ResponseWriter
	gz	*gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:	w,
		gz:	gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gz.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.gz.Close()
}

type compressReader struct {
	r	io.ReadCloser
	gz	*gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:	r,
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

func GzipRequestCompressor(next http.Handler) http.Handler{
	compressionFn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(acceptEncodingHeader), contentEncodingGZip) {
			logger.Log.Debug("Client does not support gzip compression, skipping")
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get(contentEncodingHeader), contentEncodingGZip) {
			logger.Log.Debug("Compression is not needed, skipping")
			next.ServeHTTP(w, r)
			return
		}

		logger.Log.Debug("Compression is needed")
		
		compressWriter := newCompressWriter(w)
		defer compressWriter.Close()

		compressReader, err := newCompressReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = compressReader
		defer compressReader.Close()
	
		next.ServeHTTP(compressWriter, r)
	}

	return http.HandlerFunc(compressionFn)
}