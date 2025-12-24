package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	GetByShortenValue(shortenValue string) (*model.ShortenUrl, error)
	Add(shortenUrl *model.ShortenUrl) error
}

type UrlShortenerService interface {
	Shorten(url string) (string, error)
}

type Handlers struct {
	store Repository
	urlShortener UrlShortenerService
	logger log.Logger
}

func NewHandlers(store Repository, urlShortener UrlShortenerService) *Handlers {
	return &Handlers {
		store: store,
		urlShortener: urlShortener,
	}
}

func (h *Handlers) GetUrlByShortenValue(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodGet) {
		log.Println("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		log.Println("id not specified")
		http.Error(w, "id not specified", http.StatusBadRequest)
		return 
	}

	url, err := h.store.GetByShortenValue(id)
	if err != nil {
		log.Println("internal server error")
		http.Error(w, "internal server error", http.StatusBadRequest);
		return
	}
	if url == nil {
		log.Println("url not found")
		http.Error(w, "url not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
}

func (h *Handlers) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodPost) {
		log.Println("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

 	contentType := r.Header.Get("Content-Type")
    if contentType != "text/plain" {
		log.Println("unsupported media type")
        http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
        return
    }

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("internal server error")
		http.Error(w, "internal server error", http.StatusBadRequest)
		return
	}

	originalUrl := string(bytes)
	shortenValue, err := h.urlShortener.Shorten(originalUrl)
	if err != nil {
		log.Println("failed to shortn url")
		http.Error(w, "failed to shortn url", http.StatusBadRequest)
		return
	}

	shortenUrl := model.NewShortenUrl(shortenValue, originalUrl)
	h.store.Add(shortenUrl)
	w.WriteHeader(http.StatusCreated)
}