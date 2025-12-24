package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/newmersedez/urlshort/internal/handler/config"
	"github.com/newmersedez/urlshort/internal/model"
)

type Repository interface {
	GetByShortenValue(shortenValue string) *model.ShortenUrl
	Add(shortenUrl *model.ShortenUrl)
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

func GetUrlByShortenValueHandler(baseUrl string, repo Repository) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
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

		shortenUrl := fmt.Sprintf("%s/%s", baseUrl, url.ShortenValue)
		log.Printf("Returning original %s by shorten %s", url.OriginalValue, shortenUrl)
		
		w.Header().Set("Content-Type", "text/plain")
		http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
	}
}

func ShortenUrlHandler(baseUrl string, repo Repository, urlShortener UrlShortenerService) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		originalUrl := string(bytes)
		shortenValue, err := urlShortener.Shorten(originalUrl)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := model.NewShortenUrl(*shortenValue, originalUrl)
		repo.Add(url)
		
		shortenUrl := fmt.Sprintf("%s/%s", baseUrl, url.ShortenValue)
		log.Printf("Successfully shorten %s to %s", url.OriginalValue, shortenUrl)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenUrl))
	}
}

func newRouter(h *handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", ShortenUrlHandler(h.baseURL, h.store, h.urlShortener))
	mux.HandleFunc("GET /{id}/", GetUrlByShortenValueHandler(h.baseURL, h.store))
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
