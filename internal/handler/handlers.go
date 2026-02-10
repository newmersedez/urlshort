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
	"github.com/newmersedez/urlshort/internal/repository/db"
)

type Repository interface {
	Get(ctx context.Context, id string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	AddBatch(ctx context.Context, shortenURLs []*model.ShortenURL) error
	Ping(ctx context.Context) error
	Close()
}

type Shortener interface {
	Shorten(url string) (string, error)
}

type shortenURLRequest struct {
	URL string `json:"url"`
}

type shortenBatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type shortenURLResponse struct {
	Result string `json:"result"`
}

type shortenBatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortlURL     string `json:"short_url"`
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
		return fmt.Errorf("failed to initialize handlers object: %w", err)
	}

	server, err := newServer(cfg.ServerAddr, newRouter(handler))
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
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

	router.Post("/", handler.shortenURLHandle)
	router.Post("/api/shorten", handler.enhancedShortenURLHandle)
	router.Post("/api/shorten/batch", handler.shortenBatchURLsHandle)
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

	url, err := h.store.Get(r.Context(), id)

	if err != nil {
		h.logger.Error("error retrieving URL from repository", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
}

func (h *handlers) enhancedShortenURLHandle(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.shortener.Shorten(request.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	fullShortenURL, err := url.JoinPath(h.baseURL, id)
	if err != nil {
		h.logger.Error("failed to build shorten URL", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortenURL := model.NewShortenURL(id, request.URL)
	response := shortenURLResponse{
		Result: fullShortenURL,
	}

	err = h.store.Add(r.Context(), shortenURL)
	if err != nil {
		if errors.Is(err, db.ErrUniqueViolation) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		h.logger.Error("failed to save shorten URL to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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

func (h *handlers) shortenURLHandle(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.shortener.Shorten(originalURL)
	if err != nil {
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
		return
	}

	fullShortenURL, err := url.JoinPath(h.baseURL, id)
	if err != nil {
		h.logger.Error("failed to build full shorten URL", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortenURL := model.NewShortenURL(id, originalURL)

	err = h.store.Add(r.Context(), shortenURL)
	if err != nil {
		if errors.Is(err, db.ErrUniqueViolation) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fullShortenURL))
			return
		}

		h.logger.Error("failed to save shorten URL to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortenURL))
}

func (h *handlers) shortenBatchURLsHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	var requestBody []shortenBatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	responseBody := make([]shortenBatchURLResponse, 0, len(requestBody))
	shortenURLs := make([]*model.ShortenURL, 0, len(requestBody))

	for _, item := range requestBody {
		id, err := h.shortener.Shorten(item.OriginalURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid URL %s", item.OriginalURL), http.StatusBadRequest)
			return
		}

		shortenURL, err := url.JoinPath(h.baseURL, id)
		if err != nil {
			h.logger.Error("failed to build full shorten URL", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		shortenURLs = append(shortenURLs, model.NewShortenURL(id, item.OriginalURL))
		responseBody = append(responseBody, shortenBatchURLResponse{
			CorrelationID: item.CorrelationID,
			ShortlURL:     shortenURL,
		})
	}

	err := h.store.AddBatch(r.Context(), shortenURLs)
	if err != nil {
		h.logger.Error("failed to save shorten URLs to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(responseBody)

	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handlers) pingDatabaseHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		h.logger.Error("failed to connect to the database", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
