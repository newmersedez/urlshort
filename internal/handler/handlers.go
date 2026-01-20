package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	Get(ctx context.Context, key string) (*model.ShortenURL, error)
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
}

type Shortener interface {
	Shorten(url string) (string, error)
}

type Handlers struct {
	baseURL		string
	store		Repository
	shortener	Shortener
	logger		*log.Logger
}

func NewHandler(baseURL string, store Repository, shortener Shortener, logger *log.Logger) *Handlers {
	return &Handlers{
		baseURL: baseURL,
		store: store,
		shortener: shortener,
		logger: logger,
	}
}

func (h *Handlers) GetOriginUrlHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.Println("id not specified")
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url, err := h.store.Get(ctx, id)
	
	if err != nil {
		h.logger.Printf("error retrieving URL: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if url == nil {
		h.logger.Printf("URL not found for id: %s", id)
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	h.logger.Printf("Redirecting %s to %s", id, url.Value)
	http.Redirect(w, r, url.Value, http.StatusTemporaryRedirect)
}

func (h *Handlers) ShortenURLHandle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("error reading request body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		h.logger.Println("empty URL provided")
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	shortenValue, err := h.shortener.Shorten(originalURL)
	if err != nil {
		h.logger.Printf("error shortening URL: %v", err)
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	url := model.NewShortenURL(shortenValue, originalURL)
	
	err = h.store.Add(ctx, url)
	if err != nil {
		h.logger.Printf("error storing URL: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	shortenURL := fmt.Sprintf("%s/%s", h.baseURL, url.Key)
	h.logger.Printf("Successfully shortened %s to %s", url.Value, shortenURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenURL))
}
