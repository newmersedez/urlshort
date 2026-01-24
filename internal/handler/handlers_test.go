package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/newmersedez/urlshort/internal/model"
	"github.com/newmersedez/urlshort/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanShortenValidURL(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://stackoverflow.com"))
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaPlainTextHandle(w, request)

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
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("url//string"))
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaPlainTextHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCanGetFullURLByShortenValue(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"

	repo := NewMockRepository()
	shortenURL := model.ShortenURL{
		Key:   "12345678",
		Value: "https://stackoverflow.com",
	}

	ctx := t.Context()
	err := repo.Add(ctx, &shortenURL)
	require.NoError(t, err)

	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "12345678")

	// Добавляем контекст в запрос
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Act
	h.GetOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.NotEmpty(t, res.Header.Get("Location"))
}

func TestCannotGetFullURLByShortenValueIfIdIsNotSpecified(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"

	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	w := httptest.NewRecorder()

	// Act
	h.GetOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotGetFullURLByShortenValueIfItDoesNotExist(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"

	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	request := httptest.NewRequest(http.MethodGet, "/1337", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1337")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	h.GetOriginURLHandle(w, request)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestCanShortenValidURLViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "https://practicum.yandex.ru"}`))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaJSONHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	require.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var response ShortenURLResponse
	err = json.Unmarshal(resBody, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response.Result)
	require.Contains(t, res.Header.Get("Content-Type"), "application/json")
}

func TestCannotHandleInvalidRequestBodyViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "url//string"`))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaPlainTextHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotHandleInvalidContentTypeViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "url//string"`))
	r.Header.Add("Content-Type", "text/xml")
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaPlainTextHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCannotShortenInvalidURLViaJSONHandler(t *testing.T) {
	// Arrange
	baseURL := "http://localhost:8080"
	repo := NewMockRepository()
	urlShortener := service.NewURLShortenerService()
	h := NewHandler(baseURL, repo, urlShortener)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "url//string"}`))
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	h.ShortenURLViaPlainTextHandle(w, r)

	//Assert
	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

type MockRepository struct {
	data map[string]model.ShortenURL
}

func NewMockRepository() Repository {
	return &MockRepository{
		data: make(map[string]model.ShortenURL),
	}
}

func (r *MockRepository) Get(ctx context.Context, shortenValue string) (*model.ShortenURL, error) {
	shortenURL, exists := r.data[shortenValue]
	if !exists {
		return nil, nil
	}

	return &shortenURL, nil
}

func (r *MockRepository) Add(ctx context.Context, shortenURL *model.ShortenURL) error {
	r.data[shortenURL.Key] = *shortenURL
	return nil
}

func (r *MockRepository) Dispose() {}
