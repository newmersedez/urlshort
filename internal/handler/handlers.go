package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	Get(key string) (*model.ShortenURL, error)
	Add(shortenURL *model.ShortenURL) error
}

type URLShortenerService interface {
	Shorten(url string) (string, error)
}

type Handler struct {
	baseURL      string
	store        Repository
	urlShortener URLShortenerService
	logger       *log.Logger
}

func NewHandler(baseURL string, store Repository, shortener URLShortenerService, logger *log.Logger) *Handler {
	return &Handler{
		baseURL:      baseURL,
		store:        store,
		urlShortener: shortener,
		logger:       logger,
	}
}

func (h *Handler) GetOriginUrlHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.Println("id not specified")
		http.Error(w, "id not specified", http.StatusBadRequest)
		return
	}

	url, err := h.store.Get(id)
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

func (h *Handler) ShortenURLHandle(w http.ResponseWriter, r *http.Request) {
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

	shortenValue, err := h.urlShortener.Shorten(originalURL)
	if err != nil {
		h.logger.Printf("error shortening URL: %v", err)
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	url := model.NewShortenURL(shortenValue, originalURL)
	if err := h.store.Add(url); err != nil {
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
