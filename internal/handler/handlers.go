package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/model"
	"go.uber.org/zap"
)

const (
	contentTypeHeader         	= "Content-Type"
	contentTypeApplicationJSON	= "application/json"
	contentTypeTextPlain 		= "text/plain"

	errUnsupportedContentType	= "unsupported content type"
	errInternalServerError		= "internal server error"
)

type Repository interface {
	Get(ctx context.Context, key string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	Dispose()
}

type Shortener interface {
	Shorten(url string) (string, error)
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	Result string `json:"result"`
}

type Handlers struct {
	baseURL   string
	store     Repository
	shortener Shortener
}

func NewHandler(baseURL string, store Repository, shortener Shortener) *Handlers {
	return &Handlers{
		baseURL:   baseURL,
		store:     store,
		shortener: shortener,
	}
}

func (h *Handlers) GetOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Log.Info("id is required")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url, err := h.store.Get(ctx, id)

	if err != nil {
		logger.Log.Error("error retrieving URL", zap.Error(err))
		http.Error(w, errInternalServerError, http.StatusInternalServerError)
		return
	}
	if url == nil {
		logger.Log.Error("URL not found", zap.String("id", id))
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.Value, http.StatusTemporaryRedirect)
}

func (h Handlers) ShortenURLViaJSONHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get(contentTypeHeader); contentType != contentTypeApplicationJSON {
		logger.Log.Error(errUnsupportedContentType)
		http.Error(w, errUnsupportedContentType, http.StatusUnsupportedMediaType)
		return
	}

	var request ShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Log.Error("failed to parse request body", zap.Error(err))
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		logger.Log.Error("url value is required", zap.String("url", request.URL))
		http.Error(w, "url value is required", http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortener.Shorten(request.URL)
	if err != nil {
		logger.Log.Error("invalid URL", zap.Error(err))
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url := model.NewShortenURL(shortURL, request.URL)

	err = h.store.Add(ctx, url)
	if err != nil {
		logger.Log.Error(errInternalServerError, zap.Error(err))
		http.Error(w, errInternalServerError, http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)

	response := ShortenURLResponse{
		Result: fmt.Sprintf("%s/%s", h.baseURL, url.Key),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("failed to write response body", zap.Error(err))
		http.Error(w, errInternalServerError, http.StatusInternalServerError)
		return
	}
}

func (h Handlers) ShortenURLViaPlainTextHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get(contentTypeHeader); contentType != contentTypeTextPlain {
		logger.Log.Error(errUnsupportedContentType)
		http.Error(w, errUnsupportedContentType, http.StatusUnsupportedMediaType)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("error reading request body", zap.Error(err))
		http.Error(w, "request body is not a valid plain text", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		logger.Log.Error("URL is required")
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortenValue, err := h.shortener.Shorten(originalURL)
	if err != nil {
		logger.Log.Error("error shortening URL", zap.Error(err))
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url := model.NewShortenURL(shortenValue, originalURL)

	err = h.store.Add(ctx, url)
	if err != nil {
		logger.Log.Error("error storing URL", zap.Error(err))
		http.Error(w, errInternalServerError, http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeTextPlain)
	w.WriteHeader(http.StatusCreated)

	shortenURL := fmt.Sprintf("%s/%s", h.baseURL, url.Key)
	w.Write([]byte(shortenURL))
}
