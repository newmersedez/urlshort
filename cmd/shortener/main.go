package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/newmersedez/urlshort/internal/config"
	grpchandler "github.com/newmersedez/urlshort/internal/grpc/handler"
	"github.com/newmersedez/urlshort/internal/handler"
	"github.com/newmersedez/urlshort/internal/logger"
	"github.com/newmersedez/urlshort/internal/repository"
	"github.com/newmersedez/urlshort/internal/service"
)

var buildVersion string
var buildDate string
var buildCommit string

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

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

func run() (err error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config object: %w", err)
	}

	appLog, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger object: %w", err)
	}

	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	tokenService, err := service.NewTokenService()
	if err != nil {
		return fmt.Errorf("failed to initialize token service object: %w", err)
	}

	shortenerService := service.NewURLShortenerService()

	var repo handler.Repository

	switch {
	case cfg.DatabaseDSN != "":
		repo, err = repository.NewDBRepository(ctx, cfg.DatabaseDSN, appLog)
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

	g.Go(func() error {
		defer appLog.Info("repository closed")
		<-ctx.Done()
		repo.Close()
		return nil
	})

	cleanupService := service.NewCleanupService(ctx, repo, appLog)

	g.Go(func() error {
		cleanupService.Start(ctx)
		return nil
	})

	auditService := service.NewAuditService(appLog)
	if cfg.AuditFile != "" {
		auditService.Subscribe(service.NewFileAuditObserver(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditService.Subscribe(service.NewHTTPAuditObserver(cfg.AuditURL))
	}

	server, err := handler.Serve(*cfg, repo, shortenerService, appLog, tokenService, cleanupService, auditService)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	grpcServer := grpchandler.NewShortenerServer(cfg.BaseURL, repo, shortenerService, tokenService, appLog)

	g.Go(func() error {
		appLog.Info("Starting gRPC server", "address", cfg.GRPCAddr)
		if err := grpchandler.Serve(ctx, grpcServer, cfg.GRPCAddr); err != nil {
			return fmt.Errorf("grpc server failed: %w", err)
		}
		return nil
	})

	g.Go(func() (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("a panic occurred: %v", rec)
			}
		}()
		appLog.Info("Starting HTTP server", "address", cfg.ServerAddr)
		if cfg.EnableHTTPS {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve has failed: %w", err)
	})

	g.Go(func() error {
		defer appLog.Info("server has been shutdown")
		<-ctx.Done()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			appLog.Error("an error occurred during server shutdown", "error", err)
		}
		return nil
	})

	return g.Wait()
}
