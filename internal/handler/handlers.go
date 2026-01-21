package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	Get(ctx context.Context, key string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
}

type Shortener interface {
	Shorten(url string) (string, error)
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
		logger.Log.Info("id is not specified")
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url, err := h.store.Get(ctx, id)

	if err != nil {
		logger.Log.Error("error retrieving URL", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if url == nil {
		logger.Log.Error("URL not found for id", zap.String("id", id))
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.Value, http.StatusTemporaryRedirect)
}

func (h *Handlers) ShortenURLHandle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("error reading request body", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		logger.Log.Error("empty URL provided")
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	shortenURL := fmt.Sprintf("%s/%s", h.baseURL, url.Key)
	w.Write([]byte(shortenURL))
}
