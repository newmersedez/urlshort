package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/newmersedez/urlshort/internal/handler/config"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	GetByShortenValue(shortenValue string) *model.ShortenURL
	Add(shortenURL *model.ShortenURL)
}

type URLShortenerService interface {
	Shorten(url string) (*string, error)
}

type handlers struct {
	baseURL      string
	store        Repository
	urlShortener URLShortenerService
	logger       *log.Logger
}

func Serve(cfg config.Config, store Repository, shortener URLShortenerService, logger *log.Logger) error {
	h := newHandlers(cfg.BaseURL, store, shortener, logger)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	log.Printf("Started server on address %s", cfg.ServerAddr)
	return srv.ListenAndServe()
}

func GetURLByShortenValueHandler(baseURL string, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id not specified", http.StatusBadRequest)
			return
		}

		url := repo.GetByShortenValue(id)
		if url == nil {
			log.Println("url not found")
			http.Error(w, "url not found", http.StatusBadRequest)
			return
		}

		shortenURL := fmt.Sprintf("%s/%s", baseURL, url.ShortenValue)
		log.Printf("Returning original %s by shorten %s", url.OriginalValue, shortenURL)

		w.Header().Set("Content-Type", "text/plain")
		http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
	}
}

func ShortenURLHandler(baseURL string, repo Repository, urlShortener URLShortenerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		originalURL := string(bytes)
		shortenValue, err := urlShortener.Shorten(originalURL)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := model.NewShortenURL(*shortenValue, originalURL)
		repo.Add(url)

		shortenURL := fmt.Sprintf("%s/%s", baseURL, url.ShortenValue)
		log.Printf("Successfully shorten %s to %s", url.OriginalValue, shortenURL)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenURL))
	}
}

func newRouter(h *handlers) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", ShortenURLHandler(h.baseURL, h.store, h.urlShortener))
	router.Get("/{id}", GetURLByShortenValueHandler(h.baseURL, h.store))

	return router
}

func newHandlers(baseURL string, store Repository, urlShortener URLShortenerService, logger *log.Logger) *handlers {
	return &handlers{
		baseURL:      baseURL,
		store:        store,
		urlShortener: urlShortener,
		logger:       logger,
	}
}
