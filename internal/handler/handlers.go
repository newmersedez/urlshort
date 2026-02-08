package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/middleware"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	Get(ctx context.Context, ID string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL model.ShortenURL) error
	AddBatch(ctx context.Context, shortenUrls []model.ShortenURL) error
	Ping(ctx context.Context) error
	Dispose()
}

type Shortener interface {
	Shorten(url string) (string, error)
}

type shortenURLRequest struct {
	URL string `json:"url"`
}

type shortenBatchURLRequest struct {
	CorrelationId	string `json:"correlation_id"`
	OriginalUrl		string `json:"original_url"`
}

type shortenURLResponse struct {
	Result string `json:"result"`
}

type shortenBatchURLResponse struct {
	CorrelationId	string `json:"correlation_id"`
	ShortlUrl		string `json:"short_url"`
}

type handlers struct {
	baseURL   string
	store     Repository
	shortener Shortener
	logger    *slog.Logger
}

func Serve(cfg config.Config, store Repository, shortener Shortener, logger *slog.Logger) error {
	handler, err := newHandlers(cfg.BaseURL, store, shortener, logger)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	server, err := newServer(cfg.ServerAddr, newRouter(handler))
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	return server.ListenAndServe()
}

func newHandlers(baseURL string, store Repository, shortener Shortener, logger *slog.Logger) (*handlers, error) {
	handlers := &handlers{
		baseURL:   baseURL,
		store:     store,
		shortener: shortener,
		logger:    logger,
	}

	return handlers, nil
}

func newRouter(handler *handlers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestLoggerMiddleware(handler.logger))
	router.Use(middleware.RequestCompressorMiddleware(handler.logger))

	router.Post("/", handler.shortenURLViaPlainTextHandle)
	router.Post("/api/shorten", handler.shortenURLViaJSONHandle)
	router.Post("/api/shorten/batch", handler.shortenBatchUrlsViaJSONHandle)
	router.Get("/{id}", handler.getOriginURLHandle)
	router.Get("/ping", handler.pingDatabaseHandle)

	return router
}

func newServer(address string, router *chi.Mux) (*http.Server, error) {
	if address == "" {
		return nil, errors.New("server address is not set")
	}

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

func (h *handlers) getOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	
	url, err := h.store.Get(ctx, id)

	if err != nil {
		h.logger.Error("error retrieving URL", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
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

	ID, err := h.shortener.Shorten(request.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	fullShortenURL, err := url.JoinPath(h.baseURL, ID)
	if err != nil {
		h.logger.Error("failed to get full url", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	existingURL, err := h.store.Get(ctx, ID)
	if err != nil {
		h.logger.Error("failed to get url from store", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if existingURL != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fullShortenURL))
		return
	}

	shortenURL := model.NewShortenURL(ID, request.URL)

	err = h.store.Add(ctx, shortenURL)
	if err != nil {
		h.logger.Error("failed to add shorten url to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := shortenURLResponse{
		Result: fullShortenURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
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

	ID, err := h.shortener.Shorten(originalURL)
	if err != nil {
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
		return
	}

	fullShortenURL, err := url.JoinPath(h.baseURL, ID)
	if err != nil {
		h.logger.Error("failed to get full url", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	existingURL, err := h.store.Get(ctx, ID)
	if err != nil {
		h.logger.Error("failed to get url from store", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if existingURL != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fullShortenURL))
		return
	}

	shortenURL := model.NewShortenURL(ID, originalURL)

	err = h.store.Add(ctx, shortenURL)
	if err != nil {
		h.logger.Error("error storing URL", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortenURL))
}

func (h *handlers) shortenBatchUrlsViaJSONHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	var requestBody []shortenBatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}


}

func (h *handlers) pingDatabaseHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		h.logger.Error("failed to connect to the database", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}