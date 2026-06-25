// Package handler реализует HTTP-обработчики сервиса сокращения URL.
// Пакет предоставляет маршрутизатор и набор хендлеров для работы
// с URL: сокращение, перенаправление, пакетная обработка и удаление.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/config"
	"github.com/newmersedez/urlshort/internal/middleware"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/newmersedez/urlshort/internal/repository/db"
)

// Repository описывает интерфейс для хранилища сокращённых URL.
// Реализации могут быть in-memory, файловыми или использовать базу данных.
type Repository interface {
	// Get возвращает сокращённый URL по его идентификатору.
	Get(ctx context.Context, id string) (*model.ShortenURL, error)
	// GetList возвращает все активные URL пользователя.
	GetList(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	// GetDeletedList возвращает все мягко удалённые URL.
	GetDeletedList(ctx context.Context) ([]model.ShortenURL, error)
	// Add сохраняет новый сокращённый URL.
	Add(ctx context.Context, shortenURL *model.ShortenURL) error
	// AddBatch сохраняет пакет сокращённых URL за одну операцию.
	AddBatch(ctx context.Context, shortenURLs []*model.ShortenURL) error
	// SoftDeleteBatch помечает список URL пользователя как удалённые.
	SoftDeleteBatch(ctx context.Context, userID uuid.UUID, ids []string) error
	// HardDeleteBatch физически удаляет URL из хранилища.
	HardDeleteBatch(ctx context.Context, ids []string) error
	// Stats возвращает количество сокращённых URL и уникальных пользователей.
	Stats(ctx context.Context) (urls int, users int, err error)
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
	// Close освобождает ресурсы хранилища.
	Close()
}

// ShortenerService описывает сервис генерации короткого идентификатора из URL.
type ShortenerService interface {
	// Shorten возвращает короткий идентификатор для переданного URL.
	// Ошибка возвращается, если URL имеет недопустимую схему.
	Shorten(url string) (string, error)
}

// TokenService описывает сервис управления токенами аутентификации.
type TokenService interface {
	// IsValid проверяет, является ли токен корректным и не истёкшим.
	IsValid(token string) bool
	// GetToken выпускает зашифрованный токен для переданного userID.
	GetToken(userID uuid.UUID) (string, error)
	// GetUserID расшифровывает токен и возвращает идентификатор пользователя.
	GetUserID(token string) (uuid.UUID, error)
}

// CleanupService описывает сервис фонового удаления помеченных URL.
type CleanupService interface {
	// ScheduleDelete добавляет список идентификаторов в очередь на удаление.
	ScheduleDelete(ctx context.Context, userID uuid.UUID, ids []string)
	// Start запускает фоновый цикл обработки очереди удалений.
	Start(ctx context.Context)
}

// AuditService описывает сервис записи аудит-событий.
type AuditService interface {
	// Notify асинхронно уведомляет всех подписчиков о произошедшем событии.
	Notify(event *model.AuditEvent)
}

type shortenURLRequest struct {
	URL string `json:"url"`
}

type shortenBatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type shortenURLResponse struct {
	Result string `json:"result"`
}

type shortenBatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type urlListResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

type handlers struct {
	baseURL          string
	trustedSubnet    *net.IPNet
	store            Repository
	logger           *slog.Logger
	shortenerService ShortenerService
	tokenService     TokenService
	cleanupService   CleanupService
	auditService     AuditService
}

// Serve создаёт и настраивает HTTP-сервер. Не запускает его — вызывающий код
// отвечает за запуск (ListenAndServe) и остановку (Shutdown).
func Serve(
	cfg config.Config,
	store Repository,
	shortener ShortenerService,
	logger *slog.Logger,
	tokenService TokenService,
	cleanupService CleanupService,
	auditService AuditService) (*http.Server, error) {
	var trustedSubnet *net.IPNet
	if cfg.TrustedSubnet != "" {
		var err error
		_, trustedSubnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted_subnet %q: %w", cfg.TrustedSubnet, err)
		}
	}

	handler, err := newHandlers(cfg.BaseURL, trustedSubnet, store, logger, shortener, tokenService, cleanupService, auditService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize handlers object: %w", err)
	}

	server, err := newServer(cfg.ServerAddr, newRouter(handler))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	if cfg.EnableHTTPS {
		host, _, _ := net.SplitHostPort(cfg.ServerAddr)
		tlsCfg, err := tlsConfigFor(host, cfg.TLSCertCacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
		server.TLSConfig = tlsCfg
	}

	return server, nil
}

// NewHandler создаёт http.Handler со всеми маршрутами сервиса.
// Используйте эту функцию для интеграционного тестирования и документационных примеров.
func NewHandler(
	baseURL string,
	store Repository,
	shortener ShortenerService,
	logger *slog.Logger,
	tokenService TokenService,
	cleanupService CleanupService,
	auditService AuditService) (http.Handler, error) {
	h, err := newHandlers(baseURL, nil, store, logger, shortener, tokenService, cleanupService, auditService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}
	return newRouter(h), nil
}

func newHandlers(
	baseURL string,
	trustedSubnet *net.IPNet,
	store Repository,
	logger *slog.Logger,
	shortenerService ShortenerService,
	tokenService TokenService,
	cleanupService CleanupService,
	auditService AuditService) (*handlers, error) {
	handlers := &handlers{
		baseURL:          strings.TrimRight(baseURL, "/") + "/",
		trustedSubnet:    trustedSubnet,
		store:            store,
		shortenerService: shortenerService,
		logger:           logger,
		tokenService:     tokenService,
		cleanupService:   cleanupService,
		auditService:     auditService,
	}

	return handlers, nil
}

func newRouter(handler *handlers) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestLoggerMiddleware(handler.logger))
	router.Use(middleware.RequestCompressorMiddleware(handler.logger))
	router.Use(middleware.AuthorizationMiddleware(handler.tokenService, handler.logger))

	router.Group(func(r chi.Router) {
		r.Use(middleware.TrustedSubnetMiddleware(handler.trustedSubnet))
		r.Get("/api/internal/stats", handler.getStatsHandle)
	})

	router.Get("/api/user/urls", handler.GetURLsHandle)
	router.Get("/{id}", handler.getOriginURLHandle)
	router.Get("/ping", handler.pingDatabaseHandle)
	router.Post("/", handler.shortenURLHandle)
	router.Post("/api/shorten", handler.enhancedShortenURLHandle)
	router.Post("/api/shorten/batch", handler.shortenBatchURLsHandle)
	router.Delete("/api/user/urls", handler.deleteURLsHandle)

	return router
}

func newServer(address string, router *chi.Mux) (*http.Server, error) {
	if address == "" {
		return nil, errors.New("server address is not set")
	}

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

func (h *handlers) getOriginURLHandle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	url, err := h.store.Get(r.Context(), id)

	if err != nil {
		h.logger.Error("error retrieving URL from repository", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if url == nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if url.Deleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	event := model.NewAuditEvent("follow", url.UserID.String(), url.OriginalValue)
	h.auditService.Notify(event)

	http.Redirect(w, r, url.OriginalValue, http.StatusTemporaryRedirect)
}

func (h *handlers) buildShortURL(id string) string {
	return h.baseURL + id
}

// GetURLsHandle обрабатывает GET /api/user/urls - возвращает список URL текущего пользователя.
// Требует действительный cookie с токеном аутентификации.
func (h *handlers) GetURLsHandle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	urls, err := h.store.GetList(r.Context(), userID)

	if err != nil {
		h.logger.Error("error retrieving URL from repository", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseBody := make([]urlListResponse, 0, len(urls))

	for _, item := range urls {
		responseBody = append(responseBody, urlListResponse{
			ShortURL:    h.buildShortURL(item.ID),
			OriginalURL: item.OriginalValue,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if len(responseBody) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		h.logger.Error("failed to write response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handlers) enhancedShortenURLHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var request shortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "url value is required", http.StatusBadRequest)
		return
	}

	id, err := h.shortenerService.Shorten(request.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	fullShortenURL := h.buildShortURL(id)

	shortenURL := model.NewShortenURL(userID, id, request.URL)
	response := shortenURLResponse{
		Result: fullShortenURL,
	}

	err = h.store.Add(r.Context(), shortenURL)
	if err != nil {
		if errors.Is(err, db.ErrUniqueViolation) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		h.logger.Error("failed to save shorten URL to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	event := model.NewAuditEvent("shorten", userID.String(), request.URL)
	h.auditService.Notify(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handlers) shortenURLHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request body is not a valid plain text", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	id, err := h.shortenerService.Shorten(originalURL)
	if err != nil {
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
		return
	}

	fullShortenURL := h.buildShortURL(id)

	shortenURL := model.NewShortenURL(userID, id, originalURL)

	err = h.store.Add(r.Context(), shortenURL)
	if err != nil {
		if errors.Is(err, db.ErrUniqueViolation) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fullShortenURL))
			return
		}

		h.logger.Error("failed to save shorten URL to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	event := model.NewAuditEvent("shorten", userID.String(), originalURL)
	h.auditService.Notify(event)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortenURL))
}

func (h *handlers) shortenBatchURLsHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var requestBody []shortenBatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	responseBody := make([]shortenBatchURLResponse, 0, len(requestBody))
	shortenURLs := make([]*model.ShortenURL, 0, len(requestBody))

	for _, item := range requestBody {
		id, err := h.shortenerService.Shorten(item.OriginalURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid URL %s", item.OriginalURL), http.StatusBadRequest)
			return
		}

		shortenURLs = append(shortenURLs, model.NewShortenURL(userID, id, item.OriginalURL))
		responseBody = append(responseBody, shortenBatchURLResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      h.buildShortURL(id),
		})
	}

	err := h.store.AddBatch(r.Context(), shortenURLs)
	if err != nil {
		h.logger.Error("failed to save shorten URLs to the storage", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, item := range requestBody {
		event := model.NewAuditEvent("shorten", userID.String(), item.OriginalURL)
		h.auditService.Notify(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(responseBody)

	if err != nil {
		h.logger.Error("failed to write response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handlers) pingDatabaseHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		h.logger.Error("failed to connect to the database", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) deleteURLsHandle(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, "request body is not a valid JSON", http.StatusBadRequest)
		return
	}

	if len(ids) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if err := h.store.SoftDeleteBatch(r.Context(), userID, ids); err != nil {
		h.logger.Error("failed to mark URLs as deleted", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.cleanupService.ScheduleDelete(r.Context(), userID, ids)

	w.WriteHeader(http.StatusAccepted)
}

// getStatsHandle обрабатывает GET /api/internal/stats.
func (h *handlers) getStatsHandle(w http.ResponseWriter, r *http.Request) {
	urls, users, err := h.store.Stats(r.Context())
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(statsResponse{URLs: urls, Users: users}); err != nil {
		h.logger.Error("failed to write stats response", "error", err)
	}
}
