package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"
)

var buildVersion string
var buildDate string
var buildCommit string

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}
	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

func main() {
	printBuildInfo()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config object: %w", err)
	}

	log, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger object: %w", err)
	}

	tokenService, err := service.NewTokenService()
	if err != nil {
		return fmt.Errorf("failed to initialize token service object: %w", err)
	}

	shortenerService := service.NewURLShortenerService()

	var repo handler.Repository

	switch {
	case cfg.DatabaseDSN != "":
		repo, err = repository.NewDBRepository(context.Background(), cfg.DatabaseDSN, log)
		if err != nil {
			return fmt.Errorf("failed to initialize db repository object: %w", err)
		}
	case cfg.FileStoragePath != "":
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("failed to initialize file repository object: %w", err)
		}
	default:
		repo, err = repository.NewMemoryRepository()
		if err != nil {
			return fmt.Errorf("failed to initialize in-memory repository object: %w", err)
		}
	}

	defer repo.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-quit
		cancel()
	}()

	cleanupService := service.NewCleanupService(ctx, repo, log)

	auditService := service.NewAuditService(log)
	if cfg.AuditFile != "" {
		auditService.Subscribe(service.NewFileAuditObserver(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditService.Subscribe(service.NewHTTPAuditObserver(cfg.AuditURL))
	}

	log.Info("Starting server", "address", cfg.ServerAddr)
	return handler.Serve(ctx, *cfg, repo, shortenerService, log, tokenService, cleanupService, auditService)
}
