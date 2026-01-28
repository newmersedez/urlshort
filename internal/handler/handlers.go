package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/middleware"
	"github.com/newmersedez/urlshort/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	Get(ctx context.Context, key string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	Dispose()
}

type Shortener interface {
	Shorten(url string) (string, error)
}

type shortenURLRequest struct {
	URL string `json:"url"`
}

type shortenURLResponse struct {
	Result string `json:"result"`
}

type handlers struct {
	baseURL   string
	store     Repository
	shortener Shortener
}

func Serve(cfg config.Config, store Repository, shortener Shortener) error {
	handler := newHandlers(cfg.BaseURL, store, shortener)
	router := newRouter(handler)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return server.ListenAndServe()
}

func newHandlers(baseURL string, store Repository, shortener Shortener) *handlers {
	return &handlers{
		baseURL:   baseURL,
		store:     store,
		shortener: shortener,
	}
}

func newRouter(handler *handlers) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware)
	router.Use(middleware.RequestCompressorMiddleware)

	router.Post("/", handler.shortenURLViaPlainTextHandle)
	router.Post("/api/shorten", handler.shortenURLViaJSONHandle)
	router.Get("/{id}", handler.getOriginURLHandle)

	return router
}

func (h *handlers) getOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url, err := h.store.Get(ctx, id)

	if err != nil {
		logger.Log.Error("error retrieving URL", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.Value, http.StatusTemporaryRedirect)
}

func (h *handlers) shortenURLViaJSONHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	var request shortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "url value is required", http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortener.Shorten(request.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url := model.NewShortenURL(shortURL, request.URL)

	err = h.store.Add(ctx, url)
	if err != nil {
		logger.Log.Error("failed to add shorten url to the storage", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := shortenURLResponse{
		Result: fmt.Sprintf("%s/%s", h.baseURL, url.Key),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("failed to write response body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handlers) shortenURLViaPlainTextHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request body is not a valid plain text", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortenValue, err := h.shortener.Shorten(originalURL)
	if err != nil {
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	shortenURL := model.NewShortenURL(shortenValue, originalURL)

	err = h.store.Add(ctx, shortenURL)
	if err != nil {
		logger.Log.Error("error storing URL", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	fullShortenURL, err := url.JoinPath(h.baseURL, shortenURL.Key)
	if err != nil {
		logger.Log.Error("failed to get full url: %w", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fullShortenURL))
}
