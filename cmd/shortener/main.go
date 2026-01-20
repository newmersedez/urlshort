package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"
)

func main() {
	logger := log.Default()

	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *log.Logger) error {
	cfg := config.NewConfig()

	store := repository.NewRepository()
	shortener := service.NewURLShortenerService()
	handler := handler.NewHandler(cfg.BaseURL, store, shortener, logger)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/", handler.ShortenURLHandle)
    router.Get("/{id}", handler.GetOriginUrlHandle)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Printf("Starting server on %s", cfg.ServerAddr)
	return server.ListenAndServe()
}
