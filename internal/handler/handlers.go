package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/newmersedez/urlshort/internal/handler/config"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	GetByShortenValue(shortenValue string) (*model.ShortenUrl, error)
	Add(shortenUrl *model.ShortenUrl) error
}

type UrlShortenerService interface {
	Shorten(url string) (*string, error)
}

type handlers struct {
    baseURL string
	store Repository
	urlShortener UrlShortenerService
	logger *log.Logger
}

func Serve(cfg config.Config, store Repository, shortener UrlShortenerService, logger *log.Logger) error {
	h := newHandlers(cfg.BaseUrl, store, shortener, logger)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	log.Printf("Started server on address %s", cfg.ServerAddr)
	return srv.ListenAndServe()
}

func (h *handlers) GetUrlByShortenValue(w http.ResponseWriter, r *http.Request) {
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

	shortenUrl := fmt.Sprintf("%s/%s", h.baseURL, url.ShortenValue)
	log.Printf("Returning original %s by shorten %s", url.OriginalValue, shortenUrl)
	http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
}

func (h *handlers) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodPost) {
		log.Println("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	if err != nil || shortenValue == nil {
		log.Println("failed to shortn url")
		http.Error(w, "failed to shortn url", http.StatusBadRequest)
		return
	}

	url := model.NewShortenUrl(*shortenValue, originalUrl)
	if err := h.store.Add(url); err != nil {
		log.Println("failed to shortn url")
		http.Error(w, "failed to shortn url", http.StatusBadRequest)
		return
	}
	
	shortenUrl := fmt.Sprintf("%s/%s", h.baseURL, url.ShortenValue)
	log.Printf("Successfully shorten %s to %s", url.OriginalValue, shortenUrl)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenUrl))
}

func newRouter(h *handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.ShortenUrl)
	mux.HandleFunc("GET /{id}", h.GetUrlByShortenValue)
	return mux
}


func newHandlers(baseUrl string, store Repository, urlShortener UrlShortenerService, logger *log.Logger) *handlers {
	return &handlers {
		baseURL: baseUrl,
		store: store,
		urlShortener: urlShortener,
		logger: logger,
	}
}
