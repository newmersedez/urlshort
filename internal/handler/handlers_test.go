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
	"github.com/newmersedez/urlshort/internal/handler/mocks"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCanShortenValidURL(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "https://stackoverflow.com"
	key := "12345678"

	logger := slog.Default()

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(t.Context(), key).Return(nil, nil).Once()
	mockRepo.EXPECT().Add(mock.Anything, mock.MatchedBy(func(s *model.ShortenURL) bool {
		return s != nil && s.ID == key && s.OriginalValue == originalURL
	})).Return(nil).Once()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return(key, nil).Once()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	request.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaPlainTextHandle(w, request)

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

	logger := slog.Default()
	mockRepo := mocks.NewMockRepository(t)

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return("", errors.New("invalid url")).Once()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	request.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaPlainTextHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCanGetFullURLByShortenValue(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	key := "12345678"
	originalURL := "https://stackoverflow.com"
	shortenURL := model.NewShortenURL(key, originalURL)

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(mock.Anything, key).Return(shortenURL, nil).Once()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Content-Type", "text/plain")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", key)

	// Добавляем контекст в запрос
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

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

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
	key := "12345678"

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(mock.Anything, key).Return(nil, nil).Once()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Content-Type", "text/plain")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", key)
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
	key := "12345678"

	logger := slog.Default()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return(key, nil).Once()

	mockRepo := mocks.NewMockRepository(t)
	mockRepo.EXPECT().Get(mock.Anything, key).Return(nil, nil).Once()
	mockRepo.EXPECT().Add(mock.Anything, mock.MatchedBy(func(s *model.ShortenURL) bool {
		return s != nil && s.ID == key && s.OriginalValue == originalURL
	})).Return(nil).Once()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaJSONHandle(w, r)

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

	logger := slog.Default()
	mockShortener := mocks.NewMockShortener(t)
	mockRepo := mocks.NewMockRepository(t)

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`"url": "%s"`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaJSONHandle(w, r)

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

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "text/xml")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaJSONHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusUnsupportedMediaType, res.StatusCode)
}

func TestCannotShortenInvalidURLViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	originalURL := "url//string"

	mockRepo := mocks.NewMockRepository(t)
	logger := slog.Default()

	mockShortener := mocks.NewMockShortener(t)
	mockShortener.EXPECT().Shorten(originalURL).Return("", errors.New("invalid URL")).Maybe()

	h, _ := newHandlers(baseURL, mockRepo, mockShortener, logger)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, originalURL)))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.shortenURLViaJSONHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
