package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/middleware"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	store := repository.NewRepository()
	shortener := service.NewURLShortenerService()
	handler := handler.NewHandler(cfg.BaseURL, store, shortener)

	router := chi.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware)
	router.Use(middleware.RequestCompressorMiddleware)

	router.Post("/", handler.ShortenURLViaPlainTextHandle)
	router.Post("/api/shorten", handler.ShortenURLViaJSONHandle)
	router.Get("/{id}", handler.GetOriginURLHandle)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddr))
	return server.ListenAndServe()
}
