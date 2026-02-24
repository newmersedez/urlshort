package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/handler/mocks"
	"github.com/newmersedez/urlshort/internal/middleware"
	middlewareMocks "github.com/newmersedez/urlshort/internal/middleware/mocks"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCanShortenValidURL(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "https://stackoverflow.com"
	id := "12345678"
	userID := uuid.New()

	logger := slog.Default()

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Add(mock.Anything, mock.MatchedBy(func(s *model.ShortenURL) bool {
		return s != nil && s.ID == id && s.OriginalValue == originalURL
	})).Return(nil).Once()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return(id, nil).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)
	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	request.Header.Set("Content-Type", "text/plain")

	ctx := middleware.SetUserID(t.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.shortenURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, string(resBody))
	assert.Contains(t, res.Header.Get("Content-Type"), "text/plain")
}

func TestCannotShortenInvalidURL(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "url//string"
	userID := uuid.New()

	logger := slog.Default()
	mockRepo := mocks.NewMockRepository(t)

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return("", errors.New("invalid url")).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	request.Header.Set("Content-Type", "text/plain")
	ctx := middleware.SetUserID(t.Context(), userID)
	request = request.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.shortenURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCanGetFullURLByShortenValue(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	id := "12345678"
	userID := uuid.New()
	originalURL := "https://stackoverflow.com"
	shortenURL := model.NewShortenURL(userID, id, originalURL)

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(mock.Anything, id).Return(shortenURL, nil).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Content-Type", "text/plain")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Act
	h.getOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.NotEmpty(t, res.Header.Get("Location"))
}

func TestCannotGetFullURLByShortenValueIfIdIsNotSpecified(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"

	mockShortener := mocks.NewMockShortener(t)
	mockRepo := mocks.NewMockRepository(t)
	logger := slog.Default()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	request.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Act
	h.getOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotGetFullURLByShortenValueIfItDoesNotExist(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	id := "12345678"

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(mock.Anything, id).Return(nil, nil).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Content-Type", "text/plain")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	h.getOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestCanShortenValidURLViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "https://stackoverflow.com"
	id := "12345678"
	userID := uuid.New()

	logger := slog.Default()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return(id, nil).Once()

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Add(mock.Anything, mock.MatchedBy(func(s *model.ShortenURL) bool {
		return s != nil && s.ID == id && s.OriginalValue == originalURL
	})).Return(nil).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	ctx := middleware.SetUserID(t.Context(), userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.enhancedShortenURLHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	require.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var response shortenURLResponse
	err = json.Unmarshal(resBody, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response.Result)
	require.Contains(t, res.Header.Get("Content-Type"), "application/json")
}

func TestCannotHandleInvalidRequestBodyViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "https://stackoverflow.com"
	userID := uuid.New()

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)
	mockRepo := mocks.NewMockRepository(t)

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`"url": "%s"`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	ctx := middleware.SetUserID(t.Context(), userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.enhancedShortenURLHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotHandleInvalidContentTypeViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "https://stackoverflow.com"

	mockShortener := mocks.NewMockShortener(t)
	mockRepo := mocks.NewMockRepository(t)
	logger := slog.Default()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "text/xml")
	w := httptest.NewRecorder()

	// Act
	h.enhancedShortenURLHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)
}

func TestCannotShortenInvalidURLViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "url//string"
	userID := uuid.New()

	mockRepo := mocks.NewMockRepository(t)
	logger := slog.Default()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return("", errors.New("invalid URL")).Maybe()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	ctx := middleware.SetUserID(t.Context(), userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.enhancedShortenURLHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCanShortenValidBatchURLsViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL1 := "https://stackoverflow1.com"
	originalURL2 := "https://stackoverflow2.com"
	id1 := "123"
	id2 := "456"
	userID := uuid.New()

	logger := slog.Default()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL1).Return(id1, nil).Once()
	mockShortener.EXPECT().Shorten(originalURL2).Return(id2, nil).Once()

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().AddBatch(
		mock.Anything,
		mock.MatchedBy(func(urls []*model.ShortenURL) bool {
			if len(urls) != 2 {
				return false
			}

			if urls[0].ID != id1 || urls[0].OriginalValue != originalURL1 {
				return false
			}

			if urls[1].ID != id2 || urls[1].OriginalValue != originalURL2 {
				return false
			}

			return true
		}),
	).Return(nil).Once()

	tokenService := middlewareMocks.NewMockTokenService(t)
	cleanupService := mocks.NewMockCleanupService(t)

	h, _ := newHandlers(baseURL, mockRepo, logger, mockShortener, tokenService, cleanupService)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(fmt.Sprintf(`
		[
			{
				"correlation_id": "1",
				"original_url": "%s"
			},
			{
				"correlation_id": "2",
				"original_url": "%s"
			}
		]
	`,
		originalURL1, originalURL2)))
	r.Header.Add("Content-Type", "application/json")
	ctx := middleware.SetUserID(t.Context(), userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	h.shortenBatchURLsHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	require.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var response []shortenBatchURLResponse
	err = json.Unmarshal(resBody, &response)
	require.NoError(t, err)
	require.Len(t, response, 2)
	require.Contains(t, res.Header.Get("Content-Type"), "application/json")
	require.Contains(t, response, shortenBatchURLResponse{CorrelationID: "1", ShortURL: "http://localhost:8080/123"})
	require.Contains(t, response, shortenBatchURLResponse{CorrelationID: "2", ShortURL: "http://localhost:8080/456"})
}
